package chord

import (
	"context"
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/engine/metrics"
	"github.com/strabox/caravela-sim/mocks/caravela"
	overlayMock "github.com/strabox/caravela-sim/mocks/overlay"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	overlayTypes "github.com/strabox/caravela/overlay/types"
	"math"
	"sort"
)

// chordLogTag chord's mock log tag.
const chordLogTag = "SIM-CHORD"

// Mock mocks the interactions with a Chord overlay client simulating its functionality.
// All in memory, goroutine-safe.
type Mock struct {
	collector       *metrics.Collector // Metrics collector.
	numSpeedupNodes int

	numNodes      int // Initial number of nodes for the chord.
	numSuccessors int // Number of successors for each chord node.

	speedupNodes    []speedupNodeMock      // Array of speed up nodes.
	fakeNodesRing   []overlayMock.NodeMock // Array that represent the node's chord ring.
	nodesIdIndexMap map[string]int         // ID <-> Index.
	nodesIpIndexMap map[string]int         // IP <-> Index.
}

// NewChordMock creates a new chord overlay that can be used by an application component.
// It implements the github.com/strabox/caravela/node/external Overlay interface.
func NewChordMock(numNodes, numSuccessors, numSpeedupNodes int, metricsCollector *metrics.Collector) *Mock {
	return &Mock{
		collector:       metricsCollector,
		numSpeedupNodes: numSpeedupNodes,
		numNodes:        numNodes,
		numSuccessors:   numSuccessors,
		speedupNodes:    make([]speedupNodeMock, numSpeedupNodes),
		fakeNodesRing:   make([]overlayMock.NodeMock, numNodes),
		nodesIdIndexMap: make(map[string]int),
		nodesIpIndexMap: make(map[string]int),
	}
}

// Init initializes the chord's mock structure.
func (m *Mock) Init() {
	for i := 0; i < m.numNodes; i++ {
		m.fakeNodesRing[i] = *overlayMock.NewRandomNode()
	}

	sort.Sort(m) // Sort nodes by ID (ascending order)

	speedupDivSize := m.numNodes / m.numSpeedupNodes
	divSize := 1
	speedupNodeIndex := 0
	for i := 0; i < m.numNodes; i++ {
		m.nodesIdIndexMap[m.fakeNodesRing[i].String()] = i
		m.nodesIpIndexMap[m.fakeNodesRing[i].IP()] = i

		if speedupDivSize != 0 && divSize == speedupDivSize && speedupNodeIndex < m.numSpeedupNodes {
			m.speedupNodes[speedupNodeIndex] = *newSpeedupNodeMock(i, &m.fakeNodesRing[i])
			util.Log.Debugf(util.LogTag(chordLogTag)+"Speedup Node: %d, GUID: %s, ToNodeIndex: %d",
				speedupNodeIndex, m.speedupNodes[speedupNodeIndex].String(), i)
			speedupNodeIndex++
			divSize = 1
			continue
		}
		divSize++
	}

	m.Print()
}

func (m *Mock) Print() {
	util.Log.Debugf("##################################################################")
	util.Log.Debugf("#                   CHORD's MOCK CONFIGURATIONS                  #")
	util.Log.Debugf("##################################################################")
	util.Log.Debugf("#Nodes:              %d", m.numNodes)
	util.Log.Debugf("#Speedup Nodes:      %d", len(m.speedupNodes))
	util.Log.Debugf("Speedup Window Size: %d", m.numNodes/m.numSpeedupNodes)
	util.Log.Debugf("##################################################################")
}

func (m *Mock) GetNodeMockByIndex(index int) *overlayMock.NodeMock {
	return &m.fakeNodesRing[index]
}

func (m *Mock) GetNodeMockByGUID(guid string) (int, *overlayMock.NodeMock) {
	index := m.nodesIdIndexMap[guid]
	return index, &m.fakeNodesRing[index]
}

func (m *Mock) GetNodeMockByIP(ip string) (int, *overlayMock.NodeMock) {
	index, exist := m.nodesIpIndexMap[ip]
	if !exist {
		panic(errors.New("Node's IP does not exist"))
	}
	return index, &m.fakeNodesRing[index]
}

func (m *Mock) collectLookupMessages(ctx context.Context, targetNodeIndex int) {
	fromNodeIndex, _ := m.GetNodeMockByGUID(types.NodeGUID(ctx))
	distance := (fromNodeIndex - targetNodeIndex) % m.numNodes
	if distance < 0 {
		distance = -distance
	}
	if distance != 0 {
		m.collector.IncrMessagesTradedRequest(types.RequestID(ctx), int(math.Log2(float64(distance)))/2)
	}
}

