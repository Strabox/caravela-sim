package cli

import (
	"fmt"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/simulation"
	"github.com/strabox/caravela-sim/util"
	"github.com/urfave/cli"
)

func start(c *cli.Context) {
	configFilePath := c.GlobalString("config")

	simulatorConfig, err := configuration.ReadFromFile(configFilePath)
	if err != nil {
		util.Log.Errorf("Cannot read config file %s, error: %s", configFilePath, err)
		fmt.Println("Information: Using the default configurations!!")
		simulatorConfig = configuration.Default()
	}

	overrideFileConfigs(c, simulatorConfig)

	simulatorConfig.Print()

	caravelaConfigs := caravela.Configuration()
	mySimulator := simulation.NewSimulator(simulatorConfig, caravelaConfigs)

	fmt.Println("Initializing simulation...")
	mySimulator.Init()

	fmt.Println("Starting simulation...")
	mySimulator.Start()

	fmt.Println("Simulation ended")
}

// Overrides file configurations with CLI arguments passed
func overrideFileConfigs(c *cli.Context, config *configuration.Configuration) {
	config.SimulatorLogLevel = c.GlobalString("log")
}
