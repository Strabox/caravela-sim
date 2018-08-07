package simulation

import (
	"github.com/ivpusic/grpool"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/mocks/docker"
	"github.com/strabox/caravela-sim/mocks/overlay"
	"github.com/strabox/caravela-sim/mocks/overlay/chord"
	"github.com/strabox/caravela-sim/simulation/feeder"
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
	metricsCollector *metrics.Collector   // Metric's collector.
	nodes            []*caravelaNode.Node // Array with all the nodes for the simulation.
	overlayMock      *chord.Mock          // Overlay that connects all nodes.
	workersPool      *grpool.Pool
	feeder           feeder.Feeder
	caravelaConfigs  *caravelaConfig.Configuration // CARAVELA's configurations.
	simulatorConfigs *configuration.Configuration  // Simulator's configurations.
}

func NewSimulator(simConfig *configuration.Configuration, caravelaConfigurations *caravelaConfig.Configuration) *Simulator {
	maxWorkers := 1
	if simConfig.Multithreaded() {
		maxWorkers = runtime.NumCPU() * 2
	}

	return &Simulator{
		metricsCollector: metrics.NewCollector(simConfig.TotalNumberOfNodes(), simConfig.OutputDirectoryPath()),
		nodes:            make([]*caravelaNode.Node, simConfig.TotalNumberOfNodes()),
		overlayMock:      nil,
		workersPool:      grpool.NewPool(maxWorkers, maxWorkers*5),
		feeder:           feeder.Create(simConfig),
		caravelaConfigs:  caravelaConfigurations,
		simulatorConfigs: simConfig,
	}
}

func (sim *Simulator) Init() {
	util.Log.Info(util.LogTag(simLogTag) + "Initializing...")

	// Init CARAVELA's packages structures
	caravela.Init(sim.simulatorConfigs.CaravelaLogsLevel())

	// External node's component mocks (Creation and Init)
	apiServerMock := caravela.NewAPIServerMock()
	dockerClientMock := docker.NewClientMock(docker.CreateResourceGen(sim.simulatorConfigs, sim.caravelaConfigs))
	caravelaClientMock := caravela.NewRemoteClientMock(sim, sim.metricsCollector)
	sim.overlayMock = chord.NewChordMock(sim.simulatorConfigs.TotalNumberOfNodes(),
		sim.caravelaConfigs.ChordNumSuccessors(), sim.metricsCollector)
	overlay.Init(sim.caravelaConfigs.ChordHashSizeBits() / 8)
	sim.overlayMock.Init()

	// Create the CARAVELA's nodes for the simulation
	util.Log.Info(util.LogTag(simLogTag) + "Initializing nodes...")
	for i := 0; i < sim.simulatorConfigs.NumberOfNodes; i++ {
		overlayNodeMock := sim.overlayMock.GetNodeMockByIndex(i)
		nodeConfig, _ := caravelaConfig.ObtainExternal(overlayNodeMock.IP(), sim.caravelaConfigs)
		sim.nodes[i] = caravelaNode.NewNode(nodeConfig, sim.overlayMock, caravelaClientMock, dockerClientMock,
			apiServerMock)
		sim.nodes[i].AddTrader(overlayNodeMock.Bytes())
	}

	// Start all the CARAVELA's nodes
	util.Log.Info(util.LogTag(simLogTag) + "Starting nodes functions...")
	for i := 0; i < sim.simulatorConfigs.TotalNumberOfNodes(); i++ {
		sim.nodes[i].Start(true, util.RandomIP())
	}

	// Init metric's collector
	maxNodesResources := make([]types.Resources, sim.simulatorConfigs.TotalNumberOfNodes())
	for i := range maxNodesResources {
		maxNodesResources[i] = sim.nodes[i].MaximumResourcesSim()
	}
	sim.metricsCollector.Init(maxNodesResources)

	// Init request feeder
	sim.feeder.Init(sim.metricsCollector)

	util.Log.Info(util.LogTag(simLogTag) + "Initialized")
}

