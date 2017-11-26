package main

import (
	"image/color"
	"log"
	"net/http"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	"github.com/Bo0mer/connectivity"
)

func main() {
	r := connectivity.NewRecorder()
	defer r.Stop()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		xticks := plot.TimeTicks{Format: "2006-01-02\n15:04:05"}
		spans := r.Spans()
		pts := make(plotter.XYs, 2*len(spans))
		for i, span := range spans {
			pts[2*i].X = float64(span.StartTime.Unix())
			pts[2*i+1].X = float64(span.EndTime.Unix())
			if span.Kind == connectivity.Online {
				pts[2*i].Y = 1
				pts[2*i+1].Y = 1
			} else {
				pts[2*i].Y = 0
				pts[2*i+1].Y = 0
			}
		}

		p, err := plot.New()
		if err != nil {
			log.Panic(err)
		}
		p.Title.Text = "Connectivity"
		p.X.Tick.Marker = xticks
		p.Y.Label.Text = "Connectivity\n"
		p.Add(plotter.NewGrid())

		line, points, err := plotter.NewLinePoints(pts)
		if err != nil {
			log.Panic(err)
		}
		line.Color = color.RGBA{G: 255, A: 255}
		points.Shape = draw.CircleGlyph{}
		points.Color = color.RGBA{R: 255, A: 255}

		p.Add(line, points)
		wt, err := p.WriterTo(10*vg.Centimeter, 5*vg.Centimeter, "svg")
		if err != nil {
			log.Panic(err)
		}
		wt.WriteTo(w)
	})

	if err := http.ListenAndServe("localhost:3333", nil); err != nil {
		log.Fatal(err)
	}
}
