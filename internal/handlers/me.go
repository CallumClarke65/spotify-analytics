package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
	"github.com/zmb3/spotify/v2"
)

func Me(w http.ResponseWriter, r *http.Request) {
	spotifyauth.RequireSpotifyAuth(func(w http.ResponseWriter, r *http.Request, client *spotify.Client) {
		user, err := client.CurrentUser(r.Context())
		if err != nil {
			http.Error(w, "Failed to fetch Spotify user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})(w, r)
}
