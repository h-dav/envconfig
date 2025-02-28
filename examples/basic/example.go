package main

import (
	"fmt"
	"log"

	"github.com/h-dav/envconfig/v2"
)

type config struct {
	Example     string `env:"KEY"`
	SecondKey   int    `env:"SECOND_KEY" required:"true"`
	DefaulteKey bool   `env:"DEFAULTED_KEY" default:"false"`
	Server      struct {
		Port string `env:"PORT"`
	} `prefix:"SERVER_"`
	List []string `env:"LIST"`
}

func main() {
	var cfg config

	if err := envconfig.Set("./config/example.env", &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Populated Config: %+v\n", cfg)
}
