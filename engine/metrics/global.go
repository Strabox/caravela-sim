package metrics

import (
	"github.com/strabox/caravela/api/types"
	"sync"
	"sync/atomic"
	"time"
)

type Resources struct {
	CPUClass int64
	CPUs     int64
	Memory   int64
}

// Global holds metrics information about the system's level metrics collected during a time window.
type Global struct {
	Start time.Duration `json:"StartTime"` // Start time of the collection.
	End   time.Duration `json:"EndTime"`   // End time of the collection.

	NodesMetrics []Node `json:"NodesMetrics"` // Metrics collected for each system's node.

	RunRequestsSucceeded   int64        `json:"RunRequestsSucceeded"` // Number of run requests that were successful deployed.
	RunRequestsAggregator  sync.Map     `json:"-"`
	RunRequestsCompleted   []RunRequest `json:"RunRequestsCompleted"`
	requestsCompletedMutex sync.Mutex   `json:"-"`

	ResourcesRequested Resources `json:"ResourcesRequested"`
	ResourcesAllocated Resources `json:"ResourcesAllocated"`

	// Debug Performance Metrics
	GetOffersRelayed       int64 `json:"GetOffersRelayed"`
	EmptyGetOffersMessages int64 `json:"EmptyGetOffersMessages"`
}

// NewGlobalInitial returns a structure to hold the first collection of metrics.
func NewGlobalInitial(numNodes int, startTime time.Duration, nodesMaxRes []types.Resources) *Global {
	res := &Global{
		Start: startTime,

		GetOffersRelayed:     0,
		RunRequestsSucceeded: 0,
		NodesMetrics:         make([]Node, numNodes),

		RunRequestsAggregator:  sync.Map{},
		RunRequestsCompleted:   make([]RunRequest, 0),
		requestsCompletedMutex: sync.Mutex{},

		ResourcesRequested: Resources{CPUClass: 0, CPUs: 0, Memory: 0},
		ResourcesAllocated: Resources{CPUClass: 0, CPUs: 0, Memory: 0},
	}

	for index := range res.NodesMetrics {
		res.NodesMetrics[index] = *NewNode(nodesMaxRes[index])
	}

	return res
}

// NewGlobalNext returns a structure to hold the subsequent collection of metrics based on the
// the previous window nodes.
func NewGlobalNext(numNodes int, prevGlobal *Global) *Global {
	res := &Global{
		Start: prevGlobal.EndTime(),

		RunRequestsSucceeded: 0,
		NodesMetrics:         make([]Node, numNodes),

		RunRequestsAggregator:  sync.Map{},
		RunRequestsCompleted:   make([]RunRequest, 0),
		requestsCompletedMutex: sync.Mutex{},
	}

	for index := range prevGlobal.NodesMetrics {
		res.NodesMetrics[index] = *NewNode(prevGlobal.NodesMetrics[index].MaximumResources())
	}

	return res
}

// ========================= Metrics Collector Methods ====================================

func (g *Global) MessageReceived(nodeIndex int, amount int64, requestSizeBytes int64) {
	g.NodesMetrics[nodeIndex].MessageReceived(amount, requestSizeBytes)
}

func (g *Global) GetOfferRelayed(amount int64) {
	atomic.AddInt64(&g.GetOffersRelayed, amount)
}

func (g *Global) EmptyGetOfferMessages(amount int64) {
	atomic.AddInt64(&g.EmptyGetOffersMessages, amount)
}

func (g *Global) CreateRunRequest(nodeIndex int, requestID string, resources types.Resources) {
	newRunRequest := NewRunRequest(resources)
	newRunRequest.IncrMessagesExchanged(1)
	g.RunRequestsAggregator.Store(requestID, newRunRequest)

	atomic.AddInt64(&g.ResourcesRequested.CPUs, int64(resources.CPUs))
	atomic.AddInt64(&g.ResourcesRequested.Memory, int64(resources.Memory))

	g.NodesMetrics[nodeIndex].RunRequestSubmitted()
}

