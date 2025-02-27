package envconfig_test

import (
	"testing"

	"github.com/h-dav/envconfig/v2"
)

func TestSetSuccessWithOneField(t *testing.T) {
	type Config struct {
		Example string `env:"KEY"`
	}

	var config Config

	want := Config{
		Example: "value1",
	}

	envconfig.Set("./test_data/success_with_one_field.env", &config)

	if config != want {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

func TestSetSuccessWithDefaultValueAndEmptyEnvFile(t *testing.T) {
	type Config struct {
		Example string `env:"DEFAULT_VALUE" default:"value2"`
	}

	var config Config

	want := Config{
		Example: "value2",
	}

	envconfig.Set("./test_data/success_with_one_default_value_and_empty_env_file.env", &config)

	if config != want {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

func TestSetSuccessWithRequiredField(t *testing.T) {
	type Config struct {
		Example string `env:"REQUIRED_VALUE" required:"true"`
	}

	var config Config

	want := Config{
		Example: "value3",
	}

	envconfig.Set("./test_data/success_with_required_field.env", &config)

	if config != want {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

func TestSetSuccessWithNestedStruct(t *testing.T) {
	type Config struct {
		Server struct {
			Port string `env:"PORT"`
		} `prefix:"SERVER_"`
	}

	var config Config

	var want Config
	want.Server.Port = "8080"

	envconfig.Set("./test_data/success_with_nested_struct.env", &config)

	if config != want {
		t.Errorf("got %+v, want %+v", config, want)
	}
}
