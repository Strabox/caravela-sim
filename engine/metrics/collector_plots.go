package metrics

import (
	"fmt"
	"github.com/strabox/caravela-sim/engine/metrics/graphics"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
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

const boxPlotWidth = 7
const boxPlotOverTimePNGWidth = 18 * vg.Centimeter
const boxPlotOverTimePNGHeight = 8 * vg.Centimeter

const quartilePlotOverTimePNGWidth = 12 * vg.Centimeter
const quartilePlotOverTimePNGHeight = 9 * vg.Centimeter

const linePlotOverTimePNGWidth = 10 * vg.Centimeter
const linePlotOverTimePNGHeight = 10 * vg.Centimeter

const heatMapOverTimePNGWidth = 1440
const heatMapOverTimePNGHeight = 540
const heatMapOverTimePaletteSize = 52

// resultPlot is used as a temporary structure to save gonum/plot structures and its visual label.
type resultPlot struct {
	plot  *plot.Plot
	label string
}

func (c *Collector) plotRequestsRate() {
	const title = "Requests Per Tick"
	const xLabel = "Tick"
	const yLabel = "% of Total Nodes in Requests"
	const outDir = "Requests"
	// Line Plot. Requests rates.
	linePlotRes := graphics.NewPlot(title, xLabel, yLabel, true)

	dataPoints := make([]interface{}, 0)
	deployRatesPts := make(plotter.XYs, c.simulatorConfigs.MaximumTicks())
	deploySuperTickSize := int(math.Ceil(float64(c.simulatorConfigs.MaximumTicks()) / float64(len(c.simulatorConfigs.DeployRequestsRate()))))
	deployCurrentSuperTick := 0
	for i := 0; i < c.simulatorConfigs.MaximumTicks(); i++ {
		deployRatesPts[i].X = float64(i)
		deployRatesPts[i].Y = c.simulatorConfigs.DeployRequestsRate()[deployCurrentSuperTick]
		if i != 0 && i%deploySuperTickSize == 0 {
			deployCurrentSuperTick++
		}
	}
	dataPoints = append(dataPoints, "Deploy Requests", deployRatesPts)

	stopRatesPts := make(plotter.XYs, c.simulatorConfigs.MaximumTicks())
	stopSuperTickSize := int(math.Ceil(float64(c.simulatorConfigs.MaximumTicks()) / float64(len(c.simulatorConfigs.StopRequestsRate()))))
	stopCurrentSuperTick := 0
	for i := 0; i < c.simulatorConfigs.MaximumTicks(); i++ {
		stopRatesPts[i].X = float64(i)
		stopRatesPts[i].Y = c.simulatorConfigs.StopRequestsRate()[stopCurrentSuperTick]
		if i != 0 && i%stopSuperTickSize == 0 {
			stopCurrentSuperTick++
		}
	}
	dataPoints = append(dataPoints, "Stop Requests", stopRatesPts)

	plotutil.AddLines(linePlotRes, dataPoints...)

	graphics.Save(linePlotRes, linePlotOverTimePNGWidth, 10*vg.Centimeter, generatePNGFileName(c.outputDirPath, outDir, "RequestsRates"))
}

func (c *Collector) plotResourcesAllocationEfficiency() {
	const title = "Resources Allocation Efficiency - Cumulative"
	const xLabel = "Time (Minutes)"
	const yLabel = "Allocation Efficiency"
	const outDir = "Resources"

	plotRes := graphics.NewPlot(title, xLabel, yLabel, true)

	dataPoints := make([]interface{}, 0)
	for _, simData := range c.simulations {
		pts := make(plotter.XYs, len(simData.snapshots))

		accResourcesRequested := types.Resources{CPUs: 0, Memory: 0}
		accResourcesAllocated := types.Resources{CPUs: 0, Memory: 0}
		for i := range pts {
			tickResourcesRequested := simData.snapshots[i].TotalResourcesRequested()
			tickResourcesAllocated := simData.snapshots[i].TotalResourcesAllocated()
			accResourcesRequested.CPUs += tickResourcesRequested.CPUs
			accResourcesRequested.Memory += tickResourcesRequested.Memory
			accResourcesAllocated.CPUs += tickResourcesAllocated.CPUs
			accResourcesAllocated.Memory += tickResourcesAllocated.Memory
			pts[i].X = simData.snapshots[i].EndTime().Minutes()
			pts[i].Y = ((float64(accResourcesAllocated.CPUs) / float64(accResourcesRequested.CPUs)) +
				(float64(accResourcesAllocated.Memory) / float64(accResourcesRequested.Memory))) / 2
		}
		dataPoints = append(dataPoints, visualStrategyName(simData.label), pts)
	}

	plotutil.AddLines(plotRes, dataPoints...)

	plotRes.Legend.Top = true

	graphics.Save(plotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(c.outputDirPath, outDir, "ResourcesAllocationEfficiency"))
}

func (c *Collector) plotActiveOffersByNode() {
	const title = "Offers Per Node (%s)"
	const xLabel = "Time (Minutes)"
	const yLabel = "Offers"
	const outDir = "Offers"

	// Box Plot over time.
	resultLabel := ""
	outliers := make([][]float64, 0)
	for _, simData := range c.simulations {
		if isOfferingBasedStrategy(simData.label) {
			plotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, false)

			boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
			for i, snapshot := range simData.snapshots {
				boxPlotPoints := make(plotter.Values, 0)
				boxPlotPoints = append(boxPlotPoints, snapshot.TotalTraderActiveOfferPerNode()...)
				boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), simData.snapshots[i].EndTime().Minutes(), boxPlotPoints)
			}

			logString := "Offers: "
			for i := range boxPlots {
				logString += fmt.Sprintf("<%d, %.2f%%>; ", len(boxPlots[i].Outside), (float64(len(boxPlots[i].Outside))/float64(len(boxPlots[i].Values)))*100)
				plotRes.Add(boxPlots[i])
				// Find the outliers.
				currentOutliers := make([]float64, len(boxPlots[i].Outside))
				for j := range boxPlots[i].Outside {
					currentOutliers[j] = boxPlots[i].Value(boxPlots[i].Outside[j])
				}
				outliers = append(outliers, currentOutliers)
			}
			util.Log.Info(util.LogTag(plotsLogTag) + logString)

			resultLabel = simData.label
			graphics.Save(plotRes, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
				generatePNGFileName(c.outputDirPath, outDir, "ActiveOffersDistributionBoxPlot_"+visualStrategyName(simData.label)))
		}
	}

	// Box Plots for the outliers.
	outliersBoxPlot := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(resultLabel)), xLabel, yLabel, false)
	outliersQuartilePlot := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(resultLabel)), xLabel, yLabel, false)
	for i := range outliers {
		boxPlotPoints := make(plotter.Values, 0)
		boxPlotPoints = append(boxPlotPoints, outliers[i]...)
		boxPlot, _ := plotter.NewBoxPlot(vg.Points(boxPlotWidth), float64(i+1)*c.simulatorConfigs.TicksInterval().Minutes(), boxPlotPoints)
		outliersBoxPlot.Add(boxPlot)
		quartilePlot, _ := plotter.NewQuartPlot(float64(i+1)*c.simulatorConfigs.TicksInterval().Minutes(), boxPlotPoints)
		outliersQuartilePlot.Add(quartilePlot)
	}
	graphics.Save(outliersQuartilePlot, quartilePlotOverTimePNGWidth, quartilePlotOverTimePNGHeight,
		generatePNGFileName(c.outputDirPath, outDir, "ActiveOffersOutliersDistributionQuartilePlot_"+visualStrategyName(resultLabel)))
	graphics.Save(outliersBoxPlot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
		generatePNGFileName(c.outputDirPath, outDir, "ActiveOffersOutliersDistributionBoxPlot_"+visualStrategyName(resultLabel)))
}

