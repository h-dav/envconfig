package envconfig

import (
	"bufio"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
)

type pair struct {
	key   string
	value string
}

//func SetPopulate(filename string, cfg {}interface) error {
//	if err := setVars(filename); err != nil {
//		return err
//	}
//
//	if err := populate(cfg); err != nil {
//		return err
//	}
//}

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
				break // Reached end of file
			} else {
				return err // Handle other errors
			}
		}

		pair := entryConvert(strings.Replace(line, "\n", "", -1))

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

// Populate will map values from the os into the struct passed in.
func Populate(cfg interface{}) error {
	return populate(cfg)
}

func populate(cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("input must be a non-nil pointer to a struct")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("input must be a pointer to a struct")
	}

	vType := v.Type()
	for i := 0; i < vType.NumField(); i++ {
		field := vType.Field(i)
		value := os.Getenv(field.Tag.Get("env"))
		if value == "" {
			continue
		}
		v.Field(i).SetString(value)
	}

	return nil
}
