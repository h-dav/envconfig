package envconfig

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	envExtension   = ".env"
	tomlExtensions = ".toml"
)

type parser interface {
	// parse should ingest a file and set the values as enrivonment variables.
	parse(filename string) error
}

func process(filename string) error {
	parser, err := identifyParser(filename)
	if err != nil {
		return fmt.Errorf("identify parser: %w", err)
	}

	if err := parser.parse(filename); err != nil {
		return fmt.Errorf("set environment variables: %w", err)
	}
	return nil
}

func identifyParser(filename string) (parser, error) {
	var parser parser

	switch filepath.Ext(filename) {
	case envExtension:
		parser = envFileParser{}
	case tomlExtensions:
		parser = tomlFileParser{}
	default:
		return nil, &FileTypeValidationError{Filename: filename}
	}

	return parser, nil

}

type envFileParser struct{}

func (e envFileParser) parse(filename string) error {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return &OpenFileError{Err: err}
	}
	defer file.Close() //nolint:errcheck // File closure.

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Handles empty and commented lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		entry, err := e.parseEnvLine(line)
		if err != nil {
			return fmt.Errorf("parse environment variable line: %w", err)
		}

		if err = handleTextReplacement(&entry.value); err != nil {
			return fmt.Errorf("handle text replacement: %w", err)
		}

		if err := os.Setenv(entry.key, entry.value); err != nil {
			return &SetEnvironmentVariableError{Err: err}
		}
	}

	if err := scanner.Err(); err != nil {
		return &FileReadError{Filename: filename, Err: err}
	}

	return nil
}

// parseEnvLine parses an individual .env line, and will detect comments.
func (e envFileParser) parseEnvLine(line string) (entry, error) {
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

type tomlFileParser struct{}

func (t tomlFileParser) parse(filename string) error {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return &OpenFileError{Err: err}
	}
	defer file.Close() //nolint:errcheck // File closure.

	return nil
}

// textReplacementRegex is used to detect text replacement in environment variables.
var textReplacementRegex = regexp.MustCompile(`\${[^}]+}`)

// handleTextReplacement will take a value and if it has `${[placeholder]}` as a substring, it will be replaced.
func handleTextReplacement(value *string) error {
	match := textReplacementRegex.FindStringSubmatch(*value)

	for _, m := range match {
		environmentValue := strings.TrimPrefix(m, "${")
		environmentValue = strings.TrimSuffix(environmentValue, "}")

		replacementValue := os.Getenv(environmentValue)
		if replacementValue == "" {
			return &ReplacementError{VariableName: environmentValue}
		}

		*value = strings.ReplaceAll(*value, m, replacementValue)
	}

	return nil
}
