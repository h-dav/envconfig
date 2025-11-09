package envconfig

import "reflect"

type settings struct {
	filepath        string
	activeProfile   string
	prefix          string
	source          map[string]string
	temporaryPrefix string // temporary prefix is only used we are populating nested structs
	sources []source
	decoders map[reflect.Type]DecoderFunc
}

type option func(*settings)

// WithFilepath option will cause the file provided to be used to set variables in the environment.
func WithFilepath(filepath string) option {
	return func(s *settings) {
		s.filepath = filepath
		s.sources = append(s.sources, FileSource{
			filepath: filepath,
		})
	}
}

func WithActiveProfile(activeProfile string) option {
	return func(s *settings) {
		if activeProfile == "" {
			activeProfile = "default"
		}
		s.activeProfile = activeProfile
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
		for typ, dec := range decoders {
			s.decoders[typ] = dec
		}
	}
}
