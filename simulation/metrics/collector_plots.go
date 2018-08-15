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
)

func (coll *Collector) plotRequestsSucceeded() {
	plotRes := graphics.New("Deploy Requests Success", "Time (Seconds)", "Request Succeeded", true)

	pts := make(plotter.XYs, len(coll.snapshots))
	for i := range pts {
		pts[i].X = coll.snapshots[i].EndTime().Seconds()
		pts[i].Y = float64(coll.snapshots[i].TotalRunRequestsSucceeded())
	}

	err := plotutil.AddLinePoints(plotRes, "Requests", pts)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, coll.outputDirPath+"\\"+"RequestsSucceeded.png")
}

func (coll *Collector) plotRequestsMessagesTradedPerRequest() {
	plotRes := graphics.New("Average Discover Messages", "Time (Seconds)", "#Avg Messages", true)

	pts := make(plotter.XYs, len(coll.snapshots))
	for i := range pts {
		pts[i].X = coll.snapshots[i].EndTime().Seconds()
		pts[i].Y = float64(coll.snapshots[i].RunRequestsAvgMessages())
	}

	err := plotutil.AddLinePoints(plotRes, "Total Messages", pts)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, coll.outputDirPath+"\\"+"MessagesPerRequest.png")
}

func (coll *Collector) plotFreeResources() {
	plotRes := graphics.New("Free Resources", "Time (Seconds)", "Resources Free %", true)

	availableResPts := make(plotter.XYs, len(coll.snapshots))
	requestsSucceeded := make(plotter.XYs, len(coll.snapshots))
	for i := range availableResPts {
		availableResPts[i].X = coll.snapshots[i].EndTime().Seconds()
		availableResPts[i].Y = float64(coll.snapshots[i].AllAvailableResourcesAvg())
		requestsSucceeded[i].X = coll.snapshots[i].EndTime().Seconds()
		requestsSucceeded[i].Y = coll.snapshots[i].RunRequestSuccessRatio()
	}

	err := plotutil.AddLinePoints(plotRes, "Free Resources", availableResPts,
		"Requests Succeeded", requestsSucceeded)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, coll.outputDirPath+"\\"+"FreeResources.png")
}

func (coll *Collector) plotRelayedGetOfferMessages() {
	plotRes := graphics.New("Total Relayed Get Offer", "Time (Seconds)", "#Relayed Get Offers", true)

	relayedGetOffersPts := make(plotter.XYs, len(coll.snapshots))
	for i := range relayedGetOffersPts {
		relayedGetOffersPts[i].X = coll.snapshots[i].EndTime().Seconds()
		relayedGetOffersPts[i].Y = float64(coll.snapshots[i].TotalGetOffersRelayed())
	}

	err := plotutil.AddLinePoints(plotRes, "Get Offers Relayed", relayedGetOffersPts)
	if err != nil {
		panic(errors.New("Problem with plots, error: " + err.Error()))
	}

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, coll.outputDirPath+"\\"+"GetOffersRelayed.png")
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
	fmt.Println("HM")
	return g.grid.At(c, r)
}

func (g *myGrid) X(c int) float64 {
	return 5
}

func (g *myGrid) Y(r int) float64 {
	return 10
}

func (coll *Collector) plotResourceDistribution() {
	grid := NewMyGrid()
	palette := palette.Heat(12, 1)
	heatMap := plotter.NewHeatMap(grid, palette)

	plotRes, err := plot.New()
	if err != nil {
		log.Panic(err)
	}
	plotRes.Title.Text = "Heat map"
	plotRes.X.Tick.Marker = plot.DefaultTicks{}
	plotRes.Y.Tick.Marker = plot.DefaultTicks{}
	plotRes.Add(heatMap)

	legend, err := plot.NewLegend() // Create a legend.
	if err != nil {
		log.Panic(err)
	}
	thumbs := plotter.PaletteThumbnailers(palette)
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		if i != 0 && i != len(thumbs)-1 {
			legend.Add("H", t)
			continue
		}
		var val float64
		switch i {
		case 0:
			val = heatMap.Min
		case len(thumbs) - 1:
			val = heatMap.Max
		}
		legend.Add(fmt.Sprintf("%.2g", val), t)
	}

	plotRes.X.Padding = 0
	plotRes.Y.Padding = 0
	plotRes.X.Max = 1.5
	plotRes.Y.Max = 1.5

	img := vgimg.New(650, 300)
	dc := draw.New(img)

	legend.Top = true
	// Calculate the width of the legend.
	r := legend.Rectangle(dc)
	legendWidth := r.Max.X - r.Min.X
	legend.YOffs = -plotRes.Title.Font.Extents().Height // Adjust the legend down a little.

	legend.Draw(dc)
	dc = draw.Crop(dc, 0, -legendWidth-vg.Millimeter, 0, 0) // Make space for the legend.
	plotRes.Draw(dc)

	graphics.Save(plotRes, 25*vg.Centimeter, 17*vg.Centimeter, coll.outputDirPath+"\\"+"HeatMap.png")

	/*
		w, err := os.Create(coll.outputDirPath + "\\" + "heatMap.png")
		if err != nil {
			log.Panic(err)
		}
		png := vgimg.PngCanvas{Canvas: img}
		if _, err = png.WriteTo(w); err != nil {
			log.Panic(err)
		}
	*/
}
