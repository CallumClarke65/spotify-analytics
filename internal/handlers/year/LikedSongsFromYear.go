package yearHandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
	"github.com/go-chi/chi/v5"
	"github.com/zmb3/spotify/v2"
	"go.uber.org/zap"
)

type LikedSongsFromYearRequestBody struct {
	SaveObject bool `json:"saveObject"`
}

func LikedSongsFromYear(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	var body LikedSongsFromYearRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	client := spotifyauth.ClientFromContext(r.Context())

	var userSavedTracks []spotify.FullTrack
	userSavedTracks, err = services.GetAllUserSavedTracks(r.Context(), client)

	if err != nil {
		http.Error(w, "Failed to fetch playlists", http.StatusInternalServerError)
		return
	}

	trackMap := make(map[string]services.TrackInfo)
	tracksFromYear := services.FilterTracksFromYear(userSavedTracks, year)
	for _, t := range tracksFromYear {
		info := services.GetShortTrackDetails(t)
		trackMap[info.TrackID] = info
	}

	if body.SaveObject {
		safeUsername := strings.Replace(spotifyauth.UserNameFromContext(r.Context()), " ", "_", -1)

		filename := fmt.Sprintf("liked_songs_%s_%s_%s", yearStr, safeUsername, time.Now().Format(time.RFC3339))
		err = services.WriteJsonObjectToFile(trackMap, filename)

		if err != nil {
			zap.L().Error("Failed to save songs from playlists object", zap.Error(err))
			http.Error(w, "Failed to save object", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trackMap)
}
