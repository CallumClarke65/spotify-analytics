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

type ctxKey string

const spotifyClientKey ctxKey = "spotifyClient"

var (
	authenticator *spotifyauthpkg.Authenticator
	state         = "random-state" // in production, generate per session
)

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
		spotifyauthpkg.WithScopes(
			spotifyauthpkg.ScopeUserReadEmail,
			spotifyauthpkg.ScopeUserReadPrivate,
			spotifyauthpkg.ScopePlaylistModifyPrivate,
			spotifyauthpkg.ScopePlaylistModifyPublic,
			spotifyauthpkg.ScopePlaylistReadPrivate,
			spotifyauthpkg.ScopeUserLibraryRead,
			spotifyauthpkg.ScopeUserTopRead,
		),
	)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := authenticator.AuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  token.AccessToken,
		"token_type":    token.TokenType,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry,
	})
}

func SpotifyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		bearer := r.Header.Get("Authorization")
		if bearer == "" || len(bearer) < 7 || bearer[:7] != "Bearer " {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		token := &oauth2.Token{
			AccessToken: bearer[7:], // strip "Bearer "
		}

		// Build Spotify client
		client := spotify.New(authenticator.Client(r.Context(), token))

		// Attach client to context
		ctx := context.WithValue(r.Context(), spotifyClientKey, client)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ClientFromContext(ctx context.Context) *spotify.Client {
	client, _ := ctx.Value(spotifyClientKey).(*spotify.Client)
	return client
}

func UserNameFromContext(ctx context.Context) string {
	client := ClientFromContext(ctx)
	if client == nil {
		return ""
	}
	user, err := client.CurrentUser(ctx)
	if err != nil {
		return ""
	}

	return user.DisplayName
}

func UserIDFromContext(ctx context.Context) string {
	client := ClientFromContext(ctx)
	if client == nil {
		return ""
	}
	user, err := client.CurrentUser(ctx)
	if err != nil {
		return ""
	}

	return user.ID
}
