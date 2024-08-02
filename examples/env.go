package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/h-dav/envconfig"
)

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

	fmt.Printf("Populated Config: %v", cfg)
}
