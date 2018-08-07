package metrics

import (
	"errors"
	"github.com/strabox/caravela-sim/simulation/metrics/graphics"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func (collector *Collector) plotRequestsSucceededOverTime() {
	plot := graphics.New("Requests success over time", "Time", "Request Succeeded")

	pts := make(plotter.XYs, len(collector.snapshots))
	for i := range pts {
		pts[i].X = collector.snapshots[i].StartTime().Seconds()
		pts[i].Y = float64(collector.snapshots[i].TotalRunRequestsSucceeded())
	}

	err := plotutil.AddLinePoints(plot, "Requests", pts)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	// Save the graphics to a PNG file.
	graphics.Save(plot, 22*vg.Centimeter, 15*vg.Centimeter,
		collector.outputDirPath+"\\"+"RequestsSucceeded.png")
}

func (collector *Collector) plotRequestsMessagesTradedOverTime() {
	plot := graphics.New("Average Messages Trader", "Time", "Number of Messages")

	pts := make(plotter.XYs, len(collector.snapshots))
	for i := range pts {
		pts[i].X = collector.snapshots[i].StartTime().Seconds()
		pts[i].Y = float64(collector.snapshots[i].RunRequestsAvgMessages())
	}

	err := plotutil.AddLinePoints(plot, "Total Messages", pts)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	// Save the graphics to a PNG file.
	graphics.Save(plot, 22*vg.Centimeter, 15*vg.Centimeter,
		collector.outputDirPath+"\\"+"MessagesPerRequest.png")
}

func (collector *Collector) plotAvailableResourcesOverTime() {
	plot := graphics.New("Available resources over time", "Time", "%Resources Available")

	availableResPts := make(plotter.XYs, len(collector.snapshots))
	requestsSucceeded := make(plotter.XYs, len(collector.snapshots))
	for i := range availableResPts {
		availableResPts[i].X = collector.snapshots[i].StartTime().Seconds()
		availableResPts[i].Y = float64(collector.snapshots[i].AllAvailableResourcesAvg())
		requestsSucceeded[i].X = collector.snapshots[i].StartTime().Seconds()
		requestsSucceeded[i].Y = collector.snapshots[i].RunRequestSuccessRatio()
	}

	err := plotutil.AddLinePoints(plot, "Available Resources", availableResPts,
		"Requests Succeeded", requestsSucceeded)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	// Save the graphics to a PNG file.
	graphics.Save(plot, 22*vg.Centimeter, 15*vg.Centimeter,
		collector.outputDirPath+"\\"+"AvailableResources.png")
}

func (collector *Collector) plotResourceDistributionOverTime() {
	plot := graphics.New("Available resources over time", "Time", "%Resources Available")

	availableResPts := make(plotter.XYs, len(collector.snapshots))
	requestsSucceeded := make(plotter.XYs, len(collector.snapshots))
	for i := range availableResPts {
		availableResPts[i].X = collector.snapshots[i].StartTime().Seconds()
		availableResPts[i].Y = float64(collector.snapshots[i].AllAvailableResourcesAvg())
		requestsSucceeded[i].X = collector.snapshots[i].StartTime().Seconds()
		requestsSucceeded[i].Y = collector.snapshots[i].RunRequestSuccessRatio()
	}

	err := plotutil.AddLinePoints(plot, "Available Resources", availableResPts,
		"Requests Succeeded", requestsSucceeded)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	// Save the graphics to a PNG file.
	graphics.Save(plot, 22*vg.Centimeter, 15*vg.Centimeter,
		collector.outputDirPath+"\\"+"AvailableResources.png")
}
