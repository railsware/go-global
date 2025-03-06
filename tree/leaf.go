package tree

import (
	"encoding"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
)

func (paramTree Node) writeLeafValue(destination reflect.Value) WriteErrors { //nolint:cyclop
	if !destination.CanSet() {
		return newWriteErrors("value is not writable")
	}

	// If destination implements TextUnmarshaler, it takes priority
	if handled, unmarshalerErrors := tryWriteUnmarshaler(paramTree.Value, destination); handled {
		return unmarshalerErrors
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
	case reflect.Struct:
		return newWriteErrors("cannot write param: destination should be a primitive type, not a struct")
	case reflect.Map:
		return newWriteErrors("cannot write param: destination should be a primitive type, not a map")
	case reflect.Slice:
		if destination.Type().Elem().Kind() == reflect.Uint8 {
			return writeBase64(paramTree.Value, destination)
		}
		return newWriteErrors("cannot write param: destination should be a primitive type, not a slice")
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

func writeBase64(source string, destination reflect.Value) WriteErrors {
	bytes, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return newWriteErrors(fmt.Sprintf("could not decode base64: %v", err))
	}
	destination.SetBytes(bytes)
	return WriteErrors{}
}

func tryWriteUnmarshaler(source string, destination reflect.Value) (bool, WriteErrors) {
	if destination.CanInterface() {
		if unmarshaler, ok := destination.Interface().(encoding.TextUnmarshaler); ok {
			return true, writeUnmarshaler(source, unmarshaler)
		}
	}

	// In some cases unmarshaling requires a pointer receiver. So if the value itself does not implement the interface,
	// check a pointer to it as well.
	if destination.CanAddr() {
		if unmarshaler, ok := destination.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return true, writeUnmarshaler(source, unmarshaler)
		}
	}

	return false, WriteErrors{}
}

func writeUnmarshaler(source string, unmarshaler encoding.TextUnmarshaler) WriteErrors {
	err := unmarshaler.UnmarshalText([]byte(source))
	if err != nil {
		return newWriteErrors(fmt.Sprintf("cannot write param: UnmarshalText returned error: %v", err))
	}
	return WriteErrors{}
}
