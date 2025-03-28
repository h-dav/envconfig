# envconfig

[![Go Reference](https://pkg.go.dev/badge/github.com/h-dav/envconfig.svg)](https://pkg.go.dev/github.com/h-dav/envconfig)
[![Go Report Card](https://goreportcard.com/badge/github.com/h-dav/envconfig/v2)](https://goreportcard.com/report/github.com/h-dav/envconfig/v2)
[![Test](https://github.com/h-dav/envconfig/actions/workflows/test.yml/badge.svg)](https://github.com/h-dav/envconfig/actions/workflows/test.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/h-dav/envconfig/blob/main/LICENSE)

Package `envconfig` provides functionality to easily populate your config structure by using both environment variables, and a .env file (optional).

- [Installation](#installation)
- [Features](#features)
    - [Options](#options)
    - [Supported Data Types](#supported-data-types)
- [Usage](#usage)

## Installation

```bash
go get github.com/h-dav/envconfig/v2
```

## Features

### Options

- `required`: `true` or `false`
- `default`: Default value if environment variable is not set.
- `prefix`: Used for nested structures.
- `envjson`: Used for deserialising json into config.
- Text Replacement: `${EXAMPLE}` can be used to insert other environment variables.

### Supported Data Types

- string (slice compatible)
- int (slice compatible)
- float (slice compatible)
- bool
- time.Duration

## Usage

```go
package main

import (
    "time"

    "github.com/h-dav/envconfig/v2"
)

type Config struct {
    // Fields must be exported to be populated.
    LogLevel string `env:"LOG_LEVEL" default:"info"`
    Server struct {
        Port string `env:"PORT" required:"true"`
    } `prefix:"SERVER_"`
    JSONField struct {
        First string `json:"first"`
    } `envjson:"JSON_FIELD"`
    SliceIntField []int `env:"SLICE_INT_FIELD"`
    DurationField time.Duration `env:"DURATION"`
}

func main() {
    var cfg Config

    if err := envconfig.Set("./config/default.env", &cfg); err != nil {
        ...
    }
}
```

Corresponding .env file:

```env
LOG_LEVEL=debug
SERVER_PORT=8080
JSON_FIELD={"first": "example"}
SLICE_INT_FIELD=1, 2, 3
DURATION=30s
```

> [!NOTE]
> See [test cases](./env_test.go) for more usage examples.

> [!NOTE]
> This package takes heavy inspiration from [httputil](https://github.com/nickbryan/httputil) for handling reflection.
