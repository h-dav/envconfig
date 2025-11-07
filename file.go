package envconfig

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	envExtension = ".env"
)

type parser interface {
	// parse should populate a config struct.
	parse() (map[string]string, error)
}

func (s *settings) processFilename() error {
	parser, err := identifyFileParser(s.filename)
	if err != nil {
		return fmt.Errorf("identify file parser: %w", err)
	}

	source, err := parser.parse()
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	s.source = source

	return nil
}

func identifyFileParser(filename string) (parser, error) {
	var parser parser

	switch filepath.Ext(filename) {
	case envExtension:
		parser = envFileParser{
			source:   map[string]string{},
			filename: filename,
		}
	default:
		return nil, &FileTypeValidationError{Filename: filename}
	}

	return parser, nil
}

type envFileParser struct {
	source   map[string]string
	filename string
}

func (e envFileParser) parse() (map[string]string, error) {
	file, err := os.Open(filepath.Clean(e.filename))
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
		return make(map[string]string), &FileReadError{Filename: e.filename, Err: err}
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
