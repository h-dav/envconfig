package envconfig

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	// tagPrefix is used for nested structs inside your config struct.
	tagPrefix = "prefix"

	// tagEnv is used for fetching the environment variable by name.
	tagEnv = "env"

	// tagDefault is used to set a fallback value for a config field if the environment variable is not set.
	tagDefault = "default"

	// tagRequired is used for config struct fields that are required. If the environment variable is not set, an
	// error will be returned.
	tagRequired = "required"
)

// textReplacementRegex is used to detect text replacement in environment variables.
var textReplacementRegex = regexp.MustCompile(`\${[^}]+}`)

// Set will firstly set your .env file variables into the environment variables, then populate the passed struct
// using all environment variables.
func Set(filename string, config any) error {
	if !strings.HasSuffix(filename, ".env") {
		return &FileTypeValidationError{Filename: filename}
	}

	if err := setEnvironmentVariables(filename); err != nil {
		return fmt.Errorf("setting environment variables: %w", err)
	}

	if err := populateConfig(config); err != nil {
		return fmt.Errorf("populating config struct: %w", err)
	}

	return nil
}

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

		name, value, found := strings.Cut(line, ":")
		if !found {
			name, value, found = strings.Cut(line, "=")
			if !found {
				return &ParseError{Line: line}
			}
		}

		// Clean name of environment variable.
		name = strings.TrimSpace(name)

		// Clean a value of starting whitespace and comments.
		value = strings.TrimSpace(value)
		value, _, _ = strings.Cut(value, " #")

		if err = handleTextReplacement(&value); err != nil {
			return fmt.Errorf("handling text replacement: %w", err)
		}

		if err := os.Setenv(name, value); err != nil {
			return &SetEnvironmentVariableError{Err: err}
		}
	}

	if err := scanner.Err(); err != nil {
		return &FileReadError{Err: err}
	}

	return nil
}

// handleTextReplacement will check a .env file entry value for text replacements, and fulfill the text replacement.
func handleTextReplacement(value *string) error {
	match := textReplacementRegex.FindStringSubmatch(*value)

	for _, m := range match {
		envValue := strings.TrimPrefix(m, "${")
		envValue = strings.TrimSuffix(envValue, "}")

		matchedEnvValue := os.Getenv(envValue)
		if matchedEnvValue == "" {
			return &ReplacementError{VariableName: envValue}
		}

		*value = strings.ReplaceAll(*value, m, matchedEnvValue)
	}

	return nil
}

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

		if err := handlePrefixOption(field, configFieldValue, ""); err != nil {
			return fmt.Errorf("handling prefix option: %w", err)
		}

		environmentVariableName := field.Tag.Get(tagEnv)
		if environmentVariableName == "" {
			continue
		}

		environmentVariable := fetchEnvironmentVariable(environmentVariableName, field)
		if environmentVariable == "" {
			if err := checkRequiredOption(environmentVariableName, field); err != nil {
				return fmt.Errorf("checking required option: %w", err)
			}

			continue
		}

		if err := setFieldValue(configFieldValue, environmentVariable, environmentVariableName); err != nil {
			return fmt.Errorf("setting field value: %w", err)
		}
	}

	return nil
}

func handlePrefixOption(
	field reflect.StructField,
	configFieldValue reflect.Value,
	extendedPrefix string, // extendedPrefix is not zero value when a struct is deeply nested.
) error {
	if field.Type.Kind() == reflect.Struct {
		prefixOptionValue, prefixOptionSet := field.Tag.Lookup(tagPrefix)
		if !prefixOptionSet {
			return &PrefixOptionError{
				FieldName: field.Name,
			}
		}

		if prefixOptionSet {
			if err := populateNestedConfig(configFieldValue, extendedPrefix+prefixOptionValue); err != nil {
				return fmt.Errorf("populating nested config struct: %w", err)
			}
		}
	}

	return nil
}

func populateNestedConfig(nestedConfig reflect.Value, prefix string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		if !configFieldValue.CanSet() || !configFieldValue.IsZero() {
			continue
		}

		if err := handlePrefixOption(field, configFieldValue, prefix); err != nil {
			return fmt.Errorf("handling prefix option: %w", err)
		}

		envVariableName := prefix + field.Tag.Get(tagEnv)
		if envVariableName == prefix { // Ensure tag is set.
			continue
		}

		envValue := fetchEnvironmentVariable(envVariableName, field)
		if envValue == "" {
			if err := checkRequiredOption(envVariableName, field); err != nil {
				return fmt.Errorf("checking required option: %w", err)
			}

			continue
		}

		if err := setFieldValue(configFieldValue, envValue, envVariableName); err != nil {
			return fmt.Errorf("setting field value: %w", err)
		}
	}

	return nil
}

