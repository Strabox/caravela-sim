package metrics

import (
	"errors"
	"fmt"
	"github.com/strabox/caravela-sim/engine/metrics/graphics"
	"github.com/strabox/caravela-sim/util"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"math"
	"path/filepath"
	"sort"
)

const plotsLogTag = "PLOTS"

const boxPlotWidth = 20
const boxPlotOverTimePNGWidth = 40 * vg.Centimeter
const boxPlotOverTimePNGHeight = 12 * vg.Centimeter

const linePlotOverTimePNGWidth = 12 * vg.Centimeter
const linePlotOverTimePNGHeight = 12 * vg.Centimeter

const heatMapOverTimePNGWidth = 1600
const heatMapOverTimePNGHeight = 720

// resultPlot is used as a temporary structure to save gonum/plot structures and its visual label.
type resultPlot struct {
	plot  *plot.Plot
	label string
}

func (coll *Collector) plotActiveOffersByNode() {
	util.Log.Infof(util.LogTag(plotsLogTag) + "Plot Active Offers per Node")
	// Box Plot over time.
	minY, maxY := math.MaxFloat64, float64(0)
	resPlots := make([]*resultPlot, 0)
	for _, simData := range coll.simulations {
		if isOfferingBasedStrategy(simData.label) {
			plotRes := graphics.NewPlot(fmt.Sprintf("Active Offers Per Node (%s)", visualStrategyName(simData.label)),
				"Time (hh:mm)", "#Active Offers", false)

			boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
			xLabels := make([]string, len(simData.snapshots))
			for i, snapshot := range simData.snapshots {
				boxPlotPoints := make(plotter.Values, 0)
				boxPlotPoints = append(boxPlotPoints, snapshot.TotalTraderActiveOfferPerNode()...)
				boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), float64(i), boxPlotPoints)
				xLabels[i] = util.FmtDuration(snapshot.EndTime())
			}

			for i := range boxPlots {
				util.Log.Infof(util.LogTag(plotsLogTag)+"TraderActiveOffers: Outliers %d, Ratio: %.2f%%", len(boxPlots[i].Outside), (float64(len(boxPlots[i].Outside))/float64(len(boxPlots[i].Values)))*100)
				plotRes.Add(boxPlots[i])
			}
			plotRes.NominalX(xLabels...)

			if plotRes.Y.Max > maxY {
				maxY = plotRes.Y.Max
			}
			if plotRes.Y.Min < minY {
				minY = plotRes.Y.Min
			}

			resPlots = append(resPlots, &resultPlot{plot: plotRes, label: simData.label})
		}
	}

	for _, resPlot := range resPlots {
		resPlot.plot.Y.Max = maxY
		resPlot.plot.Y.Min = minY
		graphics.Save(resPlot.plot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
			generatePNGFileName(coll.outputDirPath, "ActiveOffersDistributionBoxPlot_"+visualStrategyName(resPlot.label)))
	}

	// Heat Map.
	/* Heat Map.
	for _, simData := range coll.simulations {
		if isOfferingBasedStrategy(simData.label) {
			data := make([]float64, len(simData.snapshots)*coll.numNodes)
			lower := 0
			upper := coll.numNodes
			yTicks := make([]int, len(simData.snapshots)+1)
			for i, snapshot := range simData.snapshots {
				yTicks[i] = int(snapshot.EndTime().Seconds())
				copy(data[lower:upper], snapshot.TotalTraderActiveOfferPerNode())
				lower += coll.numNodes
				upper += coll.numNodes
			}

			yTicks[len(yTicks)-1] = math.MaxInt64
			sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

			dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), coll.numNodes, data)}

			graphics.NewHeatMap(generatePNGFileName(coll.outputDirPath, "ActiveOffersDistributionHeatMap_"+visualStrategyName(simData.label)),
				fmt.Sprintf("Active Offers Per Node (%s)", visualStrategyName(simData.label)),
				"Nodes", "Time (Seconds)", yTicks, dataGrid, palette.Heat(64, 1))
		}
	}
	*/
}

