package simulation

import (
	"github.com/strabox/caravela-sim/simulation/metrics"
	caravelaNode "github.com/strabox/caravela/node"
)

type Simulator interface {
	Init()
	Start()
	NodeByIP(ip string) (*caravelaNode.Node, int)
	NodeByGUID(guid string) (*caravelaNode.Node, int)
	Metrics() *metrics.Metrics
}
