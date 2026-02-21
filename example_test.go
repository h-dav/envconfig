package envconfig_test

import (
	"fmt"
	"os"

	"github.com/h-dav/envconfig/v3"
)

func ExampleSet() {
	type Config struct {
		Value string `config:"VALUE"`
	}

	if err := os.Setenv("VALUE", "value"); err != nil {
		panic(err)
	}
	defer os.Unsetenv("VALUE")

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
		Service string `config:"SERVICE,required"`
		Port    int    `config:"PORT,default=8080"`
	}

	if err := os.Setenv("SERVICE", "auth"); err != nil {
		panic(err)
	}
	defer os.Unsetenv("SERVICE")
	// PORT is not set, so it will use the default value.

	var cfg Config

	if err := envconfig.Set(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf("Service: %s, Port: %d\n", cfg.Service, cfg.Port)
	// Output:
	// Service: auth, Port: 8080
}

func ExampleSet_withOptions() {
	type Config struct {
		DBName string `config:"DB_NAME"`
	}

	// Example showing WithPrefix
	if err := os.Setenv("APP_DB_NAME", "users_db"); err != nil {
		panic(err)
	}
	defer os.Unsetenv("APP_DB_NAME")

	var cfg Config
	if err := envconfig.Set(&cfg, envconfig.WithPrefix("APP_")); err != nil {
		panic(err)
	}

	fmt.Println(cfg.DBName)
	// Output:
	// users_db
}
