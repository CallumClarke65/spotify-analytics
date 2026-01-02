package services

import (
	"context"
	"sort"
	"strconv"

	"github.com/go-analyze/charts"
	"github.com/zmb3/spotify/v2"
	"go.uber.org/zap"
)

func GraphTracksByYear(
	ctx context.Context,
	client *spotify.Client,
	tracks []spotify.FullTrack,
	graphTitle string,
	trackCountUnit float64,
) ([]byte, error) {
	counts := map[int]float64{}
	for _, t := range tracks {
		rd := t.Album.ReleaseDate
		if rd == "" || rd == "0" {
			continue
		}

		year, err := strconv.Atoi(rd[:4])
		if err != nil || year == 0 {
			continue
		}

		counts[year]++
	}

	// Make sure our list of years starts at the beginning of a decade
	years := make([]int, 0, len(counts))
	for y := range counts {
		years = append(years, y)
	}
	sort.Ints(years)
	minYear := years[0]
	maxYear := years[len(years)-1]
	firstDecade := minYear - (minYear % 10)
	fullYears := []int{}
	for y := firstDecade; y <= maxYear; y++ {
		fullYears = append(fullYears, y)
	}

	values := make([][]float64, 1)
	values[0] = []float64{}
	xLabels := []string{}

	for _, y := range fullYears {
		xLabels = append(xLabels, strconv.Itoa(y))

		// Use actual count if present, otherwise 0
		if count, ok := counts[y]; ok {
			values[0] = append(values[0], count)
		} else {
			values[0] = append(values[0], 0)
		}
	}

	opt := charts.NewBarChartOptionWithData(values)

	opt.XAxis.Labels = xLabels
	opt.XAxis.Title = "Release Year"
	opt.XAxis.Unit = 5

	opt.Title.Text = graphTitle
	opt.YAxis[0].Unit = trackCountUnit

	// Render chart
	painter := charts.NewPainter(charts.PainterOptions{
		Width:  1280,
		Height: 720,
	})
	if err := painter.BarChart(opt); err != nil {
		zap.L().Error("Failed to build chart", zap.Error(err))
		return nil, err
	}

	return painter.Bytes()
}
