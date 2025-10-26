// Package envconfig provides functionality to easily populate your config structure by using both environment variables, and a config file (optional).
package envconfig

import (
	"fmt"
	"os"
	"reflect"
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

// Set will parse the .env file and set the values in the environment, then populate the passed in struct
// using all environment variables.
func Set(filename string, config any) error {
	if filename != "" {
		if err := process(filename); err != nil {
			return fmt.Errorf("parse file: %w", err)
		}
	}

	if err := populateConfig(config); err != nil {
		return fmt.Errorf("populate config struct: %w", err)
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

		// Ignore fields that are not exported, or fields have a non-zero value.
		if !configFieldValue.CanSet() || !configFieldValue.IsZero() {
			continue
		}

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
