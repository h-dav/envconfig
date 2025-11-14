package envconfig

import (
	"maps"
	"reflect"
)

type settings struct {
	prefix          string
	source          map[string]string
	temporaryPrefix string // temporary prefix is only used we are populating nested structs
	sources         []source
	decoders        map[reflect.Type]DecoderFunc
}

type option func(*settings)

// WithFilepath option will cause the file provided to be used to set variables in the environment.
func WithFilepath(filepath string) option {
	return func(s *settings) {
		s.sources = append(s.sources, FileSource{
			filepath: filepath,
		})
	}
}

func WithActiveProfile(filepath, activeProfile string) option {
	return func(s *settings) {
		if activeProfile == "" {
			activeProfile = "default"
		}
		s.sources = append(s.sources, FileSource{
			filepath: filepath + activeProfile + envExtension,
		})
	}
}

// WithPrefix option will add the prefix to before every set and retrieval from env.
func WithPrefix(prefix string) option {
	return func(s *settings) {
		s.prefix = prefix
	}
}

func WithDecoders(decoders map[reflect.Type]DecoderFunc) option {
	return func(s *settings) {
		if s.decoders == nil {
			s.decoders = make(map[reflect.Type]DecoderFunc)
		}
		maps.Copy(s.decoders, decoders)
	}
}
