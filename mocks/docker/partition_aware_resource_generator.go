package docker

import (
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/util"
	caravelaConfigs "github.com/strabox/caravela/configuration"
)

type partitionAwareResourceGen struct {
	caravelaConfigs *caravelaConfigs.Configuration
}

func newPartitionAwareResourceGen(_ *configuration.Configuration, caravelaConfigs *caravelaConfigs.Configuration) (ResourcesGenerator, error) {
	return &partitionAwareResourceGen{
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
			copyPartitions.CPUClasses[cp].CPUCores[cc].RAMs = make([]caravelaConfigs.RAMPartition, len(cores.RAMs))
			for r, ram := range cores.RAMs {
				copyPartitions.CPUClasses[cp].CPUCores[cc].RAMs[r].Value = ram.Value
				copyPartitions.CPUClasses[cp].CPUCores[cc].RAMs[r].Percentage = ram.Percentage
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

			ramAcc := 0
			for r, ram := range copyPartitions.CPUClasses[cp].CPUCores[cc].RAMs {
				currentRAMPercentage := ram.Percentage
				copyPartitions.CPUClasses[cp].CPUCores[cc].RAMs[r].Percentage += ramAcc
				ramAcc += currentRAMPercentage
			}
		}
	}

	cpuClassRand := util.RandomInteger(1, 100)
	for cp, class := range copyPartitions.CPUClasses {
		if cpuClassRand <= class.Percentage {
			cpuCoresRand := util.RandomInteger(1, 100)
			for cc, cores := range copyPartitions.CPUClasses[cp].CPUCores {
				if cpuCoresRand <= cores.Percentage {
					ramRand := util.RandomInteger(1, 100)
					for _, ram := range copyPartitions.CPUClasses[cp].CPUCores[cc].RAMs {
						if ramRand <= ram.Percentage {
							return class.Value, cores.Value, ram.Value
						}
					}
				}
			}
		}
	}

	panic(errors.New("partition-fit error generating resources"))
}