func (coll *Collector) plotMemoryUsedByNode() {
	// Box Plot over time.
	minY, maxY := math.MaxFloat64, float64(0)
	resPlots := make([]*resultPlot, 0)
	for _, simData := range coll.simulations {
		plotRes := graphics.NewPlot(fmt.Sprintf("Memory Used by Node (%s)", visualStrategyName(simData.label)),
			"Time (hh:mm)", "Memory Used (bytes)", false)

		boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
		xLabels := make([]string, len(simData.snapshots))
		for i, snapshot := range simData.snapshots {
			boxPlotPoints := make(plotter.Values, 0)
			boxPlotPoints = append(boxPlotPoints, snapshot.TotalMemoryUsedByNode()...)
			boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), float64(i), boxPlotPoints)
			xLabels[i] = util.FmtDuration(snapshot.EndTime())
		}

		for i := range boxPlots {
			util.Log.Infof(util.LogTag(plotsLogTag)+"MemoryUsed: Outliers %d, Ratio: %.2f%%",
				len(boxPlots[i].Outside), (float64(len(boxPlots[i].Outside))/float64(len(boxPlots[i].Values)))*100)
			plotRes.Add(boxPlots[i])
		}
		plotRes.NominalX(xLabels...)

		if plotRes.Y.Max > maxY && !isSwarmBasedStrategy(simData.label) {
			maxY = plotRes.Y.Max
		}
		if plotRes.Y.Min < minY && !isSwarmBasedStrategy(simData.label) {
			minY = plotRes.Y.Min
		}

		resPlots = append(resPlots, &resultPlot{plot: plotRes, label: simData.label})
	}

	for _, resPlot := range resPlots {
		if !isSwarmBasedStrategy(resPlot.label) {
			resPlot.plot.Y.Max = maxY
			resPlot.plot.Y.Min = minY
		}
		graphics.Save(resPlot.plot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
			generatePNGFileName(coll.outputDirPath, "MemoryUsedDistributionBoxPlot_"+visualStrategyName(resPlot.label)))
	}
}

func (coll *Collector) plotRequestsSucceeded() {
	plotRes := graphics.NewPlot("Deploy Requests Success (Cumulative)", "Time (Seconds)", "Requests Succeeded", true)

	dataPoints := make([]interface{}, 0)
	for _, simData := range coll.simulations {
		pts := make(plotter.XYs, len(simData.snapshots))

		runRequestsSucceededAcc := int64(0)
		for i := range pts {
			runRequestsSucceededAcc += simData.snapshots[i].TotalRunRequestsSucceeded()
			pts[i].X = simData.snapshots[i].EndTime().Seconds()
			pts[i].Y = float64(runRequestsSucceededAcc)
		}
		dataPoints = append(dataPoints, visualStrategyName(simData.label), pts)
	}

	err := plotutil.AddLines(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(coll.outputDirPath, "RequestsSucceeded"))
}

