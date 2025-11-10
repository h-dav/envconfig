// Package envconfig provides functionality to easily load config into your struct.
package envconfig

import (
	"fmt"
	"maps"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

type entry struct {
	key, value string
}

// textReplacementRegex is used to detect text replacement in environment variables.
var textReplacementRegex = regexp.MustCompile(`\${[^}]+}`)

// Set will parse multiple sources for config values, and use these values to populate the passed in config struct.
func Set(config any, opts ...option) error {
	s := &settings{
		source:   map[string]string{},
		decoders: defaultDecoders,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.sources = append(s.sources, EnvironmentVariableSource{}, FlagSource{})

	if s.activeProfile != "" {
		if s.filepath == "" {
			return fmt.Errorf("assign active profile: %w", &IncompatibleOptionsError{
				FirstOption:  "WithActiveProfile()",
				SecondOption: "WithFilepath()",
				Reason:       "directory in filepath option must be provided when using active profile",
			})
		}

		dir, _ := filepath.Split(s.filepath)

		s.filepath = dir + s.activeProfile + envExtension
	}

	for _, source := range s.sources {
		values, err := source.Load()
		if err != nil {
			return fmt.Errorf("load from source: %w", err)
		}

		maps.Copy(s.source, values)
	}

	if err := s.populateStruct(config); err != nil {
		return fmt.Errorf("populate config struct: %w", err)
	}

	return nil
}

// populateStruct uses the items in settings.source to populate the passed in config struct.
func (s *settings) populateStruct(config any) error {
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

		if err := chain.Handle(field, configFieldValue, s, ""); err != nil {
			return fmt.Errorf("process field '%s': %w", field.Name, err)
		}
	}

	return nil
}

// resolveReplacement checks if a string has the pattern of ${...}, and if so, uses values in settings.source to
// replace the pattern, and returns the newly created string.
func (s *settings) resolveReplacement(value string) (string, error) {
	match := textReplacementRegex.FindStringSubmatch(value)

	for _, m := range match {
		environmentValue := strings.TrimPrefix(m, "${")
		environmentValue = strings.TrimSuffix(environmentValue, "}")

		replacementValue := s.source[environmentValue]
		if replacementValue == "" {
			return "", &ReplacementError{VariableName: environmentValue}
		}

		value = strings.ReplaceAll(value, m, replacementValue)
	}

	return value, nil
}

// populateNestedConfig populates a nested struct.
func (s *settings) populateNestedConfig(nestedConfig reflect.Value, prefix string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		if !configFieldValue.CanSet() {
			continue
		}

		// Process the field with the chain.
		if err := chain.Handle(field, configFieldValue, s, prefix); err != nil {
			return fmt.Errorf("error processing field '%s': %w", field.Name, err)
		}
	}

	return nil
}
