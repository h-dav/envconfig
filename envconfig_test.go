package envconfig_test

import (
	"bufio"
	"log"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/h-dav/envconfig/v3"
)

type SuccessWithOneField struct {
	Example string `env:"KEY"`
}
type SuccessWithOneIntField struct {
	Example int `env:"KEY"`
}

type SuccessWithDefaultValueAndEmptyEnvFile struct {
	Example string `env:"DEFAULT_VALUE" default:"value2"`
}

type SuccessWithRequiredField struct {
	Example string `env:"REQUIRED_VALUE" required:"true"`
}

type SuccessWithTextReplacement struct {
	ReplaceField string `env:"REPLACE_FIELD"`
}

type SuccessWithSettingTimeDuration struct {
	Duration time.Duration `env:"DURATION"`
}

type SuccessWithPrefixOption struct {
	Duration time.Duration `env:"DURATION"`
}

// TestSet is test cases for simple use cases,
// such as flat config structures and fundamental fields, like required, and default.
func TestSet(t *testing.T) {
	type testCase struct {
		filename string
		want     any
		assert   func(*testing.T, testCase)
	}

	testCases := map[string]testCase{
		"success with one field": {
			filename: "./test_data/success_with_one_field.env",
			want: SuccessWithOneField{
				Example: "value1",
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithOneField

				if err := envconfig.Set(&config); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
		"success with one int field": {
			filename: "./test_data/success_with_one_int_field.env",
			want: SuccessWithOneIntField{
				Example: 10,
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithOneIntField

				if err := envconfig.Set(&config); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
		"success with default value and empty env file": {
			filename: "./test_data/success_with_one_default_value_and_empty_env_file.env",
			want: SuccessWithDefaultValueAndEmptyEnvFile{
				Example: "value2",
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithDefaultValueAndEmptyEnvFile

				if err := envconfig.Set(&config); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
		"success with required field": {
			filename: "./test_data/success_with_required_field.env",
			want: SuccessWithRequiredField{
				Example: "value3",
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithRequiredField

				if err := envconfig.Set(&config); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
		// "success with text replacement": {
		// 	filename: "./test_data/success_with_text_replacement.env",
		// 	want: SuccessWithTextReplacement{
		// 		ReplaceField: "exampleField",
		// 	},
		// 	assert: func(t *testing.T, tc testCase) {
		// 		t.Helper()
		//
		// 		var config SuccessWithTextReplacement
		//
		// 		if err := envconfig.Set(&config); err != nil {
		// 			t.Fail()
		// 		}
		//
		// 		if config != tc.want {
		// 			t.Errorf("got %+v, want %+v", config, tc.want)
		// 		}
		// 	},
		// },
		"success with setting time.Duration": {
			filename: "./test_data/success_with_setting_time_Duration.env",
			want: SuccessWithSettingTimeDuration{
				Duration: 10000000000,
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithSettingTimeDuration

				if err := envconfig.Set(&config); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn,
			func(t *testing.T) {
				t.Parallel()

				loadFileIntoEnvironmentVariables(tc.filename)

				tc.assert(t, tc)
			},
		)
	}
}

func loadFileIntoEnvironmentVariables(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, _ := strings.Cut(scanner.Text(), "=")

		// Clean environment variable key.
		key = strings.TrimSpace(key)

		// Clean a value of starting whitespace and comments.
		value = strings.TrimSpace(value)
		value, _, _ = strings.Cut(value, " #")
		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}

func TestSetWithPrefix(t *testing.T) {
	type testCase struct {
		want   any
		assert func(*testing.T, testCase)
	}

	testCases := map[string]testCase{
		"success with prefix option": {
			want: SuccessWithPrefixOption{
				Duration: 10000000000,
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				os.Setenv("PREFIX_DURATION", "10s")

				var config SuccessWithPrefixOption

				if err := envconfig.Set(
					&config,
					envconfig.WithPrefix("PREFIX_"),
				); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn,
			func(t *testing.T) {
				t.Parallel()

				tc.assert(t, tc)
			},
		)
	}
}

// Slice test cases.

func TestSetSuccessWithSliceStringField(t *testing.T) {
	type Config struct {
		SliceStringField []string `env:"SLICE_STRING_FIELD"`
	}

	var config Config

	var want Config
	want.SliceStringField = []string{"first", "second", "third"}

	loadFileIntoEnvironmentVariables("./test_data/success_with_slice_string_field.env")

	envconfig.Set(&config)

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

	loadFileIntoEnvironmentVariables("./test_data/success_with_slice_int_field.env")

	envconfig.Set(&config)

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

	loadFileIntoEnvironmentVariables("./test_data/success_with_slice_float_field.env")

	envconfig.Set(&config)

	if !slices.Equal(config.SliceFloatField, want.SliceFloatField) {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

// Nested test cases.

func TestSetSuccessWithNestedStruct(t *testing.T) {
	type Config struct {
		Server struct {
			Port string `env:"PORT"`
		} `prefix:"SERVER_"`
	}

	var config Config

	var want Config
	want.Server.Port = "8080"

	loadFileIntoEnvironmentVariables("./test_data/success_with_nested_struct.env")

	envconfig.Set(&config)

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

	loadFileIntoEnvironmentVariables("./test_data/success_with_deeply_nested_struct.env")

	envconfig.Set(&config)

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

	loadFileIntoEnvironmentVariables("./test_data/success_with_thrice_deeply_nested_struct.env")

	envconfig.Set(&config)

	if config != want {
		t.Errorf("got %+v, want %+v", config, want)
	}
}

// JSON test cases.

func TestSetSuccessWithJsonField(t *testing.T) {
	type Config struct {
		JSONField struct {
			First string `json:"first"`
		} `envjson:"JSON_FIELD"`
	}

	var config Config

	var want Config
	want.JSONField.First = "example"
	loadFileIntoEnvironmentVariables("./test_data/success_with_json_field.env")

	envconfig.Set(&config)

	if config != want {
		t.Errorf("got %+v, want %+v", config, want)
	}
}
