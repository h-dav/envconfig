// Package envconfig parses a .env file and loads the values into the os and returns a populated config structure.
package envconfig

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

type pair struct {
	key   string
	value string
}

// options are included in the tag.
type options struct {
	prefix   bool
	required bool
}

const (
	optPrefix   = "prefix"
	optRequired = "required"
)

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

		key, opts, err := keyAndOptions(tag)
		if err != nil {
			return fmt.Errorf("processing key and options: %v", err)
		}

		// Check if the tag in the struct is set in the OS `required`
		if opts.required {
			if err := handleRequired(key); err != nil {
				return fmt.Errorf("handling required option: %v", err)
			}
			continue
		}

		// Populate the nested prefix struct (All nested structs must use `prefix`).
		if opts.prefix {
			if err := handlePrefix(value, i, field, key); err != nil {
				return fmt.Errorf("handling prefix option: %v", err)
			}
			continue
		}

		envValue := os.Getenv(tag)
		if envValue == "" {
			continue
		}

		value.Field(i).SetString(envValue)
	}

	return nil
}

func handlePrefix(value reflect.Value, i int, field reflect.StructField, key string) error {
	nestedStruct := value.Field(i)
	nestedType := field.Type

	for ni := 0; ni < nestedStruct.NumField(); ni++ {
		nestedField := nestedType.Field(ni)
		nestedTag := nestedField.Tag.Get("env")

		nestedKey, nestedOpts, err := keyAndOptions(key + nestedTag)
		if err != nil {
			return fmt.Errorf("processing key and option: %v", err)
		}

		if nestedOpts.required {
			if err := handleRequired(nestedKey); err != nil {
				return fmt.Errorf("checking for required value: %v", err)
			}
		}

		envValue := os.Getenv(nestedKey)
		if envValue == "" {
			continue
		}

		value.Field(i).Field(ni).SetString(envValue)
	}
	return nil
}

func handleRequired(key string) error {
	envValue := os.Getenv(key)
	if envValue == "" {
		return fmt.Errorf("value does not exist for struct tag: %v", key)
	}
	return nil
}

func keyAndOptions(tag string) (string, options, error) {
	parts := strings.Split(tag, ",")

	key, tagOpts := strings.TrimSpace(parts[0]), parts[1:]

	var opts options

	for _, o := range tagOpts {
		switch {
		case o == optPrefix:
			opts.prefix = true
		case o == optRequired:
			opts.required = true
		}
	}

	return key, opts, nil
}
