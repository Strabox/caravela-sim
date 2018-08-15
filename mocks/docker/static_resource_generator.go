package docker

import (
	"github.com/strabox/caravela-sim/configuration"
	caravelaConfigs "github.com/strabox/caravela/configuration"
)

type staticResourceGen struct {
	simConfigs *configuration.Configuration
}

func newStaticResourceGen(simConfigs *configuration.Configuration, _ *caravelaConfigs.Configuration) (ResourcesGenerator, error) {
	return &staticResourceGen{
		simConfigs: simConfigs,
	}, nil
}

func (s *staticResourceGen) Generate() (int, int) {
	return s.simConfigs.StaticGeneratorResources().CPUs, s.simConfigs.StaticGeneratorResources().RAM
}
