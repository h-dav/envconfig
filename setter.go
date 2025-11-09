package envconfig

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

var defaultDecoders = map[reflect.Type]DecoderFunc{
	reflect.TypeOf(time.Duration(0)): func(key, value string) (reflect.Value, error) {
		durationValue, err := time.ParseDuration(value)
		if err != nil {
			return reflect.Value{}, &FieldConversionError{
				FieldName:  key,
				TargetType: "time.Duration",
				Err:        err,
			}
		}

		return reflect.ValueOf(durationValue), nil
	},
	reflect.TypeOf(int(0)): func(key, value string) (reflect.Value, error) {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, &FieldConversionError{
				FieldName:  key,
				TargetType: "int",
				Err:        err,
			}
		}

		return reflect.ValueOf(intValue), nil
	},
	reflect.TypeOf(true): func(key, value string) (reflect.Value, error) {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, &FieldConversionError{
				FieldName:  key,
				TargetType: "bool",
				Err:        err,
			}
		}

		return reflect.ValueOf(boolValue), nil
	},
	reflect.TypeOf(float64(0)): func(key, value string) (reflect.Value, error) {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, &FieldConversionError{
				FieldName:  key,
				TargetType: "float",
				Err:        err,
			}
		}

		return reflect.ValueOf(floatValue), nil
	},
}

type Setter interface {
	Set(value string) error
}

// setFieldValue determines the type of a config field, and branch out to the correct
// function to populate that data type.
func (s settings) setFieldValue(
	configFieldValue reflect.Value,
	entry entry,
) error {
	fieldAddr := configFieldValue.Addr()

	if setter, ok := fieldAddr.Interface().(Setter); ok {
		return setter.Set(entry.value)
	}

	if dec, ok := s.decoders[configFieldValue.Type()]; ok {
		decodedValue, err := dec(entry.key, entry.value)
		if err != nil {
			return err
		}
		configFieldValue.Set(decodedValue)
		return nil
	}

	switch configFieldValue.Interface().(type) {
	case string:
		configFieldValue.SetString(entry.value)
	case []string:
		return setStringSliceFieldValue(configFieldValue, entry.value)
	case []int:
		return setIntSliceFieldValue(configFieldValue, entry)
	case []float64:
		return setFloatSliceFieldValue(configFieldValue, entry)
	default:
		return &UnsupportedFieldTypeError{FieldType: configFieldValue.Interface()}
	}

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
