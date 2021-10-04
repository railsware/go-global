# Go companion modules for Ruby Global gem

This repository contains modules to load configuration analogous to the [Global gem](https://github.com/railsware/global). Refere to Global's documentation for more details

## AWS Parameter Store usage example

See [main Global documentation](https://github.com/railsware/global#aws-parameter-store-1) for setup instructions.

```sh
go get github.com/railsware/go-global/aws
```

```go
import (
  "os"
  "github.com/aws/aws-sdk-go/aws/session"
  globalAWS "github.com/railsware/go-global/aws"
)

type Config struct {
  Database struct {
   URL string `json:"url"`
    PoolSize int `json:"pool_size"`
  } `json:"database"`
}

var config Config

awsSession := session.Must(session.NewSession())
awsParamPrefix := os.Getenv("AWS_PARAM_PREFIX")

err := globalAWS.LoadConfigFromParameterStore(awsSession, awsParamPrefix, &config)

// config.Database.URL loaded from /param_prefix/database/url
// config.Database.PoolSize loaded from /param_prefix/database/pool_size
```

Supported value types: `string`, `int`, `bool` ("true"/"false").
