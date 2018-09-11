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
	"math/rand"
	"sync"
	"time"
)

const logRandFeederTag = "R-FEEDER"

// randomFeeder generates a stream of user requests using a pre-defined defined requests profile.
type randomFeeder struct {
	collector        *metrics.Collector           // Metrics collector that collects system level metrics.
	reqInjectionNode sync.Map                     // Map of ContainerID<->NodeIndex.
	randomGenerator  *rand.Rand                   // Pseudo-random generator.
	simConfigs       *configuration.Configuration // Simulator's configurations.
}

// newRandomFeeder creates a new random feeder.
func newRandomFeeder(simConfigs *configuration.Configuration, _ *caravelaConfigs.Configuration, rngSeed int64) (Feeder, error) {
	return &randomFeeder{
		collector:        nil,
		reqInjectionNode: sync.Map{},
		randomGenerator:  rand.New(caravelaUtil.NewSourceSafe(rand.NewSource(rngSeed))),
		simConfigs:       simConfigs,
	}, nil
}

func (rf *randomFeeder) Init(metricsCollector *metrics.Collector, _ types.Resources) {
	rf.collector = metricsCollector
}

func (rf *randomFeeder) Start(ticksChannel <-chan chan RequestTask) {
	runReqPerTick := int(float64(rf.simConfigs.NumberOfNodes) * float64(0.025)) // Send 2.5% of cluster size in requests per node
	stopReqPerTick := int(float64(rf.simConfigs.NumberOfNodes) * float64(0.017))
	tick := 0
	for {
		select {
		case newTickChan, more := <-ticksChannel: // Send all the requests for this tickChan
			if more {
				if tick > 10 && tick < 20 {
					for s := 0; s < stopReqPerTick; s++ { // Stop Containers Requests
						rf.reqInjectionNode.Range(func(key, value interface{}) bool {
							containerID, _ := key.(string)
							injectionNode, _ := value.(*node.Node)
							newTickChan <- func(_ int, _ *node.Node, _ time.Duration) {
								err := injectionNode.StopContainers(context.Background(), []string{containerID})
								if err != nil {
									//util.Log.Infof(util.LogTag(logRandFeederTag)+"Stop container FAILED, err: %s", err)
								}
								rf.reqInjectionNode.Delete(containerID)
							}
							return false
						})
					}
				}

				if tick <= 10 || tick >= 20 {
					for r := 0; r < runReqPerTick; r++ { // Run Container Requests
						newTickChan <- func(nodeIndex int, injectedNode *node.Node, currentTime time.Duration) {
							resources := rf.generateResourcesProfile() // Generate the resources necessary for the request.
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
								rf.reqInjectionNode.Store(contStatus[0].ContainerID, injectedNode)
								rf.collector.RunRequestSucceeded()
							}
							rf.collector.ArchiveRunRequest(requestID)
						}
					}
				}

				close(newTickChan) // No more user requests for this tick
			} else { // Simulator closed ticks channel
				return // Stop feeding engine
			}
		}
		tick++
	}
}

func (rf *randomFeeder) generateResourcesProfile() types.Resources {
	copyRequestProfiles := make([]requestProfile, len(requestProfiles))
	copy(copyRequestProfiles, requestProfiles)

	acc := 0
	for i, profile := range copyRequestProfiles {
		currentPercentage := profile.Percentage
		copyRequestProfiles[i].Percentage += acc
		acc += currentPercentage
	}
	if acc != 100 {
		panic(errors.New("random feeder profiles probability does not sum 100%"))
	}

	randProfile := rf.randomGenerator.Intn(101)
	for _, profile := range copyRequestProfiles {
		if randProfile <= profile.Percentage {
			return profile.Resources
		}
	}
	panic(fmt.Errorf("random feeder problem generating resources, rand profile: %d", randProfile))
}

type requestProfile struct {
	Resources  types.Resources
	Percentage int
}

var requestProfiles = []requestProfile{
	{
		Resources: types.Resources{
			CPUClass: 0,
			CPUs:     1,
			Memory:   350,
		},
		Percentage: 35,
	},
	{
		Resources: types.Resources{
			CPUClass: 0,
			CPUs:     2,
			Memory:   1024,
		},
		Percentage: 15,
	},
	{
		Resources: types.Resources{
			CPUClass: 0,
			CPUs:     4,
			Memory:   4048,
		},
		Percentage: 10,
	},
	{
		Resources: types.Resources{
			CPUClass: 1,
			CPUs:     2,
			Memory:   750,
		},
		Percentage: 20,
	},
	{
		Resources: types.Resources{
			CPUClass: 1,
			CPUs:     3,
			Memory:   1500,
		},
		Percentage: 10,
	},
	{
		Resources: types.Resources{
			CPUClass: 1,
			CPUs:     3,
			Memory:   2500,
		},
		Percentage: 10,
	},
}
