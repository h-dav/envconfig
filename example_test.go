package envconfig_test

import (
	"fmt"
	"os"

	"github.com/h-dav/envconfig/v3"
)

func ExampleSet() {
	type Config struct {
		Value string `env:"VALUE"`
	}

	if err := os.Setenv("VALUE", "value"); err != nil {
		panic(err)
	}

	var cfg Config

	if err := envconfig.Set(&cfg); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Value)
	// Output:
	// value
}

func ExampleSet_advanced() {
	type Config struct {
		Service string `env:"SERVICE,required"`
		Port    int    `env:"PORT,default=8080"`
	}

	if err := os.Setenv("SERVICE", "auth"); err != nil {
		panic(err)
	}
	// PORT is not set, so it will use the default value.

	var cfg Config

	if err := envconfig.Set(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf("Service: %s, Port: %d\n", cfg.Service, cfg.Port)
	// Output:
	// Service: auth, Port: 8080
}
