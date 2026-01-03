package graphs

import (
	"net/http"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
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

	buf, err := services.BarChartTracksByYear(
		r.Context(),
		client,
		tracks,
		spotifyauth.UserNameFromContext(r.Context())+"'s Top Tracks - "+timeRangeStr,
		25,
	)

	if err != nil {
		zap.L().Error("Failed to render chart to bytes", zap.Error(err))
		http.Error(w, `{"error":"failed to render chart"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf)
}

func GetTopTracksYearPopularityHeatmapHandler(w http.ResponseWriter, r *http.Request) {
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

	buf, err := services.HeatmapTracksByYearAndPopularity(
		r.Context(),
		client,
		tracks,
		spotifyauth.UserNameFromContext(r.Context())+"'s Top Tracks ("+timeRangeStr+") Year vs Popularity Heatmap",
		3,
		10,
	)

	if err != nil {
		zap.L().Error("Failed to render chart to bytes", zap.Error(err))
		http.Error(w, `{"error":"failed to render chart"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf)
}
