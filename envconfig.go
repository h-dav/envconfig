// Package envconfig provides functionality to easily load config into your struct.
package envconfig

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type entry struct {
	key, value string
}

// textReplacementRegex is used to detect text replacement in environment variables.
var textReplacementRegex = regexp.MustCompile(`\${[^}]+}`)

// Set will parse the .env file and set the values in the environment, then populate the passed in struct
// using all environment variables.
func Set(config any, opts ...option) error {
	s := &settings{
		source: map[string]string{},
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.filename != "" {
		if err := s.processFilename(); err != nil {
			return fmt.Errorf("process file: %w", err)
		}
	}

	if err := s.processEnvironmentVariables(); err != nil {
		return fmt.Errorf("process environment variables: %w", err)
	}

	if err := s.populateConfig(config); err != nil {
		return fmt.Errorf("populate config: %w", err)
	}

	return nil
}

func (s settings) populateConfig(config any) error {
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

		if err := handlePrefixTag(field, configFieldValue, "", s.source); err != nil {
			return fmt.Errorf("handle prefix tag: %w", err)
		}

		environmentVariableKey := field.Tag.Get(tagEnv)
		if environmentVariableKey == "" {
			continue
		}

		value := s.source[environmentVariableKey]
		if value == "" {
			if err := checkRequiredTag(environmentVariableKey, field); err != nil {
				return fmt.Errorf("check required tag: %w", err)
			}

			value = field.Tag.Get(tagDefault)
		}

		match := textReplacementRegex.FindStringSubmatch(value)

		for _, m := range match {
			environmentValue := strings.TrimPrefix(m, "${")
			environmentValue = strings.TrimSuffix(environmentValue, "}")

			replacementValue := s.source[environmentValue]
			if replacementValue == "" {
				return &ReplacementError{VariableName: environmentValue}
			}

			value = strings.ReplaceAll(value, m, replacementValue)
		}

		if err := setFieldValue(
			configFieldValue, entry{environmentVariableKey, value}); err != nil {
			return fmt.Errorf("set field value: %w", err)
		}
	}

	return nil
}

// populateNestedConfig populates a nested struct.
func populateNestedConfig(nestedConfig reflect.Value, prefix string, source map[string]string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		if !configFieldValue.CanSet() || !configFieldValue.IsZero() {
			continue
		}

		jsonOptionValue, jsonOptionSet := field.Tag.Lookup(tagJSON)
		if jsonOptionSet {
			err := json.Unmarshal([]byte(source[jsonOptionValue]), &configFieldValue)
			if err != nil {
				return fmt.Errorf("handle JSON tag: %w", err)
			}

			continue
		}

		if err := handlePrefixTag(field, configFieldValue, prefix, source); err != nil {
			return fmt.Errorf("handle prefix tag: %w", err)
		}

		environmentVariableKey := prefix + field.Tag.Get(tagEnv)
		if environmentVariableKey == prefix { // Ensure tag is set.
			continue
		}
		if err := setFieldValue(
			configFieldValue, entry{environmentVariableKey, source[environmentVariableKey]}); err != nil {
			return fmt.Errorf("set field value: %w", err)
		}
	}

	return nil
}