func (c *Collector) plotMemoryUsedByNode() {
	const title = "Memory Used by Node (%s)"
	const xLabel = "Time (Minutes)"
	const yLabel = "Memory Used (bytes)"
	const outDir = "Memory"

	simsLabels := make([]string, 0)
	simsOutliers := make([][][]float64, 0)

	minY, maxY := math.MaxFloat64, float64(0)
	resBoxPlots := make([]*resultPlot, 0)
	resQuatPlots := make([]*resultPlot, 0)
	for _, simData := range c.simulations {
		quatPlotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, false)
		boxPlotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, false)

		quartilePlots := make([]*plotter.QuartPlot, len(simData.snapshots))
		boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
		for i, snapshot := range simData.snapshots {
			boxPlotPoints := make(plotter.Values, 0)
			boxPlotPoints = append(boxPlotPoints, snapshot.TotalMemoryUsedByNode()...)
			boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), snapshot.EndTime().Minutes(), boxPlotPoints)
			quartilePlots[i], _ = plotter.NewQuartPlot(snapshot.EndTime().Minutes(), boxPlotPoints)
		}

		logString := fmt.Sprintf("MemoryUsedByNode(%s): ", simData.label)
		simOutliers := make([][]float64, 0)
		for i := range boxPlots {
			logString += fmt.Sprintf("<%d, %.2f%%>; ", len(boxPlots[i].Outside), (float64(len(boxPlots[i].Outside))/float64(len(boxPlots[i].Values)))*100)
			boxPlotRes.Add(boxPlots[i])
			quatPlotRes.Add(quartilePlots[i])
			// Find the simsOutliers.
			currentOutliers := make([]float64, len(boxPlots[i].Outside))
			for j := range boxPlots[i].Outside {
				currentOutliers[j] = boxPlots[i].Value(boxPlots[i].Outside[j])
			}
			simOutliers = append(simOutliers, currentOutliers)
		}
		simsLabels = append(simsLabels, simData.label)
		simsOutliers = append(simsOutliers, simOutliers)
		util.Log.Info(util.LogTag(plotsLogTag) + logString)

		if boxPlotRes.Y.Max > maxY && !isSwarmBasedStrategy(simData.label) {
			maxY = boxPlotRes.Y.Max
		}
		if boxPlotRes.Y.Min < minY && !isSwarmBasedStrategy(simData.label) {
			minY = boxPlotRes.Y.Min
		}

		resBoxPlots = append(resBoxPlots, &resultPlot{plot: boxPlotRes, label: simData.label})
		resQuatPlots = append(resQuatPlots, &resultPlot{plot: quatPlotRes, label: simData.label})
	}

	for _, resPlot := range resBoxPlots {
		if !isSwarmBasedStrategy(resPlot.label) {
			resPlot.plot.Y.Max = maxY
			resPlot.plot.Y.Min = minY
		}
		graphics.Save(resPlot.plot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
			generatePNGFileName(c.outputDirPath, outDir, "MemoryUsedDistributionBoxPlot_"+visualStrategyName(resPlot.label)))
	}
	for _, resPlot := range resQuatPlots {
		if !isSwarmBasedStrategy(resPlot.label) {
			resPlot.plot.Y.Max = maxY
			resPlot.plot.Y.Min = minY
		}
		graphics.Save(resPlot.plot, quartilePlotOverTimePNGWidth, quartilePlotOverTimePNGHeight,
			generatePNGFileName(c.outputDirPath, outDir, "MemoryUsedDistributionQuartilePlot_"+visualStrategyName(resPlot.label)))
	}

	// Box Plots for the outliers.
	for i := range simsOutliers {
		outliersPlot := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simsLabels[i])), xLabel, yLabel, false)
		for j := range simsOutliers[i] {
			boxPlotPoints := make(plotter.Values, 0)
			boxPlotPoints = append(boxPlotPoints, simsOutliers[i][j]...)
			boxPlot, _ := plotter.NewBoxPlot(vg.Points(boxPlotWidth), float64(j), boxPlotPoints)
			outliersPlot.Add(boxPlot)
		}
		graphics.Save(outliersPlot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
			generatePNGFileName(c.outputDirPath, outDir, "MemoryUsedOutliersDistributionBoxPlot_"+visualStrategyName(simsLabels[i])))
	}
}

