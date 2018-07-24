package metrics

import (
	"errors"
	"fmt"
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

func (metrics *Global) Init() {
	for index := range metrics.NodesMetrics {
		metrics.NodesMetrics[index] = *NewNode()
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

func (metrics *Global) CreateRunRequest(nodeIndex int, resources types.Resources,
	currentTime time.Duration) {
	newRequest := NewRunRequest(resources)
	metrics.RunRequestsSubmitted = append(metrics.RunRequestsSubmitted, *newRequest)
}

func (metrics *Global) MsgsTradedActiveRequest(numMsgs int) {
	request, err := metrics.activeRequest()
	if err == nil {
		request.IncrementMessagesTraded(int64(numMsgs))
	}
}

func (metrics *Global) RunRequestsAvgMsgs(numMessages int) float64 {
	accMsgsTrader := int64(0)
	for _, runRequest := range metrics.RunRequestsSubmitted {
		accMsgsTrader += runRequest.TotalMessagesTraded()
	}
	return float64(accMsgsTrader) / float64(len(metrics.RunRequestsSubmitted))
}

func (metrics *Global) TotalRunRequestsSucceeded() int64 {
	return metrics.RunRequestsSucceeded
}

func (metrics *Global) RunRequests() int64 {
	return int64(len(metrics.RunRequestsSubmitted))
}

func (metrics *Global) RequestSuccessRatio() float64 {
	return float64(metrics.RunRequestsSucceeded) / float64(metrics.RunRequests())
}

func (metrics *Global) activeRequest() (*RunRequest, error) {
	if len(metrics.RunRequestsSubmitted) == 0 {
		return nil, errors.New("no request active")
	}
	return &metrics.RunRequestsSubmitted[len(metrics.RunRequestsSubmitted)-1], nil
}

func (metrics *Global) Print() {
	fmt.Print("Msgs Traded: ")
	for _, runRequest := range metrics.RunRequestsSubmitted {
		fmt.Printf("%d, ", runRequest.TotalMessagesTraded())
	}
	fmt.Println()
}
