package docker

import "github.com/strabox/caravela/configuration"

type staticResourceGen struct {
	// Nothing necessary
}

func newStaticResourceGen(_ *configuration.Configuration) (ResourcesGenerator, error) {
	return &staticResourceGen{}, nil
}

func (s *staticResourceGen) Generate() (int, int) {
	return 2, 2048
}