func (g *Global) IncrMessagesTradedRequest(requestID string, numMessages int) {
	if req, exist := g.RunRequestsAggregator.Load(requestID); exist {
		if request, ok := req.(*RunRequest); ok {
			request.IncrMessagesExchanged(int64(numMessages))
		}
	}
}

func (g *Global) ArchiveRunRequest(requestID string, succeeded bool) {
	if req, exist := g.RunRequestsAggregator.Load(requestID); exist {
		if request, ok := req.(*RunRequest); ok {
			if succeeded {
				atomic.AddInt64(&g.ResourcesAllocated.CPUs, int64(request.ResourcesRequested().CPUs))
				atomic.AddInt64(&g.ResourcesAllocated.Memory, int64(request.ResourcesRequested().Memory))
				atomic.AddInt64(&g.RunRequestsSucceeded, 1)
			}

			g.RunRequestsAggregator.Delete(requestID)

			g.requestsCompletedMutex.Lock()
			defer g.requestsCompletedMutex.Unlock()
			g.RunRequestsCompleted = append(g.RunRequestsCompleted, *request)
		}
	}
}

// ========================= Derived/Calculated Metrics ====================================

// RunRequestSuccessRatio returns the request success ratio for all the requests during this collection.
func (g *Global) RunRequestSuccessRatio() float64 {
	if g.TotalRunRequests() == 0 {
		return 0
	}
	return float64(g.RunRequestsSucceeded) / float64(g.TotalRunRequests())
}

func (g *Global) RunRequestsAvgMessages() float64 {
	if len(g.RunRequestsCompleted) == 0 {
		return 0
	}
	accMessages := int64(0)
	for _, runRequest := range g.RunRequestsCompleted {
		accMessages += runRequest.TotalMessagesExchanged()
	}
	return float64(accMessages) / float64(len(g.RunRequestsCompleted))
}

func (g *Global) TotalUsedResourcesAvg() float64 {
	result := float64(0)
	numOfNodesCalculated := float64(0)
	for _, nodeMetrics := range g.NodesMetrics {
		numOfNodesCalculated++
		if numOfNodesCalculated == 0 {
			result = nodeMetrics.UsedResourcesRatio()
		} else {
			result = (result*(numOfNodesCalculated-1) + nodeMetrics.UsedResourcesRatio()) / numOfNodesCalculated
		}
	}
	return result
}

func (g *Global) TotalFreeResourcesAvg() float64 {
	result := float64(0)
	numOfNodesCalculated := float64(0)
	for _, nodeMetrics := range g.NodesMetrics {
		numOfNodesCalculated++
		if numOfNodesCalculated == 0 {
			result = nodeMetrics.FreeResourcesRatio()
		} else {
			result = (result*(numOfNodesCalculated-1) + nodeMetrics.FreeResourcesRatio()) / numOfNodesCalculated
		}
	}
	return result
}

func (g *Global) MessagesExchangedByRequest() []float64 {
	resTotalMessages := make([]float64, len(g.RunRequestsCompleted))
	for i := range resTotalMessages {
		resTotalMessages[i] = float64(g.RunRequestsCompleted[i].TotalMessagesExchanged())
	}
	return resTotalMessages
}

func (g *Global) TotalTraderActiveOfferPerNode() []float64 {
	resTotalActiveOffers := make([]float64, len(g.NodesMetrics))
	for i := range resTotalActiveOffers {
		resTotalActiveOffers[i] = float64(g.NodesMetrics[i].TotalTraderActiveOffers())
	}
	return resTotalActiveOffers
}

func (g *Global) ResourcesUnreachableRatioNode() []float64 {
	res := make([]float64, len(g.NodesMetrics))
	for i, nodeMetric := range g.NodesMetrics {
		res[i] = nodeMetric.UnreachableResourcesRatio()
	}
	return res
}

func (g *Global) ResourcesUsedNodeRatio() []float64 {
	res := make([]float64, len(g.NodesMetrics))
	for i, nodeMetric := range g.NodesMetrics {
		res[i] = nodeMetric.UsedResourcesRatio()
	}
	return res
}

func (g *Global) TotalMemoryUsedByNode() []float64 {
	res := make([]float64, len(g.NodesMetrics))
	for i, nodeMetric := range g.NodesMetrics {
		res[i] = float64(nodeMetric.TotalMemoryUsed())
	}
	return res
}

