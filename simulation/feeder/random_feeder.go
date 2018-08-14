package feeder

import (
	"context"
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
	runReqPerTick := int(float64(rf.simConfigs.NumberOfNodes) * float64(0.1)) // Send 10% of cluster size in requests per node

	for {
		select {
		case newTickChan, more := <-ticksChannel: // Send all the requests for this tickChan
			if more {
				for i := 0; i < runReqPerTick; i++ {
					newTickChan <- func(randNodeIndex int, randNode *node.Node, currentTime time.Duration) {
						res := types.Resources{CPUs: 1, RAM: 250}

						requestID := guid.NewGUIDRandom().String()
						requestContext := context.WithValue(context.Background(), types.RequestIDKey, requestID)
						rf.collector.CreateRunRequest(randNodeIndex, requestID, res, currentTime)
						err := randNode.SubmitContainers(
							requestContext,
							[]types.ContainerConfig{
								{
									ImageKey:     util.RandomName(),
									Name:         util.RandomName(),
									PortMappings: caravela.EmptyPortMappings(),
									Args:         caravela.EmptyContainerArgs(),
									Resources: types.Resources{
										CPUs: res.CPUs,
										RAM:  res.RAM,
									},
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
