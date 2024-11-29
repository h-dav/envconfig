# envconfig

A low dependency package for .env files (or just environment variables).

## Install

```bash
go get github.com/h-dav/envconfig
```

## Usage

### Set environment variables and populate a struct:

Code:

```go
package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/h-dav/envconfig"
)

// ExampleConfig is your config struct using `env` struct tags.
type ExampleConfig struct {
	Example      string  `env:"EXAMPLE,required"`
	AnotherValue string  `env:"ANOTHER_VALUE"`
	IntExample   int     `env:"INT_EXAMPLE"`
	Int32Example int32   `env:"INT32_EXAMPLE"`
	FloatExample float64 `env:"FLOAT_EXAMPLE"`
	Service      struct {
		Port int64  `env:"PORT"`
		Name string `env:"NAME,required"`
	} `env:"prefix=HTTP_"`
	ExampleEndpoint string `env:"EXAMPLE_ENDPOINT"`
	DefaultValue    string `env:"DEFAULT_VALUE,default=thevalue"`
}

func main() {
	cfg := ExampleConfig{}
	if err := envconfig.SetPopulate(filepath.Join("config", "example.env"), &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Populated Config: %+v\n", cfg)
}
```

.env file:

```
EXAMPLE: value
ANOTHER_VALUE: v0.0.0
INT_EXAMPLE: 4
INT32_EXAMPLE: 23
#COMMENTED_EXAMPLE: test
FLOAT_EXAMPLE: 4.44
HTTP_PORT: 9999
HTTP_NAME: example_name
DNS: example.com
EXAMPLE_ENDPOINT: https://${DNS}/v1
```

Output:

```
Populated Config: {Example:value AnotherValue:v0.0.0 IntExample:4 Int32Example:23 FloatExample:4.44 Service:{Port:9999 Name:example_name} ExampleEndpoint:https://example.com/v1 DefaultValue:thevalue}
```

### Example - Set environment variables

Code:

```go
...
if err := envconfig.SetVars(filepath.Join("examples", "example.env")); err != nil {
	panic(err)
}
...
```

### Example - Populate a struct with environment variables

Code:

```go
...
type ExampleConfig struct {
	Example      string  `env:"EXAMPLE,required"`
	AnotherValue string  `env:"ANOTHER_VALUE"`
	IntExample   int     `env:"INT_EXAMPLE"`
	Int32Example int32   `env:"INT32_EXAMPLE"`
	FloatExample float64 `env:"FLOAT_EXAMPLE"`
	Service      struct {
		Port int64  `env:"PORT"`
		Name string `env:"NAME,required"`
	} `env:"prefix=HTTP_"`
	ExampleEndpoint string `env:"EXAMPLE_ENDPOINT"`
	DefaultValue    string `env:"DEFAULT_VALUE,default=thevalue"`
}

cfg := ExampleConfig{}
if err := envconfig.Populate(&cfg); err != nil {
    panic(err)
}
...
```

## Options

| Tag Option | Type       | Example                |
| ---------- | ---------- | ---------------------- |
| default    | assignment | env:"KEY,default=dval" |
| prefix     | assignment | env:"prefix=pstruct_"  |
| required   | flag       | env:"KEYNAME,required" |

## Supported types

- Int, Int32, Int64
- String
- Float64

[![Go Reference](https://pkg.go.dev/badge/github.com/h-dav/envconfig.svg)](https://pkg.go.dev/github.com/h-dav/envconfig)
