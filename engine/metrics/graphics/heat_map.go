package graphics

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"image/color"
	"log"
	"math"
	"os"
)

func NewHeatMap(fileName, title, XLabel, YLabel string, yTicks []int, grid *UnitGrid, colorPalette palette.Palette) {
	heatMap := plotter.NewHeatMap(grid, colorPalette)

	plotRes := NewPlot(title, XLabel, YLabel, false)
	plotRes.Y.Tick.Marker = integerTicks{ticks: yTicks}
	plotRes.HideX()
	plotRes.Add(heatMap)

	legend, err := plot.NewLegend() // Create a legend.
	if err != nil {
		log.Panic(err)
	}
	thumbs := plotter.PaletteThumbnailers(colorPalette)
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		if i != 0 && i != len(thumbs)-1 {
			legend.Add("", t)
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

	img := vgimg.New(1850, 940)
	dc := draw.New(img)

	legend.Top = true
	// Calculate the width of the legend.
	r := legend.Rectangle(dc)
	legendWidth := r.Max.X - r.Min.X
	legend.YOffs = -plotRes.Title.Font.Extents().Height // Adjust the legend down a little.

	legend.Draw(dc)
	dc = draw.Crop(dc, 0, -legendWidth-vg.Millimeter, 0, 0) // Make space for the legend.
	plotRes.Draw(dc)

	w, err := os.Create(fileName)
	if err != nil {
		log.Panic(err)
	}
	png := vgimg.PngCanvas{Canvas: img}
	if _, err = png.WriteTo(w); err != nil {
		log.Panic(err)
	}
}

// UnitGrid is a generic data grid.
type UnitGrid struct {
	Data mat.Matrix
}

func (g UnitGrid) Dims() (c, r int) {
	r, c = g.Data.Dims()
	return
}

func (g UnitGrid) Z(c, r int) float64 {
	return g.Data.At(r, c)
}

func (g UnitGrid) X(c int) float64 {
	_, totalColumns := g.Data.Dims()
	if c < 0 || c >= totalColumns {
		panic("heat map index out of range")
	}
	return float64(c) - float64(totalColumns-2)
}

func (g UnitGrid) Y(r int) float64 {
	totalRows, _ := g.Data.Dims()
	if r < 0 || r >= totalRows {
		panic("heat map index out of range")
	}
	return float64(totalRows) - float64(r-1)
}

type integerTicks struct {
	ticks []int
}

func (it integerTicks) Ticks(min, max float64) []plot.Tick {
	var t []plot.Tick
	ticksIndex := 0
	for i := math.Trunc(min); i <= max; i++ {
		if ticksIndex < len(it.ticks) {
			t = append(t, plot.Tick{Value: i, Label: fmt.Sprint(it.ticks[ticksIndex])})
		} else {
			t = append(t, plot.Tick{Value: i, Label: fmt.Sprint(-1)})
		}
		ticksIndex++
	}

	return t
}

type HardcodedPalette struct {
	colors []color.Color
}

func (h *HardcodedPalette) Colors() []color.Color {
	return h.colors
}

func MyHeatPalette() palette.Palette {
	colors := []color.Color{
		color.RGBA{R: 255, G: 245, B: 240, A: 255},
		color.RGBA{R: 254, G: 224, B: 210, A: 255},
		color.RGBA{R: 252, G: 187, B: 161, A: 255},
		color.RGBA{R: 252, G: 146, B: 114, A: 255},
		color.RGBA{R: 251, G: 106, B: 74, A: 255},
		color.RGBA{R: 239, G: 59, B: 44, A: 255},
		color.RGBA{R: 203, G: 24, B: 29, A: 255},
		color.RGBA{R: 203, G: 24, B: 29, A: 255},
		color.RGBA{R: 165, G: 15, B: 21, A: 255},
		color.RGBA{R: 103, G: 0, B: 13, A: 255},
	}
	return &HardcodedPalette{colors: colors}
}

func MyInvertedHeatPalette() palette.Palette {
	colors := []color.Color{
		color.RGBA{R: 103, G: 0, B: 13, A: 255},
		color.RGBA{R: 165, G: 15, B: 21, A: 255},
		color.RGBA{R: 203, G: 24, B: 29, A: 255},
		color.RGBA{R: 239, G: 59, B: 44, A: 255},
		color.RGBA{R: 251, G: 106, B: 74, A: 255},
		color.RGBA{R: 252, G: 146, B: 114, A: 255},
		color.RGBA{R: 252, G: 187, B: 161, A: 255},
		color.RGBA{R: 254, G: 224, B: 210, A: 255},
		color.RGBA{R: 255, G: 245, B: 240, A: 255},
	}
	return &HardcodedPalette{colors: colors}
}
