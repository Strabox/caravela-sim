package docker

import (
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/configuration"
)

type partitionAwareResourceGen struct {
	caravelaConfigs *configuration.Configuration
}

func newPartitionAwareResourceGen(caravelaConfigs *configuration.Configuration) (ResourcesGenerator, error) {
	return &partitionAwareResourceGen{
		caravelaConfigs: caravelaConfigs,
	}, nil
}

func (p *partitionAwareResourceGen) Generate() (int, int) {
	coresPartitions := make([]configuration.CPUCoresPartition, len(p.caravelaConfigs.CPUCoresPartitions()))
	ramPartitions := make([]configuration.RAMPartition, len(p.caravelaConfigs.RAMPartitions()))
	copy(coresPartitions, p.caravelaConfigs.CPUCoresPartitions())
	copy(ramPartitions, p.caravelaConfigs.RAMPartitions())

	acc := 0
	for i := range coresPartitions {
		currentPercentage := coresPartitions[i].Percentage
		coresPartitions[i].Percentage += acc
		acc += currentPercentage
	}

	acc = 0
	for i := range ramPartitions {
		currentPercentage := ramPartitions[i].Percentage
		ramPartitions[i].Percentage += acc
		acc += currentPercentage
	}

	randInt := util.RandomInteger(1, 100)
	prevCores := coresPartitions[0].Cores
	for i := range coresPartitions {
		if randInt <= coresPartitions[i].Percentage {
			prevCores = coresPartitions[i].Cores
			break
		}
	}

	randInt = util.RandomInteger(1, 100)
	prevRAM := ramPartitions[0].RAM
	for i := range ramPartitions {
		if randInt <= ramPartitions[i].Percentage {
			prevRAM = ramPartitions[i].RAM
			break
		}
	}

	util.Log.Debugf(util.LogTag("ResGen")+"MaxRes: <%d;%d>", prevCores, prevRAM)
	return prevCores, prevRAM
}
