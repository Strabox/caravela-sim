package simulator

import (
	"fmt"
	"github.com/strabox/caravela-sim/caravela"
	"github.com/strabox/caravela-sim/docker"
	"github.com/strabox/caravela-sim/overlay"
	"github.com/strabox/caravela-sim/overlay/chord"
	"github.com/strabox/caravela/configuration"
	caravelaNode "github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/common/guid"
	"time"
)

const timeBetweenNodeStart = 0 * time.Millisecond

type Simulator struct {
	numNodes int
	nodes    []*caravelaNode.Node
	overlay  *chord.Mock
}

func NewSimulator(numNodes int) *Simulator {
	return &Simulator{
		numNodes: numNodes,
		overlay:  chord.NewChordMock(numNodes),
		nodes:    make([]*caravelaNode.Node, numNodes),
	}
}

func (sim *Simulator) Init() {
	fmt.Println("INITIALIZING SIMULATOR...")
	// External component mocks
	apiServerMock := caravela.NewAPIServerMock(sim)
	caravelaClientMock := caravela.NewRemoteClientMock(sim)
	dockerClientMock := docker.NewClientMock()

	// Overlay initialization
	overlay.IdSizeBytes = caravela.NodeIDSizeBits / 8
	sim.overlay.Init()

	// Caravela initialization
	guid.InitializeGUID(caravela.NodeIDSizeBits)

	fmt.Println("INITIALIZING CARAVELA NODES...")
	// Caravela
	for i := 0; i < sim.numNodes; i++ {
		overlayNodeMock := sim.overlay.GetNodeMockByIndex(i)
		nodeConfig, _ := configuration.ReadFromFile(overlayNodeMock.IP(), caravela.FakeDockerAPIVersion)
		sim.nodes[i] = caravelaNode.NewNode(nodeConfig, sim.overlay, caravelaClientMock, dockerClientMock,
			apiServerMock)
		go sim.nodes[i].Start(true, caravela.LocalIP) // Each node has a unique goroutine for it
		sim.nodes[i].AddTrader(overlayNodeMock.Bytes())
		time.Sleep(timeBetweenNodeStart) // To avoid all the nodes being exactly timely synced (due to timers)
	}
	fmt.Println("SIMULATOR INITIALIZED")
}

func (sim *Simulator) Start() {
	fmt.Println("SIMULATION STARTED...")
	time.Sleep(1 * time.Minute)

	sim.nodes[2].Scheduler().Run(caravela.FakeContainerImageKey, caravela.EmptyPortMappings(),
		caravela.EmptyContainerArgs(), 1, 256)

	time.Sleep(3 * time.Minute)
	fmt.Println("SIMULATION ENDED")
}

func (sim *Simulator) NodeByIP(ip string) *caravelaNode.Node {
	index, _ := sim.overlay.GetNodeMockByIP(ip)
	return sim.nodes[index]
}

func (sim *Simulator) NodeByGUID(guid string) *caravelaNode.Node {
	index, _ := sim.overlay.GetNodeMockByGUID(guid)
	return sim.nodes[index]
}
