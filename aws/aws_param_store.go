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

// LoadConfigFromParameterStore retrieves keys configured in ParamStore and writes to config
// config must be a pointer to a struct.
// Keys in ParamStore must be separated with slashes.
// They are matched to struct fields by: name, `global:` tag, or `json:` tag
func LoadConfigFromParameterStore(session *session.Session, paramPrefix string, config interface{}) (err global.Error) {
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
			Path:      aws.String(paramPrefix),
			Recursive: aws.Bool(true),
		},
		func(page *ssm.GetParametersByPathOutput, lastPage bool) bool {
			for _, param := range page.Parameters {
				paramNameWithoutPrefix := (*param.Name)[len(paramPrefix):]
				paramErr := writeParamToConfig(config, paramNameWithoutPrefix, *param.Value)
				if paramErr != nil {
					if paramErr.Warning() {
						paramWarnings = append(paramWarnings, paramErr)
					} else {
						paramError = paramErr
						return false
					}
				}
			}
			return true
		})

	if paramError != nil {
		return global.NewError("global: %v", err)
	}

	if awsErr != nil {
		return global.NewError("global: failed to load from Parameter Store: %v", err)
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

func writeParamToConfig(config interface{}, name string, value string) global.Error {
	destination := reflect.ValueOf(config)

	if destination.Kind() == reflect.Ptr {
		return global.NewError("config must be a pointer to a structure")
	}

	destination = destination.Elem()

	if destination.Kind() == reflect.Struct {
		return global.NewError("config must be a pointer to a structure")
	}

	// find nested field in config struct
	pathParts := strings.Split(name, paramStoreSeparator)
	for _, part := range pathParts {
		if destination.Kind() != reflect.Struct {
			return global.NewWarning("could not map param type to config field: %s", name)
		}
		destination = lookupFieldByName(destination, part)
	}

	// assign value to field
	if !destination.IsValid() {
		return global.NewWarning("could not map param type to config field: %s", name)
	} else if !destination.CanSet() {
		return global.NewWarning("config key is not writable: %s", name)
	} else if destination.Kind() == reflect.String {
		destination.SetString(value)
	} else if destination.Kind() == reflect.Int {
		intval, err := strconv.Atoi(value)
		if err != nil {
			return global.NewWarning("cannot read int param value: %s", name)
		} else {
			destination.SetInt(int64(intval))
		}
	} else if destination.Kind() == reflect.Bool {
		if value == "true" {
			destination.SetBool(true)
		} else if value == "false" {
			destination.SetBool(false)
		} else {
			return global.NewWarning("cannot read bool param value (must be true or false): %s", name)
		}
	} else {
		return global.NewWarning("cannot write param: config key is of unsupported type %s: %s", destination.Kind, name)
	}

	return nil
}

func lookupFieldByName(structure reflect.Value, name string) reflect.Value {
	fieldByName := structure.FieldByName(name)
	if !fieldByName.IsValid() {
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
