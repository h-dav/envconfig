// Package envconfig provides functionality to easily populate your config structure by using both environment variables, and a .env file (optional).
package envconfig

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
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

type entry struct {
	key, value string
}

// textReplacementRegex is used to detect text replacement in environment variables.
var textReplacementRegex = regexp.MustCompile(`\${[^}]+}`)

// Set will parse the .env file and set the values in the environment, then populate the passed in struct
// using ALL environment variables.
func Set(filename string, config any) error {
	if filename != "" {
		if filepath.Ext(filename) != ".env" {
			return &FileTypeValidationError{Filename: filename}
		}

		if err := setEnvironmentVariables(filename); err != nil {
			return fmt.Errorf("set environment variables: %w", err)
		}
	}

	if err := populateConfig(config); err != nil {
		return fmt.Errorf("populate config struct: %w", err)
	}

	return nil
}

// setEnvironmentVariables will parse the file and set the values in the environment.
func setEnvironmentVariables(filename string) error {
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

		entry, err := parseEnvLine(line)
		if err != nil {
			return fmt.Errorf("parse environment variable line: %w", err)
		}

		if err = handleTextReplacement(&entry.value); err != nil {
			return fmt.Errorf("handle text replacement: %w", err)
		}

		if err := os.Setenv(entry.key, entry.value); err != nil {
			return &SetEnvironmentVariableError{Err: err}
		}
	}

	if err := scanner.Err(); err != nil {
		return &FileReadError{Filename: filename, Err: err}
	}

	return nil
}

