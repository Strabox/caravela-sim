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

func (p *partitionAwareResourceGen) Generate() (int, int) {
	copyPartitions := caravelaConfigs.ResourcesPartitions{}
	copyPartitions.CPUPowers = make([]caravelaConfigs.CPUPowerPartition, len(p.caravelaConfigs.ResourcesPartitions().CPUPowers))
	for cp, power := range p.caravelaConfigs.ResourcesPartitions().CPUPowers {
		copyPartitions.CPUPowers[cp].Value = power.Value
		copyPartitions.CPUPowers[cp].Percentage = power.Percentage
		copyPartitions.CPUPowers[cp].CPUCores = make([]caravelaConfigs.CPUCoresPartition, len(power.CPUCores))
		for cc, cores := range power.CPUCores {
			copyPartitions.CPUPowers[cp].CPUCores[cc].Value = cores.Value
			copyPartitions.CPUPowers[cp].CPUCores[cc].Percentage = cores.Percentage
			copyPartitions.CPUPowers[cp].CPUCores[cc].RAMs = make([]caravelaConfigs.RAMPartition, len(cores.RAMs))
			for r, ram := range cores.RAMs {
				copyPartitions.CPUPowers[cp].CPUCores[cc].RAMs[r].Value = ram.Value
				copyPartitions.CPUPowers[cp].CPUCores[cc].RAMs[r].Percentage = ram.Percentage
			}
		}
	}

	cpAcc := 0
	for cp, power := range copyPartitions.CPUPowers {
		currentCPPercentage := power.Percentage
		copyPartitions.CPUPowers[cp].Percentage += cpAcc
		cpAcc += currentCPPercentage

		ccAcc := 0
		for cc, cores := range copyPartitions.CPUPowers[cp].CPUCores {
			currentCCPercentage := cores.Percentage
			copyPartitions.CPUPowers[cp].CPUCores[cc].Percentage += ccAcc
			ccAcc += currentCCPercentage

			ramAcc := 0
			for r, ram := range copyPartitions.CPUPowers[cp].CPUCores[cc].RAMs {
				currentRAMPercentage := ram.Percentage
				copyPartitions.CPUPowers[cp].CPUCores[cc].RAMs[r].Percentage += ramAcc
				ramAcc += currentRAMPercentage
			}
		}
	}

	cpuPowerRand := util.RandomInteger(1, 100)
	for cp, power := range copyPartitions.CPUPowers {
		if cpuPowerRand <= power.Percentage {
			cpuCoresRand := util.RandomInteger(1, 100)
			for cc, cores := range copyPartitions.CPUPowers[cp].CPUCores {
				if cpuCoresRand <= cores.Percentage {
					ramRand := util.RandomInteger(1, 100)
					for _, ram := range copyPartitions.CPUPowers[cp].CPUCores[cc].RAMs {
						if ramRand <= ram.Percentage {
							return cores.Value, ram.Value
						}
					}
				}
			}
		}
	}

	panic(errors.New("partition-fit error generating resources"))
}
