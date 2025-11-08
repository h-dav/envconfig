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

type entry struct {
	key, value string
}

// textReplacementRegex is used to detect text replacement in environment variables.
var textReplacementRegex = regexp.MustCompile(`\${[^}]+}`)

// Set will parse multiple sources for config values, and use these values to populate the passed in config struct.
func Set(config any, opts ...option) error {
	s := &settings{
		source: map[string]string{},
	}

	for _, opt := range opts {
		opt(s)
	}

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

	if s.filepath != "" {
		if err := s.processFilepath(); err != nil {
			return fmt.Errorf("process filepath: %w", err)
		}
	}

	if err := s.processEnvironmentVariables(); err != nil {
		return fmt.Errorf("process environment variables: %w", err)
	}

	if err := s.populateStruct(config); err != nil {
		return fmt.Errorf("populate config struct: %w", err)
	}

	return nil
}

// populateStruct uses the items in settings.source to populate the passed in config struct.
func (s settings) populateStruct(config any) error {
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
			err := json.Unmarshal([]byte(s.source[jsonOptionValue]), configFieldValue.Addr().Interface())
			if err != nil {
				return fmt.Errorf("unmarshal JSON: %w", err)
			}
			continue
		}

		if err := s.handlePrefixTag(field, configFieldValue, ""); err != nil {
			return fmt.Errorf("handle prefix tag: %w", err)
		}

		key := field.Tag.Get(tagEnv)
		if key == "" {
			continue
		}

		value := s.source[key]
		if value == "" {
			if err := checkRequiredTag(key, field); err != nil {
				return fmt.Errorf("check required tag: %w", err)
			}

			value = field.Tag.Get(tagDefault)
		}

		value, err := s.resolveReplacement(value)
		if err != nil {
			return fmt.Errorf("resolve replacement: %w", err)
		}

		if err := setFieldValue(
			configFieldValue, entry{key, value}); err != nil {
			return fmt.Errorf("set field value: %w", err)
		}
	}

	return nil
}

// resolveReplacement checks if a string has the pattern of ${...}, and if so, uses values in settings.source to
// replace the pattern, and returns the newly created string.
func (s settings) resolveReplacement(value string) (string, error) {
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
func (s settings) populateNestedConfig(nestedConfig reflect.Value, prefix string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		if !configFieldValue.CanSet() || !configFieldValue.IsZero() {
			continue
		}

		jsonOptionValue, jsonOptionSet := field.Tag.Lookup(tagJSON)
		if jsonOptionSet {
			err := json.Unmarshal([]byte(s.source[jsonOptionValue]), &configFieldValue)
			if err != nil {
				return fmt.Errorf("handle JSON tag: %w", err)
			}

			continue
		}

		if err := s.handlePrefixTag(field, configFieldValue, prefix); err != nil {
			return fmt.Errorf("handle prefix tag: %w", err)
		}

		environmentVariableKey := prefix + field.Tag.Get(tagEnv)
		if environmentVariableKey == prefix { // Ensure tag is set.
			continue
		}
		if err := setFieldValue(
			configFieldValue, entry{environmentVariableKey, s.source[environmentVariableKey]}); err != nil {
			return fmt.Errorf("set field value: %w", err)
		}
	}

	return nil
}
