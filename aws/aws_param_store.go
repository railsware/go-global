package aws

import (
	"context"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/railsware/go-global/v2"
)

const paramStoreSeparator = "/"

type LoadConfigOptions struct {
	ParamPrefix string
	// If IgnoreUnmappedParams is set, a parameter with no matching config field will be silently ignored.
	IgnoreUnmappedParams bool
}

// LoadConfigFromParameterStore retrieves keys configured in ParamStore and writes to config.
// config must be a pointer to a struct.
// Keys in ParamStore must be separated with slashes.
// They are matched to struct fields by: name, `global:` tag, or `json:` tag
func LoadConfigFromParameterStore(awsConfig aws.Config, options LoadConfigOptions, globalConfig interface{}) global.Error {
	paramPaginator := ssm.NewGetParametersByPathPaginator(
		ssm.NewFromConfig(awsConfig),
		&ssm.GetParametersByPathInput{
			Path:           aws.String(options.ParamPrefix),
			Recursive:      true,
			WithDecryption: true,
		},
	)
	return LoadConfigFromParamPaginator(paramPaginator, options, globalConfig)
}

/*
ParamPaginator is an interface for aws ssm.ParamPaginator or mock implementation for unit tests
*/
type ParamPaginator interface {
	HasMorePages() bool
	NextPage(ctx context.Context, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
}

func LoadConfigFromParamPaginator(p ParamPaginator, options LoadConfigOptions, globalConfig interface{}) (err global.Error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = global.NewError("global: panic while loading config: %v, trace - %s", panicErr, debug.Stack())
			return
		}
	}()
	destination := reflect.ValueOf(globalConfig)
	if destination.Kind() != reflect.Ptr {
		return global.NewError("config must be a pointer to a structure")
	}
	if destination.Elem().Kind() != reflect.Struct {
		return global.NewError("config must be a pointer to a structure")
	}
	params := map[string]string{}
	for p.HasMorePages() {
		page, err := p.NextPage(context.Background())
		if err != nil {
			return global.NewError("global: failed to load from Parameter Store: %v", err)
		}
		for _, param := range page.Parameters {
			params[aws.ToString(param.Name)[len(options.ParamPrefix):]] = aws.ToString(param.Value)
		}
	}
	if err = populateParamsToConfig("", params, destination); err != nil && !options.IgnoreUnmappedParams {
		return err
	}
	return nil
}

func populateParamsToConfig(prefix string, params map[string]string, destination reflect.Value) global.Error {
	destinationElem := destination.Elem()
	var warnings []string
	for i := 0; i < destinationElem.NumField(); i++ {
		field := destinationElem.Field(i)
		fieldType := destinationElem.Type().Field(i)
		tag := fieldType.Tag.Get("global")
		if tag == "" {
			tag = fieldType.Tag.Get("json")
		}
		if tag == "" {
			tag = fieldType.Name
		}
		key := strings.Join([]string{prefix, tag}, paramStoreSeparator)
		if field.Kind() == reflect.Struct {
			if err := populateParamsToConfig(key, params, field.Addr()); err != nil {
				warnings = append(warnings, "failed to populate params for group "+key+". error - "+err.Error())
			}
			continue
		}
		value := params[key]
		if err := writeParamToConfig(key, field, value, params); err != nil {
			warnings = append(warnings, "failed to populate params for key "+key+". error - "+err.Error())
			continue
		}
	}
	if len(warnings) != 0 {
		return global.NewWarning("global: failed to read some parameters: %s", strings.Join(warnings, "; "))
	}
	return nil
}

func writeParamToConfig(key string, destination reflect.Value, value string, params map[string]string) global.Error {
	if !destination.CanSet() {
		return global.NewWarning("config key is not writable")
	}
	if destination.Kind() == reflect.Ptr {
		if destination.IsNil() {
			destination.Set(reflect.New(destination.Type().Elem()))
		}
		destination = destination.Elem()
	}
	switch destination.Kind() {
	case reflect.String:
		destination.SetString(value)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return global.NewWarning("cannot read bool param value (must be true or false). value - %s, err - %s", value, err)
		}
		destination.SetBool(boolValue)
	case reflect.Int:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return global.NewWarning("cannot read int param value: value - %s, err - %s", value, err)
		}
		destination.SetInt(intValue)
	case reflect.Slice:
		// case with slice configured in param store with multiple paths like */0/* */1/*
		if value == "" {
			return loadMultiKeysSlice(key, destination, params)
		}
		if destination.Type().Elem().Kind() != reflect.String {
			return global.NewWarning("slice config destination supported only for slice of strings")
		}
		values := strings.Split(value, ",")
		destination.Set(reflect.ValueOf(values))
	default:
		return global.NewWarning("cannot write param: config key is of unsupported type %s", destination.Kind())
	}
	return nil
}

func loadMultiKeysSlice(key string, destination reflect.Value, params map[string]string) global.Error {
	var length int
	indexesMap := map[string]bool{}
	for paramKey := range params {
		if !strings.HasPrefix(paramKey, key) {
			continue
		}
		index := strings.Split(strings.TrimPrefix(paramKey, key+"/"), "/")[0]
		_, ok := indexesMap[index]
		if ok {
			continue
		}
		indexesMap[index] = true
		length++
	}
	var messages []string
	elements := reflect.MakeSlice(destination.Type(), length, length)
	destination.Set(elements)
	for i := 0; i < length; i++ {
		paramKey := strings.Join([]string{key, strconv.Itoa(i)}, "/")
		elem := destination.Index(i)
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				elem.Set(reflect.New(elem.Type().Elem()))
			}
			elem = elem.Elem()
		}
		switch elem.Kind() {
		case reflect.String:
			elem.SetString(params[paramKey])
		case reflect.Struct:
			if err := populateParamsToConfig(paramKey, params, elem.Addr()); err != nil {
				return err
			}
		default:
			messages = append(messages, "failed to populate key - "+key+", wrong elem kind - "+destination.Type().Elem().Kind().String())
		}
	}
	if len(messages) != 0 {
		return global.NewWarning(strings.Join(messages, "; "))
	}
	return nil
}
