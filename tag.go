package envconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

const (
	// tagEnv is used for fetching the environment variable by name.
	tagEnv = "env"

	// tagDefault is used to set a fallback value for a config field if the environment variable is not set.
	tagDefault = "default"

	// tagRequired is used for config struct fields that are required. If the environment variable is not set, an
	// error will be returned.
	tagRequired = "required"

	// tagJSON is used for environment variables that are JSON.
	tagJSON = "envjson"

	// tagPrefix is used for nested structs inside your config struct.
	tagPrefix = "prefix"
)

// handlePrefixTag will handle nested structures that use the prefix option.
func (e environmentVariableParser) handlePrefixTag(
	field reflect.StructField,
	configFieldValue reflect.Value,
	prefix string, // prefix is not zero value when a struct is deeply nested.
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

// handleJSONTag will handle populating JSON structs via environment variables that are JSON.
func (e environmentVariableParser) handleJSONTag(
	configFieldValue reflect.Value,
	environmentKey string, // environmentKey is not zero value when a struct is deeply nested.
) error {
	environmentValue := os.Getenv(environmentKey)

	if err := json.Unmarshal([]byte(environmentValue), configFieldValue.Addr().Interface()); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}

	return nil
}

// handleJSONTag will handle populating JSON structs via environment variables that are JSON.
func (e envFileParser) handleJSONTag(
	configFieldValue reflect.Value,
	environmentKey string, // environmentKey is not zero value when a struct is deeply nested.
) error {
	err := json.Unmarshal([]byte(e.config[environmentKey]), &configFieldValue)
	if err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}

	return nil
}

// checkRequiredTag checks if a field is required and returns an error if so.
//
// This function is only called when an environment variable is not set for a field.
func checkRequiredTag(environmentVariableKey string, field reflect.StructField) error {
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