func (c *Collector) plotRequestsSucceeded() {
	const title = "Deploy Requests Success - Cumulative"
	const xLabel = "Time (Minutes)"
	const yLabel = "Requests Succeeded"
	const outDir = "Requests"

	plotRes := graphics.NewPlot(title, xLabel, yLabel, true)

	dataPoints := make([]interface{}, 0)
	for _, simData := range c.simulations {
		pts := make(plotter.XYs, len(simData.snapshots))

		runRequestsSucceededAcc := int64(0)
		for i := range pts {
			runRequestsSucceededAcc += simData.snapshots[i].TotalRunRequestsSucceeded()
			pts[i].X = simData.snapshots[i].EndTime().Minutes()
			pts[i].Y = float64(runRequestsSucceededAcc)
		}
		dataPoints = append(dataPoints, visualStrategyName(simData.label), pts)
	}

	for _, simData := range c.simulations {
		pts := make(plotter.XYs, len(simData.snapshots))

		totalRequestsAcc := int64(0)
		for i := range pts {
			totalRequestsAcc += simData.snapshots[i].TotalRunRequests()
			pts[i].X = simData.snapshots[i].EndTime().Minutes()
			pts[i].Y = float64(totalRequestsAcc)
		}
		dataPoints = append(dataPoints, "total", pts)
		break
	}

	plotutil.AddLines(plotRes, dataPoints...)

	graphics.Save(plotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(c.outputDirPath, outDir, "RequestsSucceeded"))
}

