package chord

import (
	"context"
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/mocks/caravela"
	overlayMock "github.com/strabox/caravela-sim/mocks/overlay"
	"github.com/strabox/caravela-sim/simulation/metrics"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	overlayTypes "github.com/strabox/caravela/overlay/types"
	"math"
	"sort"
)

// chordLogTag chord's mock log tag.
const chordLogTag = "SIM-CHORD"

// numSpeedupNodes amount of nodes to speed up the simulation of a lookup request.
const numSpeedupNodes = 200

// Mock mocks the interactions with a Chord overlay client simulating its functionality.
// All in memory, goroutine-safe.
type Mock struct {
	collector *metrics.Collector // Metrics collector.

	numNodes      int // Initial number of nodes for the chord.
	numSuccessors int // Number of successors for each chord node.

	speedupNodes    []speedupNodeMock      // Array of speed up nodes.
	fakeNodesRing   []overlayMock.NodeMock // Array that represent the node's chord ring.
	nodesIdIndexMap map[string]int         // ID <-> Index.
	nodesIpIndexMap map[string]int         // IP <-> Index.
}

// NewChordMock creates a new chord overlay that can be used by an application component.
// It implements the github.com/strabox/caravela/node/external Overlay interface.
func NewChordMock(numNodes, numSuccessors int, metricsCollector *metrics.Collector) *Mock {
	return &Mock{
		collector:       metricsCollector,
		numNodes:        numNodes,
		numSuccessors:   numSuccessors,
		speedupNodes:    make([]speedupNodeMock, numSpeedupNodes),
		fakeNodesRing:   make([]overlayMock.NodeMock, numNodes),
		nodesIdIndexMap: make(map[string]int),
		nodesIpIndexMap: make(map[string]int),
	}
}

// Init initializes the chord's mock structure.
func (chord *Mock) Init() {
	for i := 0; i < chord.numNodes; i++ {
		chord.fakeNodesRing[i] = *overlayMock.NewRandomNode()
	}

	sort.Sort(chord) // Sort nodes by ID (ascending order)

	speedupDivSize := chord.numNodes / numSpeedupNodes
	divSize := 1
	speedupNodeIndex := 0
	for i := 0; i < chord.numNodes; i++ {
		chord.nodesIdIndexMap[chord.fakeNodesRing[i].String()] = i
		chord.nodesIpIndexMap[chord.fakeNodesRing[i].IP()] = i

		if speedupDivSize != 0 && divSize == speedupDivSize && speedupNodeIndex < numSpeedupNodes {
			chord.speedupNodes[speedupNodeIndex] = *newSpeedupNodeMock(i, &chord.fakeNodesRing[i])
			util.Log.Debugf(util.LogTag(chordLogTag)+"Speedup Node: %d, GUID: %s, ToNodeIndex: %d",
				speedupNodeIndex, chord.speedupNodes[speedupNodeIndex].String(), i)
			speedupNodeIndex++
			divSize = 1
			continue
		}
		divSize++
	}

	chord.Print()
}

func (chord *Mock) Print() {
	util.Log.Debugf("##################################################################")
	util.Log.Debugf("#                   CHORD's MOCK CONFIGURATIONS                  #")
	util.Log.Debugf("##################################################################")
	util.Log.Debugf("#Nodes:              %d", chord.numNodes)
	util.Log.Debugf("#Speedup Nodes:      %d", len(chord.speedupNodes))
	util.Log.Debugf("Speedup Window Size: %d", chord.numNodes/numSpeedupNodes)
	util.Log.Debugf("##################################################################")
}

func (chord *Mock) GetNodeMockByIndex(index int) *overlayMock.NodeMock {
	return &chord.fakeNodesRing[index]
}

func (chord *Mock) GetNodeMockByGUID(guid string) (int, *overlayMock.NodeMock) {
	index := chord.nodesIdIndexMap[guid]
	return index, &chord.fakeNodesRing[index]
}

func (chord *Mock) GetNodeMockByIP(ip string) (int, *overlayMock.NodeMock) {
	index, exist := chord.nodesIpIndexMap[ip]
	if !exist {
		panic(errors.New("Node's IP does not exist"))
	}
	return index, &chord.fakeNodesRing[index]
}

func (chord *Mock) collectLookupMessages(ctx context.Context, targetNodeIndex int) {
	fromNodeIndex, _ := chord.GetNodeMockByGUID(types.NodeGUID(ctx))
	distance := (fromNodeIndex - targetNodeIndex) % chord.numNodes
	if distance < 0 {
		distance = -distance
	}
	if distance != 0 {
		chord.collector.IncrMessagesTradedRequest(types.RequestID(ctx), int(math.Log2(float64(distance)))/2)
	}
}