// ===============================================================================
// =							  Overlay Interface                              =
// ===============================================================================

func (m *Mock) Create(_ context.Context, _ overlayTypes.OverlayMembership) error {
	// Do Nothing (Not necessary for the engine)
	return nil
}

func (m *Mock) Join(_ context.Context, _ string, _ int, _ overlayTypes.OverlayMembership) error {
	// Do Nothing (Not necessary for the engine)
	return nil
}

func (m *Mock) Lookup(ctx context.Context, key []byte) ([]*overlayTypes.OverlayNode, error) {
	searchMockNode := overlayMock.NewNodeBytes(key)

	// Simulate the lookup using the chord with the array

	startSearchIndex := 0

	// Speeding up the lookup using the special pointer nodes (Speedup Nodes)
	if len(m.fakeNodesRing) >= m.numSpeedupNodes {
		for index, node := range m.speedupNodes {
			if searchMockNode.Smaller(node.NodeMock) {
				if index == 0 {
					break
				} else {
					startSearchIndex = m.speedupNodes[index-1].index
					break
				}
			}
		}
	}

	// Regular sequential search for the node in the fake ring section.
	res := make([]*overlayTypes.OverlayNode, m.numSuccessors)
	resIndex := 0
	targetNodeIndex := -1
	for i := startSearchIndex; i < len(m.fakeNodesRing); i++ {
		currentNode := m.fakeNodesRing[i]
		if !currentNode.Smaller(searchMockNode) {
			if targetNodeIndex == -1 {
				targetNodeIndex = i
			}
			res[resIndex] = overlayTypes.NewOverlayNode(currentNode.IP(), caravela.FakePort, currentNode.Bytes())
			resIndex++
		}
		if resIndex == m.numSuccessors {
			m.collectLookupMessages(ctx, targetNodeIndex) // Collect metrics
			return res, nil
		}
	}

	startIndex := 0
	for ; resIndex < m.numSuccessors; resIndex++ {
		if targetNodeIndex == -1 {
			targetNodeIndex = startIndex
		}
		res[resIndex] = overlayTypes.NewOverlayNode(m.fakeNodesRing[startIndex].IP(), caravela.FakePort, m.fakeNodesRing[startIndex].Bytes())
		startIndex++
	}

	m.collectLookupMessages(ctx, targetNodeIndex) // Collect metrics
	return res, nil
}

func (m *Mock) Neighbors(_ context.Context, nodeID []byte) ([]*overlayTypes.OverlayNode, error) {
	res := make([]*overlayTypes.OverlayNode, 2)
	neighMockNode := overlayMock.NewNodeBytes(nodeID)
	index := m.nodesIdIndexMap[neighMockNode.String()]
	if index == 0 {
		res[0] = overlayTypes.NewOverlayNode(m.fakeNodesRing[m.numNodes-1].IP(), caravela.FakePort, m.fakeNodesRing[m.numNodes-1].Bytes())
		res[1] = overlayTypes.NewOverlayNode(m.fakeNodesRing[1].IP(), caravela.FakePort, m.fakeNodesRing[1].Bytes())
	} else if index == m.numNodes-1 {
		res[0] = overlayTypes.NewOverlayNode(m.fakeNodesRing[m.numNodes-2].IP(), caravela.FakePort, m.fakeNodesRing[m.numNodes-2].Bytes())
		res[1] = overlayTypes.NewOverlayNode(m.fakeNodesRing[0].IP(), caravela.FakePort, m.fakeNodesRing[0].Bytes())
	} else {
		res[0] = overlayTypes.NewOverlayNode(m.fakeNodesRing[index-1].IP(), caravela.FakePort, m.fakeNodesRing[index-1].Bytes())
		res[1] = overlayTypes.NewOverlayNode(m.fakeNodesRing[index+1].IP(), caravela.FakePort, m.fakeNodesRing[index+1].Bytes())
	}
	return res, nil
}

func (m *Mock) NodeID(_ context.Context) ([]byte, error) {
	// Do Nothing (Not necessary for the engine)
	return nil, nil
}

func (m *Mock) Leave(_ context.Context) error {
	// Do Nothing (Not necessary for the engine)
	return nil
}

// ===============================================================================
// =							  Sort Interface                                 =
// ===============================================================================

func (m *Mock) Len() int {
	return len(m.fakeNodesRing)
}

func (m *Mock) Swap(i, j int) {
	m.fakeNodesRing[i], m.fakeNodesRing[j] = m.fakeNodesRing[j], m.fakeNodesRing[i]
}

func (m *Mock) Less(i, j int) bool {
	return m.fakeNodesRing[i].Smaller(&m.fakeNodesRing[j])
}
