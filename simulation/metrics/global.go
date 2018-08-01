package metrics

import (
	"errors"
	"github.com/strabox/caravela/api/types"
	"time"
)

// Global holds metrics information about the system's level metrics collected during a time window.
type Global struct {
	StartTime time.Duration `json:"StartTime"` // Start time of the collection.
	EndTime   time.Duration `json:"EndTime"`   // End time of the collection.

	RunRequestsSucceeded int64        `json:"RunRequestsSucceeded"` // Number of run requests that were successful deployed.
	NodesMetrics         []Node       `json:"NodesMetrics"`         // Metrics collected for each system's node.
	RunRequestsSubmitted []RunRequest `json:"RunRequestsSubmitted"` // Run requests submitted during the global collection.
}

// NewGlobalInitial returns a structure to hold the first collection of metrics.
func NewGlobalInitial(numNodes int, startTime time.Duration, nodesMaxRes []types.Resources) *Global {
	res := &Global{
		StartTime: startTime,

		RunRequestsSucceeded: 0,
		NodesMetrics:         make([]Node, numNodes),
		RunRequestsSubmitted: make([]RunRequest, 0),
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
		StartTime: prevGlobal.EndTimes(),

		RunRequestsSucceeded: 0,
		NodesMetrics:         make([]Node, numNodes),
		RunRequestsSubmitted: make([]RunRequest, 0),
	}

	for index := range prevGlobal.NodesMetrics {
		newNode := NewNode(prevGlobal.NodesMetrics[index].MaximumResources())
		res.NodesMetrics[index] = *newNode
	}

	return res
}

// ========================= Metrics Collector Methods ====================================

func (global *Global) RunRequestSucceeded() {
	global.RunRequestsSucceeded++
}

func (global *Global) IncrMessagesTradedRequest(numMessages int) {
	request, err := global.activeRequest()
	if err == nil {
		request.IncrMessagesTraded(int64(numMessages))
	}
}

func (global *Global) CreateRunRequest(nodeIndex int, resources types.Resources,
	currentTime time.Duration) {
	newRequest := NewRunRequest(resources)
	global.RunRequestsSubmitted = append(global.RunRequestsSubmitted, *newRequest)
}

func (global *Global) activeRequest() (*RunRequest, error) {
	if len(global.RunRequestsSubmitted) == 0 {
		return nil, errors.New("no request active")
	}
	return &global.RunRequestsSubmitted[len(global.RunRequestsSubmitted)-1], nil
}

// ========================= Derived/Calculated Metrics ====================================

// RunRequestSuccessRatio returns the request success ratio for all the requests during this collection.
func (global *Global) RunRequestSuccessRatio() float64 {
	return float64(global.RunRequestsSucceeded) / float64(global.TotalRunRequests())
}
func (global *Global) RunRequestsAvgMessages() float64 {
	accMsgsTrader := int64(0)
	for _, runRequest := range global.RunRequestsSubmitted {
		accMsgsTrader += runRequest.TotalMessagesTraded()
	}
	return float64(accMsgsTrader) / float64(len(global.RunRequestsSubmitted))
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

func (global *Global) StartTimes() time.Duration {
	return global.StartTime
}

func (global *Global) EndTimes() time.Duration {
	return global.EndTime
}

func (global *Global) SetEndTime(endTime time.Duration) {
	global.EndTime = endTime
}

func (global *Global) TotalRunRequestsSucceeded() int64 {
	return global.RunRequestsSucceeded
}

func (global *Global) TotalRunRequests() int64 {
	return int64(len(global.RunRequestsSubmitted))
}

func (global *Global) SetAvailableNodeResources(nodeIndex int, res types.Resources) {
	global.NodesMetrics[nodeIndex].SetAvailableResources(res)
}

func (global *Global) APIRequestReceived(nodeIndex int) {
	if len(global.NodesMetrics) > nodeIndex {
		global.NodesMetrics[nodeIndex].APIRequestReceived()
	}
}
