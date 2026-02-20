package envconfig_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/h-dav/envconfig/v3"
)

func TestErrorIs(t *testing.T) {
	tests := []struct {
		err    error
		target error
		want   bool
	}{
		{&envconfig.FileTypeValidationError{Filepath: "test"}, envconfig.ErrFile, true},
		{&envconfig.OpenFileError{Err: errors.New("open")}, envconfig.ErrFile, true},
		{&envconfig.FieldConversionError{FieldName: "test", Err: errors.New("conv")}, envconfig.ErrConversion, true},
		{&envconfig.UnsupportedFieldTypeError{FieldType: "test"}, envconfig.ErrUnsupported, true},
		{&envconfig.InvalidConfigTypeError{ProvidedType: "test"}, envconfig.ErrInvalidConfig, true},
		{&envconfig.RequiredFieldError{FieldName: "test"}, envconfig.ErrRequired, true},
		{&envconfig.ReplacementError{VariableName: "test"}, envconfig.ErrReplacement, true},
		{&envconfig.ParseError{Line: "test", Err: errors.New("syntax")}, envconfig.ErrParse, true},
		{&envconfig.FileReadError{Filepath: "test", Err: errors.New("read")}, envconfig.ErrFile, true},
		{&envconfig.JSONUnmarshalError{FieldName: "test", Err: errors.New("json")}, envconfig.ErrJSON, true},
		{&envconfig.MalformedTagError{Tag: "test"}, envconfig.ErrTag, true},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.want {
				t.Errorf("errors.Is(%T, %v) = %v, want %v", tt.err, tt.target, got, tt.want)
			}
		})
	}
}

func TestErrorAs(t *testing.T) {
	err := &envconfig.JSONUnmarshalError{FieldName: "TestField", RawValue: "{}", Err: errors.New("inner")}
	var target *envconfig.JSONUnmarshalError
	if !errors.As(err, &target) {
		t.Fatal("errors.As failed for JSONUnmarshalError")
	}
	if target.FieldName != "TestField" {
		t.Errorf("expected FieldName TestField, got %s", target.FieldName)
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []error{
		&envconfig.FileTypeValidationError{Filepath: "test.txt"},
		&envconfig.OpenFileError{Err: errors.New("perm")},
		&envconfig.SetEnvironmentVariableError{Err: errors.New("setenv")},
		&envconfig.FieldConversionError{FieldName: "Field", TargetType: "int", Err: errors.New("bad")},
		&envconfig.UnsupportedFieldTypeError{FieldType: 123},
		&envconfig.InvalidConfigTypeError{ProvidedType: "string"},
		&envconfig.RequiredFieldError{FieldName: "Secret"},
		&envconfig.ReplacementError{VariableName: "VAR"},
		&envconfig.ParseError{Line: "KEY=VAL", Err: errors.New("syntax")},
		&envconfig.FileReadError{Filepath: "env", Err: errors.New("io")},
		&envconfig.JSONUnmarshalError{FieldName: "Data", RawValue: "{", Err: errors.New("eof")},
		&envconfig.MalformedTagError{Tag: "env:required,bad", Err: errors.New("unknown")},
		&envconfig.FieldError{FieldName: "F", Op: "op", Err: errors.New("fail")},
		&envconfig.FieldError{FieldName: "", Op: "global", Err: errors.New("fail")},
	}

	for _, err := range tests {
		if msg := err.Error(); msg == "" {
			t.Errorf("empty error message for %T", err)
		}
	}
}
