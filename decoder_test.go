package envconfig_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/h-dav/envconfig/v3"
)

func TestJSONErrorContext(t *testing.T) {
	type Config struct {
		Data struct {
			Key string `json:"key"`
		} `env:"DATA,envjson"`
	}

	malformedJSON := `{"key": "value"` // Missing closing brace
	if err := os.Setenv("DATA", malformedJSON); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("DATA"); err != nil {
			t.Errorf("failed to unset env: %v", err)
		}
	}()

	var cfg Config
	err := envconfig.Set(&cfg)

	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	expectedPart1 := "failed to unmarshal JSON for field Data"
	if !strings.Contains(err.Error(), expectedPart1) {
		t.Errorf("error message does not contain field name. got: %v, want to contain: %s", err, expectedPart1)
	}

	// This is the part we wanted to ADD: the raw input string
	expectedValuePart := fmt.Sprintf("%q", malformedJSON)
	if !strings.Contains(err.Error(), expectedValuePart) {
		t.Errorf("error message does not contain raw input string. got: %v, want to contain: %s", err, expectedValuePart)
	}
}

func TestJSONErrorComprehensive(t *testing.T) {
	type Nested struct {
		Age int `json:"age"`
	}

	type Config struct {
		Profile struct {
			Name string `json:"name"`
		} `env:"PROFILE,envjson"`
		Settings Nested `env:"SETTINGS,envjson"`
		Valid    struct {
			OK bool `json:"ok"`
		} `env:"VALID,envjson"`
	}

	tests := []struct {
		name          string
		env           map[string]string
		expectedErr   string
		expectedField string
		expectedValue string
		verifyUnwrap  bool
	}{
		{
			name: "syntax error",
			env: map[string]string{
				"PROFILE": `{"name": "Alice"`,
			},
			expectedErr:   "failed to unmarshal JSON for field Profile with value \"{\\\"name\\\": \\\"Alice\\\"\": unexpected end of JSON input",
			expectedField: "Profile",
			expectedValue: `{"name": "Alice"`,
			verifyUnwrap:  true,
		},
		{
			name: "type mismatch",
			env: map[string]string{
				"SETTINGS": `{"age": "not-an-int"}`,
			},
			expectedErr:   "failed to unmarshal JSON for field Settings with value \"{\\\"age\\\": \\\"not-an-int\\\"}\": json: cannot unmarshal string into Go struct field Nested.age of type int",
			expectedField: "Settings",
			expectedValue: `{"age": "not-an-int"}`,
			verifyUnwrap:  true,
		},
		{
			name: "valid json remains valid",
			env: map[string]string{
				"VALID": `{"ok": true}`,
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.env {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf("failed to set env %s: %v", k, err)
				}
			}

			var cfg Config
			err := envconfig.Set(&cfg)

			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// We use strings.Contains because the error might be wrapped further by envconfig.Set
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("error string mismatch\ngot:  %v\nwant to contain: %v", err.Error(), tt.expectedErr)
			}

			if tt.verifyUnwrap {
				var jsonErr *envconfig.JSONUnmarshalError
				if !errors.As(err, &jsonErr) {
					t.Errorf("expected error to be unwrapable as *envconfig.JSONUnmarshalError, got %T: %v", err, err)
				} else {
					if jsonErr.FieldName != tt.expectedField {
						t.Errorf("JSONUnmarshalError.FieldName = %v, want %v", jsonErr.FieldName, tt.expectedField)
					}
					if jsonErr.RawValue != tt.expectedValue {
						t.Errorf("JSONUnmarshalError.RawValue = %v, want %v", jsonErr.RawValue, tt.expectedValue)
					}
				}
			}
		})
	}
}
