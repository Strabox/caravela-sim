package main

import (
	"fmt"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/simulation/simulator"
	"os"
)

func main() {
	var simulatorConfig *configuration.Configuration
	var err error

	if len(os.Args) == 2 {
		simulatorConfig, err = configuration.ReadFromFile(os.Args[1])
	} else {
		simulatorConfig, err = configuration.ReadFromDefaultFile()
	}

	if err != nil {
		fmt.Printf("Error in configuration file: %s\n", err)
		os.Exit(1)
	}

	mySimulator := simulator.NewSimulator(simulatorConfig)
	mySimulator.Init()
	mySimulator.Start()
}
