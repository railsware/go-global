# Go companion modules for Ruby Global gem

This repository contains modules to load configuration analogous to the [Global gem](https://github.com/railsware/global). Refere to Global's documentation for more details

## AWS Parameter Store usage example

See [main Global documentation](https://github.com/railsware/global#aws-parameter-store-1) for setup instructions.

```sh
go get github.com/railsware/go-global/v2/aws
```

```go
import (
  "os"
  awsConfig "github.com/aws/aws-sdk-go-v2/config"
  globalAWS "github.com/railsware/go-global/v2/aws"
)

type Config struct {
  Database struct {
    PoolSize int `json:"pool_size"`
    URLs []string `json:"urls"`
  } `json:"database"`

}

var config Config

awsConfig, err := awsConfig.LoadDefaultConfig(context.TODO())
if err != nil {
  log.Fatalf("cannot load aws config: %v", err)
}
awsParamPrefix := os.Getenv("AWS_PARAM_PREFIX")

err := globalAWS.LoadConfigFromParameterStore(
  awsConfig,
  globalAWS.LoadConfigOptions{ParamPrefix: awsParamPrefix},
  &config
)

// config.Database.URLs[0] loaded from /param_prefix/database/urls/0
// config.Database.PoolSize loaded from /param_prefix/database/pool_size
```

Supported value types: `string`, `int`, `bool` ("true"/"false").

Complex type should be either a `struct`, or a `slice`. You can arbitrarily nest structs and slices.

For structs, use `global` or `json` tag to set field name.

For slices, all subscripts in Parameter Store must be integers.

### Shorthand for running on AWS ECS or Lambda

If you use Global in AWS environments, you can DRY up the code by following a convention:

- Set up param prefix in the `AWS_PARAM_PREFIX` environment variable.
- Make sure the IAM role has permissions to the params

Then you can simply use:

```go
var config Config
// Will panic if anything goes wrong
globalAWS.MustLoadConfig(&config)
```

### Handling unmapped params

By default, Global assumes that all params from Parameter Store must be mapped to config fields, and returns a warning if some param does not correspond to a config field.

Sometimes, it is expected that the config only contains a subset of the params, and such a warning is unnecessary.

In this case, set `options.IgnoreUnmappedParams` to true. Note that other mapping issues, like a type mismatch, will still cause an error.
