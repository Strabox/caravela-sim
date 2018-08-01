package configuration

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/strabox/caravela-sim/util"
	"time"
)

// Default name of the configuration file.
const DefaultConfigFilePath = "simulation.toml"

// Default Simulator's log's level
const DefaultSimLogLevel = "panic"

// Default Simulator's out directory path
const DefaultOutDirectoryPath = "out"

// Configuration structure with initialization parameters for the simulator.
type Configuration struct {
	NumberOfNodes     int      // Number of nodes used in the simulation.
	TickInterval      duration // Interval between each simulator tick (in simulation time).
	MaxTicks          int      // Maximum number of ticks done by the simulator.
	OutDirectoryPath  string   // Path of the output's directory.
	SimulatorLogLevel string   // Log's level of the simulator.
	CaravelaLogLevel  string   // Log's level of the CARAVELA's system.
}

// Produces the configuration structure for a basic simulation.
func Default() *Configuration {
	return &Configuration{
		NumberOfNodes:     2500,
		TickInterval:      duration{Duration: 10 * time.Second},
		MaxTicks:          15,
		OutDirectoryPath:  DefaultOutDirectoryPath,
		SimulatorLogLevel: DefaultSimLogLevel,
		CaravelaLogLevel:  "info",
	}
}

// Produces configuration structure reading from the configuration file and filling the rest
// with the default values of the simulation
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

// Briefly validate the configuration to avoid/short-circuit many runtime errors due to
// typos or completely non sense configurations.
func (config *Configuration) validate() error {
	isValidLogLevel := func(logLevel string) bool {
		if logLevel == "info" || logLevel == "debug" || logLevel == "warning" ||
			logLevel == "error" || logLevel == "fatal" || logLevel == "panic" {
			return true
		} else {
			return false
		}
	}

	if config.NumberOfNodes <= 0 {
		return fmt.Errorf("the number of nodes in the simulation must be > 0: %d", config.NumberOfNodes)
	}

	if config.MaxTicks <= 0 {
		return fmt.Errorf("the number of maximum ticks must be > 0: %d", config.MaxTicks)
	}

	if !isValidLogLevel(config.CaravelaLogLevel) {
		return fmt.Errorf("invalid caravela log level: %s", config.CaravelaLogLevel)
	}

	if !isValidLogLevel(config.SimulatorLogLevel) {
		return fmt.Errorf("invalid simulator log level: %s", config.SimulatorLogLevel)
	}

	return nil
}

func (config *Configuration) TotalNumberOfNodes() int {
	return config.NumberOfNodes
}

func (config *Configuration) TicksInterval() time.Duration {
	return config.TickInterval.Duration
}

func (config *Configuration) MaximumTicks() int {
	return config.MaxTicks
}

func (config *Configuration) OutputDirectoryPath() string {
	return config.OutDirectoryPath
}

func (config *Configuration) SimulatorLogsLevel() string {
	return config.SimulatorLogLevel
}

func (config *Configuration) CaravelaLogsLevel() string {
	return config.CaravelaLogLevel
}

// Print/log the current configurations in order to debug the programs behavior.
func (config *Configuration) Print() {
	util.Log.Infof("##################################################################")
	util.Log.Infof("#               CARAVELA's SIMULATOR CONFIGURATIONS              #")
	util.Log.Infof("##################################################################")

	util.Log.Infof("#Nodes:               %d", config.TotalNumberOfNodes())
	util.Log.Infof("Tick Interval:        %s", config.TicksInterval().String())
	util.Log.Infof("Max Ticks:            %d", config.MaximumTicks())
	util.Log.Infof("Output directory:     %s", config.OutputDirectoryPath())
	util.Log.Infof("Sim's log level:      %s", config.SimulatorLogsLevel())
	util.Log.Infof("CARAVELA's log level: %s", config.CaravelaLogsLevel())

	util.Log.Infof("##################################################################")
}
