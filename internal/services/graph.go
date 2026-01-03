package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/go-analyze/charts"
	"github.com/zmb3/spotify/v2"
	"go.uber.org/zap"
)

func BarChartTracksByYear(
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

func HeatmapTracksByYearAndPopularity(
	ctx context.Context,
	client *spotify.Client,
	tracks []spotify.FullTrack,
	graphTitle string,
	yearBucketSize int, // e.g. 3
	popularityBucketSize int, // e.g. 10
) ([]byte, error) {

	type cell struct {
		x int
		y int
	}

	// --- Collect valid years and popularity ---
	years := []int{}
	for _, t := range tracks {
		if t.Album.ReleaseDate == "" {
			continue
		}
		year, err := strconv.Atoi(t.Album.ReleaseDate[:4])
		if err != nil || year == 0 {
			continue
		}
		years = append(years, year)
	}

	if len(years) == 0 {
		return nil, fmt.Errorf("no valid release years found")
	}

	sort.Ints(years)
	minYear := years[0]
	maxYear := years[len(years)-1]

	// --- Build year buckets ---
	bucketCount := int(math.Ceil(float64(maxYear-minYear+1) / float64(yearBucketSize)))
	xLabels := make([]string, bucketCount)
	bucketStartYears := make([]int, bucketCount)

	for i := 0; i < bucketCount; i++ {
		start := minYear + i*yearBucketSize
		end := start + yearBucketSize - 1

		startCentury := start / 100
		endCentury := end / 100

		if startCentury == endCentury {
			// Same century, safe to shorten last two digits
			xLabels[i] = fmt.Sprintf("%d-%02d", start, end%100)
		} else {
			// Different centuries, show full years
			xLabels[i] = fmt.Sprintf("%d-%d", start, end)
		}
		bucketStartYears[i] = start
	}

	yearBucketIndex := map[int]int{}
	for i, start := range bucketStartYears {
		yearBucketIndex[start] = i
	}

	// --- Build popularity buckets ---
	popBucketCount := int(math.Ceil(101.0 / float64(popularityBucketSize)))
	yLabels := make([]string, popBucketCount)
	for j := 0; j < popBucketCount; j++ {
		start := j * popularityBucketSize
		end := start + popularityBucketSize - 1
		if end > 100 {
			end = 100
		}
		yLabels[j] = fmt.Sprintf("%d-%d", start, end)
	}

	// --- Count cells ---
	counts := map[cell]float64{}
	for _, t := range tracks {
		if t.Album.ReleaseDate == "" {
			continue
		}
		year, err := strconv.Atoi(t.Album.ReleaseDate[:4])
		if err != nil {
			continue
		}
		pop := int(t.Popularity)

		x := (year - minYear) / yearBucketSize
		y := pop / popularityBucketSize

		if x < 0 || x >= bucketCount {
			continue
		}
		if y < 0 || y >= popBucketCount {
			continue
		}

		counts[cell{x, y}]++
	}

	// --- Convert to heatmap matrix ---
	values := make([][]float64, popBucketCount)
	for y := 0; y < popBucketCount; y++ {
		values[y] = make([]float64, bucketCount)
	}

	for c, v := range counts {
		values[c.y][c.x] = v
	}

	// --- Build chart ---
	opt := charts.NewHeatMapOptionWithData(values)
	opt.Title.Text = graphTitle
	opt.XAxis.Labels = xLabels
	opt.XAxis.Title = "Release Year"
	opt.XAxis.LabelRotation = math.Pi / 2
	opt.YAxis.Labels = yLabels
	opt.YAxis.Title = "Popularity on Spotify"

	// --- Render ---
	painter := charts.NewPainter(charts.PainterOptions{
		Width:  1280,
		Height: 720,
	})

	if err := painter.HeatMapChart(opt); err != nil {
		zap.L().Error("Failed to build heatmap", zap.Error(err))
		return nil, err
	}

	return painter.Bytes()
}
