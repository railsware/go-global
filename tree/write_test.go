package tree

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStructType struct {
	Str            string                     `json:"str"`
	Int            int                        `global:"int"`
	Bool           bool                       `json:"bool"`
	StrMap         map[string]string          `json:"strmap"`
	StrSlice       []string                   `json:"strslice"`
	IntMap         map[string]int             `json:"intmap"`
	IntSlice       []int                      `json:"intslice"`
	Nested         *testStructType            `json:"nested"`
	NestedMap      map[string]testStructType  `json:"nmap"`
	NestedMapPtr   map[string]*testStructType `json:"nmapptr"`
	NestedSlice    []testStructType           `json:"nslice"`
	NestedSlicePtr []*testStructType          `json:"nsliceptr"`
	SliceOfMap     []map[string]string        `json:"sliceofmap"`
	MapOfSlice     map[string][]string        `json:"mapofslice"`
	// Fields just for testing errors
	Float  float64        `global:"float"`
	BadMap map[int]string `json:"badmap"`
}

func TestWrite(t *testing.T) {

	var testStruct testStructType

	tree := &Node{
		Children: map[string]*Node{
			"Str":  {Value: "foo"},
			"int":  {Value: "123"},
			"bool": {Value: "true"},
			"intmap": {
				Children: map[string]*Node{
					"foo": {Value: "123"},
					"bar": {Value: "456"},
				},
			},
			"strmap": {
				Children: map[string]*Node{
					"foo": {Value: "bar"},
					"baz": {Value: "qux"},
				},
			},
			"intslice": {
				Children: map[string]*Node{
					"0": {Value: "12"},
					"1": {Value: "34"},
				},
			},
			"strslice": {
				Children: map[string]*Node{
					"0": {Value: "foo"},
					// value for 1 is deliberately missing
					"2": {Value: "bar"},
				},
			},
			"nested": {
				Children: map[string]*Node{
					"str": {Value: "nested_foo"},
				},
			},
			"nmap": {
				Children: map[string]*Node{
					"foo": {
						Children: map[string]*Node{
							"str": {Value: "nested_foo_in_map"},
						},
					},
				},
			},
			"nmapptr": {
				Children: map[string]*Node{
					"foo": {
						Children: map[string]*Node{
							"str": {Value: "nested_foo_in_map_ptr"},
						},
					},
				},
			},
			"nslice": {
				Children: map[string]*Node{
					"0": {
						Children: map[string]*Node{
							"str": {Value: "nested_foo_in_slice"},
						},
					},
				},
			},
			"nsliceptr": {
				Children: map[string]*Node{
					"0": {
						Children: map[string]*Node{
							"str": {Value: "nested_foo_in_slice_ptr"},
						},
					},
				},
			},
			"sliceofmap": {
				Children: map[string]*Node{
					"0": {
						Children: map[string]*Node{
							"foo": {Value: "map_in_slice"},
						},
					},
				},
			},
			"mapofslice": {
				Children: map[string]*Node{
					"foo": {
						Children: map[string]*Node{
							"0": {Value: "slice_in_map"},
						},
					},
				},
			},
		},
	}

	errors := tree.Write(reflect.ValueOf(&testStruct))
	require.False(t, errors.Present())

	expectedStruct := testStructType{
		Str:  "foo",
		Int:  123,
		Bool: true,
		IntMap: map[string]int{
			"foo": 123,
			"bar": 456,
		},
		StrMap: map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
		IntSlice: []int{12, 34},
		StrSlice: []string{"foo", "", "bar"},
		Nested:   &testStructType{Str: "nested_foo"},
		NestedMap: map[string]testStructType{
			"foo": {
				Str: "nested_foo_in_map",
			},
		},
		NestedMapPtr: map[string]*testStructType{
			"foo": {
				Str: "nested_foo_in_map_ptr",
			},
		},
		NestedSlice:    []testStructType{{Str: "nested_foo_in_slice"}},
		NestedSlicePtr: []*testStructType{{Str: "nested_foo_in_slice_ptr"}},
		SliceOfMap:     []map[string]string{{"foo": "map_in_slice"}},
		MapOfSlice:     map[string][]string{"foo": {"slice_in_map"}},
	}

	assert.Equal(t, expectedStruct, testStruct, "assignment works correctly")

	treeToUpdate := &Node{
		Children: map[string]*Node{
			"Str":  {Value: "bar"},
			"int":  {Value: "456"},
			"bool": {Value: "false"},
			"intmap": {
				Children: map[string]*Node{
					"bar": {Value: "4560"},
					"baz": {Value: "789"},
				},
			},
			"strmap": {
				Children: map[string]*Node{
					"baz": {Value: "bam"},
					"zip": {Value: "zap"},
				},
			},
			"intslice": {
				Children: map[string]*Node{
					"0": {Value: "56"},
					"2": {Value: "90"},
				},
			},
			"strslice": {
				Children: map[string]*Node{
					"1": {Value: "baz"},
				},
			},
			"nested": {
				Children: map[string]*Node{
					"int": {Value: "1234"},
				},
			},
			"nmap": {
				Children: map[string]*Node{
					"foo": {
						Children: map[string]*Node{
							"int": {Value: "123"},
						},
					},
					"bar": {
						Children: map[string]*Node{
							"int": {Value: "456"},
						},
					},
				},
			},
			"nmapptr": {
				Children: map[string]*Node{
					"foo": {
						Children: map[string]*Node{
							"int": {Value: "78"},
						},
					},
					"bar": {
						Children: map[string]*Node{
							"int": {Value: "90"},
						},
					},
				},
			},
			"nslice": {
				Children: map[string]*Node{
					"0": {
						Children: map[string]*Node{
							"str": {Value: "updated_foo_in_slice"},
						},
					},
				},
			},
			"nsliceptr": {
				Children: map[string]*Node{
					"0": {
						Children: map[string]*Node{
							"str": {Value: "updated_foo_in_slice_ptr"},
						},
					},
				},
			},
			"sliceofmap": {
				Children: map[string]*Node{
					"0": {
						Children: map[string]*Node{
							"foo": {Value: "updated_map_in_slice"},
						},
					},
				},
			},
			"mapofslice": {
				Children: map[string]*Node{
					"foo": {
						Children: map[string]*Node{
							"0": {Value: "updated_slice_in_map"},
						},
					},
				},
			},
		},
	}

	errors = treeToUpdate.Write(reflect.ValueOf(&testStruct))
	require.False(t, errors.Present())

	expectedMergedStruct := testStructType{
		Str:  "bar",
		Int:  456,
		Bool: false,
		IntMap: map[string]int{
			"foo": 123,
			"bar": 4560,
			"baz": 789,
		},
		StrMap: map[string]string{
			"foo": "bar",
			"baz": "bam",
			"zip": "zap",
		},
		IntSlice: []int{56, 34, 90},
		StrSlice: []string{"foo", "baz", "bar"},
		Nested:   &testStructType{Str: "nested_foo", Int: 1234},
		NestedMap: map[string]testStructType{
			"foo": {
				Str: "nested_foo_in_map",
				Int: 123,
			},
			"bar": {
				Int: 456,
			},
		},
		NestedMapPtr: map[string]*testStructType{
			"foo": {
				Str: "nested_foo_in_map_ptr",
				Int: 78,
			},
			"bar": {
				Int: 90,
			},
		},
		NestedSlice:    []testStructType{{Str: "updated_foo_in_slice"}},
		NestedSlicePtr: []*testStructType{{Str: "updated_foo_in_slice_ptr"}},
		SliceOfMap:     []map[string]string{{"foo": "updated_map_in_slice"}},
		MapOfSlice:     map[string][]string{"foo": {"updated_slice_in_map"}},
	}

	assert.Equal(t, expectedMergedStruct, testStruct, "merging changes works correctly")

}
