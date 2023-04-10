package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/railsware/go-global/v2"
	"github.com/railsware/go-global/v2/utils"
)

const paramStoreSeparator = "/"

type LoadConfigOptions struct {
	ParamPrefix string
	// If IgnoreUnmappedParams is set, a parameter with no matching config field will be silently ignored.
	IgnoreUnmappedParams bool
}

// LoadConfigFromParameterStore retrieves keys configured in ParamStore and writes to config
// config must be a pointer to a struct.
// Keys in ParamStore must be separated with slashes.
// They are matched to struct fields by: name, `global:` tag, or `json:` tag
func LoadConfigFromParameterStore(awsConfig aws.Config, options LoadConfigOptions, globalConfig interface{}) (err global.Error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = global.NewError("global: panic while loading from parameter store: %v", panicErr)
			return
		}
	}()

	reflectedConfig, err := utils.ReflectConfig(globalConfig)
	if err != nil {
		return err
	}

	paramPaginator := ssm.NewGetParametersByPathPaginator(
		ssm.NewFromConfig(awsConfig),
		&ssm.GetParametersByPathInput{
			Path:           aws.String(options.ParamPrefix),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
		},
	)

	var params []param

	for paramPaginator.HasMorePages() {
		page, err := paramPaginator.NextPage(context.Background())
		if err != nil {
			return global.NewError("global: failed to load from Parameter Store: %v", err)
		}
		for _, ssmParam := range page.Parameters {
			paramNameWithoutPrefix := (*ssmParam.Name)[len(options.ParamPrefix):]
			params = append(params, param{paramNameWithoutPrefix, *ssmParam.Value})
		}
	}

	paramTree := buildParamTree(params)

	errors := paramTree.Write(reflectedConfig)

	return errors.Join()
}

type param struct {
	path  string
	value string
}
