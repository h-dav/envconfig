package main

import (
	"log"
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
	err := envconfig.SetPopulate(filepath.Join("examples", "example.env"), &cfg)
	if err != nil {
		log.Fatal(err)
	}
}
