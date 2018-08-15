package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api/types"
	"io/ioutil"
	"os"
	"sort"
	"time"
)

// metricsTempDirName is the prefix name for the metrics collections temporary directory.
const metricsTempDirName = "metrics-"

// simDirBaseName is the prefix name for the simulations output directories.
const simulationDirBaseName = "sim-"

// simulationDirSuffixFormat is the format for the suffix of the simulations output directories.
const simulationDirSuffixFormat = "2006-01-02_15h04m05s"

// Collector aggregates all the metrics information about the system during a simulation.
type Collector struct {
	numNodes  int      // Number of simulation nodes.
	snapshots []Global // Array of global metrics snapshots.

	outputDirPath  string // Output directory path.
	tmpDirFullPath string // Temp directory to store intermediate metrics.
}

// NewCollector creates a new metric's collector.
func NewCollector(numNodes int, baseOutputDirPath string) *Collector {
	return &Collector{
		numNodes:      numNodes,
		snapshots:     make([]Global, 1),
		outputDirPath: baseOutputDirPath + "\\" + simulationDirBaseName + time.Now().Format(simulationDirSuffixFormat),
	}
}

// Init initialize the metric's collector.
func (coll *Collector) Init(nodesMaxRes []types.Resources) {
	dirFullPath, err := ioutil.TempDir("", metricsTempDirName)
	if err != nil {
		panic(errors.New("Temp directory can't be created, error: " + err.Error()))
	}
	coll.tmpDirFullPath = dirFullPath

	os.MkdirAll(coll.outputDirPath, 0644)

	coll.snapshots[0] = *NewGlobalInitial(coll.numNodes, time.Duration(0), nodesMaxRes)
}

// ================================= Metrics Collector Methods ====================================

// GetOfferRelayed increment the number of messages traded from type GetOffersRelayed.
func (coll *Collector) GetOfferRelayed(amount int64) {
	coll.activeGlobal().GetOfferRelayed(amount)
}

// RunRequestSucceeded increments the number of run requests that were fulfilled with success.
func (coll *Collector) RunRequestSucceeded() {
	coll.activeGlobal().RunRequestSucceeded()
}

// APIRequestReceived increments the number of Caravela's API requests a node received.
func (coll *Collector) APIRequestReceived(nodeIndex int) {
	coll.activeGlobal().APIRequestReceived(nodeIndex)
}

// SetAvailableNodeResources sets the available resources of a node.
func (coll *Collector) SetAvailableNodeResources(nodeIndex int, res types.Resources) {
	coll.activeGlobal().SetAvailableNodeResources(nodeIndex, res)
}

// CreateRunRequest creates a new run request in order to gather its metrics.
func (coll *Collector) CreateRunRequest(nodeIndex int, requestID string, resources types.Resources,
	currentTime time.Duration) {
	coll.activeGlobal().CreateRunRequest(nodeIndex, requestID, resources, currentTime)
}

// IncrMessagesTradedRequest increment the number of messages traded to fulfill a run request.
func (coll *Collector) IncrMessagesTradedRequest(requestID string, numMessages int) {
	coll.activeGlobal().IncrMessagesTradedRequest(requestID, numMessages)
}

// ArchiveRunRequest archives the metrics of request that was happening because it ended.
func (coll *Collector) ArchiveRunRequest(requestID string) {
	coll.activeGlobal().ArchiveRunRequest(requestID)
}

// ================================= Collector Management Methods ===================================

// CreateNewGlobalSnapshot creates a snapshot of the system's current metrics and initialize a new one.
func (coll *Collector) CreateNewGlobalSnapshot(currentTime time.Duration) {
	coll.activeGlobal().SetEndTime(currentTime)
	coll.snapshots = append(coll.snapshots, *NewGlobalNext(coll.numNodes, coll.activeGlobal()))
}

