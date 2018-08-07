package docker

import "github.com/strabox/caravela/configuration"

type partitionAwareResourceGen struct {
	caravelaConfigs *configuration.Configuration
}

func newPartitionAwareResourceGen(caravelaConfigs *configuration.Configuration) (ResourcesGenerator, error) {
	return &partitionAwareResourceGen{
		caravelaConfigs: caravelaConfigs,
	}, nil
}

func (p *partitionAwareResourceGen) Generate() (int, int) {
	/*
		coresPartition := p.caravelaConfigs.CPUCoresPartitions()
		for _, partition := range coresPartition {
			partition.
		}
		// TODO
	*/
	return 1, 256
}