func (coll *Collector) plotSystemUsedResourcesVSRequestSuccess() {
	minY, maxY := math.MaxFloat64, float64(0)
	resPlots := make([]*resultPlot, 0)
	for _, simData := range coll.simulations {
		plotRes := graphics.NewPlot(fmt.Sprintf("Used Resources Vs Request Success (%s)",
			visualStrategyName(simData.label)), "Time (Seconds)", "Cumulative Success Ratio", true)

		totalFreeResPts := make(plotter.XYs, len(simData.snapshots))
		requestsSucceeded := make(plotter.XYs, len(simData.snapshots))

		runRequestsTotalAcc := int64(0)
		runRequestsSucceededAcc := int64(0)
		for i, r := 0, 0; i < len(simData.snapshots); i++ {
			runRequestsTotalAcc += simData.snapshots[i].TotalRunRequests()
			runRequestsSucceededAcc += simData.snapshots[i].TotalRunRequestsSucceeded()

			totalFreeResPts[i].X = simData.snapshots[i].EndTime().Seconds()
			totalFreeResPts[i].Y = float64(simData.snapshots[i].TotalUsedResourcesAvg())

			runRequestSuccessRatio := float64(runRequestsSucceededAcc) / float64(runRequestsTotalAcc)
			if runRequestSuccessRatio == 0 {
				requestsSucceeded = requestsSucceeded[:len(requestsSucceeded)-1]
				continue
			}
			requestsSucceeded[r].X = simData.snapshots[i].EndTime().Seconds()
			requestsSucceeded[r].Y = runRequestSuccessRatio
			r++
		}

		err := plotutil.AddLines(plotRes, "Used Resources", totalFreeResPts, "Requests Succeeded", requestsSucceeded)
		if err != nil {
			panic(errors.New("Problem with plots, error: " + err.Error()))
		}

		if plotRes.Y.Max > maxY {
			maxY = plotRes.Y.Max
		}
		if plotRes.Y.Min < minY {
			minY = plotRes.Y.Min
		}

		resPlots = append(resPlots, &resultPlot{plot: plotRes, label: simData.label})
	}
	for _, resPlot := range resPlots {
		resPlot.plot.Y.Max = maxY
		resPlot.plot.Y.Min = minY
		graphics.Save(resPlot.plot, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight,
			generatePNGFileName(coll.outputDirPath, "UsedResourcesVsRequestSuccess_"+visualStrategyName(resPlot.label)))
	}
}

func (coll *Collector) plotMessagesExchangedByRunRequest() {
	// Box plot. Messages per request over time.
	minY, maxY := math.MaxFloat64, float64(0)
	resBoxPlots := make([]*resultPlot, 0)
	for _, simData := range coll.simulations {
		plotRes := graphics.NewPlot(fmt.Sprintf("Lookup Messages Traded Distribution (%s)",
			visualStrategyName(simData.label)), "", "#Messages", false)

		boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
		xLabels := make([]string, len(simData.snapshots))
		for i, snapshot := range simData.snapshots {
			boxPlotPoints := make(plotter.Values, 0)
			boxPlotPoints = append(boxPlotPoints, snapshot.MessagesExchangedByRequest()...)
			boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), float64(i), boxPlotPoints)
			xLabels[i] = util.FmtDuration(snapshot.EndTime())
		}

		for i := range boxPlots {
			plotRes.Add(boxPlots[i])
		}
		plotRes.NominalX(xLabels...)

		if plotRes.Y.Max > maxY {
			maxY = plotRes.Y.Max
		}
		if plotRes.Y.Min < minY {
			minY = plotRes.Y.Min
		}

		resBoxPlots = append(resBoxPlots, &resultPlot{plot: plotRes, label: simData.label})

	}
	for _, resPlot := range resBoxPlots {
		resPlot.plot.Y.Max = maxY
		resPlot.plot.Y.Min = minY
		graphics.Save(resPlot.plot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
			generatePNGFileName(coll.outputDirPath, "MessagesPerRequestDistribution_"+visualStrategyName(resPlot.label)))
	}

	// Line Plot. Average messages per request overtime.
	linePlotRes := graphics.NewPlot("Average Discover Messages per Request", "Time (Seconds)", "Messages", true)
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
			simulationPts[j].Y = runRequestsAvgMessages
			j++
		}
		dataPoints = append(dataPoints, visualStrategyName(simData.label), simulationPts)
	}
	err := plotutil.AddLines(linePlotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}
	graphics.Save(linePlotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight,
		generatePNGFileName(coll.outputDirPath, "MessagesPerRequestAvgOverTime"))
}

