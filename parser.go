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
	// parse should ingest a file and set the values in config.
	parse(config any, filename string) error
}

func process(config any, filename string) error {
	parser, err := identifyParser(filename)
	if err != nil {
		return fmt.Errorf("identify parser: %w", err)
	}

	if err := parser.parse(config, filename); err != nil {
		return fmt.Errorf("set environment variables: %w", err)
	}
	return nil
}

func identifyParser(filename string) (parser, error) {
	var parser parser

	switch filepath.Ext(filename) {
	case envExtension:
		parser = envFileParser{
			config: map[string]string{},
		}
	default:
		return nil, &FileTypeValidationError{Filename: filename}
	}

	return parser, nil
}

type envFileParser struct {
	config map[string]string
}

func (e envFileParser) parse(config any, filename string) error {
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

		e.config[entry.key] = entry.value
	}

	if err := e.setFromFileConfig(config); err != nil {
		return fmt.Errorf("set config from file: %w", err)
	}

	if err := scanner.Err(); err != nil {
		return &FileReadError{Filename: filename, Err: err}
	}

	return nil
}

func (e envFileParser) setFromFileConfig(config any) error {
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
			err := json.Unmarshal([]byte(e.config[jsonOptionValue]), &configFieldValue)
			if err != nil {
				return fmt.Errorf("handle JSON option: %w", err)
			}

			continue
		}

		if err := e.handlePrefixTag(field, configFieldValue, ""); err != nil {
			return fmt.Errorf("handle prefix option: %w", err)
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

func (e envFileParser) handlePrefixTag(
	field reflect.StructField,
	configFieldValue reflect.Value,
	prefix string,
) error {
	if field.Type.Kind() != reflect.Struct {
		return nil
	}

	prefixOptionValue, prefixOptionSet := field.Tag.Lookup(tagPrefix)
	if !prefixOptionSet {
		return &PrefixOptionError{FieldName: field.Name}
	}

	if err := e.populateNestedConfig(configFieldValue, prefix+prefixOptionValue); err != nil {
		return fmt.Errorf("populate nested config struct: %w", err)
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
				return fmt.Errorf("handle JSON option: %w", err)
			}

			continue
		}

		if err := e.handlePrefixTag(field, configFieldValue, prefix); err != nil {
			return fmt.Errorf("handle prefix option: %w", err)
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
