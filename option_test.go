package envconfig_test

import (
	"testing"

	"github.com/h-dav/envconfig/v3"
)

// TestSetWithFilepath is test cases for simple use cases,
// such as flat config structures and fundamental fields, like required, and default.
func TestSetWithFilepath(t *testing.T) {
	type testCase struct {
		filepath string
		want     any
		assert   func(*testing.T, testCase)
	}

	testCases := map[string]testCase{
		"success with one field": {
			filepath: "./test_data/success_with_one_field.env",
			want: SuccessWithOneField{
				Example: "value1",
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithOneField

				if err := envconfig.Set(&config, envconfig.WithFilepath(tc.filepath)); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
		"success with one int field": {
			filepath: "./test_data/success_with_one_int_field.env",
			want: SuccessWithOneIntField{
				Example: 10,
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithOneIntField

				if err := envconfig.Set(&config, envconfig.WithFilepath(tc.filepath)); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
		"success with default value and empty env file": {
			filepath: "./test_data/success_with_one_default_value_and_empty_env_file.env",
			want: SuccessWithDefaultValueAndEmptyEnvFile{
				Example: "value2",
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithDefaultValueAndEmptyEnvFile

				if err := envconfig.Set(&config, envconfig.WithFilepath(tc.filepath)); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
		"success with text replacement": {
			filepath: "./test_data/success_with_text_replacement.env",
			want: SuccessWithTextReplacement{
				ReplaceField: "exampleField",
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithTextReplacement

				if err := envconfig.Set(&config, envconfig.WithFilepath(tc.filepath)); err != nil {
					t.Fail()
				}

				if config != tc.want {
					t.Errorf("got %+v, want %+v", config, tc.want)
				}
			},
		},
		"success with setting time.Duration": {
			filepath: "./test_data/success_with_setting_time_Duration.env",
			want: SuccessWithSettingTimeDuration{
				Duration: 10000000000,
			},
			assert: func(t *testing.T, tc testCase) {
				t.Helper()

				var config SuccessWithSettingTimeDuration

				if err := envconfig.Set(&config, envconfig.WithFilepath(tc.filepath)); err != nil {
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
