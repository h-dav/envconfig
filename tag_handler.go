package envconfig

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	// tagEnv is used for fetching the environment variable by name.
	tagEnv = "env"

	// tagDefault is used to set a fallback value for a config field if the environment variable is not set.
	tagDefault = "default"

	// tagRequired is used for config struct fields that are required. If the environment variable is not set, an
	// error will be returned.
	tagRequired = "required"

	// tagJSON is used for environment variables that are JSON.
	tagJSON = "envjson"

	// tagPrefix is used for nested structs inside your config struct.
	tagPrefix = "prefix"
)

var chain = &PrefixTagHandler{
	BaseHandler: BaseHandler{
		next: &JSONTagHandler{
			BaseHandler: BaseHandler{
				next: &EnvTagHandler{
					BaseHandler: BaseHandler{
						next: &DefaultTagHandler{
							BaseHandler: BaseHandler{
								next: &RequiredTagHandler{},
							},
						},
					},
				},
			},
		},
	},
}

type TagHandler interface {
	SetNext(handler TagHandler) TagHandler
	Handle(field reflect.StructField, value reflect.Value, s *settings, prefix string) error
}

type BaseHandler struct {
	next TagHandler
}

func (h *BaseHandler) SetNext(handler TagHandler) TagHandler {
	h.next = handler
	return handler
}

func (h *BaseHandler) Handle(field reflect.StructField, value reflect.Value, s *settings, prefix string) error {
	if h.next != nil {
		return h.next.Handle(field, value, s, prefix)
	}
	return nil
}

type PrefixTagHandler struct {
	BaseHandler
}

func (h *PrefixTagHandler) Handle(field reflect.StructField, value reflect.Value, s *settings, prefix string) error {
	if field.Type.Kind() == reflect.Struct {
		if prefixTagValue, ok := field.Tag.Lookup(tagPrefix); ok {
			newPrefix := prefix + prefixTagValue
			if err := s.populateNestedConfig(value, newPrefix); err != nil {
				return fmt.Errorf("handle nested struct for field '%s': %w", field.Name, err)
			}
			// Stop the chain if we handle a prefix tag.
			return nil
		}
	}
	return h.BaseHandler.Handle(field, value, s, prefix)
}

type DefaultTagHandler struct {
	BaseHandler
}

func (h *DefaultTagHandler) Handle(field reflect.StructField, value reflect.Value, s *settings, prefix string) error {
	if defaultVal, ok := field.Tag.Lookup(tagDefault); ok {
		// Only set the default value if the field is still zero after other handlers have run.
		if value.IsZero() {
			if err := s.setFieldValue(value, entry{field.Name, defaultVal}); err != nil {
				return fmt.Errorf("set default value for field '%s': %w", field.Name, err)
			}
		}
	}
	return nil
}

type EnvTagHandler struct {
	BaseHandler
}

func (h *EnvTagHandler) Handle(field reflect.StructField, value reflect.Value, s *settings, prefix string) error {
	if envVar, ok := field.Tag.Lookup(tagEnv); ok {
		key := prefix + envVar
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
	return h.BaseHandler.Handle(field, value, s, prefix)
}

type JSONTagHandler struct {
	BaseHandler
}

func (h *JSONTagHandler) Handle(field reflect.StructField, value reflect.Value, s *settings, prefix string) error {
	if jsonVar, ok := field.Tag.Lookup(tagJSON); ok {
		key := prefix + jsonVar
		if jsonString, exists := s.source[key]; exists {
			if value.CanAddr() {
				if err := json.Unmarshal([]byte(jsonString), value.Addr().Interface()); err != nil {
					return fmt.Errorf("failed to unmarshal JSON for field '%s': %w", field.Name, err)
				}
			}
		}
	}
	return h.BaseHandler.Handle(field, value, s, prefix)
}

type RequiredTagHandler struct {
	BaseHandler
}

func (h *RequiredTagHandler) Handle(field reflect.StructField, value reflect.Value, s *settings, prefix string) error {
	if required, ok := field.Tag.Lookup(tagRequired); ok && (required == "true" || required == "") {
		if value.IsZero() {
			return &RequiredFieldError{
				FieldName: field.Name,
			}
		}
	}
	return nil
}
