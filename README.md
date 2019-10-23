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

`go get github.com/Syncbak-Git/go-figgy`

## Getting started

It's as simple as defining a struct, decorating it with tags, and loading it.

```Go
type Config struct{
    Server   string `ssm:"/myapp/prod/server"`
    Port     int    `ssm:"/myapp/prod/port"`
    Password string `ssm:"/myapp/prod/password,decrypt"`
}

//... meanwhile, more handwaving
cfg := Config{}
figgy.Load(ssmClient, &cfg)
```

## Runtime parameters

You can have a parameter defined at runtime by using the `LoadWithParameters` function:

``` go
type Config struct{
    Server   string `ssm:"/myapp/{{.env}}/server"`
    Port     int    `ssm:"/myapp/{{.env}}/port"`
    Password string `ssm:"/myapp/{{.env}}/password,decrypt"`
}

cfg := Config{}
figgy.LoadWithParameters(ssmClient, &cfg, figgy.P{"env": "prod"})
```

Using `Server` as an example, this will be computed to a key of `/myapp/prod/server` at runtime.

## The Future

Here are some additional features we would like to see in the near future:

- Support type conversions for map type and slices of structs
- Allow tags defined on a parent struct to influence the child field tags
  - This is similar to how the xml package handles unmarshaling