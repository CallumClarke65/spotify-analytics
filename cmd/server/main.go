package main

import (
	"log"
	"net/http"

	"github.com/CallumClarke65/spotify-analytics/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("GET /ping", handlers.Ping)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
