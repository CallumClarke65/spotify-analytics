package handlers

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

type TrackInfo struct {
	TrackID     string   `json:"track_id"`
	TrackName   string   `json:"track_name"`
	Artists     []string `json:"artists"`
	AlbumName   string   `json:"album_name"`
	ReleaseDate string   `json:"release_date"`
}

type SongsOnPlaylistsFromYearRequestBody struct {
	IgnoredPlaylistNameSubstrings []string `json:"ignoredPlaylistNameSubstrings"`
	SaveObject                    bool     `json:"saveObject"`
}

func SongsOnPlaylistsFromYear(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	var body SongsOnPlaylistsFromYearRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ignoredPlaylistNameSubstrings := make([]string, len(body.IgnoredPlaylistNameSubstrings))
	for i, s := range body.IgnoredPlaylistNameSubstrings {
		ignoredPlaylistNameSubstrings[i] = strings.ToLower(s)
	}
	zap.L().Debug("Ignoring playlists with substrings in name", zap.Strings("substrings", ignoredPlaylistNameSubstrings))

	client := spotifyauth.ClientFromContext(r.Context())

	var allPlaylists []spotify.SimplePlaylist
	allPlaylists, err = services.GetAllUserPlaylists(r.Context(), client)

	if err != nil {
		http.Error(w, "Failed to fetch playlists", http.StatusInternalServerError)
		return
	}

	trackMap := make(map[string]TrackInfo)
	for _, p := range allPlaylists {

		skip := false
		for _, sub := range ignoredPlaylistNameSubstrings {
			if strings.Contains(strings.ToLower(p.Name), sub) {
				skip = true
				break
			}
		}
		if skip {
			zap.L().Debug("Skipping ignored playlist", zap.String("playlist", p.Name))
			continue
		}

		playlistItems, err := services.GetAllPlaylistTracks(r.Context(), client, p)
		if err != nil {
			zap.L().Warn("Failed to fetch tracks for playlist", zap.String("playlist_id", string(p.ID)), zap.Error(err))
			continue
		}

		for _, item := range playlistItems {
			track := item.Track.Track

			if track.Album.ReleaseDate == "" {
				continue
			}
			releaseYear, err := strconv.Atoi(track.Album.ReleaseDate[:4])
			if err != nil || releaseYear != year {
				continue
			}

			// Skip duplicates
			if _, exists := trackMap[string(track.ID)]; exists {
				continue
			}

			artistNames := make([]string, len(track.Artists))
			for i, artist := range track.Artists {
				artistNames[i] = artist.Name
			}

			trackMap[string(track.ID)] = TrackInfo{
				TrackID:     track.ID.String(),
				TrackName:   track.Name,
				Artists:     artistNames,
				AlbumName:   track.Album.Name,
				ReleaseDate: track.Album.ReleaseDate,
			}
		}
	}

	results := make([]TrackInfo, 0, len(trackMap))
	for _, t := range trackMap {
		results = append(results, t)
	}

	if body.SaveObject {
		safeUsername := strings.Replace(spotifyauth.UserNameFromContext(r.Context()), " ", "_", -1)

		filename := fmt.Sprintf("songs_from_playlists_%s_%s_%s", yearStr, safeUsername, time.Now().Format(time.RFC3339))
		err = services.WriteJsonObjectToFile(results, filename)

		if err != nil {
			zap.L().Error("Failed to save songs from playlists object", zap.Error(err))
			http.Error(w, "Failed to save object", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
