package tree

import (
	"fmt"
	"reflect"
)

// One node of the parameter tree.
type Node struct {
	Value    string
	Children map[string]*Node
}

func (paramTree Node) Write(destination reflect.Value) WriteErrors {
	if paramTree.Children == nil {
		return paramTree.writeLeafValue(destination)
	}

	var errors WriteErrors

	if paramTree.Value != "" {
		errors.append(writeError{"ignoring self value of key that has child keys", "", true})
	}

	if destination.Kind() == reflect.Ptr {
		if destination.IsNil() {
			destination.Set(reflect.New(destination.Type().Elem()))
		}
		destination = destination.Elem()
	}

	switch destination.Kind() { //nolint:exhaustive // not covering all possible types
	case reflect.Struct:
		errors.merge(paramTree.writeIntoStruct(destination))
	case reflect.Map:
		if destination.IsNil() {
			destination.Set(reflect.MakeMap(destination.Type()))
		}
		errors.merge(paramTree.writeIntoMap(destination))
	case reflect.Slice:
		errors.merge(paramTree.writeIntoSlice(destination))
	default:
		errors.append(writeError{fmt.Sprintf("unhandleable destination type: %v", destination.Kind()), "", false})
	}

	return errors
}
