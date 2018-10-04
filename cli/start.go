package cli

import (
	"fmt"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/engine"
	"github.com/strabox/caravela-sim/engine/metrics"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/util"
	"github.com/urfave/cli"
	"time"
)

func start(c *cli.Context) {
	configFilePath := c.GlobalString("config")

	simulatorConfig, err := configuration.ReadFromFile(configFilePath)
	if err != nil {
		util.Log.Errorf("Cannot read config file %s, error: %s", configFilePath, err)
		fmt.Println("Information: Using the default configurations!!")
		simulatorConfig = configuration.Default()
	}

	overrideSimFileConfigs(c, simulatorConfig)
	simulatorConfig.Print()

	metricsCollector := metrics.NewCollector(simulatorConfig.TotalNumberOfNodes(), simulatorConfig.OutDirectoryPath)

	// Base seed for engine pseudo-random generators.
	baseRngSeed := time.Now().UnixNano()

	simEngine := engine.NewEngine(metricsCollector, simulatorConfig, baseRngSeed)

	for i, str := range simulatorConfig.CaravelaDiscoveryBackends() {
		caravelaConfigs := caravela.Configuration()
		caravelaConfigs.Caravela.DiscoveryBackend.Backend = str

		fmt.Println("Initializing engine...")
		lastSimulation := i == (len(simulatorConfig.CaravelaDiscoveryBackends()) - 1)
		simEngine.Init(true, lastSimulation, caravelaConfigs)

		fmt.Println("Starting engine...")
		simEngine.Start()
		fmt.Println("Simulation ended")
	}

	fmt.Println("Crushing engine results...")
	metricsCollector.Print() // Print the metricsCollector results and outputs the graphics.
	metricsCollector.Clear() // Clear all the temporary metric files
}

// Overrides file configurations with CLI arguments passed
func overrideSimFileConfigs(c *cli.Context, config *configuration.Configuration) {
	config.SimulatorLogLevel = c.GlobalString("log")
}
