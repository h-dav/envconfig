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

	os.Setenv("VALUE", "value")

	var cfg Config

	envconfig.Set(&cfg)

	fmt.Println(cfg.Value)
	// Output:
	// value
}
