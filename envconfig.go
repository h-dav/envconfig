// Package envconfig provides functionality to easily load config into your struct.
package envconfig

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

const envExtension = ".env"

// textReplacementRegex is used to detect text replacement in environment variables.
var textReplacementRegex = regexp.MustCompile(`\${[^}]+}`)

// Set parses multiple sources for config values and populates the passed config struct.
func Set(config any, opts ...option) error {
	s := &settings{
		source:   map[string]string{},
		decoders: defaultDecoders,
	}

	for _, opt := range opts {
		opt(s)
	}

	// Process filepaths and create FileSources
	for _, f := range s.filepaths {
		path := f
		if s.activeProfile != "" {
			dir, _ := filepath.Split(path)
			path = filepath.Join(dir, s.activeProfile+envExtension)
		}
		s.sources = append(s.sources, FileSource{Filepath: path})
	}

	if s.activeProfile != "" && len(s.filepaths) == 0 {
		return fmt.Errorf("assign active profile: %w", &IncompatibleOptionsError{
			FirstOption:  "WithActiveProfile()",
			SecondOption: "WithFilepath()",
			Reason:       "directory in filepath option must be provided when using active profile",
		})
	}

	// Add EnvironmentVariableSource and FlagSource at the end
	s.sources = append(s.sources, EnvironmentVariableSource{}, FlagSource{})

	// Load all sources
	for _, source := range s.sources {
		values, err := source.Load()
		if err != nil {
			return fmt.Errorf("load from source: %w", err)
		}

		for key, value := range values {
			s.source[key] = value
		}
	}

	if err := s.populateStruct(config); err != nil {
		return fmt.Errorf("populate config struct: %w", err)
	}

	return nil
}

// populateStruct populates the config struct using the loaded values.
func (s *settings) populateStruct(config any) error {
	configValue := reflect.ValueOf(config)
	if configValue.Kind() != reflect.Pointer || configValue.Elem().Kind() != reflect.Struct {
		return &InvalidConfigTypeError{ProvidedType: config}
	}

	// Pass s.prefix to the recursive function so top-level fields respect the prefix option.
	return s.populateStructRecursive(configValue.Elem(), s.prefix)
}

func (s *settings) populateStructRecursive(structValue reflect.Value, prefix string) error {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Handle JSON tag
		if jsonKey, ok := field.Tag.Lookup(tagJSON); ok {
			if val, exists := s.source[jsonKey]; exists {
				if err := json.Unmarshal([]byte(val), fieldValue.Addr().Interface()); err != nil {
					return fmt.Errorf("unmarshal JSON for field %s: %w", field.Name, err)
				}
				continue
			}
		}

		// Handle Prefix tag (Nested Structs)
		if prefixTag, ok := field.Tag.Lookup(tagPrefix); ok {
			if fieldValue.Kind() == reflect.Struct {
				if err := s.populateStructRecursive(fieldValue, prefix+prefixTag); err != nil {
					return fmt.Errorf("populate nested struct %s: %w", field.Name, err)
				}
				continue
			}
		}

		// Handle Env tag
		envKey := field.Tag.Get(tagEnv)
		if envKey == "" {
			continue
		}

		fullKey := prefix + envKey
		value := s.source[fullKey]

		// Handle Default / Required
		if value == "" {
			if err := checkRequiredTag(fullKey, field); err != nil {
				return err
			}
			value = field.Tag.Get(tagDefault)
		}

		// Handle Text Replacement
		if value != "" {
			var err error
			value, err = s.resolveReplacement(value)
			if err != nil {
				return err
			}
		}

		// Set Value
		if value != "" {
			// Pass fullKey and value directly, no entry struct.
			if err := s.setFieldValue(fieldValue, fullKey, value); err != nil {
				return fmt.Errorf("set field %s: %w", field.Name, err)
			}
		}
	}
	return nil
}

// resolveReplacement resolves ${VAR} patterns.
func (s *settings) resolveReplacement(value string) (string, error) {
	match := textReplacementRegex.FindStringSubmatch(value)

	for _, m := range match {
		key := strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}")

		replacement, ok := s.source[key]
		if !ok || replacement == "" {
			return "", &ReplacementError{VariableName: key}
		}

		value = strings.ReplaceAll(value, m, replacement)
	}

	return value, nil
}
