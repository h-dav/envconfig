package envconfig

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

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