func (g *Global) TotalBandwidthUsedOnReceivingByNode() []float64 {
	res := make([]float64, len(g.NodesMetrics))
	for i, nodeMetric := range g.NodesMetrics {
		res[i] = float64(nodeMetric.TotalBandwidthUsedOnReceiving())
	}
	return res
}

func (g *Global) TotalMessagesReceivedByNode() []float64 {
	res := make([]float64, len(g.NodesMetrics))
	for i, nodeMetric := range g.NodesMetrics {
		res[i] = float64(nodeMetric.TotalMessagesReceived())
	}
	return res
}

func (g *Global) TotalMessagesReceivedByAllNodes() float64 {
	acc := float64(0)
	for _, nodeMetric := range g.NodesMetrics {
		acc += nodeMetric.TotalMessagesReceived()
	}
	return float64(acc)
}

func (g *Global) TotalMessagesReceivedByMasterNode() float64 {
	totalMessagesReceived := float64(0)
	for _, nodeMetric := range g.NodesMetrics {
		nodeMessagesReceived := nodeMetric.TotalMessagesReceived()
		if nodeMessagesReceived > totalMessagesReceived {
			totalMessagesReceived = nodeMessagesReceived
		}
	}
	return totalMessagesReceived
}

func (g *Global) ResourcesAllocationEfficiency() float64 {
	cpuRatio := float64(g.ResourcesAllocated.CPUs) / float64(g.ResourcesRequested.CPUs)
	memoryRatio := float64(g.ResourcesAllocated.Memory) / float64(g.ResourcesRequested.Memory)
	return (cpuRatio + memoryRatio) / 2
}

// ================================= Getters and Setters =================================

func (g *Global) StartTime() time.Duration {
	return g.Start
}

func (g *Global) EndTime() time.Duration {
	return g.End
}

func (g *Global) SetEndTime(endTime time.Duration) {
	g.End = endTime
}

func (g *Global) TotalGetOffersRelayed() int64 {
	return g.GetOffersRelayed
}

func (g *Global) TotalEmptyGetOfferMessages() int64 {
	return g.EmptyGetOffersMessages
}

func (g *Global) TotalRunRequestsSucceeded() int64 {
	return g.RunRequestsSucceeded
}

func (g *Global) TotalRunRequests() int64 {
	return int64(len(g.RunRequestsCompleted))
}

func (g *Global) TotalResourcesRequested() types.Resources {
	return types.Resources{CPUs: int(g.ResourcesRequested.CPUs), Memory: int(g.ResourcesRequested.Memory)}
}

func (g *Global) TotalResourcesAllocated() types.Resources {
	return types.Resources{CPUs: int(g.ResourcesAllocated.CPUs), Memory: int(g.ResourcesAllocated.Memory)}
}

func (g *Global) SetNodeState(nodeIndex int, freeResources types.Resources, traderActiveOffers int64, memoryUsed int64) {
	if len(g.NodesMetrics) > nodeIndex {
		g.NodesMetrics[nodeIndex].SetNodeState(freeResources, traderActiveOffers, memoryUsed)
	}
}

// ===================================== Sort Interface =======================================
// Order the node's metrics by ascending Maximum ResRequested.

func (g *Global) Len() int {
	return len(g.NodesMetrics)
}

func (g *Global) Swap(i, j int) {
	g.NodesMetrics[i], g.NodesMetrics[j] = g.NodesMetrics[j], g.NodesMetrics[i]
}

func (g *Global) Less(i, j int) bool {
	iMaxRes := g.NodesMetrics[i].MaxResources
	jMaxRes := g.NodesMetrics[j].MaxResources

	if iMaxRes.CPUClass < jMaxRes.CPUClass {
		return true
	} else if iMaxRes.CPUClass == jMaxRes.CPUClass {
		if iMaxRes.CPUs < jMaxRes.CPUs {
			return true
		} else if iMaxRes.CPUs == jMaxRes.CPUs {
			if iMaxRes.Memory <= jMaxRes.Memory {
				return true
			}
		}
	}

	return false
}
