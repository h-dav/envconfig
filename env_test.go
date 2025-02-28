package envconfig_test

import (
	"slices"
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

func TestSetSuccessWithDeeplyNestedStruct(t *testing.T) {
	type Config struct {
		Server struct {
			Port struct {
				Value string `env:"VALUE"`
			} `prefix:"PORT_"`
		} `prefix:"SERVER_"`
	}

	var config Config

	var want Config
	want.Server.Port.Value = "1234"

	envconfig.Set("./test_data/success_with_deeply_nested_struct.env", &config)

	if config != want {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

func TestSetSuccessWithThriceDeeplyNestedStruct(t *testing.T) {
	type Config struct {
		Server struct {
			Database struct {
				Tables struct {
					First string `env:"FIRST"`
				} `prefix:"TABLES_"`
				Timezome string `env:"TIMEZONE"`
			} `prefix:"DATABASE_"`
		} `prefix:"SERVER_"`
	}

	var config Config

	var want Config
	want.Server.Database.Tables.First = "example_table"
	want.Server.Database.Timezome = "uk/london"

	envconfig.Set("./test_data/success_with_thrice_deeply_nested_struct.env", &config)

	if config != want {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

func TestSetSuccessWithSliceStringField(t *testing.T) {
	type Config struct {
		SliceStringField []string `env:"SLICE_STRING_FIELD"`
	}

	var config Config

	var want Config
	want.SliceStringField = []string{"first", "second", "third"}

	envconfig.Set("./test_data/success_with_slice_string_field.env", &config)

	if !slices.Equal(config.SliceStringField, want.SliceStringField) {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

func TestSetSuccessWithSliceIntField(t *testing.T) {
	type Config struct {
		SliceIntField []int `env:"SLICE_INT_FIELD"`
	}

	var config Config

	var want Config
	want.SliceIntField = []int{1, 2, 3}

	envconfig.Set("./test_data/success_with_slice_int_field.env", &config)

	if !slices.Equal(config.SliceIntField, want.SliceIntField) {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

func TestSetSuccessWithSliceFloatField(t *testing.T) {
	type Config struct {
		SliceFloatField []float64 `env:"SLICE_FLOAT_FIELD"`
	}

	var config Config

	var want Config
	want.SliceFloatField = []float64{1.2, 2.3, 3.4}

	envconfig.Set("./test_data/success_with_slice_float_field.env", &config)

	if !slices.Equal(config.SliceFloatField, want.SliceFloatField) {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

func TestSetSuccessWithTextReplacement(t *testing.T) {
	type Config struct {
		ReplaceField string `env:"REPLACE_FIELD"`
	}

	var config Config

	var want Config
	want.ReplaceField = "exampleField"

	envconfig.Set("./test_data/success_with_text_replacement.env", &config)

	if config != want {
		t.Errorf("got %+v, want %+v", config, want)
	}
}