func (c *Collector) plotSystemUsedResourcesVSRequestSuccess() {
	const title = "System Resources Usage Vs Deploy Requests Succeeded (%s)"
	const xLabel = "Time (Minutes)"
	const yLabel = "Ratios"
	const outDir = "Requests"

	minY, maxY := math.MaxFloat64, float64(0)
	resPlots := make([]*resultPlot, 0)
	for _, simData := range c.simulations {
		plotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, true)

		totalFreeResPts := make(plotter.XYs, len(simData.snapshots))
		requestsSucceeded := make(plotter.XYs, len(simData.snapshots))

		runRequestsTotalAcc := int64(0)
		runRequestsSucceededAcc := int64(0)
		for i, r := 0, 0; i < len(simData.snapshots); i++ {
			runRequestsTotalAcc += simData.snapshots[i].TotalRunRequests()
			runRequestsSucceededAcc += simData.snapshots[i].TotalRunRequestsSucceeded()

			totalFreeResPts[i].X = simData.snapshots[i].EndTime().Minutes()
			totalFreeResPts[i].Y = float64(simData.snapshots[i].TotalUsedResourcesAvg())

			runRequestSuccessRatio := float64(runRequestsSucceededAcc) / float64(runRequestsTotalAcc)
			if runRequestSuccessRatio == 0 {
				requestsSucceeded = requestsSucceeded[:len(requestsSucceeded)-1]
				continue
			}
			requestsSucceeded[r].X = simData.snapshots[i].EndTime().Minutes()
			requestsSucceeded[r].Y = runRequestSuccessRatio
			r++
		}

		plotutil.AddLines(plotRes, "System Usage", totalFreeResPts, "Requests Succeeded", requestsSucceeded)

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
			generatePNGFileName(c.outputDirPath, outDir, "UsedResourcesVsRequestSuccess_"+visualStrategyName(resPlot.label)))
	}
}

func (c *Collector) plotMessagesTradedByDeployRequestLinePlot() {
	const title = "Average Discover Messages per Request"
	const xLabel = "Time (Minutes)"
	const yLabel = "Average Messages"
	const outDir = "Requests"
	// Line Plot. Average messages per request overtime.
	linePlotRes := graphics.NewPlot(title, xLabel, yLabel, true)
	dataPoints := make([]interface{}, 0)
	for _, simData := range c.simulations {
		simulationPts := make(plotter.XYs, len(simData.snapshots))
		for i, j := 0, 0; i < len(simData.snapshots); i++ {
			runRequestsAvgMessages := simData.snapshots[i].RunRequestsAvgMessages()
			if runRequestsAvgMessages == 0 { // No requests happened in this snapshot.
				simulationPts = simulationPts[:len(simulationPts)-1]
				continue
			}
			simulationPts[j].X = simData.snapshots[i].EndTime().Minutes()
			simulationPts[j].Y = runRequestsAvgMessages
			j++
		}
		dataPoints = append(dataPoints, visualStrategyName(simData.label), simulationPts)
	}
	plotutil.AddLines(linePlotRes, dataPoints...)

	graphics.Save(linePlotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(c.outputDirPath, outDir, "MessagesPerRequestAvgOverTime"))
}

func (c *Collector) plotMessagesTradedByDeployRequestBoxPlot() {
	const title = "Messages Traded To Deploy Containers (%s)"
	const xLabel = "Time (Minutes)"
	const yLabel = "Messages"
	const outDir = "Requests"

	// Box plot. Messages per request over time.
	minY, maxY := math.MaxFloat64, float64(0)
	resBoxPlots := make([]*resultPlot, 0)
	resQuatPlots := make([]*resultPlot, 0)
	for _, simData := range c.simulations {
		boxPlotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, false)
		quatPlotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, false)

		quartilePlots := make([]*plotter.QuartPlot, len(simData.snapshots))
		boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
		for i, snapshot := range simData.snapshots {
			boxPlotPoints := make(plotter.Values, 0)
			boxPlotPoints = append(boxPlotPoints, snapshot.MessagesExchangedByRequest()...)
			boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), snapshot.EndTime().Minutes(), boxPlotPoints)
			quartilePlots[i], _ = plotter.NewQuartPlot(snapshot.EndTime().Minutes(), boxPlotPoints)
		}

		for i := range boxPlots {
			boxPlotRes.Add(boxPlots[i])
			quatPlotRes.Add(quartilePlots[i])
		}

		if boxPlotRes.Y.Max > maxY {
			maxY = boxPlotRes.Y.Max
		}
		if boxPlotRes.Y.Min < minY {
			minY = boxPlotRes.Y.Min
		}

		resBoxPlots = append(resBoxPlots, &resultPlot{plot: boxPlotRes, label: simData.label})
		resQuatPlots = append(resQuatPlots, &resultPlot{plot: quatPlotRes, label: simData.label})

	}
	for _, resPlot := range resBoxPlots {
		resPlot.plot.Y.Max = maxY
		resPlot.plot.Y.Min = minY
		graphics.Save(resPlot.plot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
			generatePNGFileName(c.outputDirPath, outDir, "MessagesPerRequestDistributionBoxPlot_"+visualStrategyName(resPlot.label)))
	}
	for _, resPlot := range resQuatPlots {
		resPlot.plot.Y.Max = maxY
		resPlot.plot.Y.Min = minY
		graphics.Save(resPlot.plot, quartilePlotOverTimePNGWidth, quartilePlotOverTimePNGHeight,
			generatePNGFileName(c.outputDirPath, outDir, "MessagesPerRequestDistributionQuartilePlot_"+visualStrategyName(resPlot.label)))
	}
}

