package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteParamToConfig(t *testing.T) {

	type NestedConfig struct {
		Nested       string
		AnotherField string
	}

	type Config struct {
		String       string
		Bool         bool
		Int          int
		ByJSONTag    string `json:"by_json_tag"`
		ByGlobalTag  string `global:"by_global_tag"`
		NestedSimple NestedConfig
		NestedDeep   struct {
			SecondLevel NestedConfig
		}
		NestedPtr   *NestedConfig
		StringArray []string
		StructArray []*NestedConfig
	}

	var config Config

	var parameters = []struct {
		name  string
		value string
	}{
		{"String", "string"},
		{"Bool", "true"},
		{"Int", "25"},
		{"by_json_tag", "json_string"},
		{"by_global_tag", "global_string"},
		{"NestedSimple/Nested", "nested_string"},
		{"NestedDeep/SecondLevel/Nested", "deep_string"},
		{"NestedPtr/Nested", "nested_ptr_string"},
		{"NestedPtr/AnotherField", "second_ptr_string"},
		{"StringArray/1", "array_string_1"},
		{"StringArray/0", "array_string_0"},
		{"StructArray/0/Nested", "array_struct"},
	}

	for _, param := range parameters {
		destination, err := findParamDestination(&config, param.name)
		assert.Nil(t, err, "Failed to find param destination: %v", err)
		err = writeParamToConfig(destination, param.value)
		assert.Nil(t, err, "Failed to write param: %v", err)
	}

	assert.Equal(t, "string", config.String)
	assert.Equal(t, true, config.Bool)
	assert.Equal(t, 25, config.Int)
	assert.Equal(t, "json_string", config.ByJSONTag)
	assert.Equal(t, "global_string", config.ByGlobalTag)
	assert.Equal(t, "nested_string", config.NestedSimple.Nested)
	assert.Equal(t, "deep_string", config.NestedDeep.SecondLevel.Nested)
	assert.Equal(t, "nested_ptr_string", config.NestedPtr.Nested)
	assert.Equal(t, "second_ptr_string", config.NestedPtr.AnotherField)
	assert.Len(t, config.StringArray, 2)
	assert.Equal(t, "array_string_0", config.StringArray[0])
	assert.Equal(t, "array_string_1", config.StringArray[1])
	assert.Equal(t, "array_struct", config.StructArray[0].Nested)
}
