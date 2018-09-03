package caravela

import (
	"github.com/strabox/caravela/api"
)

// APIServerMock mocks the the web server that attends the API requests.
// Dummy object only to be passed into CARAVELA's Node structure.
// It implements the github.com/strabox/caravela/api Server interface.
type APIServerMock struct {
	// Nothing
}

// NewAPIServerMock creates a new APIServerMock structure.
func NewAPIServerMock() *APIServerMock {
	return &APIServerMock{}
}

// ===============================================================================
// =							  API Server Interface                           =
// ===============================================================================

func (server *APIServerMock) Start(thisNode api.LocalNode) error {
	// Do Nothing (For now not necessary for the engine)
	// API Server always starts with success
	return nil
}

func (server *APIServerMock) Stop() {
	// Do Nothing (For now not necessary for the engine)
	// API Server always stop with success
}
