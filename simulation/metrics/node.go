package metrics

import (
	"github.com/strabox/caravela/api/types"
)

type Node struct {
	MaxRes              types.Resources `json:"MaxResources"`
	AvailableRes        types.Resources `json:"AvailableResources"`
	ApiRequestsReceived int64           `json:"ApiRequestsReceived"`
	RequestsSubmitted   int64           `json:"RequestsSubmitted"`
}

func NewNode(maxResources types.Resources) *Node {
	return &Node{
		MaxRes:              maxResources,
		AvailableRes:        maxResources,
		ApiRequestsReceived: 0,
		RequestsSubmitted:   0,
	}
}

func (node *Node) APIRequestReceived() {
	node.ApiRequestsReceived++
}

func (node *Node) APIRequestsReceived() int64 {
	return node.ApiRequestsReceived
}

func (node *Node) RunRequestSubmitted() {
	node.RequestsSubmitted++
}

func (node *Node) TotalRunRequestsSubmitted() int64 {
	return node.RequestsSubmitted
}

func (node *Node) AvailableResources() types.Resources {
	return node.AvailableRes
}

func (node *Node) SetAvailableResources(res types.Resources) {
	node.AvailableRes = res
}

func (node *Node) UsedResources() types.Resources {
	return types.Resources{
		CPUs: node.MaxRes.CPUs - node.AvailableRes.CPUs,
		RAM:  node.MaxRes.RAM - node.AvailableRes.RAM,
	}
}

func (node *Node) MaximumResources() types.Resources {
	return node.MaxRes
}
