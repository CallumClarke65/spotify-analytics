package graphs

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
	charts "github.com/go-analyze/charts"
	"github.com/zmb3/spotify/v2"
	"go.uber.org/zap"
)

func GetTopTracksByYearHandler(w http.ResponseWriter, r *http.Request) {
	client := spotifyauth.ClientFromContext(r.Context())
	if client == nil {
		http.Error(w, `{"error":"Spotify client missing in context"}`, http.StatusUnauthorized)
		return
	}

	timeRangeStr := r.URL.Query().Get("time_range")
	if timeRangeStr == "" {
		http.Error(w, `{"error":"time_range is required"}`, http.StatusBadRequest)
		return
	}

	var timeRange spotify.Range
	switch timeRangeStr {
	case "short_term":
		timeRange = spotify.ShortTermRange
	case "medium_term":
		timeRange = spotify.MediumTermRange
	case "long_term":
		timeRange = spotify.LongTermRange
	default:
		http.Error(w, `{"error":"invalid time_range"}`, http.StatusBadRequest)
		return
	}

	tracks, err := services.GetTopTracks(r.Context(), client, timeRange)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	counts := map[int]float64{}
	for _, t := range tracks {
		if t.Album.ReleaseDate != "" {
			if year, e := strconv.Atoi(t.Album.ReleaseDate[:4]); e == nil {
				counts[year]++
			}
		}
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

	opt.YAxis[0].Unit = 25

	// Render chart
	painter := charts.NewPainter(charts.PainterOptions{
		Width:  1280,
		Height: 720,
	})
	if err := painter.BarChart(opt); err != nil {
		zap.L().Error("Failed to build chart", zap.Error(err))
		http.Error(w, `{"error":"failed to generate chart"}`, http.StatusInternalServerError)
		return
	}

	buf, err := painter.Bytes()
	if err != nil {
		zap.L().Error("Failed to render chart to bytes", zap.Error(err))
		http.Error(w, `{"error":"failed to render chart"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf)
}
