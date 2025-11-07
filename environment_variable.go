package envconfig

import (
	"os"
	"strings"
)

// processEnvironmentVariables populates the config struct using all environment variables.
func (s *settings) processEnvironmentVariables()  error { //nolint:gocognit // Complexity is reasonable.
	e := environmentVariableParser{
		prefix: s.prefix,
		source: map[string]string{},
	}

	source, err := e.parse()
	if err != nil {
		return err
	}

	s.source = source

	return nil
}

type environmentVariableParser struct {
	prefix string
	source map[string]string
}

func (e environmentVariableParser) parse() (map[string]string, error) {
	all := os.Environ()

	for _, val := range all {
		key, value, found := strings.Cut(val, "=")
		if !found {
			continue
		}

		e.source[key] = value
	}

	return e.source, nil
}
