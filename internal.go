package envconfig

import (
	"encoding/json"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
)

type settings struct {
	prefix   string
	source   map[string]string
	sources  []source
	decoders map[reflect.Type]DecoderFunc
	logger   *slog.Logger
}

type entry struct {
	key, value string
}

// textReplacementRegex is used to detect text replacement in environment variables.
var textReplacementRegex = regexp.MustCompile(`\${[^}]+}`)

// populateStruct uses the items in settings.source to populate the passed in config struct.
func (s *settings) populateStruct(config any) error {
	configStruct := reflect.ValueOf(config)
	if configStruct.Kind() != reflect.Pointer || configStruct.Elem().Kind() != reflect.Struct {
		return &InvalidConfigTypeError{ProvidedType: config}
	}

	configValue := reflect.ValueOf(config).Elem()

	for i := range configValue.NumField() {
		field := configValue.Type().Field(i)
		configFieldValue := configValue.Field(i)

		// Ignore fields that are not exported.
		if !configFieldValue.CanSet() {
			continue
		}

		if err := s.HandleField(field, configFieldValue, s.prefix); err != nil {
			return &FieldError{FieldName: field.Name, Op: "process field", Err: err}
		}
	}

	return nil
}

// resolveReplacement checks if a string has the pattern of ${...}, and if so, uses values in settings.source to
// replace the pattern, and returns the newly created string.
func (s *settings) resolveReplacement(value string) (string, error) {
	match := textReplacementRegex.FindStringSubmatch(value)

	for _, m := range match {
		environmentValue := strings.TrimPrefix(m, "${")
		environmentValue = strings.TrimSuffix(environmentValue, "}")

		replacementValue := s.source[environmentValue]
		if replacementValue == "" {
			return "", &ReplacementError{VariableName: environmentValue}
		}

		value = strings.ReplaceAll(value, m, replacementValue)
	}

	return value, nil
}

// populateNestedConfig populates a nested struct.
func (s *settings) populateNestedConfig(nestedConfig reflect.Value, prefix string) error {
	for i := range nestedConfig.NumField() {
		field := nestedConfig.Type().Field(i)
		configFieldValue := nestedConfig.Field(i)

		if !configFieldValue.CanSet() {
			continue
		}

		// Process the field with the handler.
		if err := s.HandleField(field, configFieldValue, prefix); err != nil {
			return &FieldError{FieldName: field.Name, Op: "error processing field", Err: err}
		}
	}

	return nil
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
			if s.logger != nil {
				s.logger.Debug("entering nested struct", slog.String("field", field.Name), slog.String("prefix", newPrefix))
			}
			if err := s.populateNestedConfig(value, newPrefix); err != nil {
				return &FieldError{FieldName: field.Name, Op: "handle nested struct", Err: err}
			}
			// Stop if we handle a prefix tag.
			return nil
		}
	}

	// Handle EnvJSON
	if metadata.EnvJSON {
		key := prefix + metadata.Name
		if jsonString, exists := s.source[key]; exists {
			if s.logger != nil {
				s.logger.Debug("unmarshaling JSON field", slog.String("field", field.Name), slog.String("key", key))
			}
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
			if s.logger != nil {
				s.logger.Debug("setting field value", slog.String("field", field.Name), slog.String("key", key))
			}
			if err := s.setFieldValue(value, entry{key: key, value: resolvedValue}); err != nil {
				return &FieldError{FieldName: field.Name, Op: "set value", Err: err}
			}
		}
	}

	// Handle Default
	if value.IsZero() && metadata.Default != "" {
		if s.logger != nil {
			s.logger.Debug("setting default value", slog.String("field", field.Name), slog.String("default", metadata.Default))
		}
		if err := s.setFieldValue(value, entry{field.Name, metadata.Default}); err != nil {
			return &FieldError{FieldName: field.Name, Op: "set default value", Err: err}
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
