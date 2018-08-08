package docker

import (
	"errors"
	"fmt"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/util"
	caravelaConfigs "github.com/strabox/caravela/configuration"
	"log"
	"strings"
)

// Factory represents a method that creates new resource generators.
type ResourceGenFactory func(caravelaConfig *caravelaConfigs.Configuration) (ResourcesGenerator, error)

// generators holds all the registered resource generators available.
var generators = make(map[string]ResourceGenFactory)

// init initializes our predefined resource generators.
func init() {
	RegisterResourceGen("static", newStaticResourceGen)
	RegisterResourceGen("partition-aware", newPartitionAwareResourceGen)
	RegisterResourceGen("real", newRealisticallyResourceGen)
}

// RegisterResourceGen can be used to register a new resource generator in order to be available.
func RegisterResourceGen(resourceGenName string, factory ResourceGenFactory) {
	if factory == nil {
		log.Panic("nil resource generator registering")
	}
	_, exist := generators[resourceGenName]
	if exist {
		util.Log.Warnf("resource generator %s is being overridden", resourceGenName)
	}
	generators[resourceGenName] = factory
}

// CreateResourceGen is used to obtain a resource generator based on the configurations.
func CreateResourceGen(config *configuration.Configuration, caravelaConfigs *caravelaConfigs.Configuration) ResourcesGenerator {
	configuredResourceGen := config.ResourceGen()

	resourceGeneratorFactory, exist := generators[configuredResourceGen]
	if !exist {
		existingGenerators := make([]string, len(generators))
		for genName := range generators {
			existingGenerators = append(existingGenerators, genName)
		}
		err := errors.New(fmt.Sprintf("Invalid %s resource generator. Generators available: %s",
			configuredResourceGen, strings.Join(existingGenerators, ", ")))
		log.Panic(err)
	}

	resourceGenerator, err := resourceGeneratorFactory(caravelaConfigs)
	if err != nil {
		log.Panic(err)
	}

	return resourceGenerator
}
