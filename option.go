package envconfig

type settings struct {
	filename string
	prefix   string
}

type Option func(*settings)

// WithFilename option will cause the file provided to be used to set variables in the environment.
func WithFilename(filename string) Option {
	return func(s *settings) {
		s.filename = filename
	}
}

// WithPrefix option will add the prefix to before every set and retrieval to and from env.
func WithPrefix(prefix string) Option {
	return func(s *settings) {
		s.prefix = prefix
	}
}

