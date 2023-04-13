package tree

import (
	"fmt"
	"reflect"
	"strconv"
)

func (paramTree Node) writeLeafValue(destination reflect.Value) WriteErrors {
	if !destination.CanSet() {
		return newWriteErrors("value is not writable")
	}

	switch destination.Kind() { //nolint:exhaustive // we don't cover all types
	case reflect.String:
		destination.SetString(paramTree.Value)
	case reflect.Int:
		intval, err := strconv.Atoi(paramTree.Value)
		if err != nil {
			return newWriteErrors("cannot read int param value")
		}
		destination.SetInt(int64(intval))
	case reflect.Bool:
		switch paramTree.Value {
		case "true":
			destination.SetBool(true)
		case "false":
			destination.SetBool(false)
		default:
			return newWriteErrors("cannot read bool param value (must be true or false)")
		}
	default:
		err := fmt.Sprintf("cannot write param: config key is of unsupported type %s", destination.Kind())
		return newWriteErrors(err)
	}

	return WriteErrors{}
}
