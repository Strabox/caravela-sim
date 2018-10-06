package configuration

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	"time"
)

// TODO
const DefaultRequestFeeder = "random"

// TODO
const DefaultResourceGenerator = "partition-fit"

// TODO
const DefaultSpeedupNodes = 300

// Default name of the configuration file.
const DefaultConfigFilePath = "simulation.toml"

// Default Simulator's log's level
const DefaultSimLogLevel = "panic"

// TODO
const DefaultCaravelaLogLevel = "info"

// Default Simulator's out directory path
const DefaultOutDirectoryPath = "out"

// Configuration structure with initialization parameters for the simulator.
type Configuration struct {
	NumberOfNodes      int      // Number of nodes used in the engine.
	TickInterval       duration // Interval between each simulator tick (in engine time).
	MaxTicks           int      // Maximum number of ticks done by the simulator.
	Multithread        bool     // Used to leverage the multiple cores to speed up the engine.
	DiscoveryBackends  []string // The discovery backends to simulate
	RequestFeeder      requestFeeder
	ResourcesGenerator resourcesGenerator // Strategies used to generate the resources for each node.
	ChordMock          chordMock
	OutDirectoryPath   string // Path of the output's directory.
	SimulatorLogLevel  string // Log's level of the simulator.
	CaravelaLogLevel   string // Log's level of the CARAVELA's system.
}

// TODO
type requestFeeder struct {
	RequestFeeder      string // Used to feed the simulator with a series of requests.
	RequestsProfile    []RequestProfile
	DeployRequestsRate []float64
	StopRequestsRate   []float64
}

// TODO
type resourcesGenerator struct {
	ResourceGenerator string
	StaticResources   types.Resources
}

// TODO
type chordMock struct {
	SpeedupNodes int
}

// Default creates the configuration structure for a basic/default engine.
func Default() *Configuration {
	return &Configuration{
		NumberOfNodes:     10000,
		TickInterval:      duration{Duration: 20 * time.Second},
		MaxTicks:          50,
		Multithread:       true,
		OutDirectoryPath:  DefaultOutDirectoryPath,
		SimulatorLogLevel: DefaultSimLogLevel,
		CaravelaLogLevel:  DefaultCaravelaLogLevel,
		RequestFeeder: requestFeeder{
			RequestFeeder:      DefaultRequestFeeder,
			DeployRequestsRate: []float64{0.025, 0.015, 0.010, 0.035, 0.02, 0.01, 0.01, 0.05},
			StopRequestsRate:   []float64{0, 0, 0, 0, 0, 0.025, 0.015, 0.15},
			RequestsProfile: []RequestProfile{
				{CPUClass: 0, CPUs: 1, Memory: 256, Percentage: 20},
				{CPUClass: 0, CPUs: 2, Memory: 1500, Percentage: 20},
				{CPUClass: 0, CPUs: 3, Memory: 2500, Percentage: 10},
				{CPUClass: 0, CPUs: 4, Memory: 5500, Percentage: 10},
				{CPUClass: 0, CPUs: 8, Memory: 7000, Percentage: 10},
				{CPUClass: 1, CPUs: 1, Memory: 512, Percentage: 20},
				{CPUClass: 1, CPUs: 3, Memory: 2048, Percentage: 5},
				{CPUClass: 1, CPUs: 6, Memory: 5000, Percentage: 5},
			},
		},
		ResourcesGenerator: resourcesGenerator{
			ResourceGenerator: DefaultResourceGenerator,
			StaticResources: types.Resources{
				CPUClass: 0,
				CPUs:     4,
				Memory:   4096,
			},
		},
		ChordMock: chordMock{
			SpeedupNodes: DefaultSpeedupNodes,
		},
	}
}

// Produces configuration structure reading from the configuration file and filling the rest
// with the default values of the engine
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
func (c *Configuration) validate() error {
	isValidLogLevel := func(logLevel string) bool {
		if logLevel == "info" || logLevel == "debug" || logLevel == "warning" ||
			logLevel == "error" || logLevel == "fatal" || logLevel == "panic" {
			return true
		} else {
			return false
		}
	}

	if c.NumberOfNodes <= 0 {
		return fmt.Errorf("the number of nodes in the engine must be > 0: %d", c.NumberOfNodes)
	}

	if c.MaxTicks <= 0 {
		return fmt.Errorf("the number of maximum ticks must be > 0: %d", c.MaxTicks)
	}

	if len(c.DeployRequestsRate()) == 0 {
		return fmt.Errorf("the sequence of deploy requests rate must have at least one rate")
	}

	if len(c.StopRequestsRate()) == 0 {
		return fmt.Errorf("the sequence stop requests rate must have at least one rate")
	}

	if c.ChordMock.SpeedupNodes <= 0 {
		return fmt.Errorf("the number of speedup nodes must be > 0: %d", c.MaxTicks)
	}

	if !isValidLogLevel(c.CaravelaLogLevel) {
		return fmt.Errorf("invalid caravela log level: %s", c.CaravelaLogLevel)
	}

	if !isValidLogLevel(c.SimulatorLogLevel) {
		return fmt.Errorf("invalid simulator log level: %s", c.SimulatorLogLevel)
	}

	return nil
}

