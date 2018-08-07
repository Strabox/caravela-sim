package docker

import "github.com/strabox/caravela/configuration"

type equiPartitionResourceGen struct {
	caravelaConfigs *configuration.Configuration
}

func newEquiPartitionResourceGen(caravelaConfigs *configuration.Configuration) (ResourcesGenerator, error) {
	return &equiPartitionResourceGen{
		caravelaConfigs: caravelaConfigs,
	}, nil
}

func (e *equiPartitionResourceGen) Generate() (int, int) {
	// TODO
	return 1, 256
}
