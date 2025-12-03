package envconfig

import "reflect"

type settings struct {
	filepaths     []string
	activeProfile string
	prefix        string
	source        map[string]string
	sources       []Source
	decoders      map[reflect.Type]DecoderFunc
}

type option func(*settings)

// WithFilepath adds a file to be used for configuration loading.
func WithFilepath(filepath string) option {
	return func(s *settings) {
		s.filepaths = append(s.filepaths, filepath)
	}
}

// WithActiveProfile sets the active profile.
func WithActiveProfile(activeProfile string) option {
	return func(s *settings) {
		if activeProfile == "" {
			activeProfile = "default"
		}
		s.activeProfile = activeProfile
	}
}

// WithPrefix sets a global prefix for environment variables.
func WithPrefix(prefix string) option {
	return func(s *settings) {
		s.prefix = prefix
	}
}

// WithDecoders adds custom decoders for specific types.
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

// WithSource adds a custom source to the configuration loader.
func WithSource(source Source) option {
	return func(s *settings) {
		s.sources = append(s.sources, source)
	}
}