func (c *Configuration) TotalNumberOfNodes() int {
	return c.NumberOfNodes
}

func (c *Configuration) TicksInterval() time.Duration {
	return c.TickInterval.Duration
}

func (c *Configuration) MaximumTicks() int {
	return c.MaxTicks
}

func (c *Configuration) Multithreaded() bool {
	return c.Multithread
}

func (c *Configuration) CaravelaDiscoveryBackends() []string {
	return c.DiscoveryBackends
}

func (c *Configuration) DeployRequestsRate() []float64 {
	res := make([]float64, len(c.RequestFeeder.DeployRequestsRate))
	copy(res, c.RequestFeeder.DeployRequestsRate)
	return res
}

func (c *Configuration) StopRequestsRate() []float64 {
	res := make([]float64, len(c.RequestFeeder.StopRequestsRate))
	copy(res, c.RequestFeeder.StopRequestsRate)
	return res
}

func (c *Configuration) RequestsProfile() []RequestProfile {
	res := make([]RequestProfile, len(c.RequestFeeder.RequestsProfile))
	copy(res, c.RequestFeeder.RequestsProfile)
	return res
}

func (c *Configuration) Feeder() string {
	return c.RequestFeeder.RequestFeeder
}

func (c *Configuration) ChordMockSpeedupNodes() int {
	return c.ChordMock.SpeedupNodes
}

func (c *Configuration) ResourceGen() string {
	return c.ResourcesGenerator.ResourceGenerator
}

func (c *Configuration) StaticGeneratorResources() types.Resources {
	return c.ResourcesGenerator.StaticResources
}

func (c *Configuration) OutputDirectoryPath() string {
	return c.OutDirectoryPath
}

func (c *Configuration) SimulatorLogsLevel() string {
	return c.SimulatorLogLevel
}

func (c *Configuration) CaravelaLogsLevel() string {
	return c.CaravelaLogLevel
}

// Print/log the current configurations in order to debug the programs behavior.
func (c *Configuration) Print() {
	util.Log.Infof("##################################################################")
	util.Log.Infof("#               CARAVELA's SIMULATOR CONFIGURATIONS              #")
	util.Log.Infof("##################################################################")

	util.Log.Infof("Num Nodes:                %d", c.TotalNumberOfNodes())
	util.Log.Infof("Tick Interval:            %s", c.TicksInterval().String())
	util.Log.Infof("Max Ticks:                %d", c.MaximumTicks())
	util.Log.Infof("Multithread:              %t", c.Multithreaded())
	util.Log.Infof("Discovery Backends:       %v", c.CaravelaDiscoveryBackends())
	util.Log.Infof("Request Feeder:           %s", c.Feeder())
	util.Log.Infof("Output directory:         %s", c.OutputDirectoryPath())
	util.Log.Infof("Sim's log level:          %s", c.SimulatorLogsLevel())
	util.Log.Infof("CARAVELA's log level:     %s", c.CaravelaLogsLevel())

	util.Log.Infof("")

	util.Log.Infof("Request Feeder")
	util.Log.Infof("  Request Feeder:         %s", c.Feeder())
	util.Log.Infof("  Deploy Requests Rate:   %v", c.DeployRequestsRate())
	util.Log.Infof("  Stop Requests Rate:     %v", c.StopRequestsRate())
	for _, reqProfile := range c.RequestsProfile() {
		util.Log.Infof("    <<%d;%d>;%d>:         %d%%", reqProfile.CPUClass, reqProfile.CPUs, reqProfile.Memory, reqProfile.Percentage)
	}
	util.Log.Infof("")

	util.Log.Infof("Resource Generation")
	util.Log.Infof("  Resource Generator:     %s", c.ResourceGen())
	util.Log.Infof("  Static Gen ResRequested:   <<%d;%d>;%d>", c.StaticGeneratorResources().CPUClass, c.StaticGeneratorResources().CPUs, c.StaticGeneratorResources().Memory)

	util.Log.Infof("")

	util.Log.Infof("Chord Mock")
	util.Log.Infof("  Chord Mock Speedup:     %d", c.ChordMockSpeedupNodes())

	util.Log.Infof("##################################################################")
}
