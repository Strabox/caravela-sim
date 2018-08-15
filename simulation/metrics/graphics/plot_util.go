package graphics

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func New(title, xLabel, yLabel string, grid bool) *plot.Plot {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	if grid {
		p.Add(plotter.NewGrid())
	}

	return p
}

func Save(plot *plot.Plot, width, height vg.Length, outFilePath string) {
	if err := plot.Save(width, height, outFilePath); err != nil {
		panic(err)
	}
}
