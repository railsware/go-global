package tree

import (
	"fmt"
	"reflect"
	"strconv"
)

func (leaf Node) writeLeafValue(destination reflect.Value) WriteErrors {
	if !destination.CanSet() {
		return newWriteErrors("value is not writable", false)
	} else if destination.Kind() == reflect.String {
		destination.SetString(leaf.Value)
	} else if destination.Kind() == reflect.Int {
		intval, err := strconv.Atoi(leaf.Value)
		if err != nil {
			return newWriteErrors("cannot read int param value", false)
		} else {
			destination.SetInt(int64(intval))
		}
	} else if destination.Kind() == reflect.Bool {
		if leaf.Value == "true" {
			destination.SetBool(true)
		} else if leaf.Value == "false" {
			destination.SetBool(false)
		} else {
			return newWriteErrors("cannot read bool param value (must be true or false)", false)
		}
	} else {
		return newWriteErrors(fmt.Sprintf("cannot write param: config key is of unsupported type %s", destination.Kind()), false)
	}

	return WriteErrors{}
}