func (coll *Collector) plotResourcesUsedDistributionByNodesOverTime() {
	for _, simData := range coll.simulations {
		data := make([]float64, len(simData.snapshots)*coll.numNodes)
		lower := 0
		upper := coll.numNodes
		yTicks := make([]int, len(simData.snapshots)+1)
		for i, snapshot := range simData.snapshots {
			yTicks[i] = int(snapshot.EndTime().Seconds())
			copy(data[lower:upper], snapshot.ResourcesUsedNodeRatio())
			lower += coll.numNodes
			upper += coll.numNodes
		}

		yTicks[len(yTicks)-1] = math.MaxInt64
		sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

		dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), coll.numNodes, data)}

		graphics.NewHeatMap(generatePNGFileName(coll.outputDirPath, "ResourcesUsedDistribution_"+visualStrategyName(simData.label)),
			fmt.Sprintf("Resources Used Distribution (%s)", visualStrategyName(simData.label)),
			"Nodes", "Time (Seconds)", yTicks, dataGrid, palette.Heat(64, 1),
			heatMapOverTimePNGWidth, heatMapOverTimePNGHeight)
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
			copy(data[lower:upper], snapshot.ResourcesUnreachableRatioNode())
			lower += coll.numNodes
			upper += coll.numNodes
		}

		yTicks[len(yTicks)-1] = math.MaxInt64
		sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

		dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), coll.numNodes, data)}

		graphics.NewHeatMap(generatePNGFileName(coll.outputDirPath, "ResourcesUnreachableDistribution_"+visualStrategyName(simData.label)),
			fmt.Sprintf("Resources Unreachable Distribution (%s)", visualStrategyName(simData.label)),
			"Nodes", "Time (Seconds)", yTicks, dataGrid, palette.Heat(64, 1),
			heatMapOverTimePNGWidth, heatMapOverTimePNGHeight)
	}
}

