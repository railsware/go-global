package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
)

type mockParamPaginator struct {
	pages [][]types.Parameter
	count int
}

func (m *mockParamPaginator) HasMorePages() bool {
	m.count++
	return len(m.pages) >= m.count
}

func (m mockParamPaginator) NextPage(_ context.Context, _ ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	return &ssm.GetParametersByPathOutput{
		Parameters: m.pages[m.count-1],
	}, nil
}

type nested struct {
	A    string
	PtrB *string
}

type config struct {
	Number    int
	String    string
	Bool      bool
	Slice     []string
	Ptr       *string
	GlobalTag string `global:"global_tag"`
	JSONTag   string `json:"json_tag"`
	Nested    struct {
		Nested struct {
			String string
		}
		String string
	}
	MultipleKeysStringSlice    []string
	MultipleKeysPtrStringSlice []*string
	MultipleKeysStructSlice    []nested
}

func TestLoadConfigFromParamPaginator(t *testing.T) {
	expected := config{
		Number:                     23,
		String:                     "test",
		Bool:                       true,
		Slice:                      []string{"one", "two"},
		GlobalTag:                  "tag",
		JSONTag:                    "json",
		Ptr:                        aws.String("pointer"),
		MultipleKeysStringSlice:    []string{"a", "b"},
		MultipleKeysPtrStringSlice: []*string{aws.String("c"), aws.String("d")},
		MultipleKeysStructSlice: []nested{
			{
				A: "boom", PtrB: aws.String("bam"),
			},
			{
				A: "yes", PtrB: aws.String("no"),
			},
		},
	}
	expected.Nested.Nested.String = "nested nested"
	expected.Nested.String = "string"
	prefix := "/test"
	pages := [][]types.Parameter{
		{
			{
				Name:  aws.String(prefix + "/Number"),
				Value: aws.String(strconv.Itoa(expected.Number)),
			},
			{
				Name:  aws.String(prefix + "/String"),
				Value: aws.String(expected.String),
			},
			{
				Name:  aws.String(prefix + "/Ptr"),
				Value: expected.Ptr,
			},
			{
				Name:  aws.String(prefix + "/Bool"),
				Value: aws.String(strconv.FormatBool(expected.Bool)),
			},
			{
				Name:  aws.String(prefix + "/Slice"),
				Value: aws.String(strings.Join(expected.Slice, ",")),
			},
		},
		{
			{
				Name:  aws.String(prefix + "/global_tag"),
				Value: aws.String(expected.GlobalTag),
			},
			{
				Name:  aws.String(prefix + "/json_tag"),
				Value: aws.String(expected.JSONTag),
			},
		},
		{
			{
				Name:  aws.String(prefix + "/Nested/Nested/String"),
				Value: aws.String(expected.Nested.Nested.String),
			},
			{
				Name:  aws.String(prefix + "/Nested/String"),
				Value: aws.String(expected.Nested.String),
			},
		},
		{
			{
				Name:  aws.String(prefix + "/MultipleKeysStringSlice/0"),
				Value: aws.String(expected.MultipleKeysStringSlice[0]),
			},
			{
				Name:  aws.String(prefix + "/MultipleKeysStringSlice/1"),
				Value: aws.String(expected.MultipleKeysStringSlice[1]),
			},
			{
				Name:  aws.String(prefix + "/MultipleKeysPtrStringSlice/0"),
				Value: expected.MultipleKeysPtrStringSlice[0],
			},
			{
				Name:  aws.String(prefix + "/MultipleKeysPtrStringSlice/1"),
				Value: expected.MultipleKeysPtrStringSlice[1],
			},
			{
				Name:  aws.String(prefix + "/MultipleKeysStructSlice/0/A"),
				Value: aws.String(expected.MultipleKeysStructSlice[0].A),
			},
			{
				Name:  aws.String(prefix + "/MultipleKeysStructSlice/0/PtrB"),
				Value: expected.MultipleKeysStructSlice[0].PtrB,
			},
			{
				Name:  aws.String(prefix + "/MultipleKeysStructSlice/1/A"),
				Value: aws.String(expected.MultipleKeysStructSlice[1].A),
			},
			{
				Name:  aws.String(prefix + "/MultipleKeysStructSlice/1/PtrB"),
				Value: expected.MultipleKeysStructSlice[1].PtrB,
			},
		},
	}
	m := mockParamPaginator{pages: pages}
	cfg := config{}
	require.NoError(t, LoadConfigFromParamPaginator(&m, LoadConfigOptions{ParamPrefix: prefix}, &cfg))

	require.Equal(t, expected, cfg)
}

func TestLoadConfigFromParamPaginatorNoParams(t *testing.T) {
	m := mockParamPaginator{}
	cfg := config{}
	require.Error(t, LoadConfigFromParamPaginator(&m, LoadConfigOptions{}, &cfg))
	require.NoError(t, LoadConfigFromParamPaginator(&m, LoadConfigOptions{IgnoreUnmappedParams: true}, &cfg))
}
