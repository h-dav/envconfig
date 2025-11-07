package envconfig

type settings struct {
	filename string
	prefix   string
	source map[string]string
}

type option func(*settings)

// WithFilename option will cause the file provided to be used to set variables in the environment.
func WithFilename(filename string) option {
	return func(s *settings) {
		s.filename = filename
	}
}

// WithPrefix option will add the prefix to before every set and retrieval to and from env.
func WithPrefix(prefix string) option {
	return func(s *settings) {
		s.prefix = prefix
	}
}

