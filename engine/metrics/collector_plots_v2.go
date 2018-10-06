package metrics

import (
	"github.com/strabox/caravela-sim/engine/metrics/graphics"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const smartBarPlotOverTimePNGWidth = 15 * vg.Centimeter
const smartBarOverTimePNGHeight = 7 * vg.Centimeter

func (c *Collector) plotMemoryUsedByNodeV2() {
	allSimsLines := make([]interface{}, 0)
	allSimsErrorBars := make([]interface{}, 0)
	for _, simData := range c.simulations {
		simPts := make([]plotter.XYer, len(simData.snapshots))

		for j, snapshot := range simData.snapshots {
			memoryUsedPerNode := snapshot.TotalMemoryUsedByNode()
			tickPts := make(plotter.XYs, len(memoryUsedPerNode))
			for k := range memoryUsedPerNode {
				tickPts[k].X = snapshot.EndTime().Minutes()
				tickPts[k].Y = memoryUsedPerNode[k]
			}
			simPts[j] = tickPts
		}

		medMinMax, _ := plotutil.NewErrorPoints(plotutil.MedianAndMinMax, simPts...)
		allSimsLines = append(allSimsLines, visualStrategyName(simData.label), medMinMax)
		allSimsErrorBars = append(allSimsErrorBars, medMinMax)
	}

	plotRes := graphics.NewPlot("Memory Used per Node", "Time (Minutes)", "Memory Used (bytes)", false)
	plotutil.AddLinePoints(plotRes, allSimsLines...)
	plotutil.AddErrorBars(plotRes, allSimsErrorBars...)
	plotRes.Legend.Left = true
	plotRes.Legend.Top = true
	graphics.Save(plotRes, smartBarPlotOverTimePNGWidth, smartBarOverTimePNGHeight,
		generatePNGFileName(c.outputDirPath, "Memory", "MEMORY_USED_PER_NODE"))
}

func (c *Collector) plotBandwidthUsedByNodeV2() {
	allSimsLines := make([]interface{}, 0)
	allSimsErrorBars := make([]interface{}, 0)
	for _, simData := range c.simulations {
		simPts := make([]plotter.XYer, len(simData.snapshots))

		for j, snapshot := range simData.snapshots {
			memoryUsedPerNode := snapshot.TotalBandwidthUsedOnReceivingByNode()
			tickPts := make(plotter.XYs, len(memoryUsedPerNode))
			for k := range memoryUsedPerNode {
				tickPts[k].X = snapshot.EndTime().Minutes()
				tickPts[k].Y = memoryUsedPerNode[k]
			}
			simPts[j] = tickPts
		}

		medMinMax, _ := plotutil.NewErrorPoints(plotutil.MedianAndMinMax, simPts...)
		allSimsLines = append(allSimsLines, visualStrategyName(simData.label), medMinMax)
		allSimsErrorBars = append(allSimsErrorBars, medMinMax)
	}

	plotRes := graphics.NewPlot("Bandwidth Used per Node", "Time (Minutes)", "Bandwidth Used (bytes)", false)
	plotutil.AddLinePoints(plotRes, allSimsLines...)
	plotutil.AddErrorBars(plotRes, allSimsErrorBars...)
	plotRes.Legend.Left = true
	plotRes.Legend.Top = true
	graphics.Save(plotRes, smartBarPlotOverTimePNGWidth, smartBarOverTimePNGHeight,
		generatePNGFileName(c.outputDirPath, "Bandwidth", "BANDWIDTH_USED_PER_NODE"))
}

func (c *Collector) plotMessagesReceivedByNodeV2() {
	allSimsLines := make([]interface{}, 0)
	allSimsErrorBars := make([]interface{}, 0)
	for _, simData := range c.simulations {
		if !isSwarmBasedStrategy(simData.label) {
			simPts := make([]plotter.XYer, len(simData.snapshots))

			for j, snapshot := range simData.snapshots {
				memoryUsedPerNode := snapshot.TotalMessagesReceivedByNode()
				tickPts := make(plotter.XYs, len(memoryUsedPerNode))
				for k := range memoryUsedPerNode {
					tickPts[k].X = snapshot.EndTime().Minutes()
					tickPts[k].Y = memoryUsedPerNode[k]
				}
				simPts[j] = tickPts
			}

			medMinMax, _ := plotutil.NewErrorPoints(plotutil.MedianAndMinMax, simPts...)
			allSimsLines = append(allSimsLines, visualStrategyName(simData.label), medMinMax)
			allSimsErrorBars = append(allSimsErrorBars, medMinMax)
		}
	}

	plotRes := graphics.NewPlot("Messages Received per Node", "Time (Minutes)", "Messages", false)
	plotutil.AddLinePoints(plotRes, allSimsLines...)
	plotutil.AddErrorBars(plotRes, allSimsErrorBars...)
	plotRes.Legend.Left = true
	plotRes.Legend.Top = true
	graphics.Save(plotRes, smartBarPlotOverTimePNGWidth, smartBarOverTimePNGHeight,
		generatePNGFileName(c.outputDirPath, "Messages", "MESSAGES_RECEIVED_PER_NODE"))
}
