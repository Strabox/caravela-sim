package docker

import (
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	caravelaConfigs "github.com/strabox/caravela/configuration"
)

type realisticallyResourceGen struct {
	caravelaConfigs *caravelaConfigs.Configuration
}

type nodeResourcesProfile struct {
	Percentage int
	Resources  types.Resources
}

func newRealisticallyResourceGen(_ *configuration.Configuration, caravelaConfigs *caravelaConfigs.Configuration) (ResourcesGenerator, error) {
	return &realisticallyResourceGen{
		caravelaConfigs: caravelaConfigs,
	}, nil
}

func (e *realisticallyResourceGen) Generate() (int, int, int) {
	nodesProfiles := []nodeResourcesProfile{
		{
			Percentage: 50,
			Resources: types.Resources{
				CPUs: 2,
				RAM:  4096,
			},
		},
		{
			Percentage: 30,
			Resources: types.Resources{
				CPUs: 4,
				RAM:  4096,
			},
		},
		{
			Percentage: 15,
			Resources: types.Resources{
				CPUs: 4,
				RAM:  8128,
			},
		},
		{
			Percentage: 5,
			Resources: types.Resources{
				CPUs: 8,
				RAM:  16326,
			},
		},
	}

	acc := 0
	for i := range nodesProfiles {
		currentPercentage := nodesProfiles[i].Percentage
		nodesProfiles[i].Percentage += acc
		acc += currentPercentage
	}

	randInt := util.RandomInteger(1, 100)
	prevResourcesProfile := nodesProfiles[0].Resources
	for i := range nodesProfiles {
		if randInt <= nodesProfiles[i].Percentage {
			prevResourcesProfile = nodesProfiles[i].Resources
			break
		}
	}

	return 1, prevResourcesProfile.CPUs, prevResourcesProfile.RAM // TODO: Dehardcode the CPUClass
}
