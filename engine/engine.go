package engine

import (
	"fmt"
	"github.com/ivpusic/grpool"
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/engine/feeder"
	"github.com/strabox/caravela-sim/engine/metrics"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/mocks/docker"
	chordMock "github.com/strabox/caravela-sim/mocks/overlay/chord"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	caravelaConfig "github.com/strabox/caravela/configuration"
	caravelaNode "github.com/strabox/caravela/node"
	"runtime"
	"time"
)

// engineLogTag log's tag for the simulator engine.
const engineLogTag = "ENGINE"

const numOfRandomBagsOfNode = 2

// Engine represents an instance of a Caravela's simulator engine.
// It holds all the structures to control, feed and analyse a engine during a simulation.
type Engine struct {
	isInit         bool // Used to verify if the simulator is initialized.
	lastSimulation bool
	baseRngSeed    int64 // Base RNG seed for generating the node's max resources and the requests resources.

	// Engine's main components.
	nodes       []*caravelaNode.Node // Array with all the Caravela's nodes for the simulation.
	overlayMock *chordMock.Mock      // Overlay that "connects" all nodes.
	feeder      feeder.Feeder        // Used to feed the simulator with requests.
	nodesBags   [][]*caravelaNode.Node

	metricsCollector *metrics.Collector            // Metric's collector.
	workersPool      *grpool.Pool                  // Pool of Goroutines to run the simulation.
	caravelaConfigs  *caravelaConfig.Configuration // Caravela's configurations.
	simulatorConfigs *configuration.Configuration  // Simulator's configurations.
}

// NewEngine creates a new simulator instance based on the configurations of the caravela and its own.
func NewEngine(metricsCollector *metrics.Collector, simConfig *configuration.Configuration, baseRngSeed int64) *Engine {
	return &Engine{
		isInit:           false,
		baseRngSeed:      baseRngSeed,
		metricsCollector: metricsCollector,
		simulatorConfigs: simConfig,
	}
}

