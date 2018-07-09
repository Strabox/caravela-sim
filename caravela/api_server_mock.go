package caravela

import "github.com/strabox/caravela-sim/simulation"
import "github.com/strabox/caravela/configuration"
import "github.com/strabox/caravela/node/api"

type APIServerMock struct {
	sim simulation.Simulator
}

func NewAPIServerMock(sim simulation.Simulator) *APIServerMock {
	return &APIServerMock{
		sim: sim,
	}
}

/*
===============================================================================
							  API Server Interface
===============================================================================
*/

func (server *APIServerMock) Start(config *configuration.Configuration, thisNode api.Node) error {
	// Do Nothing (For now not necessary for the simulation)
	// API Server always starts with success
	return nil
}

func (server *APIServerMock) Stop() {
	// Do Nothing (For now not necessary for the simulation)
	// API Server always stop with success
}
