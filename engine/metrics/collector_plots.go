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
	"path/filepath"
	"sort"
)

func (coll *Collector) plotRequestsSucceeded() {
	plotRes := graphics.NewPlot("Deploy Containers Requests Success", "Time (Seconds)", "Requests Succeeded", true)

	dataPoints := make([]interface{}, 0)
	for _, simData := range coll.simulations {
		pts := make(plotter.XYs, len(simData.snapshots))

		runRequestsSucceededAcc := int64(0)
		for i := range pts {
			runRequestsSucceededAcc += simData.snapshots[i].TotalRunRequestsSucceeded()
			pts[i].X = simData.snapshots[i].EndTime().Seconds()
			pts[i].Y = float64(runRequestsSucceededAcc)
		}
		dataPoints = append(dataPoints, simData.label, pts)
	}

	err := plotutil.AddLinePoints(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, generatePNGFileName(filepath.Join(coll.outputDirPath, "RequestsSucceeded")))
}

func (coll *Collector) plotRequestsMessagesTradedPerRequest() {
	plotRes := graphics.NewPlot("Average Discover Messages", "Time (Seconds)", "#Avg Messages", true)

	dataPoints := make([]interface{}, 0)
	for _, simData := range coll.simulations {
		simulationPts := make(plotter.XYs, len(simData.snapshots))
		for i, j := 0, 0; i < len(simData.snapshots); i++ {
			runRequestsAvgMessages := simData.snapshots[i].RunRequestsAvgMessages()
			if runRequestsAvgMessages == 0 { // No requests happened in this snapshot.
				simulationPts = simulationPts[:len(simulationPts)-1]
				continue
			}
			simulationPts[j].X = simData.snapshots[i].EndTime().Seconds()
			simulationPts[j].Y = float64(runRequestsAvgMessages)
			j++
		}
		dataPoints = append(dataPoints, simData.label, simulationPts)
	}

	err := plotutil.AddLinePoints(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	plotRes.Legend.Top = true

	graphics.Save(plotRes, 40*vg.Centimeter, 27*vg.Centimeter, generatePNGFileName(filepath.Join(coll.outputDirPath, "MessagesPerRequest")))
}

func (coll *Collector) plotSystemFreeResourcesVSRequestSuccess() {
	for _, simData := range coll.simulations {
		plotRes := graphics.NewPlot(fmt.Sprintf("Free Resources (%s)", simData.label), "Time (Seconds)",
			"Resources Free %", true)
		availableResPts := make(plotter.XYs, len(simData.snapshots))
		requestsSucceeded := make(plotter.XYs, len(simData.snapshots))
		for i, r := 0, 0; i < len(simData.snapshots); i++ {
			availableResPts[i].X = simData.snapshots[i].EndTime().Seconds()
			availableResPts[i].Y = float64(simData.snapshots[i].AllAvailableResourcesAvg())

			runRequestSuccessRatio := simData.snapshots[i].RunRequestSuccessRatio()
			if runRequestSuccessRatio == 0 {
				requestsSucceeded = requestsSucceeded[:len(requestsSucceeded)-1]
				continue
			}
			requestsSucceeded[r].X = simData.snapshots[i].EndTime().Seconds()
			requestsSucceeded[r].Y = runRequestSuccessRatio
			r++
		}

		err := plotutil.AddLinePoints(plotRes, "Free Resources", availableResPts,
			"Requests Succeeded", requestsSucceeded)
		if err != nil {
			panic(errors.New("Problem with plots, error: " + err.Error()))
		}

		graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, generatePNGFileName(filepath.Join(coll.outputDirPath, "FreeResources_"+simData.label)))
	}
}

func (coll *Collector) plotMessagesTraderByRequestBoxPlots() {
	plotRes := graphics.NewPlot("Lookup Messages Traded Distribution", "", "#Messages", false)

	dataPoints := make([]interface{}, 0)
	for _, simData := range coll.simulations {
		boxPoints := make(plotter.Values, 0)
		for _, snapshot := range simData.snapshots {
			boxPoints = append(boxPoints, snapshot.MessagesExchangedByRequest()...)
		}
		dataPoints = append(dataPoints, simData.label, boxPoints)
	}

	err := plotutil.AddBoxPlots(plotRes, vg.Points(55), dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 15*vg.Centimeter, 15*vg.Centimeter, generatePNGFileName(filepath.Join(coll.outputDirPath, "MessagesDistribution")))
}

func (coll *Collector) plotResourcesUsedDistributionByNodesOverTime() {
	for _, simData := range coll.simulations {
		data := make([]float64, len(simData.snapshots)*coll.numNodes)
		lower := 0
		upper := coll.numNodes
		yTicks := make([]int, len(simData.snapshots)+1)
		for i, snapshot := range simData.snapshots {
			yTicks[i] = int(snapshot.EndTime().Seconds())
			sort.Sort(&snapshot) // Sort the snapshot's node's metric by the Max Resources in ascending order.
			copy(data[lower:upper], snapshot.ResourcesUsedNodeRatio())
			lower += coll.numNodes
			upper += coll.numNodes
		}

		yTicks[len(yTicks)-1] = math.MaxInt64
		sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

		dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), coll.numNodes, data)}

		graphics.NewHeatMap(generatePNGFileName("ResourcesUsedDistribution_"+simData.label), fmt.Sprintf("Resources Used Distribution (%s)", simData.label),
			"Nodes", "Time (Seconds)", coll.outputDirPath, yTicks, dataGrid, graphics.MyInvertedHeatPalette())
	}
}

