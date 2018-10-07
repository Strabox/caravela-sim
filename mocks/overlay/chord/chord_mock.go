package chord

import (
	"context"
	"github.com/ivpusic/grpool"
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/engine/metrics"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/overlay"
	"math/big"
	"runtime"
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

	ringMock        []NodeMock     // Array that represent the node's chord ring.
	nodesIdIndexMap map[string]int // ID <-> Index.
	nodesIpIndexMap map[string]int // IP <-> Index.
}

// NewChordMock creates a new chord overlay that can be used by an application component.
// It implements the github.com/strabox/caravela/node/external Overlay interface.
func NewChordMock(numNodes, numSuccessors, numSpeedupNodes int, metricsCollector *metrics.Collector) *Mock {
	return &Mock{
		collector:       metricsCollector,
		numSpeedupNodes: numSpeedupNodes,
		numNodes:        numNodes,
		numSuccessors:   numSuccessors,
		ringMock:        make([]NodeMock, numNodes),
		nodesIdIndexMap: make(map[string]int),
		nodesIpIndexMap: make(map[string]int),
	}
}

// Init initializes the chord's mock structure.
func (m *Mock) Init() {
	// Generate the Node's ID uniformly in order to easily do the perfect chord route mechanism with minimal overhead.
	maxNumGUIDs := maxNumGUIDs()
	nodeGUIDGen := big.NewInt(0)
	nodeGUIDSpace := maxNumGUIDs.Div(maxNumGUIDs, big.NewInt(int64(m.numNodes)))
	for i := 0; i < m.numNodes; i++ {
		m.ringMock[i] = *NewNodeRandomIP(nodeGUIDGen)
		nodeGUIDGen.Add(nodeGUIDGen, nodeGUIDSpace)
	}

	// Sort nodes by ID (ascending order)
	sort.Sort(m)

	// Use of speed up nodes in order quickly fill the finger's tables.
	speedupNodes := make([]speedupNodeMock, m.numSpeedupNodes) // Array of speed up nodes.
	speedupDivSize := m.numNodes / m.numSpeedupNodes
	divSize := 1
	speedupNodeIndex := 0
	for i := 0; i < m.numNodes; i++ {
		m.nodesIdIndexMap[m.ringMock[i].String()] = i
		m.nodesIpIndexMap[m.ringMock[i].IP()] = i

		if speedupDivSize != 0 && divSize == speedupDivSize && speedupNodeIndex < m.numSpeedupNodes {
			speedupNodes[speedupNodeIndex] = *newSpeedupNodeMock(i, &m.ringMock[i])
			speedupNodeIndex++
			divSize = 1
			continue
		}
		divSize++
	}

	successorNodeIndex := func(searchMockNode *NodeMock) int {
		// Simulate the lookup using the chord with the array
		startSearchIndex := 0
		// Speeding up the successor lookup using the special pointer nodes (Speedup Nodes).
		if len(m.ringMock) >= m.numSpeedupNodes {
			for index, node := range speedupNodes {
				if index > 0 {
					startSearchIndex = speedupNodes[index-1].index
				}
				if searchMockNode.Smaller(node.NodeMock) {
					break
				}
			}
		}

		// Regular sequential search for the node in the ring.
		for i := startSearchIndex; i < len(m.ringMock); i++ {
			currentNode := m.ringMock[i]
			if currentNode.HigherOrEqual(searchMockNode) {
				return i
			}
		}

		return 0
	}

	// Goroutine pool used to build the finger's tables faster.
	goroutinePool := grpool.NewPool(runtime.NumCPU(), runtime.NumCPU()*5)

	for i := range m.ringMock {
		ringNode := &m.ringMock[i]
		goroutinePool.WaitCount(1)
		goroutinePool.JobQueue <- func() {
			defer goroutinePool.JobDone()
			fingerTable := ringNode.FingerTable()
			fingersIndexes := make([]int, len(fingerTable))
			fingersGUIDs := make([]*guid.GUID, len(fingerTable))
			for k, fingerGUID := range fingerTable {
				successorIndex := successorNodeIndex(NewNodeBytes(fingerGUID.Bytes()))
				fingersIndexes[k] = successorIndex
				fingersGUIDs[k] = m.ringMock[successorIndex].guid
			}
			ringNode.SetFingersNodeIndexes(fingersIndexes, fingersGUIDs)
		}
	}

	goroutinePool.WaitAll() // Wait for all the plots to be completed.
	goroutinePool.Release() // Release goroutinePool resources.
	goroutinePool = nil     // Only to ensure madness errors.

	m.Print()
}

func (m *Mock) Print() {
	util.Log.Debugf("##################################################################")
	util.Log.Debugf("#                   CHORD's MOCK CONFIGURATIONS                  #")
	util.Log.Debugf("##################################################################")
	util.Log.Debugf("Nodes:              	%d", m.numNodes)
	util.Log.Debugf("Speedup Nodes:      	%d", m.numSpeedupNodes)
	util.Log.Debugf("Speedup Window Size: 	%d", m.numNodes/m.numSpeedupNodes)
	util.Log.Debugf("##################################################################")
}

