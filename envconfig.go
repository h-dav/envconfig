// Package envconfig provides functionality to easily load config into your struct.
package envconfig

import (
	"maps"
)

// Set will parse multiple sources for config values, and use these values to populate the passed in config struct.
func Set(config any, opts ...Option) error {
	s := &settings{
		source:   map[string]string{},
		decoders: defaultDecoders,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.sources = append(s.sources, EnvironmentVariableSource{}, FlagSource{})

	for _, source := range s.sources {
		values, err := source.Load()
		if err != nil {
			return &FieldError{Op: "load from source", Err: err}
		}

		maps.Copy(s.source, values)
	}

	if err := s.populateStruct(config); err != nil {
		return err
	}

	return nil
}
