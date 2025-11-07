package envconfig

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

const (
	envExtension = ".env"
)

type parser interface {
	// parse should populate a config struct.
	parse(config any) error
}

func (s settings) processFilename(config any) error {
	parser, err := identifyFileParser(s.filename)
	if err != nil {
		return fmt.Errorf("identify file parser: %w", err)
	}

	if err := parser.parse(config); err != nil {
		return fmt.Errorf("set config variables: %w", err)
	}
	return nil
}

func identifyFileParser(filename string) (parser, error) {
	var parser parser

	switch filepath.Ext(filename) {
	case envExtension:
		parser = envFileParser{
			config: map[string]string{},
			filename: filename,
		}
	default:
		return nil, &FileTypeValidationError{Filename: filename}
	}

	return parser, nil
}

type envFileParser struct {
	config map[string]string
	filename string
}

func (e envFileParser) parse(config any) error {
	file, err := os.Open(filepath.Clean(e.filename))
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

		entry, err := e.parseLine(line)
		if err != nil {
			return fmt.Errorf("parse line: %w", err)
		}

		if err = e.handleTextReplacement(&entry.value); err != nil {
			return fmt.Errorf("handle text replacement: %w", err)
		}

		e.config[entry.key] = entry.value
	}

	if err := e.populateFromFileConfig(config); err != nil {
		return fmt.Errorf("set config from file: %w", err)
	}

	if err := scanner.Err(); err != nil {
		return &FileReadError{Filename: e.filename, Err: err}
	}

	return nil
}

func (e envFileParser) populateFromFileConfig(config any) error {
	configStruct := reflect.ValueOf(config)
	if configStruct.Kind() != reflect.Pointer || configStruct.Elem().Kind() != reflect.Struct {
		return &InvalidConfigTypeError{ProvidedType: config}
	}

	configValue := reflect.ValueOf(config).Elem()

	for i := range configValue.NumField() {
		field := configValue.Type().Field(i)
		configFieldValue := configValue.Field(i)

		// Ignore fields that are not exported.
		if !configFieldValue.CanSet() {
			continue
		}

		jsonOptionValue, jsonOptionSet := field.Tag.Lookup(tagJSON)
		if jsonOptionSet {
			if err := e.handleJSONTag(configFieldValue, jsonOptionValue); err != nil {
				return fmt.Errorf("handle JSON tag: %w", err)
			}
			continue
		}

		if err := e.handlePrefixTag(field, configFieldValue, ""); err != nil {
			return fmt.Errorf("handle prefix tag: %w", err)
		}

		environmentVariableKey := field.Tag.Get(tagEnv)
		if environmentVariableKey == "" {
			continue
		}

		if err := setFieldValue(
			configFieldValue, entry{environmentVariableKey, e.config[environmentVariableKey]}); err != nil {
			return fmt.Errorf("set field value: %w", err)
		}
	}

	return nil
}

// populateNestedConfig populates a nested struct.
func (e envFileParser) populateNestedConfig(nestedConfig reflect.Value, prefix string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		if !configFieldValue.CanSet() || !configFieldValue.IsZero() {
			continue
		}

		jsonOptionValue, jsonOptionSet := field.Tag.Lookup(tagJSON)
		if jsonOptionSet {
			err := json.Unmarshal([]byte(e.config[jsonOptionValue]), &configFieldValue)
			if err != nil {
				return fmt.Errorf("handle JSON tag: %w", err)
			}

			continue
		}

		if err := e.handlePrefixTag(field, configFieldValue, prefix); err != nil {
			return fmt.Errorf("handle prefix tag: %w", err)
		}

		environmentVariableKey := prefix + field.Tag.Get(tagEnv)
		if environmentVariableKey == prefix { // Ensure tag is set.
			continue
		}
		if err := setFieldValue(
			configFieldValue, entry{environmentVariableKey, e.config[environmentVariableKey]}); err != nil {
			return fmt.Errorf("set field value: %w", err)
		}
	}

	return nil
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

// textReplacementRegex is used to detect text replacement in environment variables.
var textReplacementRegex = regexp.MustCompile(`\${[^}]+}`)

// handleTextReplacement will take a value and if it has `${[placeholder]}` as a substring, it will be replaced.
func (e envFileParser) handleTextReplacement(value *string) error {
	match := textReplacementRegex.FindStringSubmatch(*value)

	for _, m := range match {
		environmentValue := strings.TrimPrefix(m, "${")
		environmentValue = strings.TrimSuffix(environmentValue, "}")

		replacementValue := e.config[environmentValue]
		if replacementValue == "" {
			return &ReplacementError{VariableName: environmentValue}
		}

		*value = strings.ReplaceAll(*value, m, replacementValue)
	}

	return nil
}
