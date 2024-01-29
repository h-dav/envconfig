# envconfig

A low dependency package for parsing a .env file.

## Set environment variables and populate a struct

Code:
```go
package main

import (
	"path/filepath"
	
	"gitlab.com/hcdav/envconfig"
)

func main() {
	type ExampleConfig struct {
		Example      string `env:"EXAMPLE,required"`
		AnotherValue string `env:"ANOTHER_VALUE"`
		Service      struct {
			Port string `env:"PORT"`
			Name string `env:"NAME,required"`
		} `env:"HTTP_,prefix"`
	}

	cfg := ExampleConfig{}
	if err := envconfig.SetPopulate(filepath.Join("examples", "example.env"), &cfg); err != nil {
		panic(err)
	}
}
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
    Example      string `env:"EXAMPLE,required"`
    AnotherValue string `env:"ANOTHER_VALUE"`
    Service      struct {
        Port string `env:"PORT"`
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

- Int
- String
- Float64

[![Go Reference](https://pkg.go.dev/badge/gitlab.com/hcdav/envconfig.svg)](https://pkg.go.dev/gitlab.com/hcdav/envconfig)
