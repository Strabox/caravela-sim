package caravela

import (
	"github.com/strabox/caravela-sim/simulation"
	"github.com/strabox/caravela/api"
)

// APIServerMock mocks the the web server that attends the API requests.
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

func (server *APIServerMock) Start(thisNode api.LocalNode) error {
	// Do Nothing (For now not necessary for the simulation)
	// API Server always starts with success
	return nil
}

func (server *APIServerMock) Stop() {
	// Do Nothing (For now not necessary for the simulation)
	// API Server always stop with success
}
