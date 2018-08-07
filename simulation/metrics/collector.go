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

// Prefix name for the temp directory
const metricsTempDirName = "metrics-"

// Collector aggregates all the metrics information about the system during a simulation.
type Collector struct {
	numNodes  int      // Number of simulation nodes
	snapshots []Global // Array of global metrics snapshots

	outputDirPath  string // Output directory path
	tmpDirFullPath string // Temp directory to store intermediate metrics
}

// NewCollector creates a new metric's collector.
func NewCollector(numNodes int, outputDirPath string) *Collector {
	return &Collector{
		numNodes:      numNodes,
		snapshots:     make([]Global, 1),
		outputDirPath: outputDirPath,
	}
}

// Init initialize the metric's collector.
func (collector *Collector) Init(nodesMaxRes []types.Resources) {
	dirFullPath, err := ioutil.TempDir("", metricsTempDirName)
	if err != nil {
		panic(errors.New("Temp directory can't be created, error: " + err.Error()))
	}
	collector.tmpDirFullPath = dirFullPath

	os.Mkdir(collector.outputDirPath, 0644)

	collector.snapshots[0] = *NewGlobalInitial(collector.numNodes, time.Duration(0), nodesMaxRes)
}

// ========================= Metrics Collector Methods ====================================

func (collector *Collector) RunRequestSucceeded() {
	collector.activeGlobal().RunRequestSucceeded()
}

func (collector *Collector) APIRequestReceived(nodeIndex int) {
	collector.activeGlobal().APIRequestReceived(nodeIndex)
}

func (collector *Collector) SetAvailableNodeResources(nodeIndex int, res types.Resources) {
	collector.activeGlobal().SetAvailableNodeResources(nodeIndex, res)
}

// CreateRunRequest creates a new run request in order to gather its metrics.
func (collector *Collector) CreateRunRequest(nodeIndex int, requestID string, resources types.Resources,
	currentTime time.Duration) {
	collector.activeGlobal().CreateRunRequest(nodeIndex, requestID, resources, currentTime)
}

func (collector *Collector) IncrMessagesTradedRequest(requestID string, numMessages int) {
	collector.activeGlobal().IncrMessagesTradedRequest(requestID, numMessages)
}

func (collector *Collector) ArchiveRunRequest(requestID string) {
	collector.activeGlobal().ArchiveRunRequest(requestID)
}

// CreateNewGlobalSnapshot creates a snapshot of the system's current metrics and initialize a new one.
func (collector *Collector) CreateNewGlobalSnapshot(currentTime time.Duration) {
	collector.activeGlobal().SetEndTime(currentTime)
	collector.snapshots = append(collector.snapshots, *NewGlobalNext(collector.numNodes, collector.activeGlobal()))
}

func (collector *Collector) activeGlobal() *Global {
	return &collector.snapshots[len(collector.snapshots)-1]
}

func (collector *Collector) Persist(currentTime time.Duration) {
	collector.activeGlobal().SetEndTime(currentTime)

	for index, global := range collector.snapshots {
		jsonBytes, err := json.Marshal(&collector.snapshots[index])
		if err != nil {
			panic(errors.New("can't marshall the collector snapshot, error: " + err.Error()))
		}
		err = ioutil.WriteFile(collector.tmpDirFullPath+"\\"+global.Start.String()+".json", jsonBytes, 0644)
		if err != nil {
			panic(errors.New("can't write the collector snapshot to disk, error: " + err.Error()))
		}
	}

	newGlobal := NewGlobalNext(collector.numNodes, collector.activeGlobal())
	collector.snapshots = make([]Global, 1)
	collector.snapshots[0] = *newGlobal
}

func (collector *Collector) loadAllMetrics() {
	filesInfo, err := ioutil.ReadDir(collector.tmpDirFullPath)
	if err != nil {
		panic(errors.New("can't find the collector snapshot files, error: " + err.Error()))
	}

	for _, fileInfo := range filesInfo {
		if !fileInfo.IsDir() {
			var globalMetrics Global

			fileContent, err := ioutil.ReadFile(collector.tmpDirFullPath + "\\" + fileInfo.Name())
			if err != nil {
				panic(errors.New("can't read the snapshot file, error: " + err.Error()))
			}

			err = json.Unmarshal(fileContent, &globalMetrics)
			if err != nil {
				panic(errors.New("can't unmarshal the snapshot file content, error: " + err.Error()))
			}

			collector.snapshots = append(collector.snapshots, globalMetrics)
		}
	}

	sort.Sort(collector)
}

func (collector *Collector) Print() {
	collector.loadAllMetrics() // Load all the intermediate snapshots in memory

	totalRunRequests := int64(0)
	totalRunRequestsSucceeded := int64(0)
	for _, global := range collector.snapshots {
		totalRunRequests += global.TotalRunRequests()
		totalRunRequestsSucceeded += global.TotalRunRequestsSucceeded()
	}

	fmt.Printf("##################################################################\n")
	fmt.Printf("#                    SIMULATION RESULT METRICS                   #\n")
	fmt.Printf("##################################################################\n")
	fmt.Printf("Total Requests:         %d\n", totalRunRequests)
	fmt.Printf("Requests Succeeded:     %d\n", totalRunRequestsSucceeded)
	fmt.Printf("Request Success Ratio:  %.2f\n", float64(totalRunRequestsSucceeded)/float64(totalRunRequests))

	collector.plotGraphics()
}

func (collector *Collector) plotGraphics() {
	collector.plotRequestsSucceededOverTime()
	collector.plotRequestsMessagesTradedOverTime()
	collector.plotAvailableResourcesOverTime()
}

// Remove all the temporary files and resources used during the metrics gathering.
func (collector *Collector) Clear() {
	os.RemoveAll(collector.tmpDirFullPath) // clean up the simulation temp collector files
}

// ================================ Sort Interface =================================

func (collector *Collector) Len() int {
	return len(collector.snapshots)
}

func (collector *Collector) Swap(i, j int) {
	collector.snapshots[i], collector.snapshots[j] = collector.snapshots[j], collector.snapshots[i]
}

func (collector *Collector) Less(i, j int) bool {
	return collector.snapshots[i].StartTime() < collector.snapshots[j].StartTime()
}
