package metrics

import (
	"errors"
	"fmt"
	"github.com/strabox/caravela-sim/simulation/metrics/graphics"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"log"
	"os"
)

func (collector *Collector) plotRequestsSucceeded() {
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

func (collector *Collector) plotRequestsMessagesTraded() {
	plot := graphics.New("Average Lookup Messages", "Time (Seconds)", "#Avg Messages")

	pts := make(plotter.XYs, len(collector.snapshots))
	for i := range pts {
		pts[i].X = collector.snapshots[i].StartTime().Seconds()
		pts[i].Y = float64(collector.snapshots[i].RunRequestsAvgMessages())
	}

	err := plotutil.AddLinePoints(plot, "Total Messages", pts)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plot, 22*vg.Centimeter, 15*vg.Centimeter,
		collector.outputDirPath+"\\"+"MessagesPerRequest.png")
}

func (collector *Collector) plotAvailableResources() {
	plot := graphics.New("Free Resources", "Time (Seconds)", "Resources Free %")

	availableResPts := make(plotter.XYs, len(collector.snapshots))
	requestsSucceeded := make(plotter.XYs, len(collector.snapshots))
	for i := range availableResPts {
		availableResPts[i].X = collector.snapshots[i].StartTime().Seconds()
		availableResPts[i].Y = float64(collector.snapshots[i].AllAvailableResourcesAvg())
		requestsSucceeded[i].X = collector.snapshots[i].StartTime().Seconds()
		requestsSucceeded[i].Y = collector.snapshots[i].RunRequestSuccessRatio()
	}

	err := plotutil.AddLinePoints(plot, "Free Resources", availableResPts,
		"Requests Succeeded", requestsSucceeded)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plot, 22*vg.Centimeter, 15*vg.Centimeter,
		collector.outputDirPath+"\\"+"FreeResources.png")
}

type myGrid struct {
	grid *mat.Dense
}

func NewMyGrid() *myGrid {
	return &myGrid{
		grid: mat.NewDense(3, 4, []float64{
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
		}),
	}
}

func (g *myGrid) Dims() (int, int) {
	return 3, 4
}

func (g *myGrid) Z(c, r int) float64 {
	return g.grid.At(c, r)
}

func (g *myGrid) X(c int) float64 {
	return 2
}

func (g *myGrid) Y(r int) float64 {
	return 6
}

func (collector *Collector) plotResourceDistribution() {
	grid := NewMyGrid()
	pal := palette.Heat(12, 1)
	h := plotter.NewHeatMap(grid, pal)

	p, err := plot.New()
	if err != nil {
		log.Panic(err)
	}
	p.Title.Text = "Heat map"

	//p.X.Tick.Marker = integerTicks{}
	//p.Y.Tick.Marker = integerTicks{}

	p.Add(h)

	// Create a legend.
	l, err := plot.NewLegend()
	if err != nil {
		log.Panic(err)
	}
	thumbs := plotter.PaletteThumbnailers(pal)
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		if i != 0 && i != len(thumbs)-1 {
			l.Add("", t)
			continue
		}
		var val float64
		switch i {
		case 0:
			val = h.Min
		case len(thumbs) - 1:
			val = h.Max
		}
		l.Add(fmt.Sprintf("%.2g", val), t)
	}

	p.X.Padding = 0
	p.Y.Padding = 0
	p.X.Max = 1.5
	p.Y.Max = 1.5

	img := vgimg.New(250, 175)
	dc := draw.New(img)

	l.Top = true
	// Calculate the width of the legend.
	r := l.Rectangle(dc)
	legendWidth := r.Max.X - r.Min.X
	l.YOffs = -p.Title.Font.Extents().Height // Adjust the legend down a little.

	l.Draw(dc)
	dc = draw.Crop(dc, 0, -legendWidth-vg.Millimeter, 0, 0) // Make space for the legend.
	p.Draw(dc)
	w, err := os.Create("out/heatMap.png")
	if err != nil {
		log.Panic(err)
	}
	png := vgimg.PngCanvas{Canvas: img}
	if _, err = png.WriteTo(w); err != nil {
		log.Panic(err)
	}
}

func (collector *Collector) plotRelayedGetOfferMessages() {
	plot := graphics.New("Total Relayed Get Offer", "Time (Seconds)", "#Relayed Get Offers")

	relayedGetOffersPts := make(plotter.XYs, len(collector.snapshots))
	for i := range relayedGetOffersPts {
		relayedGetOffersPts[i].X = collector.snapshots[i].StartTime().Seconds()
		relayedGetOffersPts[i].Y = float64(collector.snapshots[i].TotalGetOffersRelayed())
	}

	err := plotutil.AddLinePoints(plot, "Get Offers Relayed", relayedGetOffersPts)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plot, 22*vg.Centimeter, 15*vg.Centimeter,
		collector.outputDirPath+"\\"+"GetOffersRelayed.png")
}
