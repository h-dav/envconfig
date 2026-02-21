package envconfig_test

import (
	"flag"
	"os"
	"testing"

	"github.com/h-dav/envconfig/v3"
)

func TestPrecedenceFull(t *testing.T) {
	// Define a flag for this test. 
	// envconfig picks up flags by their name.
	if flag.Lookup("INTEGRATION_FLAG") == nil {
		flag.String("INTEGRATION_FLAG", "", "test flag")
	}

	type Config struct {
		Value string `config:"INTEGRATION_FLAG,default=default"`
	}

	// 1. Default vs File
	t.Run("File wins over Default", func(t *testing.T) {
		tmpFile := "test1.env"
		if err := os.WriteFile(tmpFile, []byte("INTEGRATION_FLAG=file_val\n"), 0644); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		var cfg Config
		if err := envconfig.Set(&cfg, envconfig.WithFilepath(tmpFile)); err != nil {
			t.Fatal(err)
		}
		if cfg.Value != "file_val" {
			t.Errorf("got %q, want %q", cfg.Value, "file_val")
		}
	})

	// 2. File vs Env
	t.Run("Env wins over File", func(t *testing.T) {
		tmpFile := "test2.env"
		if err := os.WriteFile(tmpFile, []byte("INTEGRATION_FLAG=file_val\n"), 0644); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		t.Setenv("INTEGRATION_FLAG", "env_val")

		var cfg Config
		if err := envconfig.Set(&cfg, envconfig.WithFilepath(tmpFile)); err != nil {
			t.Fatal(err)
		}
		if cfg.Value != "env_val" {
			t.Errorf("got %q, want %q", cfg.Value, "env_val")
		}
	})

	// 3. Env vs Flag
	t.Run("Flag wins over Env", func(t *testing.T) {
		t.Setenv("INTEGRATION_FLAG", "env_val")
		
		// Manually set the flag value.
		if err := flag.Set("INTEGRATION_FLAG", "flag_val"); err != nil {
			t.Fatal(err)
		}

		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Value != "flag_val" {
			t.Errorf("got %q, want %q", cfg.Value, "flag_val")
		}
	})

	t.Run("Complex type precedence - Slices", func(t *testing.T) {
		type Config struct {
			Values []string `config:"SLICE"`
		}

		tmpFile := "test_slice.env"
		if err := os.WriteFile(tmpFile, []byte("SLICE=a,b,c\n"), 0644); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		t.Setenv("SLICE", "env1,env2")

		var cfg Config
		if err := envconfig.Set(&cfg, envconfig.WithFilepath(tmpFile)); err != nil {
			t.Fatal(err)
		}

		// Env should win completely, not merge.
		if len(cfg.Values) != 2 || cfg.Values[0] != "env1" || cfg.Values[1] != "env2" {
			t.Errorf("got %v, want [env1 env2]", cfg.Values)
		}
	})

	t.Run("Complex type precedence - Nested Structs", func(t *testing.T) {
		type Config struct {
			Server struct {
				Port int `config:"PORT"`
			} `config:",prefix=SERVER_"`
		}

		tmpFile := "test_nested.env"
		if err := os.WriteFile(tmpFile, []byte("SERVER_PORT=8080\n"), 0644); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		t.Setenv("SERVER_PORT", "9090")

		var cfg Config
		if err := envconfig.Set(&cfg, envconfig.WithFilepath(tmpFile)); err != nil {
			t.Fatal(err)
		}

		if cfg.Server.Port != 9090 {
			t.Errorf("got %d, want 9090", cfg.Server.Port)
		}
	})
}
