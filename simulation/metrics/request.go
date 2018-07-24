package metrics

import (
	"github.com/strabox/caravela/api/types"
)

type RunRequest struct {
	Resources      types.Resources `json:"Resources"`
	MessagesTraded int64           `json:"IncrementMessagesTraded"`
}

func NewRunRequest(resources types.Resources) *RunRequest {
	return &RunRequest{
		Resources:      resources,
		MessagesTraded: 0,
	}
}

func (request *RunRequest) IncrementMessagesTraded(numMsgs int64) {
	request.MessagesTraded += numMsgs
}
func (request *RunRequest) TotalMessagesTraded() int64 {
	return request.MessagesTraded
}
