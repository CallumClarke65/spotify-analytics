package spotifyauth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauthpkg "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	authenticator *spotifyauthpkg.Authenticator
	state         = "random-state" // in production, generate per session
)

// Init loads .env and sets up Spotify authenticator
func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env: %v", err)
	}

	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectURL := os.Getenv("SPOTIFY_REDIRECT_URL")

	authenticator = spotifyauthpkg.New(
		spotifyauthpkg.WithClientID(clientID),
		spotifyauthpkg.WithClientSecret(clientSecret),
		spotifyauthpkg.WithRedirectURL(redirectURL),
		spotifyauthpkg.WithScopes(spotifyauthpkg.ScopeUserReadEmail, spotifyauthpkg.ScopeUserReadPrivate),
	)
}

// LoginHandler redirects the user to Spotify's OAuth page
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := authenticator.AuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// CallbackHandler exchanges code for token and returns it as JSON
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	rState := r.URL.Query().Get("state")
	if rState != state {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	token, err := authenticator.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
		return
	}

	// Return the token as JSON so Postman can store it
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  token.AccessToken,
		"token_type":    token.TokenType,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry,
	})
}

// Middleware to require a Bearer token and provide a Spotify client
func RequireSpotifyAuth(next func(http.ResponseWriter, *http.Request, *spotify.Client)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		if bearer == "" || len(bearer) < 7 || bearer[:7] != "Bearer " {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		token := &oauth2.Token{
			AccessToken: bearer[7:], // strip "Bearer "
		}

		client := spotify.New(authenticator.Client(context.Background(), token))
		next(w, r, client)
	}
}
