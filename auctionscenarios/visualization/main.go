package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
)

func main() {
	summaries := LoadSummaries(os.Args[1])
	draw(summaries, CELLS, COMMUNICATIONS, true)
	draw(summaries, CELLS, SCORE, false)
	draw(summaries, CELLS, WAIT_TIME, false)
}

func draw(summaries Summaries, x string, y string, logScale bool) {
	dashes := map[float64][]vg.Length{
		0.05: []vg.Length{1, 4},
		0.1:  []vg.Length{4, 4, 1, 4},
		0.2:  []vg.Length{1},
		0.5:  []vg.Length{2, 1},
	}
	colors := map[int]color.Color{
		1:  color.RGBA{0, 0, 0, 255},
		2:  color.RGBA{0, 0, 255, 255},
		5:  color.RGBA{255, 0, 0, 255},
		10: color.RGBA{0, 255, 0, 255},
	}
	width := map[string]vg.Length{
		"all_rebid":             1,
		"compare_to_percentile": 2,
	}
	for _, scenario := range []string{LightLoad, HeavyLoad, RollingDeploy} {
		scenarioSubset := summaries.Filter(SCENARIO, scenario)
		fmt.Printf("Generating %s, %s, %s\n", scenario, x, y)
		p, err := plot.New()
		if err != nil {
			log.Fatalf("Couldn't make a new plot: ", err.Error())
		}
		p.Title.Text = scenario
		p.X.Label.Text = x
		p.Y.Label.Text = y
		for _, algorithm := range []string{"all_rebid", "compare_to_percentile"} {
			algorithmSubset := scenarioSubset.Filter(ALGORITHM, algorithm)
			for _, concurrency := range []int{1, 2, 5} {
				concurrencySubset := algorithmSubset.Filter(CONCURRENCY, concurrency)
				for _, biddingPoolFraction := range []float64{0.05, 0.1, 0.2} {
					subset := concurrencySubset.Filter(BIDDING_POOL_FRACTION, biddingPoolFraction)
					xy := subset.XY(x, y)
					line, err := plotter.NewLine(xy)
					if err != nil {
						log.Fatalf("failed to generate line plot:", err.Error())
					}
					line.LineStyle.Color = colors[concurrency]
					line.LineStyle.Dashes = dashes[biddingPoolFraction]
					line.LineStyle.Width = width[algorithm]
					p.Add(line)
					p.Legend.Add(fmt.Sprintf("A:%s - C:%d - f:%.1f", algorithm, concurrency, biddingPoolFraction), line)
				}
			}

		}

		p.Legend.Top = true
		p.Legend.Left = true
		if logScale {
			p.X.Tick.Marker = plot.LogTicks
			p.Y.Tick.Marker = plot.LogTicks
			p.X.Scale = plot.LogScale
			p.Y.Scale = plot.LogScale
		}

		err = p.Save(16, 16, fmt.Sprintf("%s_%s_%s.png", x, y, scenario))
		if err != nil {
			log.Fatalf("failed to save plot: %s", err.Error())
		}
	}
}
