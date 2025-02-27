package envconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// Set will firstly set your .env file variables into the environment variables, then populate the passed struct
// using all environment variables.
func Set(filename string, config any) error {
	if err := setEnvironmentVariables(filename); err != nil {
		return err
	}

	if err := populateConfig(config); err != nil {
		return err
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
		text := scanner.Text()

		if text == "" {
			continue
		}

		// Handles commented lines.
		if strings.HasPrefix(text, "#") {
			continue
		}

		split := strings.SplitN(text, ":", 2) //nolint:mnd // "Magic number" is reasonable in this case.
		for i, v := range split {
			split[i] = strings.TrimSpace(v)
		}

		if err := os.Setenv(split[0], split[1]); err != nil {
			return &SetEnvironmentVariableError{Err: err}
		}
	}

	return nil
}

func populateConfig(config any) error {
	configStruct := reflect.ValueOf(config)
	if configStruct.Kind() != reflect.Ptr || configStruct.Elem().Kind() != reflect.Struct {
		return &InvalidConfigTypeError{ProvidedType: config}
	}

	configValue := reflect.ValueOf(config).Elem()

	for i := range configValue.NumField() {
		field := configValue.Type().Field(i)
		configFieldValue := configValue.Field(i)

		if err := handlePrefixOption(field, configFieldValue); err != nil {
			return err
		}

		envVariableName := field.Tag.Get("env")
		if envVariableName == "" {
			continue
		}

		envValue := fetchEnvValue(envVariableName, field)
		if envValue == "" {
			if err := checkRequiredOption(envVariableName, field); err != nil {
				return err
			}

			continue
		}

		if err := setFieldValue(configFieldValue, envValue, envVariableName); err != nil {
			return err
		}
	}

	return nil
}

func fetchEnvValue(envVariableName string, field reflect.StructField) string {
	envValue := os.Getenv(envVariableName)

	if envValue == "" {
		defaultOptionValue, defaultOptionSet := field.Tag.Lookup("default")
		if defaultOptionSet {
			envValue = defaultOptionValue
		}
	}

	return envValue
}

func handlePrefixOption(field reflect.StructField, configFieldValue reflect.Value) error {
	if field.Type.Kind() == reflect.Struct {
		prefixOptionValue, prefixOptionSet := field.Tag.Lookup("prefix")
		if !prefixOptionSet {
			return &PrefixOptionError{
				ParamName: field.Name,
			}
		}

		if prefixOptionSet {
			if err := populateNestedConfig(configFieldValue, prefixOptionValue); err != nil {
				return err
			}
		}
	}

	return nil
}

func checkRequiredOption(envVariableName string, field reflect.StructField) error {
	requiredOptionValue, requiredOptionSet := field.Tag.Lookup("required")
	if !requiredOptionSet {
		return nil
	}

	requiredOption, err := strconv.ParseBool(requiredOptionValue)
	if requiredOption {
		return &RequiredFieldError{ParamName: envVariableName}
	} else if err != nil {
		return &InvalidOptionConversionError{
			ParamName: envVariableName,
			Option:    "required",
			Err:       err,
		}
	}

	return nil
}

func populateNestedConfig(nestedConfig reflect.Value, prefix string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		envVariableName := prefix + field.Tag.Get("env")
		if envVariableName == "" {
			continue
		}

		envValue := fetchEnvValue(envVariableName, field)
		if envValue == "" {
			if err := checkRequiredOption(envVariableName, field); err != nil {
				return err
			}

			continue
		}

		if err := setFieldValue(configFieldValue, envValue, envVariableName); err != nil {
			return err
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
		envValue, err := strconv.Atoi(envValue)
		if err != nil {
			return &ParamConversionError{
				ParamName:  envVariableName,
				TargetType: "int",
				Err:        err,
			}
		}

		configFieldValue.SetInt(int64(envValue))
	case bool:
		envValue, err := strconv.ParseBool(envValue)
		if err != nil {
			return &ParamConversionError{
				ParamName:  envVariableName,
				TargetType: "bool",
				Err:        err,
			}
		}

		configFieldValue.SetBool(envValue)
	case float64:
		envValue, err := strconv.ParseFloat(envValue, 64)
		if err != nil {
			return &ParamConversionError{
				ParamName:  envVariableName,
				TargetType: "float",
				Err:        err,
			}
		}

		configFieldValue.SetFloat(envValue)
	default:
		return &UnsupportedFieldTypeError{FieldType: configFieldValue.Interface()}
	}

	return nil
}
