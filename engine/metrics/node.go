package metrics

import (
	"github.com/strabox/caravela/api/types"
	"sync/atomic"
)

// Node represents a node in the system and it is used to collect node's level metrics of
// a CARAVELA's node.
type Node struct {
	MaxResources        types.Resources `json:"MaxResources"`        // Maximum resources available in the node.
	FreeRes             types.Resources `json:"FreeResources"`       // Current available resources in the node.
	ApiRequestsReceived int64           `json:"ApiRequestsReceived"` // Number of API requests received.
	RequestsSubmitted   int64           `json:"RequestsSubmitted"`   // Number of requests submitted in the node..
}

// NewNode creates a new structure of to hold a node's metrics.
func NewNode(maxResources types.Resources) *Node {
	return &Node{
		MaxResources:        maxResources,
		FreeRes:             maxResources,
		ApiRequestsReceived: 0,
		RequestsSubmitted:   0,
	}
}

// ========================= Metrics Collector Methods ====================================

func (n *Node) APIRequestReceived() {
	atomic.AddInt64(&n.ApiRequestsReceived, 1)
}

func (n *Node) RunRequestSubmitted() {
	atomic.AddInt64(&n.RequestsSubmitted, 1)
}

func (n *Node) SetFreeResources(freeRes types.Resources) {
	n.FreeRes = freeRes
}

// ================================= Getters  ==============================================

func (n *Node) MaximumResources() types.Resources {
	return n.MaxResources
}

func (n *Node) FreeResources() types.Resources {
	return n.FreeRes
}

func (n *Node) UsedResources() types.Resources {
	return types.Resources{
		CPUs:   n.MaxResources.CPUs - n.FreeRes.CPUs,
		Memory: n.MaxResources.Memory - n.FreeRes.Memory,
	}
}

func (n *Node) FreeResourcesRatio() float64 {
	if n.FreeRes.CPUs == 0 || n.FreeRes.Memory == 0 { // Impossible use this "free" resources.
		return 0
	}
	cpusRatio := float64(n.FreeRes.CPUs) / float64(n.MaxResources.CPUs)
	memoryRatio := float64(n.FreeRes.Memory) / float64(n.MaxResources.Memory)
	return (cpusRatio + memoryRatio) / 2
}

func (n *Node) UsedResourcesRatio() float64 {
	if n.FreeRes.CPUs == 0 || n.FreeRes.Memory == 0 { // Impossible use this "free" resources.
		return float64(1)
	}
	cpusRatio := float64(n.UsedResources().CPUs) / float64(n.MaxResources.CPUs)
	memoryRatio := float64(n.UsedResources().Memory) / float64(n.MaxResources.Memory)
	return (cpusRatio + memoryRatio) / 2
}

func (n *Node) UnreachableResourcesRatio() float64 {
	if n.FreeRes.CPUs == 0 || n.FreeRes.Memory == 0 { // Impossible use this "free" resources.
		cpusRatio := float64(n.FreeRes.CPUs) / float64(n.MaxResources.CPUs)
		memoryRatio := float64(n.FreeRes.Memory) / float64(n.MaxResources.Memory)
		return (cpusRatio + memoryRatio) / 2
	}
	return 0
}

func (n *Node) TotalAPIRequestsReceived() int64 {
	return n.ApiRequestsReceived
}

func (n *Node) TotalRunRequestsSubmitted() int64 {
	return n.RequestsSubmitted
}
