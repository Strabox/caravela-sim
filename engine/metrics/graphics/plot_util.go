package graphics

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"os"
	"path/filepath"
)

func NewPlot(title, xLabel, yLabel string, grid bool) *plot.Plot {
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
	dirPath := filepath.Dir(outFilePath)
	os.MkdirAll(dirPath, os.ModePerm)
	if err := plot.Save(width, height, outFilePath); err != nil {
		panic(err)
	}
}
