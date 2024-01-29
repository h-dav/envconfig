// Package envconfig parses a .env file and loads the values into the os and returns a populated config structure.
package envconfig

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type pair struct {
	key   string
	value string
}

// options are included in the tag.
type options struct {
	// prefix is used for nested structures.
	prefix bool
	// required fields will fail if they are not set in the OS values.
	required bool
}

const (
	OptPrefix   = "prefix"
	OptRequired = "required"
)

var (
	ErrRequiredFulfilled  = errors.New("required value is not set")
	ErrMismatchedDataType = errors.New("data types do not match")
)

// SetVars reads a .env file and then writes the values to os env variables.
func SetVars(filename string) error {
	return setVars(filename)
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
				break // Reached end of file.
			} else {
				return err // Handle other errors.
			}
		}

		pair := entryConvert(strings.TrimSuffix(line, "\n"))

		err = setEnvValue(pair)
		if err != nil {
			return err
		}
	}

	return nil
}

func entryConvert(line string) pair {
	newVar := strings.Split(line, "=")
	return pair{key: newVar[0], value: newVar[1]}
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

// Populate will map values from the OS into the struct passed in.
func Populate(cfg interface{}) error {
	return populate(cfg)
}

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

		// Check if the tag in the struct is set in the OS `required`
		if opts.required {
			if err := handleRequired(key); err != nil {
				return fmt.Errorf("handling required option: %v", err)
			}
		}

		// Populate the nested prefix struct (All nested structs must use `prefix`).
		if opts.prefix {
			if err := handlePrefix(value, i, field, key); err != nil {
				return fmt.Errorf("handling prefix option: %v", err)
			}
			continue
		}

		envValue := os.Getenv(key)
		if envValue == "" {
			continue
		}

		if err := setInStruct(envValue, value.Field(i)); err != nil {
			return fmt.Errorf("setting value in config struct: %v", err)
		}

	}

	return nil
}

func setInStruct(envValue string, value reflect.Value) error {
	switch value.Kind() {
	case reflect.Int:
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
			return ErrRequiredFulfilled
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
		return ErrRequiredFulfilled
	}

	return nil
}

// keyAndOptions separates the key within the `env` struct tag from the options
func keyAndOptions(tag string) (string, options) {
	parts := strings.Split(tag, ",")

	key, tagOpts := strings.TrimSpace(parts[0]), parts[1:]

	var opts options

	for _, o := range tagOpts {
		switch {
		case o == OptPrefix:
			opts.prefix = true
		case o == OptRequired:
			opts.required = true
		}
	}

	return key, opts
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