func (m *Mock) GetNodeMockByIndex(index int) *NodeMock {
	return &m.ringMock[index]
}

func (m *Mock) GetNodeMockByGUID(guid string) (int, *NodeMock) {
	index := m.nodesIdIndexMap[guid]
	return index, &m.ringMock[index]
}

func (m *Mock) GetNodeMockByIP(ip string) (int, *NodeMock) {
	index, exist := m.nodesIpIndexMap[ip]
	if !exist {
		panic(errors.New("Node's IP does not exist in the system!"))
	}
	return index, &m.ringMock[index]
}

// ===============================================================================
// =							  Overlay Interface                              =
// ===============================================================================

func (m *Mock) Create(_ context.Context, _ overlay.LocalNode) error {
	// Do Nothing (Not necessary for the engine)
	return nil
}

func (m *Mock) Join(_ context.Context, _ string, _ int, _ overlay.LocalNode) error {
	// Do Nothing (Not necessary for the engine)
	return nil
}

func (m *Mock) Lookup(ctx context.Context, key []byte) ([]*overlay.OverlayNode, error) {
	fromNodeGUID := types.NodeGUID(ctx)
	if fromNodeGUID == "" {
		panic("Lookup message did not have the from node GUID debug data!")
	}
	currentNodeSearchIndex, _ := m.GetNodeMockByGUID(fromNodeGUID)
	found := false
	messagesPerReqAcc := 0

	keyBigInt := big.NewInt(0)
	keyBigInt.SetBytes(key)
	fromBigInt := big.NewInt(0)
	fromBigInt.SetString(fromNodeGUID, 10)
	if keyBigInt.Cmp(fromBigInt) != 0 {
		for {
			currentNodeSearchIndex, found = m.ringMock[currentNodeSearchIndex].Lookup(currentNodeSearchIndex, key)
			messagesPerReqAcc++
			m.collector.MessageReceived(currentNodeSearchIndex, 1, findSuccessorMessageSizeREST)
			if found {
				// Reply to the node that called the Lookup.
				messagesPerReqAcc++
				fromNodeIndex, _ := m.GetNodeMockByGUID(fromNodeGUID)
				m.collector.MessageReceived(fromNodeIndex, 1, findSuccessorMessageResponseSizeREST)
				m.collector.IncrMessagesTradedRequest(types.RequestID(ctx), messagesPerReqAcc)
				break
			}
		}
	}

	res := make([]*overlay.OverlayNode, m.numSuccessors)
	successorsFound := 0
	for i := currentNodeSearchIndex; i < len(m.ringMock); i++ {
		res[successorsFound] = overlay.NewOverlayNode(m.ringMock[i].IP(), caravela.FakePort, m.ringMock[i].Bytes())
		successorsFound++
		if successorsFound == m.numSuccessors {
			break
		}
	}

	if successorsFound != m.numSuccessors {
		for i := 0; i < len(m.ringMock); i++ {
			res[successorsFound] = overlay.NewOverlayNode(m.ringMock[i].IP(), caravela.FakePort, m.ringMock[i].Bytes())
			successorsFound++
			if successorsFound == m.numSuccessors {
				break
			}
		}
	}
	return res, nil
}

func (m *Mock) Neighbors(_ context.Context, nodeID []byte) ([]*overlay.OverlayNode, error) {
	res := make([]*overlay.OverlayNode, 2)
	neighMockNode := NewNodeBytes(nodeID)
	index := m.nodesIdIndexMap[neighMockNode.String()]
	if index == 0 {
		res[0] = overlay.NewOverlayNode(m.ringMock[m.numNodes-1].IP(), caravela.FakePort, m.ringMock[m.numNodes-1].Bytes())
		res[1] = overlay.NewOverlayNode(m.ringMock[1].IP(), caravela.FakePort, m.ringMock[1].Bytes())
	} else if index == m.numNodes-1 {
		res[0] = overlay.NewOverlayNode(m.ringMock[m.numNodes-2].IP(), caravela.FakePort, m.ringMock[m.numNodes-2].Bytes())
		res[1] = overlay.NewOverlayNode(m.ringMock[0].IP(), caravela.FakePort, m.ringMock[0].Bytes())
	} else {
		res[0] = overlay.NewOverlayNode(m.ringMock[index-1].IP(), caravela.FakePort, m.ringMock[index-1].Bytes())
		res[1] = overlay.NewOverlayNode(m.ringMock[index+1].IP(), caravela.FakePort, m.ringMock[index+1].Bytes())
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
	return len(m.ringMock)
}

func (m *Mock) Swap(i, j int) {
	m.ringMock[i], m.ringMock[j] = m.ringMock[j], m.ringMock[i]
}

func (m *Mock) Less(i, j int) bool {
	return m.ringMock[i].Smaller(&m.ringMock[j])
}
