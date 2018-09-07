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
	copyPartitions := caravelaConfigs.ResourcesPartitions{}
	copyPartitions.CPUClasses = make([]caravelaConfigs.CPUClassPartition, len(p.caravelaConfigs.ResourcesPartitions().CPUClasses))
	for cp, class := range p.caravelaConfigs.ResourcesPartitions().CPUClasses {
		copyPartitions.CPUClasses[cp].Value = class.Value
		copyPartitions.CPUClasses[cp].Percentage = class.Percentage
		copyPartitions.CPUClasses[cp].CPUCores = make([]caravelaConfigs.CPUCoresPartition, len(class.CPUCores))
		for cc, cores := range class.CPUCores {
			copyPartitions.CPUClasses[cp].CPUCores[cc].Value = cores.Value
			copyPartitions.CPUClasses[cp].CPUCores[cc].Percentage = cores.Percentage
			copyPartitions.CPUClasses[cp].CPUCores[cc].Memory = make([]caravelaConfigs.MemoryPartition, len(cores.Memory))
			for r, memory := range cores.Memory {
				copyPartitions.CPUClasses[cp].CPUCores[cc].Memory[r].Value = memory.Value
				copyPartitions.CPUClasses[cp].CPUCores[cc].Memory[r].Percentage = memory.Percentage
			}
		}
	}

	cpAcc := 0
	for cp, cpuClass := range copyPartitions.CPUClasses {
		currentCPPercentage := cpuClass.Percentage
		copyPartitions.CPUClasses[cp].Percentage += cpAcc
		cpAcc += currentCPPercentage

		ccAcc := 0
		for cc, cores := range copyPartitions.CPUClasses[cp].CPUCores {
			currentCCPercentage := cores.Percentage
			copyPartitions.CPUClasses[cp].CPUCores[cc].Percentage += ccAcc
			ccAcc += currentCCPercentage

			memoryAcc := 0
			for r, memory := range copyPartitions.CPUClasses[cp].CPUCores[cc].Memory {
				currentMemoryPercentage := memory.Percentage
				copyPartitions.CPUClasses[cp].CPUCores[cc].Memory[r].Percentage += memoryAcc
				memoryAcc += currentMemoryPercentage
			}
		}
	}

	cpuClassRand := p.randomGenerator.Intn(101)
	for cp, class := range copyPartitions.CPUClasses {
		if cpuClassRand <= class.Percentage {
			cpuCoresRand := p.randomGenerator.Intn(101)
			for cc, cores := range copyPartitions.CPUClasses[cp].CPUCores {
				if cpuCoresRand <= cores.Percentage {
					memoryRand := p.randomGenerator.Intn(101)
					for _, memory := range copyPartitions.CPUClasses[cp].CPUCores[cc].Memory {
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
