package envconfig

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	// tagEnv is used for fetching the environment variable by name.
	tagEnv = "env"
)

// TagMetadata contains the parsed information from a struct tag.
type TagMetadata struct {
	Name     string
	Required bool
	Default  string
	Prefix   string
	EnvJSON  bool
}

// ParseTag parses a struct tag into TagMetadata.
func ParseTag(tag string) (TagMetadata, error) {
	if tag == "" {
		return TagMetadata{}, nil
	}

	parts := strings.Split(tag, ",")
	metadata := TagMetadata{
		Name: strings.TrimSpace(parts[0]),
	}

	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		key, value, found := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if found && key == "" {
			return TagMetadata{}, fmt.Errorf("malformed tag option: %s", part)
		}

		switch key {
		case "required":
			metadata.Required = true
		case "envjson":
			metadata.EnvJSON = true
		case "default":
			metadata.Default = value
		case "prefix":
			metadata.Prefix = value
		}
	}

	return metadata, nil
}

// HandleField parses the struct tag and populates the field with the corresponding value.
func (s *settings) HandleField(field reflect.StructField, value reflect.Value, prefix string) error {
	tag := field.Tag.Get(tagEnv)
	metadata, err := ParseTag(tag)
	if err != nil {
		return err
	}

	// Handle Prefix (nested struct)
	if field.Type.Kind() == reflect.Struct {
		if metadata.Prefix != "" {
			newPrefix := prefix + metadata.Prefix
			if err := s.populateNestedConfig(value, newPrefix); err != nil {
				return fmt.Errorf("handle nested struct for field '%s': %w", field.Name, err)
			}
			// Stop if we handle a prefix tag.
			return nil
		}
	}

	// Handle EnvJSON
	if metadata.EnvJSON {
		key := prefix + metadata.Name
		if jsonString, exists := s.source[key]; exists {
			if value.CanAddr() {
				if err := json.Unmarshal([]byte(jsonString), value.Addr().Interface()); err != nil {
					return &JSONUnmarshalError{
						FieldName: field.Name,
						RawValue:  jsonString,
						Err:       err,
					}
				}
			}
		}
	} else if metadata.Name != "" {
		// Handle Name (standard env var)
		key := prefix + metadata.Name
		if val, exists := s.source[key]; exists {
			resolvedValue, err := s.resolveReplacement(val)
			if err != nil {
				return err
			}
			if err := s.setFieldValue(value, entry{key: key, value: resolvedValue}); err != nil {
				return fmt.Errorf("set value for field '%s': %w", field.Name, err)
			}
		}
	}

	// Handle Default
	if value.IsZero() && metadata.Default != "" {
		if err := s.setFieldValue(value, entry{field.Name, metadata.Default}); err != nil {
			return fmt.Errorf("set default value for field '%s': %w", field.Name, err)
		}
	}

	// Handle Required
	if metadata.Required && value.IsZero() {
		return &RequiredFieldError{
			FieldName: field.Name,
		}
	}

	return nil
}
