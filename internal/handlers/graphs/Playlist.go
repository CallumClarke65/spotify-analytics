package graphs

import (
	"net/http"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
	"github.com/zmb3/spotify/v2"
	"go.uber.org/zap"
)

func GetPlaylistTracksYearGraphHandler(w http.ResponseWriter, r *http.Request) {
	client := spotifyauth.ClientFromContext(r.Context())
	if client == nil {
		http.Error(w, `{"error":"Spotify client missing in context"}`, http.StatusUnauthorized)
		return
	}

	playlistId := r.URL.Query().Get("playlist_id")
	if playlistId == "" {
		http.Error(w, `{"error":"playlist_id is required"}`, http.StatusBadRequest)
		return
	}

	playlist, err := client.GetPlaylist(r.Context(), spotify.ID(playlistId))
	if err != nil {
		http.Error(w, `{"error":"playlist with id `+playlistId+` not found"}`, http.StatusNotFound)
		return
	}

	tracks, err := services.GetAllPlaylistTracks(r.Context(), client, playlist.SimplePlaylist)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	buf, err := services.GraphTracksByYear(
		r.Context(),
		client,
		tracks,
		playlist.Name+" - Tracks by Year",
		5,
	)

	if err != nil {
		zap.L().Error("Failed to render chart to bytes", zap.Error(err))
		http.Error(w, `{"error":"failed to render chart"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf)
}
