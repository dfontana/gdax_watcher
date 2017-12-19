package common

import (
	"errors"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"os"
	"image/png"
	"time"
)

// Bollinger handles generating bollinger band plots from the given historical
// price data. This data should be in the form of time series structs, to
// handle time senstive data. Populate the struct, then call its methods.
type Bollinger struct {
	History	[]TimeSeries
}

// Plot will create a bollinger plot from data stored in the type,
// saved to the given path. Since Go-Chart provides a means to compute
// this itself, you don't have to call populate on this.
func (b *Bollinger) Plot(path string) error {
	if b.History == nil {
		return errors.New("No data to plot, please fill History field.")
	}
	// Reduce the time series into just x's and y's
	xv := make([]time.Time, len(b.History))
	yv := make([]float64, len(b.History))
	for i,item := range b.History {
		xv[i] = item.Time
		yv[i] = item.Data
	}

	// Construct series for plotting.
	hSeries := chart.TimeSeries {
		Name: "Prices",
		Style: chart.Style{
			Show: true,
			StrokeColor: chart.ColorBlue,
		},
		XValues: xv,
		YValues: yv,
	}

	// Build Bollinger Bands
	bbSeries := &chart.BollingerBandsSeries {
		Name: "Bollinger",
		Style: chart.Style{
			Show:        true,
			StrokeColor: drawing.ColorFromHex("426993"),
			FillColor:   drawing.ColorFromHex("426993").WithAlpha(64),
		},
		InnerSeries: hSeries,
	}

	// Figure out our Y Bounds real quick:
	hMin, hMax := minMaxbb(yv)
	lower := hMin - 50
	upper := hMax + 50

	// Plot it!
	graph := chart.Chart{
		Width: 1920,
		Height: 1080,
		DPI: 100,
		XAxis: chart.XAxis{
			Style:        chart.Style{Show: true},
			TickPosition: chart.TickPositionBetweenTicks,
			ValueFormatter: chart.TimeValueFormatter,
		},
		YAxis: chart.YAxis{
			Style: chart.Style{Show: true},
			Range: &chart.ContinuousRange{
				Max: upper,
				Min: lower,
			},
		},
		Series: []chart.Series{
			bbSeries,
			hSeries,
		},
	}

	if err := saveImagebb(graph, path); err != nil {
		return err
	}

	return nil
}

// Renders and saves graph.
func saveImagebb(graph chart.Chart, path string) error {
	// Write image to buffer
	collector := &chart.ImageWriter{}
	graph.Render(chart.PNG, collector)
	image, err := collector.Image()
	if err != nil {
		return err
	}
	
	// Save buffer to file (after encoding)
	outputFile, err := os.Create(path)
	if err != nil {
		return err
	}
	png.Encode(outputFile, image)
	outputFile.Close()
	return nil
}

// Returns the min and max from a slice
func minMaxbb(vals []float64) (float64, float64) {
	min := vals[0]
	max := vals[0]
	for _, val := range(vals){
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}
	return min, max
}