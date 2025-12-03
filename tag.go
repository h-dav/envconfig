package envconfig

import (
	"reflect"
	"strconv"
)

const (
	tagEnv      = "env"
	tagDefault  = "default"
	tagRequired = "required"
	tagJSON     = "envjson"
	tagPrefix   = "prefix"
)

// checkRequiredTag checks if a field is required and returns an error if so.
func checkRequiredTag(key string, field reflect.StructField) error {
	requiredVal, ok := field.Tag.Lookup(tagRequired)
	if !ok {
		return nil
	}

	required, err := strconv.ParseBool(requiredVal)
	if err != nil {
		return &InvalidOptionConversionError{
			FieldName: key,
			Option:    tagRequired,
			Err:       err,
		}
	}

	if required {
		return &RequiredFieldError{FieldName: key}
	}

	return nil
}