// Init initializes all the components making it ready to start the engine.
func (e *Engine) Init(reuseEngine, lastSimulation bool, caravelaConfigurations *caravelaConfig.Configuration) {
	util.Log.Info(util.LogTag(engineLogTag) + "Initializing...")
	maxWorkers := 1
	if e.simulatorConfigs.Multithreaded() {
		maxWorkers = runtime.NumCPU() * 2
	}
	e.workersPool = grpool.NewPool(maxWorkers, maxWorkers*30)
	e.nodes = make([]*caravelaNode.Node, e.simulatorConfigs.TotalNumberOfNodes())
	e.nodesBags = make([][]*caravelaNode.Node, numOfRandomBagsOfNode)
	e.caravelaConfigs = caravelaConfigurations
	e.feeder = feeder.Create(e.simulatorConfigs, caravelaConfigurations, e.baseRngSeed)
	e.lastSimulation = lastSimulation

	// Init CARAVELA's packages structures.
	caravela.Init(e.simulatorConfigs.CaravelaLogsLevel(), e.caravelaConfigs)

	// External node's component mocks (Creation and initialization).
	apiServerMock := caravela.NewAPIServerMock()
	dockerClientMock := docker.NewClientMock(docker.CreateResourceGen(e.simulatorConfigs, e.caravelaConfigs, e.baseRngSeed))
	caravelaClientMock := caravela.NewRemoteClientMock(e, e.metricsCollector)
	if !e.isInit || !reuseEngine {
		chordMock.Init(e.caravelaConfigs.ChordHashSizeBits())
		e.overlayMock = chordMock.NewChordMock(e.simulatorConfigs.TotalNumberOfNodes(),
			e.caravelaConfigs.ChordNumSuccessors(), e.simulatorConfigs.ChordMockSpeedupNodes(), e.metricsCollector)
		e.overlayMock.Init()
	}

	// Create the CARAVELA's nodes for the engine.
	util.Log.Info(util.LogTag(engineLogTag) + "Initializing nodes...")
	for i := 0; i < e.simulatorConfigs.NumberOfNodes; i++ {
		tempIndex := i
		e.workersPool.WaitCount(1)
		e.workersPool.JobQueue <- func() {
			defer e.workersPool.JobDone()

			overlayNodeMock := e.overlayMock.GetNodeMockByIndex(tempIndex)
			nodeConfig, err := caravelaConfig.ObtainExternal(overlayNodeMock.IP(), e.caravelaConfigs)
			if err != nil {
				panic(fmt.Errorf("can't make caravela configurations, error: %s", err))
			}

			e.nodes[tempIndex] = caravelaNode.NewNode(nodeConfig, e.overlayMock, caravelaClientMock, dockerClientMock, apiServerMock)
			e.nodes[tempIndex].AddTrader(overlayNodeMock.Bytes())
		}
	}
	e.workersPool.WaitAll()

	// Start the Caravela's nodes.
	util.Log.Info(util.LogTag(engineLogTag) + "Starting nodes functions...")
	for i := 0; i < e.simulatorConfigs.TotalNumberOfNodes(); i++ {
		tempIndex := i
		e.workersPool.WaitCount(1)
		e.workersPool.JobQueue <- func() {
			defer e.workersPool.JobDone()

			e.nodes[tempIndex].Start(true, util.RandomIP())
		}
	}
	e.workersPool.WaitAll()

	// Initialize metric's collector.
	maxNodesResources := make([]types.Resources, e.simulatorConfigs.TotalNumberOfNodes())
	for i := range maxNodesResources {
		tempIndex := i
		e.workersPool.WaitCount(1)
		e.workersPool.JobQueue <- func() {
			defer e.workersPool.JobDone()

			_, nodeMaxResources, _, _ := e.nodes[tempIndex].NodeInformationSim()
			maxNodesResources[tempIndex] = nodeMaxResources
		}
	}
	e.workersPool.WaitAll()
	e.metricsCollector.InitNewSimulation(e.caravelaConfigs.DiscoveryBackend(), maxNodesResources)

	// Initialize request feeder.
	systemTotalCPUs, systemTotalMemory := dockerClientMock.MaxResourcesAvailable()
	e.feeder.Init(e.metricsCollector, types.Resources{CPUs: systemTotalCPUs, Memory: systemTotalMemory})
	util.Log.Debugf(util.LogTag(engineLogTag)+"System Total ResRequested: <%d;%d>", systemTotalCPUs, systemTotalMemory)

	// Make a bag of random nodes in order to enforce time based actions more realistically.
	allNodes := make([]int, len(e.nodes))
	for i := range allNodes {
		allNodes[i] = i
	}
	for i := range e.nodesBags {
		e.nodesBags[i] = make([]*caravelaNode.Node, e.simulatorConfigs.TotalNumberOfNodes()/numOfRandomBagsOfNode)
		for j := range e.nodesBags[i] {
			randIndex := util.RandomInteger(0, len(allNodes)-1)
			e.nodesBags[i][j] = e.nodes[allNodes[randIndex]]
			allNodes = append(allNodes[:randIndex], allNodes[randIndex+1:]...)
		}
	}

	e.isInit = true
	util.Log.Info(util.LogTag(engineLogTag) + "Initialized")
}

