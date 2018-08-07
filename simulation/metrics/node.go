package metrics

import (
	"github.com/strabox/caravela/api/types"
	"sync/atomic"
)

// Node represents a node in the system and it is used to collect node's level metrics of
// a CARAVELA's node.
type Node struct {
	MaxResources        types.Resources `json:"MaxResources"`        // Maximum resources available in the node.
	AvailableRes        types.Resources `json:"AvailableResources"`  // Current available resources in the node.
	ApiRequestsReceived int64           `json:"ApiRequestsReceived"` // Number of API requests received.
	RequestsSubmitted   int64           `json:"RequestsSubmitted"`   // Number of requests submitted in the node..
}

// NewNode creates a new structure of to hold a node's metrics.
func NewNode(maxResources types.Resources) *Node {
	return &Node{
		MaxResources:        maxResources,
		AvailableRes:        maxResources,
		ApiRequestsReceived: 0,
		RequestsSubmitted:   0,
	}
}

// ========================= Metrics Collector Methods ====================================

func (node *Node) APIRequestReceived() {
	atomic.AddInt64(&node.ApiRequestsReceived, 1)
}

func (node *Node) RunRequestSubmitted() {
	atomic.AddInt64(&node.RequestsSubmitted, 1)
}

func (node *Node) SetAvailableResources(res types.Resources) {
	node.AvailableRes = res
}

// ================================= Getters  ==============================================

func (node *Node) MaximumResources() types.Resources {
	return node.MaxResources
}

func (node *Node) AvailableResources() types.Resources {
	return node.AvailableRes
}

func (node *Node) UsedResources() types.Resources {
	return types.Resources{
		CPUs: node.MaxResources.CPUs - node.AvailableRes.CPUs,
		RAM:  node.MaxResources.RAM - node.AvailableRes.RAM,
	}
}

func (node *Node) APIRequestsReceived() int64 {
	return node.ApiRequestsReceived
}

func (node *Node) TotalRunRequestsSubmitted() int64 {
	return node.RequestsSubmitted
}
