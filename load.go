package envconfig

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"
	"strings"
)

// source represents a provider of configuration key-value pairs.
type source interface {
	Load() (map[string]string, error)
	SourceType() string
}

// FlagSource loads configuration from command-line flags.
type FlagSource struct{}

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

func (s FlagSource) SourceType() string { return "Flag" }

const (
	envExtension = ".env"
)

// parser represents a file format parser for configuration.
type parser interface {
	parse() (map[string]string, error)
}

// FileSource loads configuration from a specific file.
type FileSource struct {
	filepath string
}

func (s FileSource) Load() (map[string]string, error) {
	p, err := identifyFileParser(s.filepath)
	if err != nil {
		return nil, err
	}

	return p.parse()
}

func (s FileSource) SourceType() string { return "File (" + s.filepath + ")" }

// identifyFileParser determines the parser to use based on the filepath extension.
func identifyFileParser(f string) (parser, error) {
	switch filepath.Ext(f) {
	case envExtension:
		return envFileParser{
			source:   make(map[string]string),
			filepath: f,
		}, nil
	default:
		return nil, &FileTypeValidationError{Filepath: f}
	}
}

type envFileParser struct {
	source   map[string]string
	filepath string
}

func (e envFileParser) parse() (map[string]string, error) {
	file, err := os.Open(filepath.Clean(e.filepath))
	if err != nil {
		return nil, &OpenFileError{Err: err}
	}
	defer file.Close() //nolint:errcheck

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Handles empty and commented lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		entry, err := e.parseLine(line)
		if err != nil {
			return nil, err
		}

		e.source[entry.key] = entry.value
	}

	if err := scanner.Err(); err != nil {
		return nil, &FileReadError{Filepath: e.filepath, Err: err}
	}

	return e.source, nil
}

// parseLine parses an individual .env line.
func (e envFileParser) parseLine(line string) (entry, error) {
	key, value, found := strings.Cut(line, "=")
	if !found {
		return entry{}, &ParseError{Line: line, Err: ErrSyntax}
	}

	return entry{
		key:   strings.TrimSpace(key),
		value: strings.TrimSpace(strings.Split(value, " #")[0]),
	}, nil
}

// EnvironmentVariableSource loads configuration from environment variables.
type EnvironmentVariableSource struct{}

func (s EnvironmentVariableSource) Load() (map[string]string, error) {
	source := make(map[string]string)
	for _, val := range os.Environ() {
		key, value, found := strings.Cut(val, "=")
		if found {
			source[key] = value
		}
	}

	return source, nil
}

func (s EnvironmentVariableSource) SourceType() string { return "Environment" }