// fetchEnvironmentVariable returns the environment variable value. This also handles the default option tag.
func fetchEnvironmentVariable(envVariableName string, field reflect.StructField) string {
	environmentVariable := os.Getenv(envVariableName)

	if environmentVariable == "" {
		defaultOptionValue, defaultOptionSet := field.Tag.Lookup(tagDefault)
		if defaultOptionSet {
			environmentVariable = defaultOptionValue
		}
	}

	return environmentVariable
}

func checkRequiredOption(envVariableName string, field reflect.StructField) error {
	requiredOptionValue, requiredOptionSet := field.Tag.Lookup(tagRequired)
	if !requiredOptionSet {
		return nil
	}

	requiredOption, err := strconv.ParseBool(requiredOptionValue)
	if requiredOption {
		return &RequiredFieldError{FieldName: envVariableName}
	} else if err != nil {
		return &InvalidOptionConversionError{
			FieldName: envVariableName,
			Option:    tagRequired,
			Err:       err,
		}
	}

	return nil
}

func setFieldValue(
	configFieldValue reflect.Value,
	envValue string,
	envVariableName string,
) error {
	switch configFieldValue.Interface().(type) {
	case string:
		configFieldValue.SetString(envValue)
	case int:
		return setIntFieldValue(configFieldValue, envValue, envVariableName)
	case bool:
		return setBoolFieldValue(configFieldValue, envValue, envVariableName)
	case float64:
		return setFloatFieldValue(configFieldValue, envValue, envVariableName)
	case []string:
		return setStringSliceFieldValue(configFieldValue, envValue)
	case []int:
		return setIntSliceFieldValue(configFieldValue, envValue, envVariableName)
	case []float64:
		return setFloatSliceFieldValue(configFieldValue, envValue, envVariableName)
	default:
		return &UnsupportedFieldTypeError{FieldType: configFieldValue.Interface()}
	}

	return nil
}

func setIntFieldValue(
	configFieldValue reflect.Value,
	envValue string,
	envVariableName string,
) error {
	intValue, err := strconv.Atoi(envValue)
	if err != nil {
		return &FieldConversionError{
			FieldName:  envVariableName,
			TargetType: "int",
			Err:        err,
		}
	}

	configFieldValue.SetInt(int64(intValue))

	return nil
}

func setBoolFieldValue(
	configFieldValue reflect.Value,
	envValue string,
	envVariableName string,
) error {
	boolValue, err := strconv.ParseBool(envValue)
	if err != nil {
		return &FieldConversionError{
			FieldName:  envVariableName,
			TargetType: "bool",
			Err:        err,
		}
	}

	configFieldValue.SetBool(boolValue)

	return nil
}

func setFloatFieldValue(
	configFieldValue reflect.Value,
	envValue string,
	envVariableName string,
) error {
	floatValue, err := strconv.ParseFloat(envValue, 64)
	if err != nil {
		return &FieldConversionError{
			FieldName:  envVariableName,
			TargetType: "float",
			Err:        err,
		}
	}

	configFieldValue.SetFloat(floatValue)

	return nil
}

func setStringSliceFieldValue(configFieldValue reflect.Value, envValue string) error {
	values := strings.Split(envValue, ",")
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
	envValue string,
	envVariableName string,
) error {
	values := strings.Split(envValue, ",")
	slice := reflect.MakeSlice(configFieldValue.Type(), len(values), len(values))

	for i, v := range values {
		v = strings.TrimSpace(v)

		parsed, err := strconv.Atoi(v)
		if err != nil {
			return &FieldConversionError{
				FieldName:  envVariableName,
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
	envValue string,
	envVariableName string,
) error {
	values := strings.Split(envValue, ",")
	slice := reflect.MakeSlice(configFieldValue.Type(), len(values), len(values))

	for i, v := range values {
		v = strings.TrimSpace(v)

		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return &FieldConversionError{
				FieldName:  envVariableName,
				TargetType: "[]float64",
				Err:        err,
			}
		}

		slice.Index(i).SetFloat(parsed)
	}

	configFieldValue.Set(slice)

	return nil
}
