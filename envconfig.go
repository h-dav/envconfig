// Package envconfig provides functionality to easily load config into your struct.
package envconfig

import (
	"fmt"
)

type entry struct {
	key, value string
}

// Set will parse the .env file and set the values in the environment, then populate the passed in struct
// using all environment variables.
func Set(config any, opts ...option) error {
	s := &settings{}

	for _, opt := range opts {
		opt(s)
	}

	if s.filename != "" {
		if err := s.processFilename(config); err != nil {
			return fmt.Errorf("parse file: %w", err)
		}
	}

	if err := s.processEnvironmentVariables(config); err != nil {
		return fmt.Errorf("populate config struct: %w", err)
	}

	return nil
}
