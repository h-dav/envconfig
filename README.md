# envconfig

A low dependency package for .env files.

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
	} `env:"HTTP_,prefix"`
}

func main() {
	cfg := ExampleConfig{}
	if err := envconfig.SetPopulate(filepath.Join("examples", "example.env"), &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Populated Config: %v", cfg)
}
```

.env file:

```
EXAMPLE=value
ANOTHER_VALUE=v0.0.0
INT_EXAMPLE=4
INT32_EXAMPLE=23
FLOAT_EXAMPLE=4.44
HTTP_PORT=9999
HTTP_NAME=example_name
```

Output:

```
Populated Config: {value v0.0.0 4 23 4.44 {9999 example_name}}
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
    } `env:"HTTP_,prefix"`
}

cfg := ExampleConfig{}
if err := envconfig.Populate(&cfg); err != nil {
    panic(err)
}
...
```

## Supported types

- Int, Int32, Int64
- String
- Float64

[![Go Reference](https://pkg.go.dev/badge/github.com/h-dav/envconfig.svg)](https://pkg.go.dev/github.com/h-dav/envconfig)
