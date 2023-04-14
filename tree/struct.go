package tree

import (
	"reflect"
)

func (paramTree Node) writeIntoStruct(destination reflect.Value) WriteErrors {
	var errors WriteErrors
	for fieldName, childTree := range paramTree.Children {
		structField := lookupFieldByName(destination, fieldName)
		if !structField.IsValid() {
			errors.append(writeError{"unknown field", fieldName, true})
			continue
		}
		errors.mergeChildErrors(fieldName, childTree.Write(structField))
	}
	return errors
}

func lookupFieldByName(structure reflect.Value, name string) reflect.Value {
	fieldByName := structure.FieldByName(name)
	if fieldByName.IsValid() {
		return fieldByName
	}

	// This might be inefficient, but fine for one-time loading of a not-crazy-big config
	for fieldIndex := 0; fieldIndex < structure.NumField(); fieldIndex++ {
		fieldTag := structure.Type().Field(fieldIndex).Tag
		if fieldTag.Get("global") == name {
			return structure.Field(fieldIndex)
		}
		if fieldTag.Get("json") == name {
			return structure.Field(fieldIndex)
		}
	}

	return reflect.Value{}
}
