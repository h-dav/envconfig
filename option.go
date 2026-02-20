package envconfig

import (
	"maps"
	"reflect"
)

type settings struct {
	prefix   string
	source   map[string]string
	sources  []source
	decoders map[reflect.Type]DecoderFunc
}

// Option is a functional option for configuring the Set function.
type Option func(*settings)

// WithFilepath causes the file at the provided path to be loaded into the environment variables.
func WithFilepath(filepath string) Option {
	return func(s *settings) {
		s.sources = append(s.sources, FileSource{
			filepath: filepath,
		})
	}
}

// WithActiveProfile loads a profile-specific environment file.
// It constructs the filename as: path + activeProfile + ".env".
// If activeProfile is empty, it defaults to "default".
func WithActiveProfile(path, activeProfile string) Option {
	return func(s *settings) {
		if activeProfile == "" {
			activeProfile = "default"
		}
		s.sources = append(s.sources, FileSource{
			filepath: path + activeProfile + envExtension,
		})
	}
}

// WithPrefix adds a prefix to all environment variable lookups.
func WithPrefix(prefix string) Option {
	return func(s *settings) {
		s.prefix = prefix
	}
}

// WithDecoders registers custom decoders for specific types.
func WithDecoders(decoders map[reflect.Type]DecoderFunc) Option {
	return func(s *settings) {
		if s.decoders == nil {
			s.decoders = make(map[reflect.Type]DecoderFunc)
		}
		maps.Copy(s.decoders, decoders)
	}
}
