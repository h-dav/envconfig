package envconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

// handlePrefixOption will handle nested structures that use the prefix option.
func handlePrefixOption(
	field reflect.StructField,
	configFieldValue reflect.Value,
	prefix string, // extendedPrefix is not zero value when a struct is deeply nested.
) error {
	if field.Type.Kind() != reflect.Struct {
		return nil
	}

	prefixOptionValue, prefixOptionSet := field.Tag.Lookup(tagPrefix)
	if !prefixOptionSet {
		return &PrefixOptionError{FieldName: field.Name}
	}

	if err := populateNestedConfig(configFieldValue, prefix+prefixOptionValue); err != nil {
		return fmt.Errorf("populate nested config struct: %w", err)
	}

	return nil
}

// handleJSONOption will handle populating JSON structs via environment variables that are JSON.
func handleJSONOption(
	configFieldValue reflect.Value,
	environmentKey string, // environmentKey is not zero value when a struct is deeply nested.
) error {
	if err := populateJSON(configFieldValue, environmentKey); err != nil {
		return fmt.Errorf("populate JSON config struct: %w", err)
	}

	return nil
}

// populateJSON will populate the JSON struct.
func populateJSON(configFieldValue reflect.Value, environmentVariableKey string) error {
	environmentValue := os.Getenv(environmentVariableKey)

	if err := json.Unmarshal([]byte(environmentValue), configFieldValue.Addr().Interface()); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}

	return nil
}

// checkRequiredOption checks if a field is required and returns an error if so.
//
// This function is only called when an environment variable is not set for a field.
func checkRequiredOption(environmentVariableKey string, field reflect.StructField) error {
	requiredOptionValue, requiredOptionSet := field.Tag.Lookup(tagRequired)
	if !requiredOptionSet {
		return nil
	}

	requiredOption, err := strconv.ParseBool(requiredOptionValue)
	if requiredOption {
		return &RequiredFieldError{FieldName: environmentVariableKey}
	} else if err != nil {
		return &InvalidOptionConversionError{
			FieldName: environmentVariableKey,
			Option:    tagRequired,
			Err:       err,
		}
	}

	return nil
}
