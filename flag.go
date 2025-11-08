package envconfig

import (
	"flag"
)

func (s *settings) processFlags() error {
	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		s.source[f.Name] = f.Value.String()
	})

	return nil
}
