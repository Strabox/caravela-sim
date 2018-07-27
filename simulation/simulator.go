package simulation

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

const simLogTag = "SIMULATOR"

type Simulator struct {
	metricsCollector *metrics.Collector            // Metrics collector
	nodes            []*caravelaNode.Node          // Array with all the nodes for the simulation
	overlayMock      *chord.Mock                   // Overlay that connects all nodes
	caravelaConfigs  *caravelaConfig.Configuration // CARAVELA's configurations
	simulatorConfigs *configuration.Configuration  // Simulator's configurations
}

func NewSimulator(simConfig *configuration.Configuration,
	caravelaConfigurations *caravelaConfig.Configuration) *Simulator {
	return &Simulator{
		metricsCollector: metrics.NewMetrics(simConfig.TotalNumberOfNodes(), simConfig.OutputDirectoryPath()),
		nodes:            make([]*caravelaNode.Node, simConfig.TotalNumberOfNodes()),
		overlayMock:      nil,
		caravelaConfigs:  caravelaConfigurations,
		simulatorConfigs: simConfig,
	}
}

func (sim *Simulator) Init() {
	util.Log.Info(util.LogTag(simLogTag) + "Initializing...")

	// Init CARAVELA's packages
	caravela.Init(sim.simulatorConfigs.CaravelaLogsLevel())

	// External component mocks creation and initialization
	apiServerMock := caravela.NewAPIServerMock()
	dockerClientMock := docker.NewClientMock()
	caravelaClientMock := caravela.NewRemoteClientMock(sim, sim.metricsCollector)
	sim.overlayMock = chord.NewChordMock(sim.simulatorConfigs.TotalNumberOfNodes(),
		sim.caravelaConfigs.ChordNumSuccessors(), sim.metricsCollector)
	overlay.Init(sim.caravelaConfigs.ChordHashSizeBits() / 8)
	sim.overlayMock.Init()

	util.Log.Info(util.LogTag(simLogTag) + "Initializing nodes...")
	// Create the CARAVELA's nodes for the simulation
	for i := 0; i < sim.simulatorConfigs.NumberOfNodes; i++ {
		overlayNodeMock := sim.overlayMock.GetNodeMockByIndex(i)
		nodeConfig, _ := caravelaConfig.ObtainExternal(overlayNodeMock.IP(), sim.caravelaConfigs)
		sim.nodes[i] = caravelaNode.NewNode(nodeConfig, sim.overlayMock, caravelaClientMock, dockerClientMock,
			apiServerMock)
		sim.nodes[i].AddTrader(overlayNodeMock.Bytes())
	}
	// Start all the CARAVELA's nodes
	for i := 0; i < sim.simulatorConfigs.TotalNumberOfNodes(); i++ {
		sim.nodes[i].Start(true, util.RandomIP())
	}

	// Init metricsCollector gatherer
	maxNodesResources := make([]types.Resources, sim.simulatorConfigs.TotalNumberOfNodes())
	for i := range maxNodesResources {
		maxNodesResources[i] = sim.nodes[i].MaximumResourcesSim()
	}
	sim.metricsCollector.Init(maxNodesResources)

	//time.Sleep(sim.simulatorConfigs.TimeBeforeSimulationStart()) //Deprecated
	util.Log.Info(util.LogTag(simLogTag) + "Initialized")
}

func (sim *Simulator) Start() {
	util.Log.Info(util.LogTag(simLogTag) + "Simulation started...")

	const ticksPerPersist = 3
	runReqPerTick := int(float64(len(sim.nodes)) * float64(0.1)) // Send 10% of cluster size in requests per node
	currentTime, lastTimeRefreshes, numTicks := 0*time.Second, 0*time.Second, 0

	for {
		util.Log.Infof(util.LogTag(simLogTag)+"Current Simulation Time: %.2f, Tick: %d, Ticks Remaining: %d",
			currentTime.Seconds(), numTicks, sim.simulatorConfigs.MaximumTicks()-numTicks)

		// ============= Inject the requests in the nodes, introducing the liveness. =============
		for i := 0; i < runReqPerTick; i++ {
			nodeIndex, node := sim.randomNode()

			res := types.Resources{CPUs: 1, RAM: 256}
			sim.metricsCollector.CreateRunRequest(nodeIndex, res, currentTime)

			err := node.SubmitContainers(util.RandomName(),
				caravela.EmptyPortMappings(), caravela.EmptyContainerArgs(), res.CPUs, res.RAM)
			if err == nil {
				sim.metricsCollector.RunRequestSucceeded()
			}
		}

		// ============ Do the actions dependent on time (e.g. timer dependent actions) ===========

		// Refresh offers
		if (currentTime - lastTimeRefreshes) >= sim.caravelaConfigs.RefreshingInterval() {
			// Necessary because the tick interval can be greater than the refresh interval.
			timesToRefresh := int((currentTime - lastTimeRefreshes) / sim.caravelaConfigs.RefreshingInterval())
			for _, node := range sim.nodes {
				for i := 0; i < timesToRefresh; i++ {
					node.RefreshOffersSim()
				}
			}
			lastTimeRefreshes = currentTime
		}

		// TODO: Spread offers ??
		// TODO: Advertise resources offers ??

		// =============== Update metricsCollector with node's current information ================
		for i := range sim.nodes {
			sim.metricsCollector.SetAvailableNodeResources(i, sim.nodes[i].AvailableResourcesSim())
		}

		// ================== Update the simulation time using the tick mechanism =================
		currentTime = currentTime + sim.simulatorConfigs.TicksInterval()
		numTicks++
		if numTicks == sim.simulatorConfigs.MaximumTicks() {
			break
		}
		if (numTicks % ticksPerPersist) == 0 {
			sim.metricsCollector.Persist(currentTime)
			continue
		}
		sim.metricsCollector.CreateNewSnapshot(currentTime)
	}

	util.Log.Info(util.LogTag(simLogTag) + "Simulation Ended")
	sim.finished()
}

// finished is called when the simulation ends/stops.
func (sim *Simulator) finished() {
	sim.clear()                  // Clear all the simulation nodes (clear all the memory) ...
	sim.metricsCollector.Print() // Print the metricsCollector results
	sim.metricsCollector.Clear() // Clear all the temporary metric files
}

// randomNode returns a random node from the simulated active nodes.
func (sim *Simulator) randomNode() (int, *caravelaNode.Node) {
	randIndex := util.RandomInteger(0, len(sim.nodes)-1)
	return randIndex, sim.nodes[randIndex]
}

// clear releases all the memory of the simulated structures, nodes, etc
func (sim *Simulator) clear() {
	sim.nodes = nil
	sim.overlayMock = nil
	runtime.GC() // Force the GC to run in order to release the memory
}

func (sim *Simulator) NodeByIP(ip string) (*caravelaNode.Node, int) {
	index, _ := sim.overlayMock.GetNodeMockByIP(ip)
	return sim.nodes[index], index
}

func (sim *Simulator) NodeByGUID(guid string) (*caravelaNode.Node, int) {
	index, _ := sim.overlayMock.GetNodeMockByGUID(guid)
	return sim.nodes[index], index
}
