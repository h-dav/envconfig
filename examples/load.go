package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/h-dav/envconfig"
)

type ExampleConfig struct {
	Example      string `env:"EXAMPLE"`
	AnotherValue string `env:"ANOTHER_VALUE"`
}

func main() {
	load()
	example := ExampleConfig{}
	err := envconfig.Populate(&example)
	if err != nil {
		return
	}
	fmt.Println(example)
}

func load() {
	err := envconfig.SetVars(filepath.Join("examples", "example.env"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(os.Getenv("EXAMPLE"))
}