// ===============================================================================
// =							  Overlay Interface                              =
// ===============================================================================

func (chord *Mock) Create(_ context.Context, _ overlayTypes.OverlayMembership) error {
	// Do Nothing (Not necessary for the simulation)
	return nil
}

func (chord *Mock) Join(_ context.Context, _ string, _ int, _ overlayTypes.OverlayMembership) error {
	// Do Nothing (Not necessary for the simulation)
	return nil
}

func (chord *Mock) Lookup(ctx context.Context, key []byte) ([]*overlayTypes.OverlayNode, error) {
	searchMockNode := overlayMock.NewNodeBytes(key)

	// Simulate the lookup using the chord with the array

	startSearchIndex := 0

	// Speeding up the lookup using the special pointer nodes (Speedup Nodes)
	if len(chord.fakeNodesRing) >= numSpeedupNodes {
		for index, node := range chord.speedupNodes {
			if searchMockNode.Smaller(node.NodeMock) {
				if index == 0 {
					break
				} else {
					startSearchIndex = chord.speedupNodes[index-1].index
					break
				}
			}
		}
	}

	// Regular sequential search for the node in the fake ring section.
	res := make([]*overlayTypes.OverlayNode, chord.numSuccessors)
	resIndex := 0
	targetNodeIndex := -1
	for i := startSearchIndex; i < len(chord.fakeNodesRing); i++ {
		currentNode := chord.fakeNodesRing[i]
		if !currentNode.Smaller(searchMockNode) {
			if targetNodeIndex == -1 {
				targetNodeIndex = i
			}
			res[resIndex] = overlayTypes.NewOverlayNode(currentNode.IP(), caravela.FakePort, currentNode.Bytes())
			resIndex++
		}
		if resIndex == chord.numSuccessors {
			chord.collectLookupMessages(ctx, targetNodeIndex) // Collect metrics
			return res, nil
		}
	}

	startIndex := 0
	for ; resIndex < chord.numSuccessors; resIndex++ {
		if targetNodeIndex == -1 {
			targetNodeIndex = startIndex
		}
		res[resIndex] = overlayTypes.NewOverlayNode(chord.fakeNodesRing[startIndex].IP(), caravela.FakePort, chord.fakeNodesRing[startIndex].Bytes())
		startIndex++
	}

	chord.collectLookupMessages(ctx, targetNodeIndex) // Collect metrics
	return res, nil
}

func (chord *Mock) Neighbors(_ context.Context, nodeID []byte) ([]*overlayTypes.OverlayNode, error) {
	res := make([]*overlayTypes.OverlayNode, 2)
	neighMockNode := overlayMock.NewNodeBytes(nodeID)
	index := chord.nodesIdIndexMap[neighMockNode.String()]
	if index == 0 {
		res[0] = overlayTypes.NewOverlayNode(chord.fakeNodesRing[chord.numNodes-1].IP(), caravela.FakePort, chord.fakeNodesRing[chord.numNodes-1].Bytes())
		res[1] = overlayTypes.NewOverlayNode(chord.fakeNodesRing[1].IP(), caravela.FakePort, chord.fakeNodesRing[1].Bytes())
	} else if index == chord.numNodes-1 {
		res[0] = overlayTypes.NewOverlayNode(chord.fakeNodesRing[chord.numNodes-2].IP(), caravela.FakePort, chord.fakeNodesRing[chord.numNodes-2].Bytes())
		res[1] = overlayTypes.NewOverlayNode(chord.fakeNodesRing[0].IP(), caravela.FakePort, chord.fakeNodesRing[0].Bytes())
	} else {
		res[0] = overlayTypes.NewOverlayNode(chord.fakeNodesRing[index-1].IP(), caravela.FakePort, chord.fakeNodesRing[index-1].Bytes())
		res[1] = overlayTypes.NewOverlayNode(chord.fakeNodesRing[index+1].IP(), caravela.FakePort, chord.fakeNodesRing[index+1].Bytes())
	}
	return res, nil
}

func (chord *Mock) NodeID(_ context.Context) ([]byte, error) {
	// Do Nothing (Not necessary for the simulation)
	return nil, nil
}

func (chord *Mock) Leave(_ context.Context) error {
	// Do Nothing (Not necessary for the simulation)
	return nil
}

// ===============================================================================
// =							  Sort Interface                                 =
// ===============================================================================

func (chord *Mock) Len() int {
	return len(chord.fakeNodesRing)
}

func (chord *Mock) Swap(i, j int) {
	chord.fakeNodesRing[i], chord.fakeNodesRing[j] = chord.fakeNodesRing[j], chord.fakeNodesRing[i]
}

func (chord *Mock) Less(i, j int) bool {
	return chord.fakeNodesRing[i].Smaller(&chord.fakeNodesRing[j])
}
