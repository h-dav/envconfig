package main

import (
	"fmt"
	"log"
	"path/filepath"

	"gitlab.com/harry.davidson/envconfig"
)

func main() {
	type ExampleConfig struct {
		Example      string `env:"EXAMPLE"`
		AnotherValue string `env:"ANOTHER_VALUE"`
		Service      struct {
			Port string `env:"PORT"`
		} `env:"HTTP_,prefix"`
	}

	cfg := ExampleConfig{}
	err := envconfig.SetPopulate(filepath.Join("examples", "example.env"), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg)
}
