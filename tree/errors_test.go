package tree

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllKindsOfErrors(t *testing.T) {

	tree := &Node{
		Children: map[string]*Node{
			"Str":   {Value: "foo"},
			"int":   {Value: "not an int"},
			"float": {Value: "1.23"},
			"bool":  {Value: "yes"},
			"intmap": {
				Children: map[string]*Node{
					"foo": {Value: "bad int"},
				},
			},
			"intslice": {
				Children: map[string]*Node{
					"0":   {Value: "bad int"},
					"foo": {Value: "bar"},
				},
			},
			"nested": {
				Value: "ignored",
				Children: map[string]*Node{
					"bad_field": {Value: "foo"},
				},
			},
			"badmap": {
				Children: map[string]*Node{
					"0": {Value: "foo"},
				},
			},
		},
	}

	var destination testStructType

	errors := tree.Write(reflect.ValueOf(&destination))

	expectedErrors := []writeError{
		{
			msg:         "cannot read int param value",
			path:        "int",
			isPathError: false,
		},
		{
			msg:         "cannot read bool param value (must be true or false)",
			path:        "bool",
			isPathError: false,
		},
		{
			msg:         "cannot write param: config key is of unsupported type float64",
			path:        "float",
			isPathError: false,
		},
		{
			msg:         "cannot read int param value",
			path:        "intmap/foo",
			isPathError: false,
		},
		{
			msg:         "cannot read int param value",
			path:        "intslice/0",
			isPathError: false,
		},
		{
			msg:         "not a numeric index",
			path:        "intslice/foo",
			isPathError: true,
		},
		{
			msg:         "ignoring self value of key that has child keys",
			path:        "nested",
			isPathError: true,
		},
		{
			msg:         "unknown field",
			path:        "nested/bad_field",
			isPathError: true,
		},
		{
			msg:         "can only write to maps with string keys",
			path:        "badmap",
			isPathError: false,
		},
	}

	assert.ElementsMatch(t, expectedErrors, errors.errors)
}

func TestJoinWarnings(t *testing.T) {
	warnings := WriteErrors{
		errors: []writeError{
			{
				"warning for foo",
				"foo",
				true,
			},
			{
				"warning for bar",
				"bar",
				true,
			},
		},
	}

	mergedError := warnings.Join()

	assert.Equal(t, "global: foo: warning for foo, bar: warning for bar", mergedError.Error())
	assert.True(t, mergedError.Warning())
}

func TestJoinErrors(t *testing.T) {
	errors := WriteErrors{
		errors: []writeError{
			{
				"error for foo",
				"foo",
				false,
			},
			{
				"warning for bar",
				"bar",
				true,
			},
		},
	}

	mergedError := errors.Join()

	assert.Equal(t, "global: foo: error for foo, bar: warning for bar", mergedError.Error())
	assert.False(t, mergedError.Warning())
}
