package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/ivpusic/grpool"
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela/api/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

// metricsTempDirName is the prefix name for the metrics collections temporary directory.
const metricsTempDirName = "metrics-"

// simDirBaseName is the prefix name for the simulations output directories.
const simulationDirBaseName = "sim-"

// simulationDirSuffixFormat is the format for the suffix of the simulations output directories.
const simulationDirSuffixFormat = "2006-01-02_15h04m05s"

// simulationData represents a complete simulation data.
type simulationData struct {
	label          string   // Label to identify the simulation.
	snapshots      []Global // System snapshots over time of the simulation.
	tmpDirFullPath string   // Temporary directory to store metrics for the simulation.
}

// ===================================== Sort Interface =======================================

func (sim *simulationData) Len() int {
	return len(sim.snapshots)
}

func (sim *simulationData) Swap(i, j int) {
	sim.snapshots[i], sim.snapshots[j] = sim.snapshots[j], sim.snapshots[i]
}

func (sim *simulationData) Less(i, j int) bool {
	return sim.snapshots[i].StartTime() < sim.snapshots[j].StartTime()
}

// Collector aggregates all the metrics information about the system during a engine.
type Collector struct {
	numNodes       int               // Number of engine nodes.
	currSimulation *simulationData   // Current simulation.
	simulations    []*simulationData // Contains all the simulations identified by its label.

	simulatorConfigs *configuration.Configuration
	outputDirPath    string // Output directory path.
}

// NewCollector creates a new metric's collector.
func NewCollector(numNodes int, baseOutputDirPath string, simulatorConfigs *configuration.Configuration) *Collector {
	return &Collector{
		numNodes:       numNodes,
		currSimulation: nil,
		simulations:    make([]*simulationData, 0),

		simulatorConfigs: simulatorConfigs,
		outputDirPath:    filepath.Join(baseOutputDirPath, simulationDirBaseName+time.Now().Format(simulationDirSuffixFormat)),
	}
}

// InitNewSimulation initialize the metric's collector.
func (c *Collector) InitNewSimulation(simLabel string, nodesMaxRes []types.Resources) {
	newSimulation := &simulationData{
		label:     simLabel,
		snapshots: make([]Global, 1),
	}
	c.currSimulation = newSimulation

	dirFullPath, err := ioutil.TempDir("", metricsTempDirName+newSimulation.label+"-")
	if err != nil {
		panic(errors.New("Temp directory can't be created, error: " + err.Error()))
	}
	newSimulation.tmpDirFullPath = dirFullPath

	os.MkdirAll(c.outputDirPath, 0644)

	newSimulation.snapshots[0] = *NewGlobalInitial(c.numNodes, time.Duration(0), nodesMaxRes)
}

// ================================= Metrics Collector Methods ====================================

// GetOfferRelayed increment the number of messages traded from type GetOffersRelayed.
func (c *Collector) GetOfferRelayed(amount int64) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.GetOfferRelayed(amount)
	}
}

func (c *Collector) EmptyGetOfferMessage(amount int64) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.EmptyGetOfferMessages(amount)
	}
}

// MessageReceived increments the number of messages received by the node.
func (c *Collector) MessageReceived(nodeIndex int, amount int64, requestSizeBytes int64) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.MessageReceived(nodeIndex, amount, requestSizeBytes)
	}
}

// SetNodeState sets the available resources of a node.
func (c *Collector) SetNodeState(nodeIndex int, freeResources types.Resources, traderActiveOffers int64, memoryOccupied int64) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.SetNodeState(nodeIndex, freeResources, traderActiveOffers, memoryOccupied)
	}
}

// CreateRunRequest creates a new run request in order to gather its metrics.
func (c *Collector) CreateRunRequest(nodeIndex int, requestID string, resources types.Resources) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.CreateRunRequest(nodeIndex, requestID, resources)
	}
}

// IncrMessagesTradedRequest increment the number of messages traded to fulfill a run request.
func (c *Collector) IncrMessagesTradedRequest(requestID string, numMessages int) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.IncrMessagesTradedRequest(requestID, numMessages)
	}
}

// ArchiveRunRequest archives the metrics of request that was happening because it ended.
func (c *Collector) ArchiveRunRequest(requestID string, succeeded bool) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.ArchiveRunRequest(requestID, succeeded)
	}
}

// ================================= Collector Management Methods ===================================

// CreateNewGlobalSnapshot creates a snapshot of the system's current metrics and initialize a new one.
func (c *Collector) CreateNewGlobalSnapshot(currentTime time.Duration) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.SetEndTime(currentTime)
		c.currSimulation.snapshots = append(c.currSimulation.snapshots, *NewGlobalNext(c.numNodes, activeGlobal))
	}
}

