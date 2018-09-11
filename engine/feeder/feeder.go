package feeder

import (
	"github.com/strabox/caravela-sim/engine/metrics"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node"
	"time"
)

type RequestTask func(randNodeIndex int, randNode *node.Node, currentTime time.Duration)

type Feeder interface {
	Init(metricsCollector *metrics.Collector, systemTotalResources types.Resources)
	Start(ticksChannel <-chan chan RequestTask)
}
