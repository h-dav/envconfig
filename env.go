package envconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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

		err = handleTextReplacement(&split[1])
		if err != nil {
			return err
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

		environmentVariable := fetchEnvironmentVariable(envVariableName, field)
		if environmentVariable == "" {
			if err := checkRequiredOption(envVariableName, field); err != nil {
				return err
			}

			continue
		}

		if err := setFieldValue(configFieldValue, environmentVariable, envVariableName); err != nil {
			return err
		}
	}

	return nil
}

func handleTextReplacement(value *string) error {
	re := regexp.MustCompile(`\${[^}]+}`)
	match := re.FindStringSubmatch(*value)

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

func populateNestedConfig(nestedConfig reflect.Value, prefix string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		envVariableName := prefix + field.Tag.Get("env")
		if envVariableName == "" {
			continue
		}

		envValue := fetchEnvironmentVariable(envVariableName, field)
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

func fetchEnvironmentVariable(envVariableName string, field reflect.StructField) string {
	environmentVariable := os.Getenv(envVariableName)

	if environmentVariable == "" {
		defaultOptionValue, defaultOptionSet := field.Tag.Lookup("default")
		if defaultOptionSet {
			environmentVariable = defaultOptionValue
		}
	}

	return environmentVariable
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

func setFieldValue( //nolint:gocognit,gocyclo // Switch case required high complexity to cover supported types.
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
	case []string:
		values := strings.Split(envValue, ",")
		slice := reflect.MakeSlice(configFieldValue.Type(), len(values), len(values))

		for i, v := range values {
			v = strings.TrimSpace(v)
			slice.Index(i).SetString(strings.TrimSpace(v))
		}

		configFieldValue.Set(slice)
	case []int:
		values := strings.Split(envValue, ",")
		slice := reflect.MakeSlice(configFieldValue.Type(), len(values), len(values))

		for i, v := range values {
			v = strings.TrimSpace(v)

			parsed, err := strconv.Atoi(v)
			if err != nil {
				return &ParamConversionError{
					ParamName:  envVariableName,
					TargetType: "[]int",
					Err:        err,
				}
			}

			slice.Index(i).SetInt(int64(parsed))
		}

		configFieldValue.Set(slice)
	case []float64:
		values := strings.Split(envValue, ",")
		slice := reflect.MakeSlice(configFieldValue.Type(), len(values), len(values))

		for i, v := range values {
			v = strings.TrimSpace(v)

			parsed, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return &ParamConversionError{
					ParamName:  envVariableName,
					TargetType: "[]float64",
					Err:        err,
				}
			}

			slice.Index(i).SetFloat(parsed)
		}

		configFieldValue.Set(slice)
	default:
		return &UnsupportedFieldTypeError{FieldType: configFieldValue.Interface()}
	}

	return nil
}
