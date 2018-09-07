package docker

/*

type realisticallyResourceGen struct {
	randomGenerator *rand.Rand						// Pseudo-random generator.
	caravelaConfigs *caravelaConfigs.Configuration	// Caravela's configurations.
}

type nodeResourcesProfile struct {
	Percentage int
	Resources  types.Resources
}

func newRealisticallyResourceGen(_ *configuration.Configuration, caravelaConfigs *caravelaConfigs.Configuration, rngSeed int64) (ResourcesGenerator, error) {
	return &realisticallyResourceGen{
		randomGenerator: rand.New(caravelaUtil.NewSourceSafe(rand.NewSource(rngSeed))),
		caravelaConfigs: caravelaConfigs,
	}, nil
}

func (r *realisticallyResourceGen) Generate() (int, int, int) {
	nodesProfiles := []nodeResourcesProfile{
		{
			Percentage: 50,
			Resources: types.Resources{
				CPUClass: 0,
				CPUs: 2,
				Memory:  4096,
			},
		},
		{
			Percentage: 30,
			Resources: types.Resources{
				CPUClass: 0,
				CPUs: 4,
				Memory:  4096,
			},
		},
		{
			Percentage: 15,
			Resources: types.Resources{
				CPUClass: 0,
				CPUs: 4,
				Memory:  8128,
			},
		},
		{
			Percentage: 5,
			Resources: types.Resources{
				CPUClass: 0,
				CPUs: 8,
				Memory:  16326,
			},
		},
	}

	acc := 0
	for i := range nodesProfiles {
		currentPercentage := nodesProfiles[i].Percentage
		nodesProfiles[i].Percentage += acc
		acc += currentPercentage
	}

	randInt := r.randomGenerator.Intn(101)
	prevResourcesProfile := nodesProfiles[0].Resources
	for i := range nodesProfiles {
		if randInt <= nodesProfiles[i].Percentage {
			prevResourcesProfile = nodesProfiles[i].Resources
			break
		}
	}

	return 0, prevResourcesProfile.CPUs, prevResourcesProfile.Memory // TODO: Dehardcode the CPUClass
}

*/
