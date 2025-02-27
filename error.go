package envconfig

import (
	"fmt"
)

// OpenFileError is when the .env config file fails to open.
type OpenFileError struct {
	Err error
}

func (e *OpenFileError) Error() string {
	return fmt.Sprintf("failed to open config file: %v", e.Err)
}

func (e *OpenFileError) Unwrap() error {
	return e.Err
}

// SetEnvironmentVariableError is when the value is failed to be set in the environment.
type SetEnvironmentVariableError struct {
	Err error
}

func (e *SetEnvironmentVariableError) Error() string {
	return fmt.Sprintf("failed to set environment variable: %v", e.Err)
}

func (e *SetEnvironmentVariableError) Unwrap() error {
	return e.Err
}

// ParamConversionError is when a parameter on the config struct fails to be set.
type ParamConversionError struct {
	ParamName  string
	TargetType string
	Err        error
}

func (e *ParamConversionError) Error() string {
	return fmt.Sprintf("failed to convert parameter %q to %s: %v", e.ParamName, e.TargetType, e.Err)
}

func (e *ParamConversionError) Unwrap() error {
	return e.Err
}

// UnsupportedFieldTypeError is when the a field type on the config struct is not compatible.
type UnsupportedFieldTypeError struct {
	FieldType any
}

// Error satisfies the error interface for UnsupportedFieldTypeError.
func (e *UnsupportedFieldTypeError) Error() string {
	return fmt.Sprintf("unsupported field type: %T", e.FieldType)
}

// InvalidConfigTypeError is when config is not a pointer to a struct.
type InvalidConfigTypeError struct {
	ProvidedType any
}

func (e *InvalidConfigTypeError) Error() string {
	return fmt.Sprintf("output must be a pointer to a struct, got %T", e.ProvidedType)
}

type RequiredFieldError struct {
	ParamName string
}

func (e *RequiredFieldError) Error() string {
	return fmt.Sprintf("required field is not set in OS: %q", e.ParamName)
}

type InvalidOptionConversionError struct {
	ParamName string
	Option    string
	Err       error
}

func (e *InvalidOptionConversionError) Error() string {
	return fmt.Sprintf("invalid option %s conversion for param %q: %v", e.Option, e.ParamName, e.Err)
}

func (e *InvalidOptionConversionError) Unwrap() error {
	return e.Err
}

type PrefixOptionError struct {
	ParamName any
}

// Error satisfies the error interface for UnsupportedFieldTypeError.
func (e *PrefixOptionError) Error() string {
	return fmt.Sprintf("prefix option is not set for nested struct field: %q", e.ParamName)
}
