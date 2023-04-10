package tree

import (
	"reflect"
	"strconv"
)

func (paramTree Node) writeIntoSlice(destination reflect.Value) WriteErrors {
	indexedParams := make(map[int]*Node)
	maxIndex := -1
	var errors WriteErrors
	for stringIndex, childTree := range paramTree.Children {
		index, err := strconv.Atoi(stringIndex)
		if err != nil || index < 0 {
			errors.append(writeError{"not a numeric index", stringIndex, true})
			continue
		}
		indexedParams[index] = childTree
		if index > maxIndex {
			maxIndex = index
		}
	}
	if destination.Cap() <= maxIndex {
		// grow destination array to match
		destination.SetLen(destination.Cap())
		additionalLength := maxIndex - destination.Cap() + 1
		additionalElements := reflect.MakeSlice(destination.Type(), additionalLength, additionalLength)
		destination.Set(reflect.AppendSlice(destination, additionalElements))
	} else if destination.Len() <= maxIndex {
		destination.SetLen(maxIndex + 1)
	}
	for index, childTree := range indexedParams {
		childErrors := childTree.Write(destination.Index(index))
		if childErrors.Present() {
			errors.mergeChildErrors(strconv.Itoa(index), childErrors)
		}
	}
	return errors
}
