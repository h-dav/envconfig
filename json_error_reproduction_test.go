package envconfig_test

import (
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
	os.Setenv("DATA", malformedJSON)
	defer os.Unsetenv("DATA")

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
