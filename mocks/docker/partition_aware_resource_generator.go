package docker

import (
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/configuration"
	caravelaConfigs "github.com/strabox/caravela/configuration"
	caravelaUtil "github.com/strabox/caravela/util"
	"math/rand"
)

// partitionAwareResourceGen is a resource generator for the maximum resources of a node.
type partitionAwareResourceGen struct {
	randomGenerator *rand.Rand                     // Pseudo-random generator.
	caravelaConfigs *caravelaConfigs.Configuration // Caravela's configurations.
}

// newPartitionAwareResourceGen creates a new partition aware maximum resources generator.
func newPartitionAwareResourceGen(_ *configuration.Configuration, caravelaConfigs *caravelaConfigs.Configuration, rngSeed int64) (ResourcesGenerator, error) {
	return &partitionAwareResourceGen{
		randomGenerator: rand.New(caravelaUtil.NewSourceSafe(rand.NewSource(rngSeed))),
		caravelaConfigs: caravelaConfigs,
	}, nil
}

func (p *partitionAwareResourceGen) Generate() (int, int, int) {
	resourcesPartitions := p.caravelaConfigs.ResourcesPartitions()

	cpAcc := 0
	for cp, cpuClass := range resourcesPartitions.CPUClasses {
		currentCPPercentage := cpuClass.Percentage
		resourcesPartitions.CPUClasses[cp].Percentage += cpAcc
		cpAcc += currentCPPercentage

		ccAcc := 0
		for cc, cores := range resourcesPartitions.CPUClasses[cp].CPUCores {
			currentCCPercentage := cores.Percentage
			resourcesPartitions.CPUClasses[cp].CPUCores[cc].Percentage += ccAcc
			ccAcc += currentCCPercentage

			memoryAcc := 0
			for r, memory := range resourcesPartitions.CPUClasses[cp].CPUCores[cc].Memory {
				currentMemoryPercentage := memory.Percentage
				resourcesPartitions.CPUClasses[cp].CPUCores[cc].Memory[r].Percentage += memoryAcc
				memoryAcc += currentMemoryPercentage
			}
		}
	}

	cpuClassRand := p.randomGenerator.Intn(101)
	for cp, class := range resourcesPartitions.CPUClasses {
		if cpuClassRand <= class.Percentage {
			cpuCoresRand := p.randomGenerator.Intn(101)
			for cc, cores := range resourcesPartitions.CPUClasses[cp].CPUCores {
				if cpuCoresRand <= cores.Percentage {
					memoryRand := p.randomGenerator.Intn(101)
					for _, memory := range resourcesPartitions.CPUClasses[cp].CPUCores[cc].Memory {
						if memoryRand <= memory.Percentage {
							return class.Value, cores.Value, memory.Value
						}
					}
				}
			}
		}
	}

	panic(errors.New("partition-fit error generating resources"))
}
