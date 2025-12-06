package yearHandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
	"github.com/go-chi/chi/v5"
	"github.com/zmb3/spotify/v2"
)

type HasSaveObject interface {
	GetSaveObject() bool
}

type TrackFetcher[B any] func(ctx context.Context, client *spotify.Client, body B) ([]spotify.FullTrack, error)

func BaseYearHandler[B HasSaveObject](fetch TrackFetcher[B]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		yearStr := chi.URLParam(r, "year")
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "Invalid year", http.StatusBadRequest)
			return
		}

		var body B
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		client := spotifyauth.ClientFromContext(r.Context())
		tracks, err := fetch(r.Context(), client, body)
		if err != nil {
			http.Error(w, "Failed to fetch tracks", http.StatusInternalServerError)
			return
		}

		filtered := services.FilterTracksFromYear(tracks, year)

		var result []services.TrackInfo
		for _, t := range filtered {
			info := services.GetShortTrackDetails(t)
			result = append(result, info)
		}

		sort.Slice(result, func(i, j int) bool {
			return result[i].Popularity > result[j].Popularity
		})

		if body.GetSaveObject() {
			username := strings.ReplaceAll(spotifyauth.UserNameFromContext(r.Context()), " ", "_")
			filename := fmt.Sprintf("songs_%d_%s_%s", year, username, time.Now().Format(time.RFC3339))
			_ = services.WriteJsonObjectToFile(result, filename)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
