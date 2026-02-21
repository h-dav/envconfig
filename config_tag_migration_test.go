package envconfig_test

import (
	"os"
	"testing"

	"github.com/h-dav/envconfig/v3"
)

func TestConfigTagMigration(t *testing.T) {
	type Config struct {
		Value string `config:"MIGRATION_TEST"`
	}

	os.Setenv("MIGRATION_TEST", "success")
	defer os.Unsetenv("MIGRATION_TEST")

	var cfg Config
	err := envconfig.Set(&cfg)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if cfg.Value != "success" {
		t.Errorf("expected Value 'success', got '%s'", cfg.Value)
	}
}
