package envconfig

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Source is the interface that wraps the Load method.
//
// Load returns a map of key-value pairs representing the configuration.
type Source interface {
	Load() (map[string]string, error)
}

// FlagSource loads configuration from command-line flags.
type FlagSource struct{}

// Load parses command-line flags and returns them as a map.
// It will parse flags if they haven't been parsed yet.
func (s FlagSource) Load() (map[string]string, error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	source := make(map[string]string)

	flag.Visit(func(f *flag.Flag) {
		source[f.Name] = f.Value.String()
	})

	return source, nil
}

// EnvironmentVariableSource loads configuration from environment variables.
type EnvironmentVariableSource struct {}

// Load loads all environment variables.
func (s EnvironmentVariableSource) Load() (map[string]string, error) {
	source := make(map[string]string)
	all := os.Environ()

	for _, val := range all {
		key, value, found := strings.Cut(val, "=")
		if !found {
			continue
		}

		source[key] = value
	}

	return source, nil
}

// FileSource loads configuration from a file.
type FileSource struct {
	Filepath string
}

// Load loads the file and parses it based on its extension.
func (s FileSource) Load() (map[string]string, error) {
	parser, err := identifyFileParser(s.Filepath)
	if err != nil {
		return nil, fmt.Errorf("identify file parser: %w", err)
	}

	source, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	return source, nil
}

type parser interface {
	Parse() (map[string]string, error)
}

func identifyFileParser(f string) (parser, error) {
	switch filepath.Ext(f) {
	case ".env":
		return envFileParser{filepath: f}, nil
	default:
		return nil, &FileTypeValidationError{Filepath: f}
	}
}

type envFileParser struct {
	filepath string
}

func (e envFileParser) Parse() (map[string]string, error) {
	file, err := os.Open(filepath.Clean(e.filepath))
	if err != nil {
		return nil, &OpenFileError{Err: err}
	}
	defer file.Close()

	source := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Handles empty and commented lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, err := parseEnvLine(line)
		if err != nil {
			return nil, fmt.Errorf("parse line: %w", err)
		}

		source[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, &FileReadError{Filepath: e.filepath, Err: err}
	}

	return source, nil
}

func parseEnvLine(line string) (key, value string, err error) {
	k, v, found := strings.Cut(line, "=")
	if !found {
		return "", "", &ParseError{Line: line}
	}

	k = strings.TrimSpace(k)
	v = strings.TrimSpace(v)
	v, _, _ = strings.Cut(v, " #")

	return k, v, nil
}
