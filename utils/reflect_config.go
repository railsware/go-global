package utils

import (
	"reflect"

	"github.com/railsware/go-global/v2"
)

func ReflectConfig(config interface{}) (reflect.Value, global.Error) {
	reflectedConfig := reflect.ValueOf(config)

	if reflectedConfig.Kind() != reflect.Ptr {
		return reflect.Value{}, global.NewError("config must be a pointer to a structure")
	}

	reflectedConfig = reflectedConfig.Elem()

	if reflectedConfig.Kind() != reflect.Struct && reflectedConfig.Kind() != reflect.Array {
		return reflect.Value{}, global.NewError("config must be a pointer to a structure or array")
	}

	return reflectedConfig, nil
}
