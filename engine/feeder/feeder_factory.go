package feeder

import (
	"errors"
	"fmt"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/util"
	"log"
	"strings"
)

// Factory represents a method that creates new requests feeders.
type Factory func(config *configuration.Configuration) (Feeder, error)

// feeders holds all the registered requests feeder available.
var feeders = make(map[string]Factory)

// init initializes our predefined request feeders.
func init() {
	Register("random", newRandomFeeder)
}

// Register can be used to register a new request feeder in order to be available.
func Register(feederName string, factory Factory) {
	if factory == nil {
		log.Panic("nil request feeder registering")
	}
	_, exist := feeders[feederName]
	if exist {
		util.Log.Warnf("request feeder %s is being overridden", feederName)
	}
	feeders[feederName] = factory
}

// Create is used to obtain a request feeder based on the configurations.
func Create(config *configuration.Configuration) Feeder {
	configuredFeeder := config.Feeder()

	feederFactory, exist := feeders[configuredFeeder]
	if !exist {
		existingFeeders := make([]string, len(feeders))
		for feederName := range feeders {
			existingFeeders = append(existingFeeders, feederName)
		}
		err := errors.New(fmt.Sprintf("Invalid %s request feeder. Feeders available: %s",
			configuredFeeder, strings.Join(existingFeeders, ", ")))
		log.Panic(err)
	}

	feeder, err := feederFactory(config)
	if err != nil {
		log.Panic(err)
	}

	return feeder
}