func (c *Collector) plotResourcesUsedDistributionByNodesOverTime() {
	const title = "Resources Used Distribution (%s)"
	const xLabel = "Nodes"
	const yLabel = "Time (Minutes)"
	const outDir = "Resources"

	for _, simData := range c.simulations {
		data := make([]float64, len(simData.snapshots)*c.numNodes)
		lower := 0
		upper := c.numNodes
		yTicks := make([]int, len(simData.snapshots)+1)
		for i, snapshot := range simData.snapshots {
			yTicks[i] = int(snapshot.EndTime().Minutes())
			copy(data[lower:upper], snapshot.ResourcesUsedNodeRatio())
			lower += c.numNodes
			upper += c.numNodes
		}

		yTicks[len(yTicks)-1] = math.MaxInt64
		sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

		dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), c.numNodes, data)}

		graphics.NewHeatMap(generatePNGFileName(c.outputDirPath, outDir, "ResourcesUsedDistribution_"+visualStrategyName(simData.label)),
			fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, yTicks, dataGrid, palette.Heat(heatMapOverTimePaletteSize, 1),
			heatMapOverTimePNGWidth, heatMapOverTimePNGHeight)
	}
}

func (c *Collector) plotResourcesUnreachableDistributionByNodesOverTime() {
	const title = "Resources Unreachable Distribution (%s)"
	const xLabel = "Nodes"
	const yLabel = "Time (Minutes)"
	const outDir = "Resources"

	for _, simData := range c.simulations {
		data := make([]float64, len(simData.snapshots)*c.numNodes)
		lower := 0
		upper := c.numNodes
		yTicks := make([]int, len(simData.snapshots)+1)
		for i, snapshot := range simData.snapshots {
			yTicks[i] = int(snapshot.EndTime().Minutes())
			copy(data[lower:upper], snapshot.ResourcesUnreachableRatioNode())
			lower += c.numNodes
			upper += c.numNodes
		}

		yTicks[len(yTicks)-1] = math.MaxInt64
		sort.Sort(sort.Reverse(sort.IntSlice(yTicks)))

		dataGrid := &graphics.UnitGrid{Data: mat.NewDense(len(simData.snapshots), c.numNodes, data)}

		graphics.NewHeatMap(generatePNGFileName(c.outputDirPath, outDir, "ResourcesUnreachableDistribution_"+visualStrategyName(simData.label)),
			fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, yTicks, dataGrid, palette.Heat(heatMapOverTimePaletteSize, 1),
			heatMapOverTimePNGWidth, heatMapOverTimePNGHeight)
	}
}

