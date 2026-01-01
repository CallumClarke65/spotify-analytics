package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
)

// Me godoc
// @Summary Get current Spotify user
// @Description Returns information about the currently authenticated Spotify user
// @Tags user
// @Produce json
// @Success 200 {object} map[string]interface{} "Spotify user info"
// @Failure 401 {string} string "Spotify client missing in context"
// @Failure 500 {string} string "Failed to fetch Spotify user"
// @Security ApiKeyAuth
// @Router /me [get]
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
