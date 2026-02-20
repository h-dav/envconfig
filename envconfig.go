// Package envconfig provides functionality to easily load config into your struct.
package envconfig

import (
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
	// We add them after options so they can be overridden if needed, 
	// or they can be the base. Actually, the precedence is determined by the order in s.sources.
	s.sources = append(s.sources, EnvironmentVariableSource{}, FlagSource{})

	for _, src := range s.sources {
		values, err := src.Load()
		if err != nil {
			return &FieldError{Op: "load from source", Err: err}
		}

		maps.Copy(s.source, values)
	}

	return s.populateStruct(config)
}