func (c *Collector) plotBandwidthUsedByNode() {
	const title = "Bandwidth Used on Receiving (%s)"
	const xLabel = "Time (Minutes)"
	const yLabel = "Bandwidth (bytes)"
	const outDir = "Bandwidth"

	simsLabels := make([]string, 0)
	simsOutliers := make([][][]float64, 0)

	// Box Plot. Bandwidth received per node.
	func() {
		minY, maxY := math.MaxFloat64, float64(0)
		resBoxPlots := make([]*resultPlot, 0)
		resQuatPlots := make([]*resultPlot, 0)
		for _, simData := range c.simulations {
			boxPlotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, false)
			quatPlotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, false)

			boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
			quartilePlots := make([]*plotter.QuartPlot, len(simData.snapshots))
			for i, snapshot := range simData.snapshots {
				boxPlotPoints := make(plotter.Values, 0)
				boxPlotPoints = append(boxPlotPoints, snapshot.TotalBandwidthUsedOnReceivingByNode()...)
				boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), snapshot.EndTime().Minutes(), boxPlotPoints)
				quartilePlots[i], _ = plotter.NewQuartPlot(simData.snapshots[i].EndTime().Minutes(), boxPlotPoints)
			}

			logString := fmt.Sprintf("BandwidthPerNode(%s): ", simData.label)
			simOutliers := make([][]float64, 0)
			for i := range boxPlots {
				logString += fmt.Sprintf("<%d, %.2f%%>; ", len(boxPlots[i].Outside), (float64(len(boxPlots[i].Outside))/float64(len(boxPlots[i].Values)))*100)
				boxPlotRes.Add(boxPlots[i])
				quatPlotRes.Add(quartilePlots[i])
				// Find the simsOutliers.
				currentOutliers := make([]float64, len(boxPlots[i].Outside))
				for j := range boxPlots[i].Outside {
					currentOutliers[j] = boxPlots[i].Value(boxPlots[i].Outside[j])
				}
				simOutliers = append(simOutliers, currentOutliers)
			}
			simsLabels = append(simsLabels, simData.label)
			simsOutliers = append(simsOutliers, simOutliers)
			util.Log.Info(util.LogTag(plotsLogTag) + logString)

			if boxPlotRes.Y.Max > maxY && !isSwarmBasedStrategy(simData.label) {
				maxY = boxPlotRes.Y.Max
			}
			if boxPlotRes.Y.Min < minY && !isSwarmBasedStrategy(simData.label) {
				minY = boxPlotRes.Y.Min
			}

			resBoxPlots = append(resBoxPlots, &resultPlot{plot: boxPlotRes, label: simData.label})
			resQuatPlots = append(resQuatPlots, &resultPlot{plot: quatPlotRes, label: simData.label})
		}

		for _, resPlot := range resBoxPlots {
			if !isSwarmBasedStrategy(resPlot.label) {
				resPlot.plot.Y.Max = maxY
				resPlot.plot.Y.Min = minY
			}

			graphics.Save(resPlot.plot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
				generatePNGFileName(c.outputDirPath, outDir, "BandwidthPerNodeBoxPlot_"+visualStrategyName(resPlot.label)))
		}
		for _, resPlot := range resQuatPlots {
			if !isSwarmBasedStrategy(resPlot.label) {
				resPlot.plot.Y.Max = maxY
				resPlot.plot.Y.Min = minY
			}
			graphics.Save(resPlot.plot, quartilePlotOverTimePNGWidth, quartilePlotOverTimePNGHeight,
				generatePNGFileName(c.outputDirPath, outDir, "BandwidthPerNodeQuartilePlot_"+visualStrategyName(resPlot.label)))
		}
	}()

	func() { // Box Plots for the outliers.
		for i := range simsOutliers {
			outliersPlot := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simsLabels[i])), xLabel, yLabel, false)
			for j := range simsOutliers[i] {
				boxPlotPoints := make(plotter.Values, 0)
				boxPlotPoints = append(boxPlotPoints, simsOutliers[i][j]...)
				boxPlot, _ := plotter.NewBoxPlot(vg.Points(boxPlotWidth), float64(j), boxPlotPoints)
				outliersPlot.Add(boxPlot)
			}
			graphics.Save(outliersPlot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
				generatePNGFileName(c.outputDirPath, outDir, "BandwidthPerNodeOutliersBoxPlot_"+visualStrategyName(simsLabels[i])))
		}
	}()
}

