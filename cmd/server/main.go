package main

import (
	"log"
	"net/http"

	"github.com/CallumClarke65/spotify-analytics/internal/handlers"
	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
)

func main() {
	mux := http.NewServeMux()
	spotifyauth.Init()

	// Routes
	mux.HandleFunc("GET /ping", handlers.Ping)

	mux.HandleFunc("/login", spotifyauth.LoginHandler)
	mux.HandleFunc("/callback", spotifyauth.CallbackHandler)

	mux.HandleFunc("GET /me", handlers.Me)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
