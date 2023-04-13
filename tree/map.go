package tree

import (
	"reflect"
)

func (paramTree Node) writeIntoMap(destination reflect.Value) WriteErrors {
	if destination.Type().Key().Kind() != reflect.String {
		return newWriteErrors("can only write to maps with string keys")
	}
	var errors WriteErrors
	if destination.Type().Elem().Kind() == reflect.Ptr {
		// pointers are easy as their values are settable
		for key, childTree := range paramTree.Children {
			pointer := destination.MapIndex(reflect.ValueOf(key))
			if !pointer.IsValid() || pointer.IsNil() {
				pointer = reflect.New(destination.Type().Elem().Elem())
				destination.SetMapIndex(reflect.ValueOf(key), pointer)
			}
			errors.mergeChildErrors(key, childTree.Write(pointer.Elem()))
		}
	} else {
		// need to create a copy of the value and write it into the map
		for key, childTree := range paramTree.Children {
			newValue := reflect.New(destination.Type().Elem())
			oldValue := destination.MapIndex(reflect.ValueOf(key))
			if oldValue.IsValid() {
				newValue.Elem().Set(oldValue)
			}
			errors.mergeChildErrors(key, childTree.Write(newValue.Elem()))
			destination.SetMapIndex(reflect.ValueOf(key), newValue.Elem())
		}
	}
	return errors
}