func (coll *Collector) plotMessagesDistributionByNodes() {
	util.Log.Infof(util.LogTag(plotsLogTag) + "Plot Messages Distribution By Node")
	// Heat Map.
	/*
		for _, simData := range coll.simulations {
			data := make([]float64, len(simData.snapshots)*coll.numNodes)
			lower := 0
			upper := coll.numNodes
			yTicks := make([]int, len(simData.snapshots)+1)
			for i, snapshot := range simData.snapshots {
				yTicks[i] = int(snapshot.EndTime().Seconds())
				copy(data[lower:upper], snapshot.TotalMessagesReceivedByNode())
				lower += coll.numNodes
				upper += coll.numNodes
			}

			yTicks[len(yTicks)-1] = math.MaxInt64
			sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

			dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), coll.numNodes, data)}

			graphics.NewHeatMap(generatePNGFileName(coll.outputDirPath, "MessagesDistributionHeatMap_"+visualStrategyName(simData.label)),
				fmt.Sprintf("Messages Distribution (%s)", visualStrategyName(simData.label)),
				"Nodes", "Time (Seconds)", yTicks, dataGrid, palette.Heat(64, 1))
		}
	*/
	// Box Plot. #Messages received per node.
	func() {
		minY, maxY := math.MaxFloat64, float64(0)
		resBoxPlots := make([]*resultPlot, 0)
		for _, simData := range coll.simulations {
			plotRes := graphics.NewPlot(fmt.Sprintf("Messages Received per Node (%s)",
				visualStrategyName(simData.label)), "Time (hh:mm)", "#Messages", false)

			boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
			xLabels := make([]string, len(simData.snapshots))
			for i, snapshot := range simData.snapshots {
				boxPlotPoints := make(plotter.Values, 0)
				boxPlotPoints = append(boxPlotPoints, snapshot.TotalMessagesReceivedByNode()...)
				xLabels[i] = util.FmtDuration(snapshot.EndTime())
				boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), float64(i), boxPlotPoints)
			}

			util.Log.Infof(util.LogTag(plotsLogTag)+"Strategy (%s)", simData.label)
			for i := range boxPlots {
				util.Log.Infof(util.LogTag(plotsLogTag)+"MessagesDistribution: Outliers %d, Ratio: %.2f%%",
					len(boxPlots[i].Outside), (float64(len(boxPlots[i].Outside))/float64(len(boxPlots[i].Values)))*100)
				plotRes.Add(boxPlots[i])
			}
			plotRes.NominalX(xLabels...)

			if plotRes.Y.Max > maxY && !isSwarmBasedStrategy(simData.label) {
				maxY = plotRes.Y.Max
			}
			if plotRes.Y.Min < minY && !isSwarmBasedStrategy(simData.label) {
				minY = plotRes.Y.Min
			}

			resBoxPlots = append(resBoxPlots, &resultPlot{plot: plotRes, label: simData.label})
		}
		for _, resPlot := range resBoxPlots {
			if !isSwarmBasedStrategy(resPlot.label) {
				resPlot.plot.Y.Max = maxY
				resPlot.plot.Y.Min = minY
			}
			graphics.Save(resPlot.plot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
				generatePNGFileName(coll.outputDirPath, "MessagesDistributionBoxPlot_"+visualStrategyName(resPlot.label)))
		}
	}()

	// Box Plot. Bandwidth received per node.
	func() {
		minY, maxY := math.MaxFloat64, float64(0)
		resBoxPlots := make([]*resultPlot, 0)
		for _, simData := range coll.simulations {
			plotRes := graphics.NewPlot(fmt.Sprintf("Bandwidth Used on Receiving (%s)",
				visualStrategyName(simData.label)), "Time (hh:mm)", "Bandwidth (bytes)", false)

			boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
			xLabels := make([]string, len(simData.snapshots))
			for i, snapshot := range simData.snapshots {
				boxPlotPoints := make(plotter.Values, 0)
				boxPlotPoints = append(boxPlotPoints, snapshot.TotalBandwidthUsedOnReceivingByNode()...)
				xLabels[i] = util.FmtDuration(snapshot.EndTime())
				boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), float64(i), boxPlotPoints)
			}

			util.Log.Infof(util.LogTag(plotsLogTag)+"Strategy (%s)", simData.label)
			for i := range boxPlots {
				util.Log.Infof(util.LogTag(plotsLogTag)+"MessagesBandwidth: Outliers %d, Ratio: %.2f%%", len(boxPlots[i].Outside), (float64(len(boxPlots[i].Outside))/float64(len(boxPlots[i].Values)))*100)
				plotRes.Add(boxPlots[i])
			}
			plotRes.NominalX(xLabels...)

			if plotRes.Y.Max > maxY && !isSwarmBasedStrategy(simData.label) {
				maxY = plotRes.Y.Max
			}
			if plotRes.Y.Min < minY && !isSwarmBasedStrategy(simData.label) {
				minY = plotRes.Y.Min
			}

			resBoxPlots = append(resBoxPlots, &resultPlot{plot: plotRes, label: simData.label})
		}
		for _, resPlot := range resBoxPlots {
			if !isSwarmBasedStrategy(resPlot.label) {
				resPlot.plot.Y.Max = maxY
				resPlot.plot.Y.Min = minY
			}

			graphics.Save(resPlot.plot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
				generatePNGFileName(coll.outputDirPath, "MessagesBandwidthBoxPlot_"+visualStrategyName(resPlot.label)))
		}
	}()
}

