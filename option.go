package envconfig

type settings struct {
	filepath        string
	activeProfile   string
	prefix          string
	source          map[string]string
	temporaryPrefix string // temporary prefix is only used we are populating nested structs
}

type option func(*settings)

// WithFilepath option will cause the file provided to be used to set variables in the environment.
func WithFilepath(filepath string) option {
	return func(s *settings) {
		s.filepath = filepath
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

// WithPrefix option will add the prefix to before every set and retrieval to and from env.
func WithPrefix(prefix string) option {
	return func(s *settings) {
		s.prefix = prefix
	}
}

