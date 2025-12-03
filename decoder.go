package envconfig

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// DecoderFunc is a function that converts a string value to a specific type.
type DecoderFunc func(key, value string) (reflect.Value, error)

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

// Setter is an interface that can be implemented by types that want to self-configure.
type Setter interface {
	Set(value string) error
}

// setFieldValue sets the value of a field.
func (s *settings) setFieldValue(fieldValue reflect.Value, key, value string) error {
	fieldAddr := fieldValue.Addr()

	if setter, ok := fieldAddr.Interface().(Setter); ok {
		return setter.Set(value)
	}

	if dec, ok := s.decoders[fieldValue.Type()]; ok {
		decodedValue, err := dec(key, value)
		if err != nil {
			return err
		}
		fieldValue.Set(decodedValue)
		return nil
	}

	switch fieldValue.Interface().(type) {
	case string:
		fieldValue.SetString(value)
	case []string:
		return setStringSliceFieldValue(fieldValue, value)
	case []int:
		return setIntSliceFieldValue(fieldValue, key, value)
	case []float64:
		return setFloatSliceFieldValue(fieldValue, key, value)
	default:
		return &UnsupportedFieldTypeError{FieldType: fieldValue.Interface()}
	}

	return nil
}

func setStringSliceFieldValue(fieldValue reflect.Value, value string) error {
	values := strings.Split(value, ",")
	slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

	for i, v := range values {
		v = strings.TrimSpace(v)
		slice.Index(i).SetString(v)
	}

	fieldValue.Set(slice)
	return nil
}

func setIntSliceFieldValue(fieldValue reflect.Value, key, value string) error {
	values := strings.Split(value, ",")
	slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

	for i, v := range values {
		v = strings.TrimSpace(v)

		parsed, err := strconv.Atoi(v)
		if err != nil {
			return &FieldConversionError{
				FieldName:  key,
				TargetType: "[]int",
				Err:        err,
			}
		}

		slice.Index(i).SetInt(int64(parsed))
	}

	fieldValue.Set(slice)
	return nil
}

func setFloatSliceFieldValue(fieldValue reflect.Value, key, value string) error {
	values := strings.Split(value, ",")
	slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

	for i, v := range values {
		v = strings.TrimSpace(v)

		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return &FieldConversionError{
				FieldName:  key,
				TargetType: "[]float64",
				Err:        err,
			}
		}

		slice.Index(i).SetFloat(parsed)
	}

	fieldValue.Set(slice)
	return nil
}
