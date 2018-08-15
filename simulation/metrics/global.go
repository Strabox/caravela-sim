package metrics

import (
	"github.com/strabox/caravela/api/types"
	"sync"
	"sync/atomic"
	"time"
)

// Global holds metrics information about the system's level metrics collected during a time window.
type Global struct {
	Start time.Duration `json:"StartTime"` // Start time of the collection.
	End   time.Duration `json:"EndTime"`   // End time of the collection.

	GetOffersRelayed     int64  `json:"GetOffersRelayed"`
	RunRequestsSucceeded int64  `json:"RunRequestsSucceeded"` // Number of run requests that were successful deployed.
	NodesMetrics         []Node `json:"NodesMetrics"`         // Metrics collected for each system's node.

	RunRequestsAggregator  sync.Map     `json:"-"`
	RunRequestsCompleted   []RunRequest `json:"RunRequestsCompleted"`
	requestsCompletedMutex sync.Mutex   `json:"-"`
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
	}

	for index := range res.NodesMetrics {
		newNode := NewNode(nodesMaxRes[index])
		res.NodesMetrics[index] = *newNode
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
		newNode := NewNode(prevGlobal.NodesMetrics[index].MaximumResources())
		res.NodesMetrics[index] = *newNode
	}

	return res
}

// ========================= Metrics Collector Methods ====================================

func (global *Global) GetOfferRelayed(amount int64) {
	atomic.AddInt64(&global.GetOffersRelayed, amount)
}

func (global *Global) RunRequestSucceeded() {
	atomic.AddInt64(&global.RunRequestsSucceeded, 1)
}

func (global *Global) CreateRunRequest(nodeIndex int, requestID string, resources types.Resources,
	currentTime time.Duration) {

	newRunRequest := NewRunRequest(resources)
	newRunRequest.IncrMessagesTraded(1)
	global.RunRequestsAggregator.Store(requestID, newRunRequest)

	global.NodesMetrics[nodeIndex].RunRequestSubmitted()
}

func (global *Global) IncrMessagesTradedRequest(requestID string, numMessages int) {
	if req, exist := global.RunRequestsAggregator.Load(requestID); exist {
		if request, ok := req.(*RunRequest); ok {
			request.IncrMessagesTraded(int64(numMessages))
		}
	}
}

func (global *Global) ArchiveRunRequest(requestID string) {
	if req, exist := global.RunRequestsAggregator.Load(requestID); exist {
		if request, ok := req.(*RunRequest); ok {
			global.RunRequestsAggregator.Delete(requestID)

			global.requestsCompletedMutex.Lock()
			defer global.requestsCompletedMutex.Unlock()
			global.RunRequestsCompleted = append(global.RunRequestsCompleted, *request)
		}
	}
}

// ========================= Derived/Calculated Metrics ====================================

// RunRequestSuccessRatio returns the request success ratio for all the requests during this collection.
func (global *Global) RunRequestSuccessRatio() float64 {
	return float64(global.RunRequestsSucceeded) / float64(global.TotalRunRequests())
}

func (global *Global) RunRequestsAvgMessages() float64 {
	accMsgsTrader := int64(0)
	for _, runRequest := range global.RunRequestsCompleted {
		accMsgsTrader += runRequest.TotalMessagesTraded()
	}
	return float64(accMsgsTrader) / float64(len(global.RunRequestsCompleted))
}

func (global *Global) AllAvailableResourcesAvg() float64 {
	maxCPUsAvailable, maxRAMAvailable, CPUsAvailable, RAMAvailable := 0, 0, 0, 0
	for _, nodeMetrics := range global.NodesMetrics {
		maxCPUsAvailable += nodeMetrics.MaximumResources().CPUs
		maxRAMAvailable += nodeMetrics.MaximumResources().RAM
		CPUsAvailable += nodeMetrics.AvailableResources().CPUs
		RAMAvailable += nodeMetrics.AvailableResources().RAM
	}

	percentageCPUAvailable := float64(CPUsAvailable) / float64(maxCPUsAvailable)
	percentageRAMAvailable := float64(RAMAvailable) / float64(maxRAMAvailable)
	return (percentageCPUAvailable + percentageRAMAvailable) / 2
}

// ============================ Getters and Setters ========================================

func (global *Global) StartTime() time.Duration {
	return global.Start
}

func (global *Global) EndTime() time.Duration {
	return global.End
}

func (global *Global) SetEndTime(endTime time.Duration) {
	global.End = endTime
}

func (global *Global) TotalGetOffersRelayed() int64 {
	return global.GetOffersRelayed
}

func (global *Global) TotalRunRequestsSucceeded() int64 {
	return global.RunRequestsSucceeded
}

func (global *Global) TotalRunRequests() int64 {
	return int64(len(global.RunRequestsCompleted))
}

func (global *Global) SetAvailableNodeResources(nodeIndex int, res types.Resources) {
	global.NodesMetrics[nodeIndex].SetAvailableResources(res)
}

func (global *Global) APIRequestReceived(nodeIndex int) {
	if len(global.NodesMetrics) > nodeIndex {
		global.NodesMetrics[nodeIndex].APIRequestReceived()
	}
}
