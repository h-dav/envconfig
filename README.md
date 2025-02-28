# envconfig

[![Go Reference](https://pkg.go.dev/badge/github.com/h-dav/envconfig.svg)](https://pkg.go.dev/github.com/h-dav/envconfig)
[![Go Report Card](https://goreportcard.com/badge/github.com/h-dav/envconfig/v2)](https://goreportcard.com/report/github.com/h-dav/envconfig/v2)
[![Test](https://github.com/h-dav/envconfig/actions/workflows/test.yml/badge.svg)](https://github.com/h-dav/envconfig/actions/workflows/test.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/h-dav/envconfig/blob/main/LICENSE)

A low dependency package for .env files.

## Install

```bash
go get github.com/h-dav/envconfig
```

## Usage

Create your config structure:

```go
type Config struct {
    LogLevel string `env:"LOG_LEVEL" default:"info"`
    Server struct {
        Port string `env:"PORT" required:"true"`
    } `prefix:"SERVER_"`
}
```

Then just read pass your config structure along with the path to your .env file to envconfig.Set:

```go
var cfg Config
err := envconfig.Set("./config/default.env", &cfg)
```

The corresponding .env file for this example:

```env
LOG_LEVEL: debug
SERVER_PORT: 8080
```

## Options

- `required`: `true` or `false`
- `default`: Fall back value if environment variable is not set.
- `prefix`: Used for nested structures.
- Text Replacement: `${EXAMPLE}` can be used to insert other environment variables.

## Supported data types

- string
- int
- float
- bool
- []string
- []int
- []float

> [!NOTE]
> This package takes heavy inspiration from [httputil](https://github.com/nickbryan/httputil) for handling reflection.
