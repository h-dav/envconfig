package envconfig_test

import (
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/h-dav/envconfig/v3"
)

func TestSet(t *testing.T) {
	t.Run("basic types", func(t *testing.T) {
		type Config struct {
			String  string        `env:"STRING"`
			Int     int           `env:"INT"`
			Float   float64       `env:"FLOAT"`
			Bool    bool          `env:"BOOL"`
			Duration time.Duration `env:"DURATION"`
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
			Default string `env:"NON_EXISTENT,default=fallback"`
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
			Required string `env:"MISSING,required"`
		}

		var cfg Config
		if err := envconfig.Set(&cfg); err == nil {
			t.Error("Set() expected error for missing required field, got nil")
		}
	})

	t.Run("text replacement", func(t *testing.T) {
		type Config struct {
			Host string `env:"HOST"`
			URL  string `env:"URL"`
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
			Strings []string  `env:"STRINGS"`
			Ints    []int     `env:"INTS"`
			Floats  []float64 `env:"FLOATS"`
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
				Port int `env:"PORT"`
			} `env:",prefix=SERVER_"`
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

	t.Run("envjson", func(t *testing.T) {
		type Config struct {
			Data struct {
				Key string `json:"key"`
			} `env:"DATA,envjson"`
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
			Value string `env:"VALUE"`
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
			} `env:"DATA,envjson"`
		}

		t.Setenv("DATA", `{"key": "value"`) // missing closing brace

		var cfg Config
		if err := envconfig.Set(&cfg); err == nil {
			t.Error("Set() expected error for malformed JSON, got nil")
		}
	})
}

func TestSetWithPrefix(t *testing.T) {
	type Config struct {
		Value string `env:"VALUE"`
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
