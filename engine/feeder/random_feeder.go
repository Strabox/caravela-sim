package feeder

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/engine/metrics"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	caravelaConfigs "github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/common/guid"
	caravelaUtil "github.com/strabox/caravela/util"
	"math"
	"math/rand"
	"sync"
	"time"
)

const logRandFeederTag = "R-FEEDER"

type containerRunning struct {
	containerID  string
	injectedNode *node.Node
}

type containersPerProfile struct {
	mutex             sync.Mutex
	containersRunning []*containerRunning
}

func newContainerPerProfile() *containersPerProfile {
	return &containersPerProfile{
		mutex:             sync.Mutex{},
		containersRunning: make([]*containerRunning, 0),
	}
}

func (c *containersPerProfile) AddRequest(containerRunning *containerRunning) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.containersRunning = append(c.containersRunning, containerRunning)
}

func (c *containersPerProfile) RemoveRequest() (*containerRunning, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.containersRunning) == 0 {
		return nil, errors.New("no request running")
	}

	res := c.containersRunning[len(c.containersRunning)-1]
	c.containersRunning = c.containersRunning[:len(c.containersRunning)-1]
	return res, nil
}

// randomFeeder generates a stream of user requests using a pre-defined defined requests profile.
type randomFeeder struct {
	collector            *metrics.Collector // Metrics collector that collects system level metrics.
	reqProfiles          map[int]*containersPerProfile
	randomGenerator      *rand.Rand                   // Pseudo-random generator.
	systemTotalResources types.Resources              // Caravela's maximum resources.
	simConfigs           *configuration.Configuration // Simulator's configurations.
}

// newRandomFeeder creates a new random feeder.
func newRandomFeeder(simConfigs *configuration.Configuration, _ *caravelaConfigs.Configuration, rngSeed int64) (Feeder, error) {
	return &randomFeeder{
		collector:       nil,
		reqProfiles:     make(map[int]*containersPerProfile),
		randomGenerator: rand.New(caravelaUtil.NewSourceSafe(rand.NewSource(rngSeed))),
		simConfigs:      simConfigs,
	}, nil
}

func (rf *randomFeeder) Init(metricsCollector *metrics.Collector, systemTotalResources types.Resources) {
	rf.collector = metricsCollector
	rf.systemTotalResources = systemTotalResources
	for i := range rf.simConfigs.RequestsProfile() {
		rf.reqProfiles[i] = newContainerPerProfile()
	}

}

func (rf *randomFeeder) Start(ticksChannel <-chan chan RequestTask) {
	totalResourcesSubmitted := types.Resources{CPUs: 0, Memory: 0}
	totalResourcesReleased := types.Resources{CPUs: 0, Memory: 0}
	submitRequests := rf.simConfigs.DeployRequestsRate()
	stopRequests := rf.simConfigs.StopRequestsRate()
	for i := range submitRequests {
		submitRequests[i] = float64(rf.simConfigs.NumberOfNodes) * (float64(submitRequests[i] / 100))
		stopRequests[i] = float64(rf.simConfigs.NumberOfNodes) * (float64(stopRequests[i] / 100))
	}
	superTicksSize := int(math.Ceil(float64(rf.simConfigs.MaximumTicks()) / float64(len(submitRequests))))

	currentSuperTick := 0
	tick := 0
	for {
		select {
		case newTickChan, more := <-ticksChannel: // Send all the requests for this tickChan
			if more {

				for r := 0; r < int(submitRequests[currentSuperTick]); r++ { // Run Container Requests

					profile, resources := rf.generateResourcesProfile() // Generate the resources necessary for the request.
					totalResourcesSubmitted.CPUs += resources.CPUs
					totalResourcesSubmitted.Memory += resources.Memory

					newTickChan <- func(nodeIndex int, injectedNode *node.Node, currentTime time.Duration) {
						requestID := guid.NewGUIDRandom().String() // Generate a GUID for tracking the request inside Caravela.
						requestCtx := context.WithValue(context.Background(), types.RequestIDKey, requestID)
						rf.collector.CreateRunRequest(nodeIndex, requestID, resources, currentTime)
						contStatus, err := injectedNode.SubmitContainers(
							requestCtx,
							[]types.ContainerConfig{{
								ImageKey:     util.RandomName(),
								Name:         util.RandomName(),
								PortMappings: caravela.EmptyPortMappings(),
								Args:         caravela.EmptyContainerArgs(),
								Resources:    resources,
								GroupPolicy:  types.SpreadGroupPolicy,
							}})
						if err == nil {
							rf.reqProfiles[profile].AddRequest(&containerRunning{containerID: contStatus[0].ContainerID, injectedNode: injectedNode})
							rf.collector.RunRequestSucceeded()
						}
						rf.collector.ArchiveRunRequest(requestID)
					}

				}

				for s := 0; s < int(stopRequests[currentSuperTick]); s++ { // Stop Containers Requests
					profile, resProfile := rf.generateResourcesProfile()

					newTickChan <- func(_ int, _ *node.Node, _ time.Duration) {
						containerToRemove, err := rf.reqProfiles[profile].RemoveRequest()
						if err == nil {
							err := containerToRemove.injectedNode.StopContainers(context.Background(), []string{containerToRemove.containerID})
							if err == nil {
								totalResourcesReleased.CPUs += resProfile.CPUs
								totalResourcesReleased.Memory += resProfile.Memory
							}
						}
					}
				}

				close(newTickChan) // No more user requests for this tick
			} else { // Simulator closed ticks channel
				util.Log.Infof(util.LogTag(logRandFeederTag)+"Total Resources Submitted: <%d,%d>", totalResourcesSubmitted.CPUs, totalResourcesSubmitted.Memory)
				util.Log.Infof(util.LogTag(logRandFeederTag)+"Total Resources Released:  <%d,%d>", totalResourcesReleased.CPUs, totalResourcesReleased.Memory)
				return // Stop feeding engine
			}
		}
		tick++
		if tick%superTicksSize == 0 {
			currentSuperTick++
		}
	}
}

// TODO
func (rf *randomFeeder) generateResourcesProfile() (int, types.Resources) {
	requestProfiles := rf.simConfigs.RequestsProfile()

	acc := 0
	for i, profile := range requestProfiles {
		currentPercentage := profile.Percentage
		requestProfiles[i].Percentage += acc
		acc += currentPercentage
	}
	if acc != 100 {
		panic(errors.New("random feeder request profiles probability does not sum 100%"))
	}

	randProfile := rf.randomGenerator.Intn(101)
	for i, profile := range requestProfiles {
		if randProfile <= profile.Percentage {
			return i, types.Resources{
				CPUClass: types.CPUClass(profile.CPUClass),
				CPUs:     profile.CPUs,
				Memory:   profile.Memory,
			}
		}
	}
	panic(fmt.Errorf("random feeder problem generating resources, rand profile: %d", randProfile))
}
