package envconfig_test

import (
	"os"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/h-dav/envconfig/v3"
)

func TestSet(t *testing.T) {
	t.Run("basic types", func(t *testing.T) {
		type Config struct {
			String  string        `config:"STRING"`
			Int     int           `config:"INT"`
			Float   float64       `config:"FLOAT"`
			Bool    bool          `config:"BOOL"`
			Duration time.Duration `config:"DURATION"`
		}

		t.Setenv("STRING", "value")
		t.Setenv("INT", "10")
		t.Setenv("FLOAT", "1.2")
		t.Setenv("BOOL", "true")
		t.Setenv("DURATION", "10s")

		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		want := Config{
			String:   "value",
			Int:      10,
			Float:    1.2,
			Bool:     true,
			Duration: 10 * time.Second,
		}

		if diff := cmp.Diff(want, cfg); diff != "" {
			t.Errorf("Set() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("default values", func(t *testing.T) {
		type Config struct {
			Default string `config:"NON_EXISTENT,default=fallback"`
		}

		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if cfg.Default != "fallback" {
			t.Errorf("got %q, want %q", cfg.Default, "fallback")
		}
	})

	t.Run("required fields", func(t *testing.T) {
		type Config struct {
			Required string `config:"MISSING,required"`
		}

		var cfg Config
		if err := envconfig.Set(&cfg); err == nil {
			t.Error("Set() expected error for missing required field, got nil")
		}
	})

	t.Run("text replacement", func(t *testing.T) {
		type Config struct {
			Host string `config:"HOST"`
			URL  string `config:"URL"`
		}

		t.Setenv("HOST", "localhost")
		t.Setenv("URL", "http://${HOST}:8080")

		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if cfg.URL != "http://localhost:8080" {
			t.Errorf("got %q, want %q", cfg.URL, "http://localhost:8080")
		}
	})

	t.Run("slices", func(t *testing.T) {
		type Config struct {
			Strings []string  `config:"STRINGS"`
			Ints    []int     `config:"INTS"`
			Floats  []float64 `config:"FLOATS"`
		}

		t.Setenv("STRINGS", "a,b,c")
		t.Setenv("INTS", "1,2,3")
		t.Setenv("FLOATS", "1.1,2.2")

		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if !slices.Equal(cfg.Strings, []string{"a", "b", "c"}) {
			t.Errorf("Strings: got %v", cfg.Strings)
		}
		if !slices.Equal(cfg.Ints, []int{1, 2, 3}) {
			t.Errorf("Ints: got %v", cfg.Ints)
		}
		if !slices.Equal(cfg.Floats, []float64{1.1, 2.2}) {
			t.Errorf("Floats: got %v", cfg.Floats)
		}
	})

	t.Run("nested structs", func(t *testing.T) {
		type Config struct {
			Server struct {
				Port int `config:"PORT"`
			} `config:",prefix=SERVER_"`
		}

		t.Setenv("SERVER_PORT", "8080")

		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if cfg.Server.Port != 8080 {
			t.Errorf("got %d, want 8080", cfg.Server.Port)
		}
	})

	t.Run("deeply nested structs", func(t *testing.T) {
		type Config struct {
			Server struct {
				Database struct {
					User string `config:"USER"`
				} `config:",prefix=DB_"`
			} `config:",prefix=SERVER_"`
		}

		t.Setenv("SERVER_DB_USER", "admin")

		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if cfg.Server.Database.User != "admin" {
			t.Errorf("got %q, want %q", cfg.Server.Database.User, "admin")
		}
	})

	t.Run("complex slices", func(t *testing.T) {
		type Config struct {
			Empty []string `config:"EMPTY"`
			Space []string `config:"SPACE"`
		}

		t.Setenv("EMPTY", "")
		t.Setenv("SPACE", "a, ,b")

		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if len(cfg.Empty) != 1 || cfg.Empty[0] != "" {
			t.Errorf("Empty: got %v", cfg.Empty)
		}
		if len(cfg.Space) != 3 || cfg.Space[1] != "" {
			t.Errorf("Space: got %v", cfg.Space)
		}
	})

	t.Run("json", func(t *testing.T) {
		type Config struct {
			Data struct {
				Key string `json:"key"`
			} `config:"DATA,json"`
		}

		t.Setenv("DATA", `{"key": "value"}`)

		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if cfg.Data.Key != "value" {
			t.Errorf("got %q, want %q", cfg.Data.Key, "value")
		}
	})

	t.Run("invalid config type - not a pointer", func(t *testing.T) {
		type Config struct {
			Value string `config:"VALUE"`
		}
		var cfg Config
		if err := envconfig.Set(cfg); err == nil {
			t.Error("Set() expected error for non-pointer, got nil")
		}
	})

	t.Run("invalid config type - not a struct", func(t *testing.T) {
		var val string
		if err := envconfig.Set(&val); err == nil {
			t.Error("Set() expected error for non-struct pointer, got nil")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		type Config struct {
			Data struct {
				Key string `json:"key"`
			} `config:"DATA,json"`
		}

		t.Setenv("DATA", `{"key": "value"`) // missing closing brace

		var cfg Config
		if err := envconfig.Set(&cfg); err == nil {
			t.Error("Set() expected error for malformed JSON, got nil")
		}
	})
}

func TestPrecedence(t *testing.T) {
	type Config struct {
		Value string `config:"VALUE,default=default_val"`
	}

	t.Run("Default only", func(t *testing.T) {
		var cfg Config
		if err := envconfig.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Value != "default_val" {
			t.Errorf("got %q, want %q", cfg.Value, "default_val")
		}
	})

	t.Run("File overwrites Default", func(t *testing.T) {
		// Create temporary .env file
		tmpFile := "precedence_test.env"
		content := "VALUE=file_val\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
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

	t.Run("Env overwrites File", func(t *testing.T) {
		tmpFile := "precedence_test.env"
		content := "VALUE=file_val\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		t.Setenv("VALUE", "env_val")

		var cfg Config
		if err := envconfig.Set(&cfg, envconfig.WithFilepath(tmpFile)); err != nil {
			t.Fatal(err)
		}
		if cfg.Value != "env_val" {
			t.Errorf("got %q, want %q", cfg.Value, "env_val")
		}
	})
}

func TestSetWithPrefix(t *testing.T) {
	type Config struct {
		Value string `config:"VALUE"`
	}

	t.Setenv("APP_VALUE", "hello")

	var cfg Config
	if err := envconfig.Set(&cfg, envconfig.WithPrefix("APP_")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if cfg.Value != "hello" {
		t.Errorf("got %q, want %q", cfg.Value, "hello")
	}
}
