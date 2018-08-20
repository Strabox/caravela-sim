package feeder

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/simulation/metrics"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/common/guid"
	"time"
)

type RandomFeeder struct {
	collector  *metrics.Collector
	simConfigs *configuration.Configuration
}

func newRandomFeeder(simConfigs *configuration.Configuration) (Feeder, error) {
	return &RandomFeeder{
		collector:  nil,
		simConfigs: simConfigs,
	}, nil
}

func (rf *RandomFeeder) Init(metricsCollector *metrics.Collector) {
	rf.collector = metricsCollector
}

func (rf *RandomFeeder) Start(ticksChannel <-chan chan RequestTask) {
	runReqPerTick := int(float64(rf.simConfigs.NumberOfNodes) * float64(0.025)) // Send 2.5% of cluster size in requests per node
	for {
		select {
		case newTickChan, more := <-ticksChannel: // Send all the requests for this tickChan
			if more {
				for i := 0; i < runReqPerTick; i++ {
					newTickChan <- func(randNodeIndex int, randNode *node.Node, currentTime time.Duration) {
						resources := rf.generateResourcesProfile() // Generate the resources necessary for the request.
						requestID := guid.NewGUIDRandom().String() // Generate a UUID for tracking the request inside Caravela.
						requestContext := context.WithValue(context.Background(), types.RequestIDKey, requestID)
						rf.collector.CreateRunRequest(randNodeIndex, requestID, resources, currentTime)
						err := randNode.SubmitContainers(
							requestContext,
							[]types.ContainerConfig{
								{
									ImageKey:     util.RandomName(),
									Name:         util.RandomName(),
									PortMappings: caravela.EmptyPortMappings(),
									Args:         caravela.EmptyContainerArgs(),
									Resources:    resources,
								},
							})
						if err == nil {
							rf.collector.RunRequestSucceeded()
						}
						rf.collector.ArchiveRunRequest(requestID)
					}
				}
				close(newTickChan) // No more requests for this tick
			} else { // Simulator closed ticks channel
				return // Stop feeding simulation
			}
		}
	}
}

func (rf *RandomFeeder) generateResourcesProfile() types.Resources {
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

	randProfile := util.RandomInteger(1, 100)
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
			CPUs: 1,
			RAM:  256,
		},
		Percentage: 50,
	},
	{
		Resources: types.Resources{
			CPUs: 2,
			RAM:  800,
		},
		Percentage: 35,
	},
	{
		Resources: types.Resources{
			CPUs: 3,
			RAM:  1500,
		},
		Percentage: 10,
	},
	{
		Resources: types.Resources{
			CPUs: 3,
			RAM:  2500,
		},
		Percentage: 5,
	},
}
