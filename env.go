// Package envconfig parses a .env file and loads the values into the os and returns a populated config structure.
package envconfig

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type pair struct {
	key   string
	value string
}

type options struct {
	prefix       bool
	required     bool
	defaultValue string
}

// Options are included in the tag.
const (
	// OptPrefix is used for nested structures.
	OptPrefix = "prefix"
	// OptRequired will fail if the value is not set in the env variables.
	OptRequired = "required"
	// OptDefault will be the value that is fallen back to if an env variable is not set.
	OptDefault = "default"
)

var (
	// ErrRequiredNotFulfilled occurs when an environment variable is not populated but is required by the config structure.
	ErrRequiredNotFulfilled = errors.New("required value is not set")
	// ErrMismatchedDataType occurs when an environment variables value cannot be parsed into the config structure's data type.
	ErrMismatchedDataType = errors.New("data types do not match")
	// ErrTextReplacement occurs when an environment variable is not set at the point of text replacement.
	ErrTextReplacement = errors.New("required value is not set for text replacement")
)

// SetVars reads a .env file and then writes the values to os env variables.
func SetVars(filename string) error {
	return setVars(filename)
}

// Populate will map values from the environment variables into the struct passed in.
func Populate(cfg interface{}) error {
	return populate(cfg)
}

// SetPopulate combines the functionality of SetVars and Populate.
func SetPopulate(filename string, cfg interface{}) error {
	if err := setVars(filename); err != nil {
		return err
	}

	if err := populate(cfg); err != nil {
		return err
	}

	return nil
}

func setVars(filename string) error {
	reader, file, err := streamFile(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		pair, err := entryConvert(strings.TrimSuffix(line, "\n"))
		if err != nil {
			return err
		}

		err = setEnvValue(pair)
		if err != nil {
			return err
		}
	}

	return nil
}

func entryConvert(line string) (pair, error) {
	newVar := strings.Split(line, "=")
	value := newVar[1]

	for {
		re := regexp.MustCompile(`\$\{([^}]*)\}`)
		match := re.FindStringSubmatch(value)

		if len(match) != 0 {
			matchedEnvValue := os.Getenv(match[1])
			if matchedEnvValue == "" {
				return pair{}, errors.New("failed")
			}

			value = strings.Replace(value, match[0], matchedEnvValue, 1)

			continue
		}
		break
	}

	return pair{key: newVar[0], value: value}, nil
}

func setEnvValue(pair pair) error {
	err := os.Setenv(pair.key, pair.value)
	if err != nil {
		return err
	}

	return nil
}

// Remember to close the file after reading from Reader is done.
func streamFile(filename string) (*bufio.Reader, *os.File, error) {
	file, err := os.Open(filename)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return bufio.NewReader(file), file, nil
}

// populate populates the config structure passed in.
func populate(cfg interface{}) error {
	value := reflect.ValueOf(cfg)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return errors.New("input must be a non-nil pointer to a struct")
	}

	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return errors.New("input must be a pointer to a struct")
	}

	vType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := vType.Field(i)
		tag := field.Tag.Get("env")

		key, opts := keyAndOptions(tag)

		// Check if the env struct key for the field is set in the environment variables.
		if opts.required {
			if err := handleRequired(key); err != nil {
				return fmt.Errorf("handling required option: %v", err)
			}
		}

		// Populate the nested struct (All nested structs must use `prefix`).
		if opts.prefix {
			if err := handlePrefix(value, i, field, key); err != nil {
				return fmt.Errorf("handling prefix option: %v", err)
			}
			continue
		}

		envValue := os.Getenv(key)
		if envValue == "" {
			if opts.defaultValue == "" {
				continue
			}
			envValue = opts.defaultValue
		}

		if err := setInStruct(envValue, value.Field(i)); err != nil {
			return fmt.Errorf("setting value in config struct: %v", err)
		}
	}

	return nil
}

func setInStruct(envValue string, value reflect.Value) error {
	switch value.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		intValue, err := strconv.Atoi(envValue)
		if err != nil {
			return ErrMismatchedDataType
		}
		value.SetInt(int64(intValue))
	case reflect.Bool:
		result := false
		if envValue == "true" || envValue == "1" {
			result = true
		}
		value.SetBool(result)
	case reflect.Float64:
		floatValue, err := strconv.ParseFloat(envValue, 64)
		if err != nil {
			return ErrMismatchedDataType
		}
		value.SetFloat(floatValue)
	default:
		value.SetString(envValue)
	}

	return nil
}

func handlePrefix(value reflect.Value, i int, field reflect.StructField, key string) error {
	nestedStruct := value.Field(i)
	nestedType := field.Type

	for ni := 0; ni < nestedStruct.NumField(); ni++ {
		nestedField := nestedType.Field(ni)
		nestedTag := nestedField.Tag.Get("env")

		nestedKey, nestedOpts := keyAndOptions(key + nestedTag)

		envValue := os.Getenv(nestedKey)

		if envValue == "" && nestedOpts.required {
			return ErrRequiredNotFulfilled
		}

		if envValue == "" {
			continue
		}

		if err := setInStruct(envValue, value.Field(i).Field(ni)); err != nil {
			return fmt.Errorf("setting nested struct value: %v", err)
		}
	}

	return nil
}

func handleRequired(key string) error {
	envValue := os.Getenv(key)
	if envValue == "" {
		return ErrRequiredNotFulfilled
	}

	return nil
}

// keyAndOptions separates the key within the `env` struct tag key from the options.
func keyAndOptions(tag string) (string, options) {
	parts := strings.Split(tag, ",")

	key, tagOpts := strings.TrimSpace(parts[0]), parts[1:]

	var opts options

	if strings.HasPrefix(key, "prefix=") {
		opts.prefix = true
		key = strings.Split(key, "prefix=")[1]
	}

	// For options that do not need a value.
	for _, o := range tagOpts {
		switch {
		case o == OptRequired:
			opts.required = true
		case strings.HasPrefix(o, "default="):
			opts.defaultValue = strings.Split(o, "default=")[1]
		}
	}

	return key, opts
}
