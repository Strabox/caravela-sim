package metrics

import (
	"github.com/strabox/caravela/api/types"
	"sync/atomic"
)

// RunRequest represents a request, to deploy a container, that was submitted in the system.
// It collects request level's metrics for the request.
type RunRequest struct {
	ResRequested   types.Resources `json:"ResRequested"`   // ResRequested necessary for the container.
	MessagesTraded int64           `json:"MessagesTraded"` // Messages traded in the system to handle the request.
}

// NewRunRequest creates a new structure to hold the information about a request.
func NewRunRequest(resourcesRequested types.Resources) *RunRequest {
	return &RunRequest{
		ResRequested:   resourcesRequested,
		MessagesTraded: 0,
	}
}

// ========================= Metrics Collector Methods ====================================

// IncrMessagesExchanged increments the number of messages necessary to handle the request.
func (r *RunRequest) IncrMessagesExchanged(numMessages int64) {
	atomic.AddInt64(&r.MessagesTraded, numMessages)
}

// ============================ Getters and Setters ========================================

// TotalMessagesExchanged returns the total number of messages necessary to handle the request.
func (r *RunRequest) TotalMessagesExchanged() int64 {
	return r.MessagesTraded
}

func (r *RunRequest) ResourcesRequested() types.Resources {
	return r.ResRequested
}