// Start starts the simulator engine.
func (e *Engine) Start() {
	const ticksPerSnapshot = 9

	if !e.isInit {
		panic(errors.New("simulator is not initialized"))
	}

	util.Log.Info(util.LogTag(engineLogTag) + "Simulation started...")
	realStartTime := time.Now()

	simCurrentTime, numTicks := 0*time.Second, 0
	simLastTimeRefreshes, simLastTimeSpread := make([]time.Duration, numOfRandomBagsOfNode), make([]time.Duration, numOfRandomBagsOfNode)
	for i := range simLastTimeRefreshes {
		simLastTimeRefreshes[i] = time.Duration(i) * e.caravelaConfigs.RefreshingInterval()
		simLastTimeSpread[i] = time.Duration(i) * e.caravelaConfigs.SpreadOffersInterval()
	}
	ticksChan := make(chan chan feeder.RequestTask)

	go e.feeder.Start(ticksChan) // Start request feeder.

	for {
		util.Log.Infof(util.LogTag(engineLogTag)+"Sim Time: %.2f, Tick: %d, Ticks Remaining: %d",
			simCurrentTime.Seconds(), numTicks, e.simulatorConfigs.MaximumTicks()-numTicks)

		// 1st. Inject the requests in the nodes, introducing the liveness.
		e.acceptRequests(ticksChan, simCurrentTime)

		// 2nd. Do the actions dependent on time (e.g. actions fired by timers).
		simLastTimeRefreshes, simLastTimeSpread = e.fireTimerActions(simCurrentTime, simLastTimeRefreshes, simLastTimeSpread)

		// 3rd. Update metrics with system's current information.
		e.updateMetrics()

		// 4th. Update the engine time using the tick mechanism.
		simCurrentTime = simCurrentTime + e.simulatorConfigs.TicksInterval()
		numTicks++
		if numTicks == e.simulatorConfigs.MaximumTicks() {
			break
		}

		if numTicks != 0 && (numTicks%ticksPerSnapshot) == 0 {
			e.metricsCollector.Persist(simCurrentTime)
		}
	}

	close(ticksChan) // Alert feeder that the engine has ended.
	e.metricsCollector.EndSimulation(simCurrentTime)

	util.Log.Info(util.LogTag(engineLogTag) + "Simulation Ended")
	util.Log.Infof(util.LogTag(engineLogTag)+"Duration: Hours: %.2fh | Min: %.2fm | Sec: %.2fs",
		(time.Now().Sub(realStartTime)).Hours(), (time.Now().Sub(realStartTime)).Minutes(), (time.Now().Sub(realStartTime)).Seconds())
	e.release()
}

// acceptRequests receives requests from the feeder to be injected in the simulated caravela.
func (e *Engine) acceptRequests(ticksChan chan<- chan feeder.RequestTask, currentTime time.Duration) {
	const requestChanSize = 30
	defer e.workersPool.WaitAll()

	newTickChan := make(chan feeder.RequestTask, requestChanSize)
	ticksChan <- newTickChan

	for {
		select {
		case requestTask, more := <-newTickChan:
			if more {
				e.workersPool.WaitCount(1)
				e.workersPool.JobQueue <- func() {
					defer e.workersPool.JobDone()

					nodeIndex, node := e.selectInjectedNode()
					requestTask(nodeIndex, node, currentTime)
				}
			} else {
				return
			}
		}
	}
}

// fireTimerActions runs the actions dependent on the real time triggers/timers.
func (e *Engine) fireTimerActions(currentTime time.Duration, lastTimeRefreshes, lastTimeSpreadOffers []time.Duration) ([]time.Duration, []time.Duration) {
	defer e.workersPool.WaitAll()

	// 1. Refresh Offers.
	for i := range lastTimeRefreshes {
		if lastTimeRefreshes[i] < currentTime && (currentTime-lastTimeRefreshes[i]) >= e.caravelaConfigs.RefreshingInterval() {
			// Necessary because the tick interval can be greater than the refresh interval.
			timesToRefresh := int((currentTime - lastTimeRefreshes[i]) / e.caravelaConfigs.RefreshingInterval())
			util.Log.Debugf(util.LogTag(engineLogTag)+"Refreshing at %d, Times: %d", currentTime.Seconds(), timesToRefresh)

			for _, node := range e.nodesBags[i] {
				tempNode := node
				e.workersPool.WaitCount(1)
				e.workersPool.JobQueue <- func() {
					defer e.workersPool.JobDone()
					for i := 0; i < timesToRefresh; i++ {
						tempNode.RefreshOffersSim()
					}
				}
			}

			lastTimeRefreshes[i] = currentTime
		}
	}

	// 2. Spread Offers.
	for i := range lastTimeSpreadOffers {
		if lastTimeSpreadOffers[i] < currentTime && (currentTime-lastTimeSpreadOffers[i]) >= e.caravelaConfigs.SpreadOffersInterval() {
			// Necessary because the tick interval can be greater than the spread offers interval.
			timesToSpread := int((currentTime - lastTimeSpreadOffers[i]) / e.caravelaConfigs.SpreadOffersInterval())
			util.Log.Debugf(util.LogTag(engineLogTag)+"Spreading at %d, Times: %d", currentTime.Seconds(), timesToSpread)

			for _, node := range e.nodesBags[i] {
				tempNode := node
				e.workersPool.WaitCount(1)
				e.workersPool.JobQueue <- func() {
					defer e.workersPool.JobDone()
					for i := 0; i < timesToSpread; i++ {
						tempNode.SpreadOffersSim()
					}
				}
			}

			lastTimeSpreadOffers[i] = currentTime
		}
	}

	return lastTimeRefreshes, lastTimeSpreadOffers
}

