package aws

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

// MustLoadConfig is a conventional method for loading the config during initialization.
// - uses "default" AWS configuration
// - uses the environment variable AWS_PARAM_PREFIX as the param prefix
// - ignores unmapped params
// - panics if anything goes wrong
// The suggested application is in the initialization of an AWS ECS service or Lambda function.
func MustLoadConfig(config interface{}) {
	if reflect.ValueOf(config).Kind() != reflect.Ptr {
		panic(errors.New("paramstore.MustLoadConfig: must pass pointer to the config struct"))
	}

	awsConfig, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("paramstore.MustLoadConfig: Cannot load AWS config: %v", err))
	}
	awsParamPrefix := os.Getenv("AWS_PARAM_PREFIX")
	if awsParamPrefix == "" {
		panic(errors.New("paramstore.MustLoadConfig: AWS_PARAM_PREFIX is required"))
	}
	configErr := LoadConfigFromParameterStore(
		awsConfig,
		LoadConfigOptions{
			ParamPrefix:          awsParamPrefix,
			IgnoreUnmappedParams: true,
		},
		config,
	)
	if configErr != nil {
		panic(fmt.Errorf("paramstore.MustLoadConfig: failed to load config: %v", configErr))
	}
}
