// Package envconfig provides functionality to easily load config into your struct.
package envconfig

import (
	"log/slog"
	"maps"
)

// Set parses values from multiple sources (Environment, Flags, and optional Files) 
// and populates the provided config struct.
func Set(config any, opts ...Option) error {
	s := &settings{
		source:   make(map[string]string),
		decoders: defaultDecoders,
	}

	for _, opt := range opts {
		opt(s)
	}

	// Default sources: Environment variables and Flags.
	s.sources = append(s.sources, EnvironmentVariableSource{}, FlagSource{})

	for _, src := range s.sources {
		if s.logger != nil {
			s.logger.Debug("loading values from source", slog.String("source", src.SourceType()))
		}
		values, err := src.Load()
		if err != nil {
			return &FieldError{Op: "load from source", Err: err}
		}

		maps.Copy(s.source, values)
	}

	return s.populateStruct(config)
}
