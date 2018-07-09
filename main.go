package main

import (
	"github.com/strabox/caravela-sim/caravela"
	"github.com/strabox/caravela-sim/simulation/simulator"
)

func main() {
	caravela.PrintSimulatorBanner()
	caravela.SetLogs()
	mySimulator := simulator.NewSimulator(5000)
	mySimulator.Init()
	mySimulator.Start()
}
