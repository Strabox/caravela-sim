package metrics

import (
	"errors"
	"fmt"
	"github.com/strabox/caravela-sim/engine/metrics/graphics"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"math"
	"sort"
)

func (coll *Collector) plotRequestsSucceeded() {
	plotRes := graphics.NewPlot("Deploy Containers Requests Success", "Time (Seconds)", "Requests Succeeded", true)

	dataPoints := make([]interface{}, 0)
	for simLabel, simData := range coll.simulations {
		pts := make(plotter.XYs, len(simData.snapshots))
		for i := range pts {
			pts[i].X = simData.snapshots[i].EndTime().Seconds()
			pts[i].Y = float64(simData.snapshots[i].TotalRunRequestsSucceeded())
		}
		dataPoints = append(dataPoints, simLabel, pts)
	}

	err := plotutil.AddLinePoints(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, generatePNGFileName(coll.outputDirPath+"\\"+"RequestsSucceeded"))
}

func (coll *Collector) plotRequestsMessagesTradedPerRequest() {
	plotRes := graphics.NewPlot("Average Discover Messages", "Time (Seconds)", "#Avg Messages", true)

	dataPoints := make([]interface{}, 0)
	for simLabel, simData := range coll.simulations {
		pts := make(plotter.XYs, len(simData.snapshots))
		for i := range pts {
			pts[i].X = simData.snapshots[i].EndTime().Seconds()
			pts[i].Y = float64(simData.snapshots[i].RunRequestsAvgMessages())
		}
		dataPoints = append(dataPoints, simLabel, pts)
	}

	err := plotutil.AddLinePoints(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, generatePNGFileName(coll.outputDirPath+"\\"+"MessagesPerRequest"))
}

func (coll *Collector) plotFreeResources() {
	for simLabel, simData := range coll.simulations {
		plotRes := graphics.NewPlot(fmt.Sprintf("Free Resources (%s)", simLabel), "Time (Seconds)",
			"Resources Free %", true)
		availableResPts := make(plotter.XYs, len(simData.snapshots))
		requestsSucceeded := make(plotter.XYs, len(simData.snapshots))
		for i := range availableResPts {
			availableResPts[i].X = simData.snapshots[i].EndTime().Seconds()
			availableResPts[i].Y = float64(simData.snapshots[i].AllAvailableResourcesAvg())
			requestsSucceeded[i].X = simData.snapshots[i].EndTime().Seconds()
			requestsSucceeded[i].Y = simData.snapshots[i].RunRequestSuccessRatio()
		}

		err := plotutil.AddLinePoints(plotRes, "Free Resources", availableResPts,
			"Requests Succeeded", requestsSucceeded)
		if err != nil {
			panic(errors.New("Problem with plots, error: " + err.Error()))
		}

		graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, generatePNGFileName(coll.outputDirPath+"\\"+"FreeResources_"+simLabel))
	}
}

func (coll *Collector) plotLookupMessagesPercentiles() {
	plotRes := graphics.NewPlot("Lookup Messages Traded Distribution", "", "#Messages", false)

	dataPoints := make([]interface{}, 0)
	for simLabel, simData := range coll.simulations {
		boxPoints := make(plotter.Values, 0)
		for _, snapshot := range simData.snapshots {
			boxPoints = append(boxPoints, snapshot.RequestsMessagesExchanged()...)
		}
		dataPoints = append(dataPoints, simLabel, boxPoints)
	}

	err := plotutil.AddBoxPlots(plotRes, vg.Points(55), dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 15*vg.Centimeter, 15*vg.Centimeter, generatePNGFileName(coll.outputDirPath+"\\"+"MessagesDistribution"))
}

func (coll *Collector) plotResourceDistribution() {
	for simLabel, simData := range coll.simulations {
		data := make([]float64, len(simData.snapshots)*coll.numNodes)
		lower := 0
		upper := coll.numNodes
		yTicks := make([]int, len(simData.snapshots)+1)
		for i, snapshot := range simData.snapshots {
			yTicks[i] = int(snapshot.EndTime().Seconds())
			sort.Sort(&snapshot) // Sort the snapshot's node's metric by the Max Resources
			copy(data[lower:upper], snapshot.ResourcesUsedNodeRatio())
			lower += coll.numNodes
			upper += coll.numNodes
		}

		yTicks[len(yTicks)-1] = math.MaxInt64
		sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

		dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), coll.numNodes, data)}

		graphics.NewHeatMap(generatePNGFileName("ResourceDistribution_"+simLabel), fmt.Sprintf("Resources Distribution (%s)", simLabel),
			"Nodes", "Time (Seconds)", coll.outputDirPath, yTicks, dataGrid)
	}
}

// ======================================= Debug Performance Plots ===============================

func (coll *Collector) plotRelayedGetOfferMessages() {
	plotRes := graphics.NewPlot("Total Relayed Get Offer", "Time (Seconds)", "#Relayed Get Offers", true)

	dataPoints := make([]interface{}, 0)
	for simLabel, simData := range coll.simulations {
		relayedGetOffersPts := make(plotter.XYs, len(simData.snapshots))
		for i := range relayedGetOffersPts {
			relayedGetOffersPts[i].X = simData.snapshots[i].EndTime().Seconds()
			relayedGetOffersPts[i].Y = float64(simData.snapshots[i].TotalGetOffersRelayed())
		}
		dataPoints = append(dataPoints, simLabel, relayedGetOffersPts)
	}

	err := plotutil.AddLinePoints(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, generatePNGFileName(coll.outputDirPath+"\\"+"Debug_GetOffersRelayed"))
}

func (coll *Collector) plotEmptyGetOfferMessages() {
	plotRes := graphics.NewPlot("Empty Get Offer Messages", "Time (Seconds)", "#Empty Get Offer", true)

	dataPoints := make([]interface{}, 0)
	for simLabel, simData := range coll.simulations {
		emptyGetOffersPts := make(plotter.XYs, len(simData.snapshots))
		for i := range emptyGetOffersPts {
			emptyGetOffersPts[i].X = simData.snapshots[i].EndTime().Seconds()
			emptyGetOffersPts[i].Y = float64(simData.snapshots[i].TotalEmptyGetOfferMessages())
		}
		dataPoints = append(dataPoints, simLabel, emptyGetOffersPts)
	}

	err := plotutil.AddLinePoints(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, generatePNGFileName(coll.outputDirPath+"\\"+"Debug_EmptyGetOffer"))
}

// ============================================== Auxiliary =========================================

func generatePNGFileName(fileName string) string {
	return fileName + ".png"
}
