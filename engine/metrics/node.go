package metrics

import (
	"github.com/strabox/caravela/api/types"
	"sync/atomic"
)

// Node represents a node in the system and it is used to collect node's level metrics of a CARAVELA's node.
type Node struct {
	MaxResources         types.Resources `json:"MaxResources"`         // Maximum resources available in the node.
	FreeResource         types.Resources `json:"FreeResources"`        // Current available resources in the node.
	MessagesReceived     int64           `json:"MessagesReceived"`     // Number of API requests received.
	MessagesReceivedSize int64           `json:"MessagesReceivedSize"` // Total size of all received messages API + Chord.
	MemoryUsed           int64           `json:"MemoryUsed"`           // Total memory occupied by the Caravela's logic components.
	RequestsSubmitted    int64           `json:"RequestsSubmitted"`    // Number of requests submitted in the node.
	TraderActiveOffers   int64           `json:"TraderActiveOffers"`   // Number of active offers in the node.
}

// NewNode creates a new structure of to hold a node's metrics.
func NewNode(maxResources types.Resources) *Node {
	return &Node{
		MaxResources:         maxResources,
		FreeResource:         maxResources,
		MessagesReceived:     0,
		MessagesReceivedSize: 0,
		RequestsSubmitted:    0,
		MemoryUsed:           0,
		TraderActiveOffers:   0,
	}
}

// ========================= Metrics Collector Methods ====================================

func (n *Node) MessageReceived(amountMessages int64, requestSizeBytes int64) {
	atomic.AddInt64(&n.MessagesReceived, amountMessages)
	atomic.AddInt64(&n.MessagesReceivedSize, requestSizeBytes)
}

func (n *Node) RunRequestSubmitted() {
	atomic.AddInt64(&n.RequestsSubmitted, 1)
}

func (n *Node) SetNodeState(freeResources types.Resources, traderActiveOffers int64, memoryUsed int64) {
	n.FreeResource = freeResources
	n.TraderActiveOffers = traderActiveOffers
	n.MemoryUsed = memoryUsed
}

// ================================== Getters  =============================================

func (n *Node) MaximumResources() types.Resources {
	return n.MaxResources
}

func (n *Node) FreeResources() types.Resources {
	return n.FreeResource
}

func (n *Node) UsedResources() types.Resources {
	return types.Resources{
		CPUs:   n.MaxResources.CPUs - n.FreeResource.CPUs,
		Memory: n.MaxResources.Memory - n.FreeResource.Memory,
	}
}
func (n *Node) FreeResourcesRatio() float64 {
	if n.FreeResource.CPUs == 0 || n.FreeResource.Memory == 0 { // Impossible use this "free" resources.
		return 0
	}
	cpusRatio := float64(n.FreeResource.CPUs) / float64(n.MaxResources.CPUs)
	memoryRatio := float64(n.FreeResource.Memory) / float64(n.MaxResources.Memory)
	return (cpusRatio + memoryRatio) / 2
}

func (n *Node) UsedResourcesRatio() float64 {
	if n.FreeResource.CPUs == 0 || n.FreeResource.Memory == 0 { // Impossible use this "free" resources.
		return float64(1)
	}
	cpusRatio := float64(n.UsedResources().CPUs) / float64(n.MaxResources.CPUs)
	memoryRatio := float64(n.UsedResources().Memory) / float64(n.MaxResources.Memory)
	return (cpusRatio + memoryRatio) / 2
}

func (n *Node) UnreachableResourcesRatio() float64 {
	if n.FreeResource.CPUs == 0 || n.FreeResource.Memory == 0 { // Impossible use this "free" resources.
		cpusRatio := float64(n.FreeResource.CPUs) / float64(n.MaxResources.CPUs)
		memoryRatio := float64(n.FreeResource.Memory) / float64(n.MaxResources.Memory)
		return (cpusRatio + memoryRatio) / 2
	}
	return 0
}

func (n *Node) TotalBandwidthUsedOnReceiving() float64 {
	return float64(n.MessagesReceivedSize)
}

func (n *Node) TotalMemoryUsed() float64 {
	return float64(n.MemoryUsed)
}

func (n *Node) TotalMessagesReceived() float64 {
	return float64(n.MessagesReceived)
}

func (n *Node) TotalRunRequestsSubmitted() float64 {
	return float64(n.RequestsSubmitted)
}

func (n *Node) TotalTraderActiveOffers() float64 {
	return float64(n.TraderActiveOffers)
}
