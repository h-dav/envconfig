package main

import (
	"fmt"
	"github.com/h-dav/envconfig/v2"
	"log"
)

type Config struct {
	Example     string `env:"KEY"`
	SecondKey   int    `env:"SECOND_KEY" required:"true"`
	DefaulteKey bool   `env:"DEFAULTED_KEY" default:"false"`
	Server      struct {
		Port string `env:"PORT"`
	} `prefix:"SERVER_"`
}

func main() {
	var cfg Config

	if err := envconfig.Set("./config/example.env", &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Populated Config: %+v\n", cfg)
}
