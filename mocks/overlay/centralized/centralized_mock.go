package centralized

import (
	overlayMock "github.com/strabox/caravela-sim/mocks/overlay"
	"github.com/strabox/caravela-sim/simulation/metrics"
)

// Centralized's mock log tag
const centralizedLogTag = "SIM-CENTRA"

type Mock struct {
	collector *metrics.Collector // Metrics collector.

	numNodes int // Initial number of nodes for the chord.

	fakeNodesRing   []overlayMock.NodeMock // Array that represent the node's chord ring.
	nodesIdIndexMap map[string]int         // ID <-> Index.
	nodesIpIndexMap map[string]int         // IP <-> Index.
}

func NewChordMock(numNodes int, metricsCollector *metrics.Collector) *Mock {
	return &Mock{
		collector:       metricsCollector,
		numNodes:        numNodes,
		fakeNodesRing:   make([]overlayMock.NodeMock, numNodes),
		nodesIdIndexMap: make(map[string]int),
		nodesIpIndexMap: make(map[string]int),
	}
}
