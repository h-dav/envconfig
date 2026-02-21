package envconfig_test

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/h-dav/envconfig/v3"
)

func TestWithLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	type Config struct {
		Name string `config:"NAME"`
	}

	os.Setenv("NAME", "test-user")
	defer os.Unsetenv("NAME")

	var cfg Config
	err := envconfig.Set(&cfg, envconfig.WithLogger(logger))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if cfg.Name != "test-user" {
		t.Errorf("expected NAME test-user, got %s", cfg.Name)
	}

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("expected log output, got empty")
	}

	// Simple check for presence of debug message
	expectedPart := "setting field value"
	if !bytes.Contains(buf.Bytes(), []byte(expectedPart)) {
		t.Errorf("log output does not contain expected message: %s", expectedPart)
	}
}
