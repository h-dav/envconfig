package envconfig

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type source interface {
	Load() (map[string]string, error)
}

type FlagSource struct{}

func (s FlagSource) Load() (map[string]string, error) {
	flag.Parse()

	source := make(map[string]string)

	flag.Visit(func(f *flag.Flag) {
		source[f.Name] = f.Value.String()
	})

	return source, nil
}

const (
	envExtension = ".env"
)

type parser interface {
	parse() (map[string]string, error)
}

type FileSource struct {
	filepath string
}

func (s FileSource) Load() (map[string]string, error) {
	parser, err := identifyFileParser(s.filepath)
	if err != nil {
		return nil, fmt.Errorf("identify file parser: %w", err)
	}

	source, err := parser.parse()
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}


	return source, nil
}

// identifyFileParser determines the parser to use based on the filepath received.
func identifyFileParser(f string) (parser, error) {
	var parser parser

	switch filepath.Ext(f) {
	case envExtension:
		parser = envFileParser{
			source:   map[string]string{},
			filepath: f,
		}
	default:
		return nil, &FileTypeValidationError{Filepath: f}
	}

	return parser, nil
}

type envFileParser struct {
	source   map[string]string
	filepath string
}

func (e envFileParser) parse() (map[string]string, error) {
	file, err := os.Open(filepath.Clean(e.filepath))
	if err != nil {
		return make(map[string]string), &OpenFileError{Err: err}
	}
	defer file.Close() //nolint:errcheck // File closure.

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Handles empty and commented lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		entry, err := e.parseLine(line)
		if err != nil {
			return make(map[string]string), fmt.Errorf("parse line: %w", err)
		}

		e.source[entry.key] = entry.value
	}

	if err := scanner.Err(); err != nil {
		return make(map[string]string), &FileReadError{Filepath: e.filepath, Err: err}
	}

	return e.source, nil
}

// parseLine parses an individual .env line, and will detect comments.
func (e envFileParser) parseLine(line string) (entry, error) {
	key, value, found := strings.Cut(line, "=")
	if !found {
		return entry{}, &ParseError{Line: line}
	}

	// Clean environment variable key.
	key = strings.TrimSpace(key)

	// Clean a value of starting whitespace and comments.
	value = strings.TrimSpace(value)
	value, _, _ = strings.Cut(value, " #")

	return entry{key: key, value: value}, nil
}

type EnvironmentVariableSource struct {
	prefix string
}

// processEnvironmentVariables populates the config struct using all environment variables.
func (s EnvironmentVariableSource) Load() (map[string]string, error) { //nolint:gocognit // Complexity is reasonable.
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
