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
		return WriteErrors{}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return writeInt(paramTree.Value, destination)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return writeUint(paramTree.Value, destination)
	case reflect.Float64, reflect.Float32:
		return writeFloat(paramTree.Value, destination)
	case reflect.Bool:
		return writeBool(paramTree.Value, destination)
	default:
		err := fmt.Sprintf("cannot write param: config key is of unsupported type %s", destination.Kind())
		return newWriteErrors(err)
	}
}

func writeInt(source string, destination reflect.Value) WriteErrors {
	intval, err := strconv.ParseInt(source, 10, destination.Type().Bits())
	if err != nil {
		return newWriteErrors(fmt.Sprintf("cannot read %v param value: %v", destination.Kind(), err))
	}
	destination.SetInt(intval)
	return WriteErrors{}
}

func writeUint(source string, destination reflect.Value) WriteErrors {
	uintval, err := strconv.ParseUint(source, 10, destination.Type().Bits())
	if err != nil {
		return newWriteErrors(fmt.Sprintf("cannot read %v param value: %v", destination.Kind(), err))
	}
	destination.SetUint(uintval)
	return WriteErrors{}
}

func writeFloat(source string, destination reflect.Value) WriteErrors {
	floatval, err := strconv.ParseFloat(source, destination.Type().Bits())
	if err != nil {
		return newWriteErrors(fmt.Sprintf("cannot read %v param value: %v", destination.Kind(), err))
	}
	destination.SetFloat(floatval)
	return WriteErrors{}
}

func writeBool(source string, destination reflect.Value) WriteErrors {
	switch source {
	case "true":
		destination.SetBool(true)
	case "false":
		destination.SetBool(false)
	default:
		return newWriteErrors("cannot read bool param value (must be true or false)")
	}
	return WriteErrors{}
}
