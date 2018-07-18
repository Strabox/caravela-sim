package simulator

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/mocks/docker"
	overlayMock "github.com/strabox/caravela-sim/mocks/overlay"
	"github.com/strabox/caravela-sim/mocks/overlay/chord"
	"github.com/strabox/caravela-sim/util"
	caravelaConfig "github.com/strabox/caravela/configuration"
	caravelaNode "github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/common/guid"
	"os"
	"time"
)

type Simulator struct {
	nodes   []*caravelaNode.Node // Array with all the nodes for the simulation
	overlay *chord.Mock          // Overlay mock that "connects" all the nodes

	config *configuration.Configuration // Configurations for the simulator
	log    *logger.Logger               // Object used for Simulator logging
}

func NewSimulator(config *configuration.Configuration) *Simulator {
	log := logger.New()
	return &Simulator{
		overlay: chord.NewChordMock(config.NumberOfNodes, log),
		nodes:   make([]*caravelaNode.Node, config.NumberOfNodes),

		config: config,
		log:    log,
	}
}

func (sim *Simulator) Init() {
	PrintSimulatorBanner()
	sim.log.Print(util.LogTag("SIMULATOR") + "INITIALIZING SIMULATOR...")

	// Init caravela logs
	caravela.InitLogs(sim.config.CaravelaLogLevel)

	// Init simulator logs
	sim.log.Level = util.LogLevel(sim.config.SimulatorLogLevel)
	sim.log.Formatter = util.LogFormatter(true, true)
	sim.log.Out = os.Stdout

	// Caravela configurations
	caravelaConfigs, _ := caravelaConfig.ReadFromFile("")

	// External component mocks creation and initialization
	apiServerMock := caravela.NewAPIServerMock(sim)
	caravelaClientMock := caravela.NewRemoteClientMock(sim)
	dockerClientMock := docker.NewClientMock()
	overlayMock.SetNodeIDSizeBytes(caravelaConfigs.Overlay.ChordHashSizeBits / 8)
	sim.overlay.Init()

	// Caravela GUID initialization
	guid.InitializeGUID(caravelaConfigs.Overlay.ChordHashSizeBits)

	sim.log.Print(util.LogTag("SIMULATOR") + "INITIALIZING CARAVELA NODES...")
	// Set up Caravela nodes and start them
	for i := 0; i < sim.config.NumberOfNodes; i++ {
		overlayNodeMock := sim.overlay.GetNodeMockByIndex(i)
		nodeConfig, _ := caravelaConfig.ReadFromFile(overlayNodeMock.IP())
		sim.nodes[i] = caravelaNode.NewNode(nodeConfig, sim.overlay, caravelaClientMock, dockerClientMock,
			apiServerMock)
		go sim.nodes[i].Start(true, caravela.LocalIP) // Each node has a unique goroutine for it!!!!!!!
		sim.nodes[i].AddTrader(overlayNodeMock.Bytes())
		time.Sleep(sim.config.TimeBetweenNodeStart.Duration) // To avoid all the nodes being exactly timely synced (due to timers)
	}
	sim.log.Print(util.LogTag("SIMULATOR") + "SIMULATOR INITIALIZED")
}

func (sim *Simulator) Start() {
	time.Sleep(sim.config.TimeBeforeStartSimulating.Duration)
	sim.log.Print(util.LogTag("SIMULATOR") + "SIMULATION STARTED...")

	for _, node := range sim.nodes {
		go node.SubmitContainers(caravela.FakeContainerImageKey, caravela.EmptyPortMappings(),
			caravela.EmptyContainerArgs(), 1, 512)
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(5 * time.Minute)
	sim.log.Print(util.LogTag("SIMULATOR") + "SIMULATION ENDED")
}

func (sim *Simulator) NodeByIP(ip string) *caravelaNode.Node {
	index, _ := sim.overlay.GetNodeMockByIP(ip)
	return sim.nodes[index]
}

func (sim *Simulator) NodeByGUID(guid string) *caravelaNode.Node {
	index, _ := sim.overlay.GetNodeMockByGUID(guid)
	return sim.nodes[index]
}