func (coll *Collector) plotTotalMessagesTradedInSystem() {
	plotRes := graphics.NewPlot("Total Messages Traded", "", "Messages Traded", false)

	barWidth := vg.Points(20)

	barOffset := []float64{0, -25, 25, -50, 50, -100, 100}
	color := 0
	for i, simData := range coll.simulations {
		messagesTradedAcc := float64(0)
		for _, snapshot := range simData.snapshots {
			messagesTradedAcc += snapshot.TotalMessagesReceivedByAllNodes()
		}

		barChart, err := plotter.NewBarChart(plotter.Values{messagesTradedAcc}, barWidth)
		if err != nil {
			panic(fmt.Errorf("error creating bar chart, %s", err))
		}

		barChart.LineStyle.Width = vg.Length(0)
		barChart.Color = plotutil.Color(color)
		barChart.Offset = vg.Points(barOffset[i])

		plotRes.Add(barChart)

		plotRes.Legend.Add(visualStrategyName(simData.label), barChart)

		color++
	}

	plotRes.Legend.Top = true
	plotRes.NominalX("Full Simulation")

	graphics.Save(plotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(coll.outputDirPath, "MessagesTradedInSystem"))
}

func (coll *Collector) plotMasterNodeMessagesReceivedOverTime() {
	// Line plot (Accumulated).
	plotResCum := graphics.NewPlot("Master Node Messages Received (Cumulative)", "Time (Seconds)", "Messages Received", true)
	dataPointsCum := make([]interface{}, 0)
	for _, simData := range coll.simulations {
		if isSwarmBasedStrategy(simData.label) {
			pts := make(plotter.XYs, len(simData.snapshots))
			accMessagesReceived := float64(0)
			for i := range pts {
				accMessagesReceived += simData.snapshots[i].TotalMessagesReceivedByMasterNode()
				pts[i].X = simData.snapshots[i].EndTime().Seconds()
				pts[i].Y = accMessagesReceived
			}
			dataPointsCum = append(dataPointsCum, visualStrategyName(simData.label), pts)
		}
	}
	err := plotutil.AddLines(plotResCum, dataPointsCum...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}
	graphics.Save(plotResCum, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight,
		generatePNGFileName(coll.outputDirPath, "MasterNodeMessagesReceivedCumulative"))

	// Line plot (Instantaneous).
	plotResInst := graphics.NewPlot("Master Node Messages Received (Instantaneous)", "Time (Seconds)", "Messages Received", true)
	dataPointsInst := make([]interface{}, 0)
	for _, simData := range coll.simulations {
		if isSwarmBasedStrategy(simData.label) {
			pts := make(plotter.XYs, len(simData.snapshots))
			for i := range pts {
				pts[i].X = simData.snapshots[i].EndTime().Seconds()
				pts[i].Y = simData.snapshots[i].TotalMessagesReceivedByMasterNode()
			}
			dataPointsInst = append(dataPointsInst, visualStrategyName(simData.label), pts)
		}
	}
	err = plotutil.AddLines(plotResInst, dataPointsInst...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}
	graphics.Save(plotResInst, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight,
		generatePNGFileName(coll.outputDirPath, "MasterNodeMessagesReceivedInstantaneous"))
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
		dataPoints = append(dataPoints, visualStrategyName(simData.label), relayedGetOffersPts)
	}

	err := plotutil.AddLines(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(coll.outputDirPath, "Debug_GetOffersRelayed"))
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
		dataPoints = append(dataPoints, visualStrategyName(simData.label), emptyGetOffersPts)
	}

	err := plotutil.AddLines(plotRes, dataPoints...)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(coll.outputDirPath, "Debug_EmptyGetOffer"))
}

// ======================================== Auxiliary Functions ====================================

func isOfferingBasedStrategy(strategyName string) bool {
	if strategyName == "chord-single-offer" ||
		strategyName == "chord-multiple-offer" ||
		strategyName == "chord-multiple-offer-updates" {
		return true
	}
	return false
}

func isSwarmBasedStrategy(strategyName string) bool {
	if strategyName == "swarm" {
		return true
	}
	return false
}

func visualStrategyName(strategyName string) string {
	if strategyName == "chord-multiple-offer" || strategyName == "chord-multiple-offer-updates" {
		return "multi-offer"
	} else if strategyName == "chord-single-offer" {
		return "single-offer"
	}
	return strategyName
}

func generatePNGFileName(pathNames ...string) string {
	return filepath.Join(pathNames...) + ".png"
}