func (c *Collector) plotMessagesDistributionByNodes() {
	const title = "Messages Received per Node (%s)"
	const xLabel = "Time (Minutes)"
	const yLabel = "Messages"
	const outDir = "Messages"

	simsLabels := make([]string, 0)
	simsOutliers := make([][][]float64, 0)

	minY, maxY := math.MaxFloat64, float64(0)
	resBoxPlots := make([]*resultPlot, 0)
	resQuatPlots := make([]*resultPlot, 0)
	for _, simData := range c.simulations {
		quatPlotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, false)
		boxPlotRes := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simData.label)), xLabel, yLabel, false)

		boxPlots := make([]*plotter.BoxPlot, len(simData.snapshots))
		quartilePlots := make([]*plotter.QuartPlot, len(simData.snapshots))
		for i, snapshot := range simData.snapshots {
			boxPlotPoints := make(plotter.Values, 0)
			boxPlotPoints = append(boxPlotPoints, snapshot.TotalMessagesReceivedByNode()...)
			boxPlots[i], _ = plotter.NewBoxPlot(vg.Points(boxPlotWidth), snapshot.EndTime().Minutes(), boxPlotPoints)
			quartilePlots[i], _ = plotter.NewQuartPlot(snapshot.EndTime().Minutes(), boxPlotPoints)
		}

		logString := fmt.Sprintf("MessagesDistributionPerNode(%s): ", simData.label)
		simOutliers := make([][]float64, 0)
		for i := range boxPlots {
			logString += fmt.Sprintf("<%d, %.2f%%>; ", len(boxPlots[i].Outside), (float64(len(boxPlots[i].Outside))/float64(len(boxPlots[i].Values)))*100)
			boxPlotRes.Add(boxPlots[i])
			quatPlotRes.Add(quartilePlots[i])
			// Find the simsOutliers.
			currentOutliers := make([]float64, len(boxPlots[i].Outside))
			for j := range boxPlots[i].Outside {
				currentOutliers[j] = boxPlots[i].Value(boxPlots[i].Outside[j])
			}
			simOutliers = append(simOutliers, currentOutliers)
		}
		simsLabels = append(simsLabels, simData.label)
		simsOutliers = append(simsOutliers, simOutliers)
		util.Log.Info(util.LogTag(plotsLogTag) + logString)

		if boxPlotRes.Y.Max > maxY && !isSwarmBasedStrategy(simData.label) {
			maxY = boxPlotRes.Y.Max
		}
		if boxPlotRes.Y.Min < minY && !isSwarmBasedStrategy(simData.label) {
			minY = boxPlotRes.Y.Min
		}

		resBoxPlots = append(resBoxPlots, &resultPlot{plot: boxPlotRes, label: simData.label})
		resQuatPlots = append(resQuatPlots, &resultPlot{plot: quatPlotRes, label: simData.label})
	}

	for _, resPlot := range resBoxPlots {
		if !isSwarmBasedStrategy(resPlot.label) {
			resPlot.plot.Y.Max = maxY
			resPlot.plot.Y.Min = minY
		}
		graphics.Save(resPlot.plot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
			generatePNGFileName(c.outputDirPath, outDir, "MessagesPerNodeDistributionBoxPlot_"+visualStrategyName(resPlot.label)))
	}

	for _, resPlot := range resQuatPlots {
		if !isSwarmBasedStrategy(resPlot.label) {
			resPlot.plot.Y.Max = maxY
			resPlot.plot.Y.Min = minY
		}
		graphics.Save(resPlot.plot, quartilePlotOverTimePNGWidth, quartilePlotOverTimePNGHeight,
			generatePNGFileName(c.outputDirPath, outDir, "MessagesPerNodeDistributionQuartilePlot_"+visualStrategyName(resPlot.label)))
	}

	// Box Plots for the outliers.
	for i := range simsOutliers {
		outliersPlot := graphics.NewPlot(fmt.Sprintf(title, visualStrategyName(simsLabels[i])), xLabel, yLabel, false)
		for j := range simsOutliers[i] {
			boxPlotPoints := make(plotter.Values, 0)
			boxPlotPoints = append(boxPlotPoints, simsOutliers[i][j]...)
			boxPlot, _ := plotter.NewBoxPlot(vg.Points(boxPlotWidth), float64(j), boxPlotPoints)
			outliersPlot.Add(boxPlot)
		}
		graphics.Save(outliersPlot, boxPlotOverTimePNGWidth, boxPlotOverTimePNGHeight,
			generatePNGFileName(c.outputDirPath, outDir, "MessagesPerNodeOutliersDistributionBoxPlot_"+visualStrategyName(simsLabels[i])))
	}
}

func (c *Collector) plotTotalMessagesTradedInSystem() {
	const title = "Total Messages Traded"
	const xLabel = ""
	const yLabel = "Messages Traded"
	const outDir = "Messages"

	plotRes := graphics.NewPlot(title, xLabel, yLabel, false)

	barWidth := vg.Points(20)

	barOffset := []float64{0, -25, 25, -50, 50, -100, 100}
	color := 0
	for i, simData := range c.simulations {
		messagesTradedAcc := float64(0)
		for _, snapshot := range simData.snapshots {
			messagesTradedAcc += snapshot.TotalMessagesReceivedByAllNodes()
		}

		barChart, _ := plotter.NewBarChart(plotter.Values{messagesTradedAcc}, barWidth)

		barChart.LineStyle.Width = vg.Length(0)
		barChart.Color = plotutil.Color(color)
		barChart.Offset = vg.Points(barOffset[i])

		plotRes.Add(barChart)

		plotRes.Legend.Add(visualStrategyName(simData.label), barChart)

		color++
	}

	plotRes.Legend.Top = true
	plotRes.NominalX("Full Simulation")

	graphics.Save(plotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(c.outputDirPath, outDir, "MessagesTradedInSystem"))
}

