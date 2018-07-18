package configuration

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"time"
)

//Directory path to where search for the default configuration file. (Directory of binary execution)
const configurationFilePath = ""

//Default name of the configuration file.
const configurationFileName = "simulation.toml"

/*
Configuration structure used to initialize the simulator.
*/
type Configuration struct {
	NumberOfNodes             int
	TimeBetweenNodeStart      duration
	TimeBeforeStartSimulating duration
	SimulatorLogLevel         string

	CaravelaLogLevel string
}

/*
Produces the configuration structure for a basic simulation.
*/
func Default() *Configuration {
	return &Configuration{
		NumberOfNodes:             1000,
		TimeBetweenNodeStart:      duration{Duration: 100 * time.Millisecond},
		TimeBeforeStartSimulating: duration{Duration: 1 * time.Minute},
		SimulatorLogLevel:         "debug",

		CaravelaLogLevel: "debug",
	}
}

/*
Produces configuration structure reading from the configuration file and filling the rest
with the default values of the simulation
*/
func ReadFromDefaultFile() (*Configuration, error) {
	return ReadFromFile(configurationFilePath + configurationFileName)
}

/*
Produces configuration structure reading from the configuration file and filling the rest
with the default values of the simulation
*/
func ReadFromFile(configFilePath string) (*Configuration, error) {
	config := Default()

	if _, err := toml.DecodeFile(configFilePath, config); err != nil {
		return nil, err
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

/*
Briefly validate the configuration to avoid/short-circuit many runtime errors due to
typos or completely non sense configurations.
*/
func (config *Configuration) validate() error {
	isValidLogLevel := func(logLevel string) bool {
		if logLevel == "info" || logLevel == "debug" || logLevel == "warning" || logLevel == "error" || logLevel == "fatal" || logLevel == "panic" {
			return true
		} else {
			return false
		}
	}

	if config.NumberOfNodes <= 0 {
		return fmt.Errorf("the number of nodes in the simulation must be > 0: %d", config.NumberOfNodes)
	}

	if !isValidLogLevel(config.CaravelaLogLevel) {
		return fmt.Errorf("invalid caravela log level: %s", config.CaravelaLogLevel)
	}

	if !isValidLogLevel(config.SimulatorLogLevel) {
		return fmt.Errorf("invalid simulator log level: %s", config.SimulatorLogLevel)
	}

	return nil
}
