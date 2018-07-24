package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api/types"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"io/ioutil"
	"os"
	"sort"
	"time"
)

// Prefix name for the temp directory
const metricsTempDirName = "metrics-"

type Metrics struct {
	numNodes  int      // Number of simulation nodes
	snapshots []Global // Array of global metrics snapshots

	outputDirPath  string // Output directory path
	tmpDirFullPath string // Temp directory to store intermediate metrics
}

func NewMetrics(numNodes int, outputDirPath string) *Metrics {
	return &Metrics{
		numNodes:      numNodes,
		snapshots:     make([]Global, 1),
		outputDirPath: outputDirPath,
	}
}

func (metrics *Metrics) Init() {
	dirFullPath, err := ioutil.TempDir("", metricsTempDirName)
	if err != nil {
		panic(errors.New("Temp directory can't be created, error: " + err.Error()))
	}
	metrics.tmpDirFullPath = dirFullPath

	os.Mkdir(metrics.outputDirPath, 0644)

	metrics.snapshots[0] = *NewGlobal(metrics.numNodes, time.Duration(0))
	metrics.snapshots[0].Init()
}

func (metrics *Metrics) RunRequestSucceeded() {
	metrics.activeGlobal().RunRequestSucceeded()
}

func (metrics *Metrics) MsgsTradedActiveRequest(numMsgs int) {
	metrics.activeGlobal().MsgsTradedActiveRequest(numMsgs)
}

func (metrics *Metrics) CreateNewSnapshot(currentTime time.Duration) {
	metrics.activeGlobal().SetEndTime(currentTime)
	newGlobal := NewGlobal(metrics.numNodes, currentTime)
	newGlobal.Init()
	metrics.snapshots = append(metrics.snapshots, *newGlobal)
}

func (metrics *Metrics) CreateRunRequest(nodeIndex int, resources types.Resources,
	currentTime time.Duration) {
	metrics.activeGlobal().CreateRunRequest(nodeIndex, resources, currentTime)
}

func (metrics *Metrics) activeGlobal() *Global {
	return &metrics.snapshots[len(metrics.snapshots)-1]
}

func (metrics *Metrics) Persist(currentTime time.Duration) {
	metrics.activeGlobal().SetEndTime(currentTime)

	for index, global := range metrics.snapshots {
		jsonBytes, err := json.Marshal(&metrics.snapshots[index])
		if err != nil {
			panic(errors.New("can't marshall the metrics snapshot, error: " + err.Error()))
		}
		err = ioutil.WriteFile(metrics.tmpDirFullPath+"\\"+global.StartTime.String()+".json", jsonBytes, 0644)
		if err != nil {
			panic(errors.New("can't write the metrics snapshot to disk, error: " + err.Error()))
		}
	}

	newGlobal := NewGlobal(metrics.numNodes, currentTime)
	newGlobal.Init()
	metrics.snapshots = make([]Global, 1)
	metrics.snapshots[0] = *newGlobal
}

func (metrics *Metrics) loadAllMetrics() {
	filesInfo, err := ioutil.ReadDir(metrics.tmpDirFullPath)
	if err != nil {
		panic(errors.New("can't find the metrics snapshot files, error: " + err.Error()))
	}

	for _, fileInfo := range filesInfo {
		if !fileInfo.IsDir() {
			var globalMetrics Global

			fileContent, err := ioutil.ReadFile(metrics.tmpDirFullPath + "\\" + fileInfo.Name())
			if err != nil {
				panic(errors.New("can't read the snapshot file, error: " + err.Error()))
			}

			err = json.Unmarshal(fileContent, &globalMetrics)
			if err != nil {
				panic(errors.New("can't unmarshal the snapshot file content, error: " + err.Error()))
			}

			metrics.snapshots = append(metrics.snapshots, globalMetrics)
		}
	}

	sort.Sort(metrics)
}

func (metrics *Metrics) Print() {
	metrics.loadAllMetrics() // Load all the intermediate snapshots in memory

	totalRunRequests := int64(0)
	totalRunRequestsSucceeded := int64(0)
	for _, global := range metrics.snapshots {
		totalRunRequests += global.RunRequests()
		totalRunRequestsSucceeded += global.TotalRunRequestsSucceeded()
	}

	fmt.Printf("##################################################################\n")
	fmt.Printf("#                    SIMULATION RESULT METRICS                   #\n")
	fmt.Printf("##################################################################\n")
	fmt.Printf("Total Requests: %d\n", totalRunRequests)
	fmt.Printf("Requests Succeeded: %d\n", totalRunRequestsSucceeded)
	fmt.Printf("Request Success Ratio: %.2f\n", float64(totalRunRequestsSucceeded)/float64(totalRunRequests))

	metrics.plot()
}

func (metrics *Metrics) plot() {
	p, err := plot.New()
	p.Title.Text = "Requests success over time"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Request Succeeded"

	pts := make(plotter.XYs, len(metrics.snapshots))
	for i := range pts {
		pts[i].X = metrics.snapshots[i].StartTimes().Seconds()
		pts[i].Y = float64(metrics.snapshots[i].TotalRunRequestsSucceeded())
	}

	err = plotutil.AddLinePoints(p,
		"Requests", pts)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	// Save the plot to a PNG file.
	if err := p.Save(7*vg.Inch, 5*vg.Inch, metrics.outputDirPath+"\\"+"points.png"); err != nil {
		panic(err)
	}
}

func (metrics *Metrics) Clear() {
	os.RemoveAll(metrics.tmpDirFullPath) // clean up the simulation temp metrics files
}

// ===============================================================================
// =							  Sort Interface                                 =
// ===============================================================================

func (metrics *Metrics) Len() int {
	return len(metrics.snapshots)
}

func (metrics *Metrics) Swap(i, j int) {
	metrics.snapshots[i], metrics.snapshots[j] = metrics.snapshots[j], metrics.snapshots[i]
}

func (metrics *Metrics) Less(i, j int) bool {
	return metrics.snapshots[i].StartTimes() < metrics.snapshots[j].StartTimes()
}
