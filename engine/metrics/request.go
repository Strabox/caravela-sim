package metrics

import (
	"github.com/strabox/caravela/api/types"
	"sync/atomic"
)

// RunRequest represents a request, to deploy a container, that was submitted in the system.
// It collects request level's metrics for the request.
type RunRequest struct {
	Resources      types.Resources `json:"Resources"`      // Resources necessary for the container.
	MessagesTraded int64           `json:"MessagesTraded"` // Messages traded in the system to handle the request.
}

// NewRunRequest creates a new structure to hold the information about a request.
func NewRunRequest(resources types.Resources) *RunRequest {
	return &RunRequest{
		Resources:      resources,
		MessagesTraded: 0,
	}
}

// ========================= Metrics Collector Methods ====================================

// IncrMessagesExchanged increments the number of messages necessary to handle the request.
func (request *RunRequest) IncrMessagesExchanged(numMessages int64) {
	atomic.AddInt64(&request.MessagesTraded, numMessages)
}

// ============================ Getters and Setters ========================================

// TotalMessagesExchanged returns the total number of messages necessary to handle the request.
func (request *RunRequest) TotalMessagesExchanged() int64 {
	return request.MessagesTraded
}