func (c *Collector) plotMasterNodeMessagesReceivedOverTime() {
	const xLabel = "Time (Minutes)"
	const yLabel = "Messages Received"
	const outDir = "SwarmMaster"

	// Line plot (Accumulated).
	plotResCum := graphics.NewPlot("Master Node Messages Received (Cumulative)", xLabel, yLabel, true)
	dataPointsCum := make([]interface{}, 0)
	for _, simData := range c.simulations {
		if isSwarmBasedStrategy(simData.label) {
			pts := make(plotter.XYs, len(simData.snapshots))
			accMessagesReceived := float64(0)
			for i := range pts {
				accMessagesReceived += simData.snapshots[i].TotalMessagesReceivedByMasterNode()
				pts[i].X = simData.snapshots[i].EndTime().Minutes()
				pts[i].Y = accMessagesReceived
			}
			dataPointsCum = append(dataPointsCum, visualStrategyName(simData.label), pts)
		}
	}
	plotutil.AddLines(plotResCum, dataPointsCum...)

	graphics.Save(plotResCum, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(c.outputDirPath, outDir, "MasterNodeMessagesReceivedCumulative"))

	// Line plot (Instantaneous).
	plotResInst := graphics.NewPlot("Master Node Messages Received (Instantaneous)", xLabel, yLabel, true)
	dataPointsInst := make([]interface{}, 0)
	for _, simData := range c.simulations {
		if isSwarmBasedStrategy(simData.label) {
			pts := make(plotter.XYs, len(simData.snapshots))
			for i := range pts {
				pts[i].X = simData.snapshots[i].EndTime().Minutes()
				pts[i].Y = simData.snapshots[i].TotalMessagesReceivedByMasterNode()
			}
			dataPointsInst = append(dataPointsInst, visualStrategyName(simData.label), pts)
		}
	}
	plotutil.AddLines(plotResInst, dataPointsInst...)

	graphics.Save(plotResInst, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(c.outputDirPath, outDir, "MasterNodeMessagesReceivedInstantaneous"))
}

// ======================================= Debug Performance Plots ===============================

func (c *Collector) plotRelayedGetOfferMessages() {
	plotRes := graphics.NewPlot("Total Relayed Get Offer", "Time (Minutes)", "#Relayed Get Offers", true)

	dataPoints := make([]interface{}, 0)
	for _, simData := range c.simulations {
		relayedGetOffersPts := make(plotter.XYs, len(simData.snapshots))
		for i := range relayedGetOffersPts {
			relayedGetOffersPts[i].X = simData.snapshots[i].EndTime().Minutes()
			relayedGetOffersPts[i].Y = float64(simData.snapshots[i].TotalGetOffersRelayed())
		}
		dataPoints = append(dataPoints, visualStrategyName(simData.label), relayedGetOffersPts)
	}

	plotutil.AddLines(plotRes, dataPoints...)

	graphics.Save(plotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(c.outputDirPath, "Debug_GetOffersRelayed"))
}

func (c *Collector) plotEmptyGetOfferMessages() {
	plotRes := graphics.NewPlot("Empty Get Offer Messages", "Time (Minutes)", "#Empty Get Offer", true)

	dataPoints := make([]interface{}, 0)
	for _, simData := range c.simulations {
		emptyGetOffersPts := make(plotter.XYs, len(simData.snapshots))
		for i := range emptyGetOffersPts {
			emptyGetOffersPts[i].X = simData.snapshots[i].EndTime().Minutes()
			emptyGetOffersPts[i].Y = float64(simData.snapshots[i].TotalEmptyGetOfferMessages())
		}
		dataPoints = append(dataPoints, visualStrategyName(simData.label), emptyGetOffersPts)
	}

	plotutil.AddLines(plotRes, dataPoints...)

	graphics.Save(plotRes, linePlotOverTimePNGWidth, linePlotOverTimePNGHeight, generatePNGFileName(c.outputDirPath, "Debug_EmptyGetOffer"))
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
		return "Multi-Offer"
	} else if strategyName == "chord-single-offer" {
		return "Single-Offer"
	} else if strategyName == "chord-random" {
		return "Random"
	} else if strategyName == "swarm" {
		return "Swarm"
	}
	return strategyName
}

func generatePNGFileName(pathNames ...string) string {
	return filepath.Join(pathNames...) + ".png"
}