func (coll *Collector) plotResourcesUnreachableDistributionByNodesOverTime() {
	for _, simData := range coll.simulations {
		data := make([]float64, len(simData.snapshots)*coll.numNodes)
		lower := 0
		upper := coll.numNodes
		yTicks := make([]int, len(simData.snapshots)+1)
		for i, snapshot := range simData.snapshots {
			yTicks[i] = int(snapshot.EndTime().Seconds())
			sort.Sort(&snapshot) // Sort the snapshot's node's metric by the Max Resources in ascending order.
			copy(data[lower:upper], snapshot.ResourcesUnreachableRatioNode())
			lower += coll.numNodes
			upper += coll.numNodes
		}

		yTicks[len(yTicks)-1] = math.MaxInt64
		sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

		dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), coll.numNodes, data)}

		graphics.NewHeatMap(generatePNGFileName("ResourcesUnreachableDistribution_"+simData.label), fmt.Sprintf("Resources Unreachable Distribution (%s)", simData.label),
			"Nodes", "Time (Seconds)", coll.outputDirPath, yTicks, dataGrid, graphics.MyInvertedHeatPalette())
	}
}

func (coll *Collector) plotMessagesAPIDistributionByNodesOverTime() {
	for _, simData := range coll.simulations {
		data := make([]float64, len(simData.snapshots)*coll.numNodes)
		lower := 0
		upper := coll.numNodes
		yTicks := make([]int, len(simData.snapshots)+1)
		for i, snapshot := range simData.snapshots {
			yTicks[i] = int(snapshot.EndTime().Seconds())
			sort.Sort(&snapshot) // Sort the snapshot's node's metric by the Max Resources in ascending order.
			copy(data[lower:upper], snapshot.TotalAPIMessagesReceivedByNode())
			lower += coll.numNodes
			upper += coll.numNodes
		}

		yTicks[len(yTicks)-1] = math.MaxInt64
		sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

		dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), coll.numNodes, data)}

		graphics.NewHeatMap(generatePNGFileName("MessagesAPIDistribution_"+simData.label), fmt.Sprintf("Messages Distribution (%s)", simData.label),
			"Nodes", "Time (Seconds)", coll.outputDirPath, yTicks, dataGrid, graphics.MyHeatPalette())
	}
}

func (coll *Collector) plotTotalMessagesTradedInSystem() {
	plotRes := graphics.NewPlot("Total Messages Traded", "Main Simulation", "#Messages Traded", false)

	barWidth := vg.Points(25)

	barOffset := float64(0)
	color := 0
	for i, simData := range coll.simulations {
		messagesTradedAcc := float64(0)
		for _, snapshot := range simData.snapshots {
			messagesTradedAcc += snapshot.TotalAPIMessagesReceivedByAllNodes()
		}

		barChart, err := plotter.NewBarChart(plotter.Values{messagesTradedAcc}, barWidth)
		if err != nil {
			panic(fmt.Errorf("error creating bar chart, %s", err))
		}

		barChart.LineStyle.Width = vg.Length(0)
		barChart.Color = plotutil.Color(color)
		barChart.Offset = vg.Points(barOffset)

		plotRes.Add(barChart)

		plotRes.Legend.Add(simData.label, barChart)
		if i%2 == 0 {
			barOffset = -30
		} else {
			barOffset = 30
		}

		color++
	}

	plotRes.Legend.Top = true
	plotRes.NominalX("Main Simulation")

	graphics.Save(plotRes, 13*vg.Centimeter, 14*vg.Centimeter, generatePNGFileName(filepath.Join(coll.outputDirPath, "MessagesTradedInSystem")))
}

// ======================================= Debug Performance Plots ===============================

func (coll *Collector) plotRelayedGetOfferMessages() {
	plotRes := graphics.NewPlot("Total Relayed Get Offer", "Time (Seconds)", "#Relayed Get Offers", true)

	dataPoints := make([]interface{}, 0)
	for _, simData := range coll.simulations {
		relayedGetOffersPts := make(plotter.XYs, len(simData.snapshots))
		for i := range relayedGetOffersPts {
			relayedGetOffersPts[i].X = simData.snapshots[i].EndTime().Seconds()
			relayedGetOffersPts[i].Y = float64(simData.snapshots[i].TotalGetOffersRelayed())
		}
		dataPoints = append(dataPoints, simData.label, relayedGetOffersPts)
	}

	err := plotutil.AddLinePoints(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, generatePNGFileName(filepath.Join(coll.outputDirPath, "Debug_GetOffersRelayed")))
}

func (coll *Collector) plotEmptyGetOfferMessages() {
	plotRes := graphics.NewPlot("Empty Get Offer Messages", "Time (Seconds)", "#Empty Get Offer", true)

	dataPoints := make([]interface{}, 0)
	for _, simData := range coll.simulations {
		emptyGetOffersPts := make(plotter.XYs, len(simData.snapshots))
		for i := range emptyGetOffersPts {
			emptyGetOffersPts[i].X = simData.snapshots[i].EndTime().Seconds()
			emptyGetOffersPts[i].Y = float64(simData.snapshots[i].TotalEmptyGetOfferMessages())
		}
		dataPoints = append(dataPoints, simData.label, emptyGetOffersPts)
	}

	err := plotutil.AddLinePoints(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, generatePNGFileName(filepath.Join(coll.outputDirPath, "Debug_EmptyGetOffer")))
}

// ======================================== Auxiliary Functions ====================================

func generatePNGFileName(fileName string) string {
	return fileName + ".png"
}
