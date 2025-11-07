package envconfig

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// processEnvironmentVariables populates the config struct using all environment variables.
func (s settings) processEnvironmentVariables(config any) error { //nolint:gocognit // Complexity is reasonable.
	e := environmentVariableParser{
		prefix: s.prefix,
	}

	if err := e.parse(config); err != nil {
		return err
	}

	return nil
}

type environmentVariableParser struct {
	prefix string
}

func (e environmentVariableParser) parse(config any) error {
	configStruct := reflect.ValueOf(config)
	if configStruct.Kind() != reflect.Pointer || configStruct.Elem().Kind() != reflect.Struct {
		return &InvalidConfigTypeError{ProvidedType: config}
	}

	configValue := reflect.ValueOf(config).Elem()

	for i := range configValue.NumField() {
		field := configValue.Type().Field(i)
		configFieldValue := configValue.Field(i)

		// Ignore fields that are not exported, or fields have a non-zero value.
		if !configFieldValue.CanSet() || !configFieldValue.IsZero() {
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

		environmentVariable := e.fetchEnvironmentVariable(e.prefix+environmentVariableKey, field)
		if environmentVariable == "" {
			if err := checkRequiredTag(environmentVariableKey, field); err != nil {
				return fmt.Errorf("check required tag: %w", err)
			}

			continue
		}

		if err := e.handleTextReplacement(&environmentVariable); err != nil {
			return fmt.Errorf("handle text replacement: %w", err)
		}

		if err := setFieldValue(configFieldValue, entry{environmentVariableKey, environmentVariable}); err != nil {
			return fmt.Errorf("set field value: %w", err)
		}
	}

	return nil
}


// populateNestedConfig populates a nested struct.
func (e environmentVariableParser) populateNestedConfig(nestedConfig reflect.Value, prefix string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		if !configFieldValue.CanSet() || !configFieldValue.IsZero() {
			continue
		}

		jsonOptionValue, jsonOptionSet := field.Tag.Lookup(tagJSON)
		if jsonOptionSet {
			err := e.handleJSONTag(configFieldValue, prefix+jsonOptionValue)
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

		environmentValue := e.fetchEnvironmentVariable(environmentVariableKey, field)
		if environmentValue == "" {
			if err := checkRequiredTag(environmentVariableKey, field); err != nil {
				return fmt.Errorf("check required option: %w", err)
			}

			continue
		}

		if err := setFieldValue(configFieldValue, entry{environmentVariableKey, environmentValue}); err != nil {
			return fmt.Errorf("set field value: %w", err)
		}
	}

	return nil
}

// fetchEnvironmentVariable returns the environment variable value. This also handles the default option tag.
func (e environmentVariableParser) fetchEnvironmentVariable(environmentVariableKey string, field reflect.StructField) string {
	environmentVariable := os.Getenv(environmentVariableKey)

	if environmentVariable != "" {
		return environmentVariable
	}

	defaultOptionValue, defaultOptionSet := field.Tag.Lookup(tagDefault)
	if defaultOptionSet {
		return defaultOptionValue
	}

	return environmentVariable
}

func (e environmentVariableParser) handleTextReplacement(value *string) error {
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
