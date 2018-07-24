package simulator

import (
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/mocks/docker"
	"github.com/strabox/caravela-sim/mocks/overlay"
	"github.com/strabox/caravela-sim/mocks/overlay/chord"
	"github.com/strabox/caravela-sim/simulation/metrics"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	caravelaConfig "github.com/strabox/caravela/configuration"
	caravelaNode "github.com/strabox/caravela/node"
	"runtime"
	"time"
)

const ticksPerPersist = 3
const simLogTag = "SIMULATOR"

type Simulator struct {
	tickInterval time.Duration    // Tick interval for simulation time
	currentTime  time.Duration    // Current simulation tick time
	metrics      *metrics.Metrics // Metrics gatherer

	nodes   []*caravelaNode.Node // Array with all the nodes for the simulation
	overlay *chord.Mock          // Overlay mock that "connects" all the nodes

	caravelaConfigurations *caravelaConfig.Configuration // CARAVELA's configurations
	simConfig              *configuration.Configuration  // Simulator's configurations
}

func NewSimulator(simConfig *configuration.Configuration,
	caravelaConfigurations *caravelaConfig.Configuration) *Simulator {
	return &Simulator{
		tickInterval: simConfig.TickInterval.Duration,
		currentTime:  0 * time.Second,
		metrics:      metrics.NewMetrics(simConfig.NumberOfNodes, simConfig.OutDirectoryPath),

		overlay: chord.NewChordMock(simConfig.NumberOfNodes, caravelaConfigurations.ChordNumSuccessors()),
		nodes:   make([]*caravelaNode.Node, simConfig.NumberOfNodes),

		caravelaConfigurations: caravelaConfigurations,
		simConfig:              simConfig,
	}
}

func (sim *Simulator) Init() {
	util.Log.Info(util.LogTag(simLogTag) + "Initializing...")

	// Init metrics
	sim.metrics.Init()
	sim.overlay.SetSimulator(sim)

	// Init caravela packages
	caravela.Init(sim.simConfig.CaravelaLogLevel)

	// External component mocks creation and initialization
	apiServerMock := caravela.NewAPIServerMock(sim)
	caravelaClientMock := caravela.NewRemoteClientMock(sim)
	dockerClientMock := docker.NewClientMock()
	overlay.Init(sim.caravelaConfigurations.ChordHashSizeBits() / 8)
	sim.overlay.Init()

	util.Log.Info(util.LogTag(simLogTag) + "Initializing nodes...")
	// Create all the CARAVELA's nodes
	for i := 0; i < sim.simConfig.NumberOfNodes; i++ {
		overlayNodeMock := sim.overlay.GetNodeMockByIndex(i)
		nodeConfig, _ := caravelaConfig.ObtainExternal(overlayNodeMock.IP(), sim.caravelaConfigurations)
		sim.nodes[i] = caravelaNode.NewNode(nodeConfig, sim.overlay, caravelaClientMock, dockerClientMock,
			apiServerMock)
		sim.nodes[i].AddTrader(overlayNodeMock.Bytes())
	}
	// Start all the CARAVELA's nodes
	for i := 0; i < sim.simConfig.NumberOfNodes; i++ {
		sim.nodes[i].Start(true, util.RandomIP())
	}

	time.Sleep(sim.simConfig.TimeBeforeSimulation.Duration)
	util.Log.Info(util.LogTag(simLogTag) + "Initialized")
}

func (sim *Simulator) Start() {
	util.Log.Info(util.LogTag(simLogTag) + "Simulation started...")

	// Send 10% of cluster size in requests per node
	nodesPerTick := int(float64(len(sim.nodes)) * float64(0.1))
	numTicks := 0

	for {
		util.Log.Infof(util.LogTag(simLogTag)+"Current Simulation Time: %.2f, Tick: %d, Ticks Remaining: %d",
			sim.currentTime.Seconds(), numTicks, sim.simConfig.MaxTicks-numTicks)

		// Do the tick actions
		for i := 0; i < nodesPerTick; i++ {
			nodeIndex, node := sim.randomNode()

			res := types.Resources{CPUs: 1, RAM: 256}
			sim.metrics.CreateRunRequest(nodeIndex, res, sim.currentTime)

			err := node.SubmitContainers(util.RandomName(),
				caravela.EmptyPortMappings(), caravela.EmptyContainerArgs(), res.CPUs, res.RAM)
			if err == nil {
				sim.metrics.RunRequestSucceeded()
			}
		}

		sim.currentTime = sim.currentTime + sim.tickInterval
		numTicks++
		if numTicks == sim.simConfig.MaxTicks {
			break
		}
		if (numTicks % ticksPerPersist) == 0 {
			sim.metrics.Persist(sim.currentTime)
			continue
		}
		sim.metrics.CreateNewSnapshot(sim.currentTime)
	}

	util.Log.Info(util.LogTag(simLogTag) + "Simulation Ended")
	sim.tearDown()      // Clear all the simulation nodes (clear all the memory) ...
	sim.metrics.Print() // Print the metrics results
	sim.metrics.Clear() // Clear all the temporary metric files
}

// randomNode returns a random node from the active nodes
func (sim *Simulator) randomNode() (int, *caravelaNode.Node) {
	randIndex := util.RandomInteger(0, len(sim.nodes)-1)
	return randIndex, sim.nodes[randIndex]
}

func (sim *Simulator) tearDown() {
	sim.nodes = nil
	sim.overlay = nil
	runtime.GC() // Force the GC to run in order to release the memory
}

func (sim *Simulator) NodeByIP(ip string) (*caravelaNode.Node, int) {
	index, _ := sim.overlay.GetNodeMockByIP(ip)
	return sim.nodes[index], index
}

func (sim *Simulator) NodeByGUID(guid string) (*caravelaNode.Node, int) {
	index, _ := sim.overlay.GetNodeMockByGUID(guid)
	return sim.nodes[index], index
}

func (sim *Simulator) Metrics() *metrics.Metrics {
	return sim.metrics
}