// updateMetrics updates all the collector's metrics.
func (e *Engine) updateMetrics() {
	defer e.workersPool.WaitAll()

	for i, node := range e.nodes {
		tempI := i
		tempNode := node

		e.workersPool.WaitCount(1)
		e.workersPool.JobQueue <- func() {
			defer e.workersPool.JobDone()

			nodeFreeResources, nodeMaxResources, numActiveOffers, _ := tempNode.NodeInformationSim()
			e.assertNodeState(nodeFreeResources, nodeMaxResources)
			e.metricsCollector.SetNodeState(tempI, nodeFreeResources, int64(numActiveOffers), int64(tempNode.DebugSizeBytes()))
		}
	}
}

// selectInjectedNode selects a node to inject the user's requests.
func (e *Engine) selectInjectedNode() (int, *caravelaNode.Node) {
	var nodeIndex = 0
	var node *caravelaNode.Node = nil
	if e.caravelaConfigs.DiscoveryBackend() == "swarm" { // Inject the requests in the master node.
		nodeIndex = 0
		node = e.nodes[0]
	} else {
		nodeIndex, node = e.randomNode() // Inject the request in a random node of the system.
	}
	return nodeIndex, node
}

// release releases all the memory of the engine structures, nodes, etc.
func (e *Engine) release() {
	util.Log.Info(util.LogTag(engineLogTag) + "Clearing engine objects...")
	e.workersPool.Release()
	e.feeder = nil
	e.nodes = nil
	e.workersPool = nil
	if e.lastSimulation {
		e.overlayMock = nil
	}
	runtime.GC() // Force the GC to run in order to release the memory
	util.Log.Info(util.LogTag(engineLogTag) + "FINISHED")
}

// NodeByIP returns the caravela node and index given the node's IP address.
func (e *Engine) NodeByIP(ip string) (*caravelaNode.Node, int) {
	index, _ := e.overlayMock.GetNodeMockByIP(ip)
	return e.nodes[index], index
}

// NodeByGUID returns the caravela node and index given the node's GUID.
func (e *Engine) NodeByGUID(guid string) (*caravelaNode.Node, int) {
	index, _ := e.overlayMock.GetNodeMockByGUID(guid)
	return e.nodes[index], index
}

// randomNode returns a random node from the simulated active nodes.
func (e *Engine) randomNode() (int, *caravelaNode.Node) {
	randIndex := util.RandomInteger(0, len(e.nodes)-1)
	return randIndex, e.nodes[randIndex]
}

func (e *Engine) assertNodeState(freeResources, maximumResources types.Resources) {
	if freeResources.CPUs < 0 {
		panic(errors.New("negative free CPUs"))
	}
	if freeResources.Memory < 0 {
		panic(errors.New("negative free Memory"))
	}

	if freeResources.CPUs > maximumResources.CPUs {
		panic(errors.New("over free CPUs"))
	}
	if freeResources.Memory > maximumResources.Memory {
		panic(errors.New("over free Memory"))
	}
}
