package chord

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela-sim/mocks/caravela"
	overlayMock "github.com/strabox/caravela-sim/mocks/overlay"
	"github.com/strabox/caravela-sim/util"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/overlay"
	"sort"
)

type Mock struct {
	log *logger.Logger

	numNodes        int                    // Initial number of nodes for the chord
	fakeNodesRing   []overlayMock.NodeMock // The array that represent the node's chord ring
	nodesIdIndexMap map[string]int         // ID <-> Index
	nodesIpIndexMap map[string]int         // IP <-> Index
}

func NewChordMock(numNodes int, log *logger.Logger) *Mock {
	return &Mock{
		log:             log,
		numNodes:        numNodes,
		fakeNodesRing:   make([]overlayMock.NodeMock, numNodes),
		nodesIdIndexMap: make(map[string]int),
		nodesIpIndexMap: make(map[string]int),
	}
}

func (chord *Mock) Init() {
	for i := 0; i < chord.numNodes; i++ {
		chord.fakeNodesRing[i] = *overlayMock.NewRandomNode()
	}

	sort.Sort(chord)

	// After sort because the index of the nodes changed
	for i := 0; i < chord.numNodes; i++ {
		chord.nodesIdIndexMap[chord.fakeNodesRing[i].String()] = i
		chord.nodesIpIndexMap[chord.fakeNodesRing[i].IP()] = i
	}
}

func (chord *Mock) Print() {
	for _, node := range chord.fakeNodesRing {
		chord.log.Printf(util.LogTag("SIMULATOR")+"IP: %s, ID: %s", node.IP(), node.String())
	}
}

func (chord *Mock) GetNodeMockByIndex(index int) *overlayMock.NodeMock {
	return &chord.fakeNodesRing[index]
}

func (chord *Mock) GetNodeMockByGUID(id string) (int, *overlayMock.NodeMock) {
	index := chord.nodesIdIndexMap[id]
	return index, &chord.fakeNodesRing[index]
}

func (chord *Mock) GetNodeMockByIP(ip string) (int, *overlayMock.NodeMock) {
	index := chord.nodesIpIndexMap[ip]
	return index, &chord.fakeNodesRing[index]
}

/*
===============================================================================
							  Overlay Interface
===============================================================================
*/

func (chord *Mock) Create(thisNode nodeAPI.OverlayMembership) error {
	// Do Nothing (For now not necessary for the simulation)
	return nil
}

func (chord *Mock) Join(overlayNodeIP string, overlayNodePort int, thisNode nodeAPI.OverlayMembership) error {
	// Do Nothing (For now not necessary for the simulation)
	return nil
}

func (chord *Mock) Lookup(key []byte) ([]*overlay.Node, error) {
	const numOfRes = 3
	searchMockNode := overlayMock.NewNode(key)
	chord.log.Printf(util.LogTag("SIMULATOR")+"LOOKUP, %s", searchMockNode.String())

	res := make([]*overlay.Node, numOfRes)
	resIndex := 0
	for _, node := range chord.fakeNodesRing {
		if !node.Smaller(searchMockNode) {
			res[resIndex] = overlay.NewNode(node.IP(), caravela.FakePort, node.Bytes())
			resIndex++
		}
		if resIndex == numOfRes {
			return res, nil
		}
	}

	startIndex := 0
	for ; resIndex < numOfRes; resIndex++ {
		res[resIndex] = overlay.NewNode(chord.fakeNodesRing[startIndex].IP(), caravela.FakePort, chord.fakeNodesRing[startIndex].Bytes())
		startIndex++
	}

	return res, nil
}

func (chord *Mock) Neighbors(nodeID []byte) ([]*overlay.Node, error) {
	res := make([]*overlay.Node, 2)
	neighMockNode := overlayMock.NewNode(nodeID)
	index := chord.nodesIdIndexMap[neighMockNode.String()]
	if index == 0 {
		res[0] = overlay.NewNode(chord.fakeNodesRing[chord.numNodes-1].IP(), caravela.FakePort, chord.fakeNodesRing[chord.numNodes-1].Bytes())
		res[1] = overlay.NewNode(chord.fakeNodesRing[1].IP(), caravela.FakePort, chord.fakeNodesRing[1].Bytes())
	} else if index == chord.numNodes-1 {
		res[0] = overlay.NewNode(chord.fakeNodesRing[chord.numNodes-2].IP(), caravela.FakePort, chord.fakeNodesRing[chord.numNodes-2].Bytes())
		res[1] = overlay.NewNode(chord.fakeNodesRing[0].IP(), caravela.FakePort, chord.fakeNodesRing[0].Bytes())
	} else {
		res[0] = overlay.NewNode(chord.fakeNodesRing[index-1].IP(), caravela.FakePort, chord.fakeNodesRing[index-1].Bytes())
		res[1] = overlay.NewNode(chord.fakeNodesRing[index+1].IP(), caravela.FakePort, chord.fakeNodesRing[index+1].Bytes())
	}
	return nil, nil
}

func (chord *Mock) NodeID() ([]byte, error) {
	// Do Nothing (For now not necessary for the simulation)
	return nil, nil
}

func (chord *Mock) Leave() error {
	// Do Nothing (For now not necessary for the simulation)
	return nil
}

/*
===============================================================================
							  Sort Interface
===============================================================================
*/

func (chord *Mock) Len() int {
	return len(chord.fakeNodesRing)
}

func (chord *Mock) Swap(i, j int) {
	chord.fakeNodesRing[i], chord.fakeNodesRing[j] = chord.fakeNodesRing[j], chord.fakeNodesRing[i]
}

func (chord *Mock) Less(i, j int) bool {
	return chord.fakeNodesRing[i].Smaller(&chord.fakeNodesRing[j])
}