// Persist is used to persist the in memory metrics into JSON files, in order to save memory.
func (c *Collector) Persist(currentTime time.Duration) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.SetEndTime(currentTime)

		for index, global := range c.currSimulation.snapshots {
			jsonBytes, err := json.Marshal(&c.currSimulation.snapshots[index])
			if err != nil {
				panic(errors.New("can't marshall the collector snapshot, error: " + err.Error()))
			}
			err = ioutil.WriteFile(filepath.Join(c.currSimulation.tmpDirFullPath, global.Start.String()+".json"), jsonBytes, 0644)
			if err != nil {
				panic(errors.New("can't write the collector snapshot to disk, error: " + err.Error()))
			}
		}

		newGlobal := NewGlobalNext(c.numNodes, activeGlobal)
		c.currSimulation.snapshots = make([]Global, 1)
		c.currSimulation.snapshots[0] = *newGlobal
	}
}

// EndSimulation is called when the simulator's engine stops and there is no need to gather more metrics.
func (c *Collector) EndSimulation(endTime time.Duration) {
	if activeGlobal, err := c.activeGlobal(); err == nil {
		activeGlobal.SetEndTime(endTime)
		c.simulations = append(c.simulations, c.currSimulation)
		c.currSimulation = nil
	}
}

// Clear removes all the temporary files and resources used during the metrics gathering.
func (c *Collector) Clear() {
	for _, simData := range c.simulations {
		os.RemoveAll(simData.tmpDirFullPath) // clean up the engine temp collector files
	}
}

// Print is used to gather all the metrics of the engine into memory, consolidating them
// in order to produce results into the console and into the files.
func (c *Collector) Print() {
	c.loadAllMetrics() // Load all the intermediate snapshots in memory

	for _, simData := range c.simulations {
		totalRunRequests := int64(0)
		totalRunRequestsSucceeded := int64(0)
		for _, global := range simData.snapshots {
			totalRunRequests += global.TotalRunRequests()
			totalRunRequestsSucceeded += global.TotalRunRequestsSucceeded()
		}

		fmt.Printf("##################################################################\n")
		fmt.Printf("#          SIMULATION RESULT METRICS (%s)     #\n", simData.label)
		fmt.Printf("##################################################################\n")
		fmt.Printf("Requests:               %d\n", totalRunRequests)
		fmt.Printf("Requests Succeeded:     %d\n", totalRunRequestsSucceeded)
		fmt.Printf("Requests Success Ratio: %.2f\n", float64(totalRunRequestsSucceeded)/float64(totalRunRequests))
	}

	c.plotGraphics() // Plot the graphics for the simulations
}

// activeGlobal returns the current global snapshot that is gathering metrics.
func (c *Collector) activeGlobal() (*Global, error) {
	if c.currSimulation == nil {
		return nil, errors.New("no active simulation")
	}
	return &c.currSimulation.snapshots[len(c.currSimulation.snapshots)-1], nil
}

// loadAllMetrics is used to fill the collector with all the metrics that were persisted into disk,
// in order to be analysed after that.
func (c *Collector) loadAllMetrics() {
	for _, simData := range c.simulations {
		filesInfo, err := ioutil.ReadDir(simData.tmpDirFullPath)
		if err != nil {
			panic(errors.New("can't find the collector snapshot files, error: " + err.Error()))
		}

		for _, fileInfo := range filesInfo {
			if !fileInfo.IsDir() {
				var globalMetrics Global

				fileContent, err := ioutil.ReadFile(filepath.Join(simData.tmpDirFullPath, fileInfo.Name()))
				if err != nil {
					panic(errors.New("can't read the snapshot file, error: " + err.Error()))
				}

				err = json.Unmarshal(fileContent, &globalMetrics)
				if err != nil {
					panic(errors.New("can't unmarshal the snapshot file content, error: " + err.Error()))
				}

				simData.snapshots = append(simData.snapshots, globalMetrics)
			}
		}
	}

	for _, simData := range c.simulations {
		sort.Sort(simData) // Sort the global snapshots by the sim time of them.
		for _, snapshot := range simData.snapshots {
			sort.Sort(&snapshot)
		}
	}
}

// plotGraphics plots all the charts/plots based on the metrics collected.
func (c *Collector) plotGraphics() {
	// Goroutine pool used to plot the graphics in parallel.
	maxWorkers := runtime.NumCPU() + 1
	goroutinePool := grpool.NewPool(maxWorkers, maxWorkers*3)

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotMemoryUsedByNodeV2()
		c.plotBandwidthUsedByNodeV2()
		c.plotMessagesReceivedByNodeV2()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotMessagesTradedByDeployRequestLinePlot()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotRequestsRate()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotBandwidthUsedByNode()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotResourcesAllocationEfficiency()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotTotalMessagesTradedInSystem()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotRequestsSucceeded()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotSystemUsedResourcesVSRequestSuccess()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotMessagesTradedByDeployRequestBoxPlot()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotResourcesUsedDistributionByNodesOverTime()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotMessagesDistributionByNodes()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotActiveOffersByNode()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotMasterNodeMessagesReceivedOverTime()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitCount(1)
	goroutinePool.JobQueue <- func() {
		c.plotMemoryUsedByNode()
		goroutinePool.JobDone()
	}

	goroutinePool.WaitAll() // Wait for all the plots to be completed.
	goroutinePool.Release() // Release goroutinePool resources.
	goroutinePool = nil
}
