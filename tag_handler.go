package envconfig

import (
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
			return TagMetadata{}, &MalformedTagError{Tag: tag}
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
