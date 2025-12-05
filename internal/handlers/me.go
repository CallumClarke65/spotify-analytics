package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
)

func Me(w http.ResponseWriter, r *http.Request) {
	client := spotifyauth.ClientFromContext(r.Context())
	if client == nil {
		http.Error(w, "Spotify client missing in context", http.StatusUnauthorized)
		return
	}

	user, err := client.CurrentUser(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch Spotify user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