// parseEnvLine parses an individual .env line, and detect comments.
func parseEnvLine(line string) (entry, error) {
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

// handleTextReplacement will check a .env file entry value for text replacements, and fulfill the text replacement.
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

// populateConfig populated the config struct using all environment variables.
func populateConfig(config any) error { //nolint:gocognit // Complexity is reasonable.
	configStruct := reflect.ValueOf(config)
	if configStruct.Kind() != reflect.Ptr || configStruct.Elem().Kind() != reflect.Struct {
		return &InvalidConfigTypeError{ProvidedType: config}
	}

	configValue := reflect.ValueOf(config).Elem()

	for i := range configValue.NumField() {
		field := configValue.Type().Field(i)
		configFieldValue := configValue.Field(i)

		// Ensure the field is exported and the field is not already populated.
		if !configFieldValue.CanSet() || !configFieldValue.IsZero() {
			continue
		}

		// Check if tagJSON option is set.
		jsonOptionValue, jsonOptionSet := field.Tag.Lookup(tagJSON)
		if jsonOptionSet {
			if err := handleJSONOption(configFieldValue, jsonOptionValue); err != nil {
				return fmt.Errorf("handle JSON option: %w", err)
			}

			continue
		}

		if err := handlePrefixOption(field, configFieldValue, ""); err != nil {
			return fmt.Errorf("handle prefix option: %w", err)
		}

		environmentVariableKey := field.Tag.Get(tagEnv)
		if environmentVariableKey == "" {
			continue
		}

		environmentVariable := fetchEnvironmentVariable(environmentVariableKey, field)
		if environmentVariable == "" {
			if err := checkRequiredOption(environmentVariableKey, field); err != nil {
				return fmt.Errorf("check required option: %w", err)
			}

			continue
		}

		if err := setFieldValue(configFieldValue, entry{environmentVariableKey, environmentVariable}); err != nil {
			return fmt.Errorf("set field value: %w", err)
		}
	}

	return nil
}

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

// populateNestedConfig populates a nested struct.
func populateNestedConfig(nestedConfig reflect.Value, prefix string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		if !configFieldValue.CanSet() || !configFieldValue.IsZero() {
			continue
		}

		jsonOptionValue, jsonOptionSet := field.Tag.Lookup(tagJSON)
		if jsonOptionSet {
			err := handleJSONOption(configFieldValue, prefix+jsonOptionValue)
			if err != nil {
				return fmt.Errorf("handle JSON option: %w", err)
			}

			continue
		}

		if err := handlePrefixOption(field, configFieldValue, prefix); err != nil {
			return fmt.Errorf("handle prefix option: %w", err)
		}

		environmentVariableKey := prefix + field.Tag.Get(tagEnv)
		if environmentVariableKey == prefix { // Ensure tag is set.
			continue
		}

		environmentValue := fetchEnvironmentVariable(environmentVariableKey, field)
		if environmentValue == "" {
			if err := checkRequiredOption(environmentVariableKey, field); err != nil {
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
func fetchEnvironmentVariable(environmentVariableKey string, field reflect.StructField) string {
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

// setFieldValue determines the type of a config field, and branch out to the correct
// function to populate that data type.
func setFieldValue(
	configFieldValue reflect.Value,
	entry entry,
) error {
	switch configFieldValue.Interface().(type) {
	case string:
		configFieldValue.SetString(entry.value)
	case int:
		return setIntFieldValue(configFieldValue, entry)
	case bool:
		return setBoolFieldValue(configFieldValue, entry)
	case float64:
		return setFloatFieldValue(configFieldValue, entry)
	case []string:
		return setStringSliceFieldValue(configFieldValue, entry.value)
	case []int:
		return setIntSliceFieldValue(configFieldValue, entry)
	case []float64:
		return setFloatSliceFieldValue(configFieldValue, entry)
	case time.Duration:
		return setDurationFieldValue(configFieldValue, entry)
	default:
		return &UnsupportedFieldTypeError{FieldType: configFieldValue.Interface()}
	}

	return nil
}

func setIntFieldValue(
	configFieldValue reflect.Value,
	entry entry,
) error {
	intValue, err := strconv.Atoi(entry.value)
	if err != nil {
		return &FieldConversionError{
			FieldName:  entry.key,
			TargetType: "int",
			Err:        err,
		}
	}

	configFieldValue.SetInt(int64(intValue))

	return nil
}

func setBoolFieldValue(
	configFieldValue reflect.Value,
	entry entry,
) error {
	boolValue, err := strconv.ParseBool(entry.value)
	if err != nil {
		return &FieldConversionError{
			FieldName:  entry.key,
			TargetType: "bool",
			Err:        err,
		}
	}

	configFieldValue.SetBool(boolValue)

	return nil
}

func setFloatFieldValue(
	configFieldValue reflect.Value,
	entry entry,
) error {
	floatValue, err := strconv.ParseFloat(entry.value, 64)
	if err != nil {
		return &FieldConversionError{
			FieldName:  entry.key,
			TargetType: "float",
			Err:        err,
		}
	}

	configFieldValue.SetFloat(floatValue)

	return nil
}

func setStringSliceFieldValue(configFieldValue reflect.Value, environmentValue string) error {
	values := strings.Split(environmentValue, ",")
	slice := reflect.MakeSlice(configFieldValue.Type(), len(values), len(values))

	for i, v := range values {
		v = strings.TrimSpace(v)
		slice.Index(i).SetString(v)
	}

	configFieldValue.Set(slice)

	return nil
}

func setIntSliceFieldValue(
	configFieldValue reflect.Value,
	entry entry,
) error {
	values := strings.Split(entry.value, ",")
	slice := reflect.MakeSlice(configFieldValue.Type(), len(values), len(values))

	for i, v := range values {
		v = strings.TrimSpace(v)

		parsed, err := strconv.Atoi(v)
		if err != nil {
			return &FieldConversionError{
				FieldName:  entry.key,
				TargetType: "[]int",
				Err:        err,
			}
		}

		slice.Index(i).SetInt(int64(parsed))
	}

	configFieldValue.Set(slice)

	return nil
}

func setFloatSliceFieldValue(
	configFieldValue reflect.Value,
	entry entry,
) error {
	values := strings.Split(entry.value, ",")
	slice := reflect.MakeSlice(configFieldValue.Type(), len(values), len(values))

	for i, v := range values {
		v = strings.TrimSpace(v)

		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return &FieldConversionError{
				FieldName:  entry.key,
				TargetType: "[]float64",
				Err:        err,
			}
		}

		slice.Index(i).SetFloat(parsed)
	}

	configFieldValue.Set(slice)

	return nil
}

func setDurationFieldValue(
	configFieldValue reflect.Value,
	entry entry,
) error {
	durationValue, err := time.ParseDuration(entry.value)
	if err != nil {
		return &FieldConversionError{
			FieldName:  entry.key,
			TargetType: "time.Duration",
			Err:        err,
		}
	}

	configFieldValue.Set(reflect.ValueOf(durationValue))

	return nil
}