// Persist is used to persist the in memory metrics into JSON files, in order to save memory.
func (coll *Collector) Persist(currentTime time.Duration) {
	coll.activeGlobal().SetEndTime(currentTime)

	for index, global := range coll.snapshots {
		jsonBytes, err := json.Marshal(&coll.snapshots[index])
		if err != nil {
			panic(errors.New("can't marshall the collector snapshot, error: " + err.Error()))
		}
		err = ioutil.WriteFile(coll.tmpDirFullPath+"\\"+global.Start.String()+".json", jsonBytes, 0644)
		if err != nil {
			panic(errors.New("can't write the collector snapshot to disk, error: " + err.Error()))
		}
	}

	newGlobal := NewGlobalNext(coll.numNodes, coll.activeGlobal())
	coll.snapshots = make([]Global, 1)
	coll.snapshots[0] = *newGlobal
}

// Stop is called when the simulation stops and there is no need to gather more metrics.
func (coll *Collector) End(endTime time.Duration) {
	coll.activeGlobal().SetEndTime(endTime)
}

// Clear removes all the temporary files and resources used during the metrics gathering.
func (coll *Collector) Clear() {
	os.RemoveAll(coll.tmpDirFullPath) // clean up the simulation temp collector files
}

// Print is used to gather all the metrics of the simulation into memory, consolidating them
// in order to produce results into the console and into the files.
func (coll *Collector) Print() {
	coll.loadAllMetrics() // Load all the intermediate snapshots in memory

	totalRunRequests := int64(0)
	totalRunRequestsSucceeded := int64(0)
	for _, global := range coll.snapshots {
		totalRunRequests += global.TotalRunRequests()
		totalRunRequestsSucceeded += global.TotalRunRequestsSucceeded()
	}

	fmt.Printf("##################################################################\n")
	fmt.Printf("#                    SIMULATION RESULT METRICS                   #\n")
	fmt.Printf("##################################################################\n")
	fmt.Printf("#Requests:               %d\n", totalRunRequests)
	fmt.Printf("#Requests Succeeded:     %d\n", totalRunRequestsSucceeded)
	fmt.Printf("#Requests Success Ratio: %.2f\n", float64(totalRunRequestsSucceeded)/float64(totalRunRequests))

	coll.plotGraphics()
}

// activeGlobal returns the current global snapshot that is gathering metrics.
func (coll *Collector) activeGlobal() *Global {
	return &coll.snapshots[len(coll.snapshots)-1]
}

// loadAllMetrics is used to fill the collector with all the metrics that were persisted into disk,
// in order to be analysed after that.
func (coll *Collector) loadAllMetrics() {
	filesInfo, err := ioutil.ReadDir(coll.tmpDirFullPath)
	if err != nil {
		panic(errors.New("can't find the collector snapshot files, error: " + err.Error()))
	}

	for _, fileInfo := range filesInfo {
		if !fileInfo.IsDir() {
			var globalMetrics Global

			fileContent, err := ioutil.ReadFile(coll.tmpDirFullPath + "\\" + fileInfo.Name())
			if err != nil {
				panic(errors.New("can't read the snapshot file, error: " + err.Error()))
			}

			err = json.Unmarshal(fileContent, &globalMetrics)
			if err != nil {
				panic(errors.New("can't unmarshal the snapshot file content, error: " + err.Error()))
			}

			coll.snapshots = append(coll.snapshots, globalMetrics)
		}
	}

	sort.Sort(coll) // Sort the global snapshots by the sim time of them.
}

// plotGraphics plots all the charts/plots based on the metrics collected.
func (coll *Collector) plotGraphics() {
	coll.plotRequestsSucceeded()
	coll.plotRequestsMessagesTradedPerRequest()
	coll.plotFreeResources()
	coll.plotRelayedGetOfferMessages()
	coll.plotResourceDistribution()
}

// ===================================== Sort Interface =======================================

func (coll *Collector) Len() int {
	return len(coll.snapshots)
}

func (coll *Collector) Swap(i, j int) {
	coll.snapshots[i], coll.snapshots[j] = coll.snapshots[j], coll.snapshots[i]
}

func (coll *Collector) Less(i, j int) bool {
	return coll.snapshots[i].StartTime() < coll.snapshots[j].StartTime()
}
