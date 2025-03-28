package envconfig

import (
	"errors"
	"fmt"
)

// FileTypeValidationError occurs when the .env config file fails to open.
type FileTypeValidationError struct {
	Filename string
}

// Error satisfies the error interface for FileTypeValidationError.
func (e *FileTypeValidationError) Error() string {
	return fmt.Sprintf("file extension is not a valid environment file: %q", e.Filename)
}

// OpenFileError occurs when the .env config file fails to open.
type OpenFileError struct {
	Err error
}

// Error satisfies the error interface for OpenFileError.
func (e *OpenFileError) Error() string {
	return fmt.Sprintf("failed to open config file: %v", e.Err)
}

// Unwrap allows OpenFileError to be used with errors.Is and errors.As.
func (e *OpenFileError) Unwrap() error { return e.Err }

// SetEnvironmentVariableError occurs when the value is failed to be set in the environment.
type SetEnvironmentVariableError struct {
	Err error
}

// Error satisfies the error interface for SetEnvironmentVariableError.
func (e *SetEnvironmentVariableError) Error() string {
	return fmt.Sprintf("failed to set environment variable: %v", e.Err.Error())
}

// Unwrap allows SetEnvironmentVariableError to be used with errors.Is and errors.As.
func (e *SetEnvironmentVariableError) Unwrap() error { return e.Err }

// FieldConversionError occurs when a field on the config struct fails to be set.
type FieldConversionError struct {
	FieldName  string
	TargetType string
	Err        error
}

// Error satisfies the error interface for FieldConversionError.
func (e *FieldConversionError) Error() string {
	return fmt.Sprintf("failed to convert field %v to %v: %v", e.FieldName, e.TargetType, e.Err.Error())
}

// Unwrap allows FieldConversionError to be used with errors.Is and errors.As.
func (e *FieldConversionError) Unwrap() error { return e.Err }

// UnsupportedFieldTypeError occurs when the a field type on the config struct is not compatible.
type UnsupportedFieldTypeError struct {
	FieldType any
}

// Error satisfies the error interface for UnsupportedFieldTypeError.
func (e *UnsupportedFieldTypeError) Error() string {
	return fmt.Sprintf("unsupported field type: %T", e.FieldType)
}

// InvalidConfigTypeError occurs when config is not a pointer to a struct.
type InvalidConfigTypeError struct {
	ProvidedType any
}

// Error satisfies the error interface for InvalidConfigTypeError.
func (e *InvalidConfigTypeError) Error() string {
	return fmt.Sprintf("output must be a pointer to a struct, got %T", e.ProvidedType)
}

// RequiredFieldError occurs when a required field is not set and in the environment variables.
type RequiredFieldError struct {
	FieldName string
}

// Error satisfies the error interface for RequiredFieldError.
func (e *RequiredFieldError) Error() string {
	return fmt.Sprintf("required field is not set in environment variables: %v", e.FieldName)
}

// InvalidOptionConversionError occurs when an option is invalid for a field.
type InvalidOptionConversionError struct {
	FieldName string
	Option    string
	Err       error
}

// Error satisfies the error interface for InvalidOptionConversionError.
func (e *InvalidOptionConversionError) Error() string {
	return fmt.Sprintf("invalid option %v conversion for field %v: %v", e.Option, e.FieldName, e.Err.Error())
}

// Unwrap allows InvalidOptionConversionError to be used with errors.Is and errors.As.
func (e *InvalidOptionConversionError) Unwrap() error { return e.Err }

// PrefixOptionError occurs when the prefix tag is invalid or not set on a nested struct.
type PrefixOptionError struct {
	FieldName any
}

// Error satisfies the error interface for PrefixOptionError.
func (e *PrefixOptionError) Error() string {
	return fmt.Sprintf("prefix option is not set for nested struct field: %v", e.FieldName)
}

// ReplacementError occurs when the environment variable being used for replacement is not set.
type ReplacementError struct {
	VariableName string
}

// Error satisfies the error interface for ReplacementError.
func (e *ReplacementError) Error() string {
	return fmt.Sprintf("environment variable for replacement is not set: %v", e.VariableName)
}

// ParseError occurs when a line from the .env config file has been parsed incorrectly.
type ParseError struct {
	Line string
	Err  error
}

// ErrSyntax indicates that a line is invalid syntax.
var ErrSyntax = errors.New("invalid syntax")

// Error statisfies the error interface for ParseError.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse line: %v: %v", e.Line, e.Err.Error())
}

// FileReadError occurs when an error occurs when scanning the .env file.
type FileReadError struct {
	Filename string
	Err      error
}

// Error satisfies the error interface for FileReadError.
func (e *FileReadError) Error() string {
	return fmt.Sprintf("reading %v: %v", e.Filename, e.Err.Error())
}

// Unwrap allows FileReadError to be used with errors.Is and errors.As.
func (e *FileReadError) Unwrap() error { return e.Err }
