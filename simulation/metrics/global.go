package metrics

import (
	"errors"
	"github.com/strabox/caravela/api/types"
	"time"
)

// Global holds metrics information about the system metrics gathered during a time window.
type Global struct {
	StartTime time.Duration `json:"StartTime"`
	EndTime   time.Duration `json:"EndTime"`

	RunRequestsSucceeded int64        `json:"RunRequestsSucceeded"`
	NodesMetrics         []Node       `json:"NodesMetrics"`
	RunRequestsSubmitted []RunRequest `json:"RunRequestsSubmitted"`
}

func NewGlobal(numNodes int, startTime time.Duration) *Global {
	return &Global{
		StartTime: startTime,

		RunRequestsSucceeded: 0,
		NodesMetrics:         make([]Node, numNodes),
		RunRequestsSubmitted: make([]RunRequest, 0),
	}
}

func (metrics *Global) Init(nodesMaxRes []types.Resources) {
	for index := range metrics.NodesMetrics {
		newNode := NewNode(nodesMaxRes[index])
		metrics.NodesMetrics[index] = *newNode
	}
}

func (metrics *Global) StartTimes() time.Duration {
	return metrics.StartTime
}

func (metrics *Global) SetEndTime(endTime time.Duration) {
	metrics.EndTime = endTime
}

func (metrics *Global) RunRequestSucceeded() {
	metrics.RunRequestsSucceeded++
}

func (metrics *Global) APIRequestReceived(nodeIndex int) {
	if len(metrics.NodesMetrics) > nodeIndex {
		metrics.NodesMetrics[nodeIndex].APIRequestReceived()
	}
}

func (metrics *Global) MsgsTradedActiveRequest(numMsgs int) {
	request, err := metrics.activeRequest()
	if err == nil {
		request.IncrementMessagesTraded(int64(numMsgs))
	}
}

func (metrics *Global) RunRequestsAvgMsgs() float64 {
	accMsgsTrader := int64(0)
	for _, runRequest := range metrics.RunRequestsSubmitted {
		accMsgsTrader += runRequest.TotalMessagesTraded()
	}
	return float64(accMsgsTrader) / float64(len(metrics.RunRequestsSubmitted))
}

func (metrics *Global) AllAvailableResourcesAvg() float64 {
	maxCPUsAvailable, maxRAMAvailable, totalCPUsAvailable, totalRAMAvailable := 0, 0, 0, 0
	for _, nodeMetrics := range metrics.NodesMetrics {
		maxCPUsAvailable += nodeMetrics.MaximumResources().CPUs
		maxRAMAvailable += nodeMetrics.MaximumResources().RAM
		totalCPUsAvailable += nodeMetrics.AvailableResources().CPUs
		totalRAMAvailable += nodeMetrics.AvailableResources().RAM
	}

	percentageCPUAvailable := float64(totalCPUsAvailable) / float64(maxCPUsAvailable)
	percentageRAMAvailable := float64(totalRAMAvailable) / float64(maxRAMAvailable)
	return (percentageCPUAvailable + percentageRAMAvailable) / 2
}

func (metrics *Global) TotalRunRequestsSucceeded() int64 {
	return metrics.RunRequestsSucceeded
}

func (metrics *Global) PercentageRunRequestsSucceeded() float64 {
	return float64(metrics.RunRequestsSucceeded) / float64(len(metrics.RunRequestsSubmitted))
}

func (metrics *Global) TotalRunRequests() int64 {
	return int64(len(metrics.RunRequestsSubmitted))
}

func (metrics *Global) SetAvailableNodeResources(nodeIndex int, res types.Resources) {
	metrics.NodesMetrics[nodeIndex].SetAvailableResources(res)
}

func (metrics *Global) RequestSuccessRatio() float64 {
	return float64(metrics.RunRequestsSucceeded) / float64(metrics.TotalRunRequests())
}

func (metrics *Global) CreateRunRequest(nodeIndex int, resources types.Resources,
	currentTime time.Duration) {
	newRequest := NewRunRequest(resources)
	metrics.RunRequestsSubmitted = append(metrics.RunRequestsSubmitted, *newRequest)
}

func (metrics *Global) activeRequest() (*RunRequest, error) {
	if len(metrics.RunRequestsSubmitted) == 0 {
		return nil, errors.New("no request active")
	}
	return &metrics.RunRequestsSubmitted[len(metrics.RunRequestsSubmitted)-1], nil
}