func (sim *Simulator) Start() {
	const ticksPerPersist = 3
	util.Log.Info(util.LogTag(simLogTag) + "Simulation started...")
	startTime := time.Now()
	currentSimTime, lastSimTimeRefreshes, numTicks := 0*time.Second, 0*time.Second, 0
	ticksChan := make(chan chan feeder.RequestTask)

	go sim.feeder.Start(ticksChan) // Start request feeder.

	for {
		util.Log.Infof(util.LogTag(simLogTag)+"Sim Time: %.2f, Tick: %d, Ticks Remaining: %d",
			currentSimTime.Seconds(), numTicks, sim.simulatorConfigs.MaximumTicks()-numTicks)

		// 1st. Inject the requests in the nodes, introducing the liveness.
		sim.acceptRequests(ticksChan, currentSimTime)

		// 2nd. Do the actions dependent on time (e.g. actions fired by timers).
		lastSimTimeRefreshes = sim.fireTimerActions(currentSimTime, lastSimTimeRefreshes)

		// 3rd. Update metrics with system's current information.
		sim.updateMetrics()

		// 4th. Update the simulation time using the tick mechanism.
		currentSimTime = currentSimTime + sim.simulatorConfigs.TicksInterval()
		numTicks++
		if numTicks == sim.simulatorConfigs.MaximumTicks() {
			close(ticksChan) // Alert feeder that the simulation has ended.
			break
		}
		if (numTicks % ticksPerPersist) == 0 {
			sim.metricsCollector.Persist(currentSimTime)
			continue
		}
		sim.metricsCollector.CreateNewGlobalSnapshot(currentSimTime)
	}

	util.Log.Info(util.LogTag(simLogTag) + "Simulation Ended")
	util.Log.Infof("Duration: Hours: %.2fh | Min: %.2fm | Sec: %.2fs", (time.Now().Sub(startTime)).Hours(),
		(time.Now().Sub(startTime)).Minutes(), (time.Now().Sub(startTime)).Seconds())
	sim.finished()
}

func (sim *Simulator) acceptRequests(ticksChan chan<- chan feeder.RequestTask, currentTime time.Duration) {
	const requestChanSize = 30
	defer sim.workersPool.WaitAll()

	newTickChan := make(chan feeder.RequestTask, requestChanSize)
	ticksChan <- newTickChan

	for {
		select {
		case requestTask, more := <-newTickChan:
			if more {
				sim.workersPool.WaitCount(1)
				sim.workersPool.JobQueue <- func() {
					defer sim.workersPool.JobDone()
					nodeIndex, node := sim.randomNode()
					requestTask(nodeIndex, node, currentTime)
				}
			} else {
				return
			}
		}
	}
}

func (sim *Simulator) fireTimerActions(currentTime, lastTimeRefreshes time.Duration) time.Duration {
	defer sim.workersPool.WaitAll()

	// Refresh offers
	if (currentTime - lastTimeRefreshes) >= sim.caravelaConfigs.RefreshingInterval() {
		// Necessary because the tick interval can be greater than the refresh interval.
		timesToRefresh := int((currentTime - lastTimeRefreshes) / sim.caravelaConfigs.RefreshingInterval())

		for _, node := range sim.nodes {
			tempNode := node
			sim.workersPool.WaitCount(1)
			sim.workersPool.JobQueue <- func() {
				defer sim.workersPool.JobDone()
				for i := 0; i < timesToRefresh; i++ {
					tempNode.RefreshOffersSim()
				}
			}
		}

		lastTimeRefreshes = currentTime
	}

	// TODO: Spread offers ??
	// TODO: Advertise resources offers ??

	return lastTimeRefreshes
}

func (sim *Simulator) updateMetrics() {
	defer sim.workersPool.WaitAll()

	for i, node := range sim.nodes {
		tempI := i
		tempNode := node

		sim.workersPool.WaitCount(1)
		sim.workersPool.JobQueue <- func() {
			defer sim.workersPool.JobDone()
			sim.metricsCollector.SetAvailableNodeResources(tempI, tempNode.AvailableResourcesSim())
		}
	}
}

// finished is called when the simulation ends/stops.
func (sim *Simulator) finished() {
	util.Log.Info(util.LogTag(simLogTag) + "Clearing simulation objects...")
	sim.release() // Clear all the simulation nodes (release all the memory) ...
	util.Log.Info(util.LogTag(simLogTag) + "Crushing simulation results...")
	sim.metricsCollector.Print() // Print the metricsCollector results and outputs the graphics.
	sim.metricsCollector.Clear() // Clear all the temporary metric files
	util.Log.Info(util.LogTag(simLogTag) + "FINISHED")
}

// release all the memory of the simulation structures, nodes, etc.
func (sim *Simulator) release() {
	sim.workersPool.Release()
	sim.feeder = nil
	sim.workersPool = nil
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

// randomNode returns a random node from the simulated active nodes.
func (sim *Simulator) randomNode() (int, *caravelaNode.Node) {
	randIndex := util.RandomInteger(0, len(sim.nodes)-1)
	return randIndex, sim.nodes[randIndex]
}
