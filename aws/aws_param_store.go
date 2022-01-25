package aws

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/railsware/go-global"
)

const paramStoreSeparator = "/"

type LoadConfigOptions struct {
	ParamPrefix string
	// If IgnoreUnmappedParams is set, a parameter with no matching config field will be silently ignored.
	IgnoreUnmappedParams bool
}

// LoadConfigFromParameterStore retrieves keys configured in ParamStore and writes to config
// config must be a pointer to a struct.
// Keys in ParamStore must be separated with slashes.
// They are matched to struct fields by: name, `global:` tag, or `json:` tag
func LoadConfigFromParameterStore(session *session.Session, options LoadConfigOptions, config interface{}) (err global.Error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = global.NewError("global: panic while loading from parameter store: %v", panicErr)
			return
		}
	}()

	client := ssm.New(session)

	var paramWarnings []global.Error
	var paramError global.Error

	awsErr := client.GetParametersByPathPages(
		&ssm.GetParametersByPathInput{
			Path:           aws.String(options.ParamPrefix),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
		},
		func(page *ssm.GetParametersByPathOutput, lastPage bool) bool {
			for _, param := range page.Parameters {
				paramNameWithoutPrefix := (*param.Name)[len(options.ParamPrefix):]
				destination, err := findParamDestination(config, paramNameWithoutPrefix)
				if err != nil {
					if !err.Warning() {
						paramError = global.NewError("%s: %v", paramNameWithoutPrefix, err)
						return false
					} else if !options.IgnoreUnmappedParams {
						paramWarnings = append(paramWarnings, global.NewWarning("%s: %v", paramNameWithoutPrefix, err))
					}
					continue
				}

				err = writeParamToConfig(destination, *param.Value)
				if err != nil {
					if err.Warning() {
						paramWarnings = append(paramWarnings, global.NewWarning("%s: %v", paramNameWithoutPrefix, err))
					} else {
						paramError = err
						return false
					}
				}
			}
			return true
		})

	if paramError != nil {
		return global.NewError("global: %v", paramError)
	}

	if awsErr != nil {
		return global.NewError("global: failed to load from Parameter Store: %v", awsErr)
	}

	if paramWarnings != nil {
		var warningMessages []string
		for _, warning := range paramWarnings {
			warningMessages = append(warningMessages, warning.Error())
		}

		return global.NewWarning("global: failed to read some parameters: %s", strings.Join(warningMessages, "; "))
	}

	return nil
}

func findParamDestination(config interface{}, name string) (reflect.Value, global.Error) {
	destination := reflect.ValueOf(config)

	if destination.Kind() != reflect.Ptr {
		return reflect.Value{}, global.NewError("config must be a pointer to a structure")
	}

	destination = destination.Elem()

	if destination.Kind() != reflect.Struct && destination.Kind() != reflect.Array {
		return reflect.Value{}, global.NewError("config must be a pointer to a structure or array")
	}

	// find nested field in config struct
	pathParts := strings.Split(name, paramStoreSeparator)
	for _, part := range pathParts {
		if destination.Kind() == reflect.Struct {
			destination = lookupFieldByName(destination, part)
		} else if destination.Kind() == reflect.Slice {
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 {
				return reflect.Value{}, global.NewWarning("could not map param to array index")
			}
			if destination.Cap() <= index {
				// grow destination array to match
				destination.SetLen(destination.Cap())
				additionalLength := index - destination.Cap() + 1
				additionalElements := reflect.MakeSlice(destination.Type(), additionalLength, additionalLength)
				destination.Set(reflect.AppendSlice(destination, additionalElements))
			} else if destination.Len() <= index {
				destination.SetLen(index + 1)
			}
			destination = destination.Index(index)
		} else {
			return reflect.Value{}, global.NewWarning("could not map param to config field")
		}
		// resolve pointer, if struct was nil
		if destination.Kind() == reflect.Ptr {
			if destination.IsNil() {
				destination.Set(reflect.New(destination.Type().Elem()))
			}
			destination = destination.Elem()
		}
	}

	// assign value to field
	if !destination.IsValid() {
		return reflect.Value{}, global.NewWarning("could not map param to config field")
	}

	return destination, nil
}

func writeParamToConfig(destination reflect.Value, value string) global.Error {
	if !destination.CanSet() {
		return global.NewWarning("config key is not writable")
	} else if destination.Kind() == reflect.String {
		destination.SetString(value)
	} else if destination.Kind() == reflect.Int {
		intval, err := strconv.Atoi(value)
		if err != nil {
			return global.NewWarning("cannot read int param value")
		} else {
			destination.SetInt(int64(intval))
		}
	} else if destination.Kind() == reflect.Bool {
		if value == "true" {
			destination.SetBool(true)
		} else if value == "false" {
			destination.SetBool(false)
		} else {
			return global.NewWarning("cannot read bool param value (must be true or false)")
		}
	} else {
		return global.NewWarning("cannot write param: config key is of unsupported type %s", destination.Kind())
	}

	return nil
}

func lookupFieldByName(structure reflect.Value, name string) reflect.Value {
	fieldByName := structure.FieldByName(name)
	if fieldByName.IsValid() {
		return fieldByName
	}

	// TODO might be inefficient, but fine for one-time loading of a not-crazy-big config
	for i := 0; i < structure.NumField(); i++ {
		fieldTag := structure.Type().Field(i).Tag
		if fieldTag.Get("global") == name {
			return structure.Field(i)
		}
		if fieldTag.Get("json") == name {
			return structure.Field(i)
		}
	}

	return reflect.Value{}
}
