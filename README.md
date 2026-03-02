# go-figgy

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/Syncbak-Git/go-figgy)
[![CircleCI](https://circleci.com/gh/Syncbak-Git/go-figgy/tree/master.svg?style=shield)](https://circleci.com/gh/Syncbak-Git/go-figgy/tree/master)
[![codecov](https://codecov.io/gh/Syncbak-Git/go-figgy/branch/master/graph/badge.svg)](https://codecov.io/gh/Syncbak-Git/go-figgy)
[![Go Report Card](https://goreportcard.com/badge/github.com/Syncbak-Git/go-figgy)](https://goreportcard.com/report/github.com/Syncbak-Git/go-figgy)

## Why is this a thing?!
We wanted to experiment with AWS's Parameter Store as a centralized system for managing out configurations.  Turns out, it is a lot of work loading them and pushing the values into configuration structs to be used by other components.

Our solution was to use Go's awesome tag feature to ease the burden of using the SSM SDK directly.  This allows us to define our configuration in the struct itself and populate the struct's fields with values when loaded!

TLDR: Tags are awesome and injecting configuration from AWS into our structs with them is even awesomer!

## Install

```
go get github.com/Syncbak-Git/go-figgy/v2
```

## Getting started

It's as simple as defining a struct, decorating it with tags, and loading it.

### Using SSM Parameter Store

```go
import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ssm"
    figgy "github.com/Syncbak-Git/go-figgy/v2"
)

type Config struct{
    Server   string `ssm:"/myapp/prod/server"`
    Port     int    `ssm:"/myapp/prod/port"`
    Password string `ssm:"/myapp/prod/password,decrypt"`
}

sess := session.Must(session.NewSession())
client := figgy.NewSSMClient(ssm.New(sess), true)

cfg := Config{}
figgy.Load(client, &cfg)
```

The second argument to `NewSSMClient` controls whether parameters are fetched with `WithDecryption` enabled. Set to `true` if your parameters include SecureString values.

### Using Secrets Manager

```go
import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/secretsmanager"
    figgy "github.com/Syncbak-Git/go-figgy/v2"
)

type Config struct{
    Server   string `ssm:"myapp-prod-server"`
    Port     int    `ssm:"myapp-prod-port"`
    Password string `ssm:"myapp-prod-password"`
}

sess := session.Must(session.NewSession())
client := figgy.NewSecretsManagerClient(secretsmanager.New(sess))

cfg := Config{}
figgy.Load(client, &cfg)
```

Note: Secrets Manager values are always encrypted, so the `,decrypt` tag option is not needed. The `ssm` struct tag name is used for both backends.

### Backend differences

| | SSM Parameter Store | Secrets Manager |
|---|---|---|
| Constructor | `figgy.NewSSMClient(api, decrypt)` | `figgy.NewSecretsManagerClient(api)` |
| Key format | Path-based (`/myapp/prod/key`) | Flat names (`myapp-prod-key`) |
| `,decrypt` tag | Honored (controls `WithDecryption`) | Ignored (always encrypted) |
| Batch size | 10 parameters per API call | 1 secret per API call |
| Cost | Free (standard parameters) | $0.40/secret/month |

## Runtime parameters

You can have a parameter defined at runtime by using the `LoadWithParameters` function:

```go
type Config struct{
    Server   string `ssm:"/myapp/{{.env}}/server"`
    Port     int    `ssm:"/myapp/{{.env}}/port"`
    Password string `ssm:"/myapp/{{.env}}/password,decrypt"`
}

cfg := Config{}
figgy.LoadWithParameters(client, &cfg, figgy.P{"env": "prod"})
```

Using `Server` as an example, this will be computed to a key of `/myapp/prod/server` at runtime.

## Custom backends

`Load` and `LoadWithParameters` accept any implementation of the `figgy.Client` interface:

```go
type Client interface {
    GetValues(keys []string) (map[string]string, error)
    BatchSize() int
}
```

This makes it easy to use environment variables for local development, in-memory maps for testing, or any other key-value source.

## Migrating from v1

v2 introduces a `Client` interface abstraction. The `Load` and `LoadWithParameters` functions no longer accept `ssmiface.SSMAPI` directly.

```go
// v1
figgy.Load(ssmClient, &cfg)

// v2
figgy.Load(figgy.NewSSMClient(ssmClient, true), &cfg)
```

## The Future

Here are some additional features we would like to see in the near future:

- Support type conversions for map type and slices of structs
- Allow tags defined on a parent struct to influence the child field tags
  - This is similar to how the xml package handles unmarshaling
