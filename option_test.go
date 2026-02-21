package envconfig_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h-dav/envconfig/v3"
)

func TestSetWithFilepath(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testCases := []struct {
			name     string
			filepath string
			want     any
			config   any
		}{
			{
				name:     "one field",
				filepath: "./test_data/success_with_one_field.env",
				config:   &SuccessWithOneField{},
				want:     &SuccessWithOneField{Example: "value1"},
			},
			{
				name:     "one int field",
				filepath: "./test_data/success_with_one_int_field.env",
				config:   &SuccessWithOneIntField{},
				want:     &SuccessWithOneIntField{Example: 10},
			},
			{
				name:     "default value",
				filepath: "./test_data/success_with_one_default_value_and_empty_env_file.env",
				config:   &SuccessWithDefaultValueAndEmptyEnvFile{},
				want:     &SuccessWithDefaultValueAndEmptyEnvFile{Example: "value2"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if err := envconfig.Set(tc.config, envconfig.WithFilepath(tc.filepath)); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
				if diff := cmp.Diff(tc.want, tc.config); diff != "" {
					t.Errorf("Set() mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("file errors", func(t *testing.T) {
		testCases := []struct {
			name     string
			filepath string
		}{
			{
				name:     "missing file",
				filepath: "./test_data/non_existent.env",
			},
			{
				name:     "invalid extension",
				filepath: "./test_data/invalid.txt",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var cfg SuccessWithOneField
				if err := envconfig.Set(&cfg, envconfig.WithFilepath(tc.filepath)); err == nil {
					t.Error("Set() expected error for invalid file, got nil")
				}
			})
		}
	})

	t.Run("with prefix", func(t *testing.T) {
		type Config struct {
			Value string `config:"VALUE"`
		}
		// Create a temporary .env file
		filepath := "./test_data/with_prefix.env"
		content := "APP_VALUE=hello\n"
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		defer func() { _ = os.Remove(filepath) }()

		var cfg Config
		if err := envconfig.Set(&cfg, envconfig.WithFilepath(filepath), envconfig.WithPrefix("APP_")); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if cfg.Value != "hello" {
			t.Errorf("got %q, want %q", cfg.Value, "hello")
		}
	})
}

func TestWithActiveProfile(t *testing.T) {
	type Config struct {
		Value string `config:"VALUE"`
	}

	// Create profile files
	if err := os.WriteFile("./test_data/dev.env", []byte("VALUE=dev_val\n"), 0644); err != nil {
		t.Fatalf("failed to create dev.env: %v", err)
	}
	defer func() { _ = os.Remove("./test_data/dev.env") }()

	if err := os.WriteFile("./test_data/default.env", []byte("VALUE=default_val\n"), 0644); err != nil {
		t.Fatalf("failed to create default.env: %v", err)
	}
	defer func() { _ = os.Remove("./test_data/default.env") }()

	t.Run("specific profile", func(t *testing.T) {
		var cfg Config
		if err := envconfig.Set(&cfg, envconfig.WithActiveProfile("./test_data/", "dev")); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		if cfg.Value != "dev_val" {
			t.Errorf("got %q, want %q", cfg.Value, "dev_val")
		}
	})

	t.Run("default fallback", func(t *testing.T) {
		var cfg Config
		if err := envconfig.Set(&cfg, envconfig.WithActiveProfile("./test_data/", "")); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		if cfg.Value != "default_val" {
			t.Errorf("got %q, want %q", cfg.Value, "default_val")
		}
	})
}
