package aws

import (
	"testing"

	"github.com/railsware/go-global/v2/tree"
	"github.com/stretchr/testify/assert"
)

func TestBuildTree(t *testing.T) {
	params := []param{
		{"String", "string"},
		{"NestedSimple/Nested", "nested_string"},
		{"NestedDeep/SecondLevel/Nested", "deep_string"},
		{"NestedDeep/SecondLevel/Nested2", "deep_string_2"},
	}

	expectedTree := &tree.Node{
		Children: map[string]*tree.Node{
			"String": {Value: "string"},
			"NestedSimple": {
				Children: map[string]*tree.Node{
					"Nested": {Value: "nested_string"},
				},
			},
			"NestedDeep": {
				Children: map[string]*tree.Node{
					"SecondLevel": {
						Children: map[string]*tree.Node{
							"Nested":  {Value: "deep_string"},
							"Nested2": {Value: "deep_string_2"},
						},
					},
				},
			},
		},
	}

	paramTree := buildParamTree(params)

	assert.Equal(t, expectedTree, paramTree)
}
