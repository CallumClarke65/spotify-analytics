package main

import (
	"net/http"
	"time"

	"github.com/CallumClarke65/spotify-analytics/internal/handlers"
	yearHandlers "github.com/CallumClarke65/spotify-analytics/internal/handlers/year"
	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type ZapLogFormatter struct{}

func (f *ZapLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	logger := zap.L().With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
	)

	if client := spotifyauth.ClientFromContext(r.Context()); client != nil {
		if user, err := client.CurrentUser(r.Context()); err == nil {
			logger = logger.With(zap.String("spotify_user", user.ID))
		}
	}

	return &ZapLogEntry{logger: logger}
}

type ZapLogEntry struct {
	logger *zap.Logger
}

func (l *ZapLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.logger.Info("Request completed",
		zap.Int("status", status),
		zap.Int("bytes", bytes),
		zap.Duration("duration", elapsed),
	)
}

func (l *ZapLogEntry) Panic(v interface{}, stack []byte) {
	l.logger.Error("Panic recovered",
		zap.Any("panic", v),
		zap.ByteString("stack", stack),
	)
}

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	spotifyauth.Init()

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestLogger(&ZapLogFormatter{}))

	r.Get("/ping", handlers.Ping)
	r.Get("/login", spotifyauth.LoginHandler)
	r.Get("/callback", spotifyauth.CallbackHandler)

	r.Group(func(r chi.Router) {
		r.Use(spotifyauth.SpotifyAuthMiddleware)

		r.Get("/me", handlers.Me)

		r.Post("/year/{year}/songsFromPlaylists", yearHandlers.SongsOnPlaylistsFromYear)
		r.Post("/year/{year}/likedSongs", yearHandlers.LikedSongsFromYear)
		r.Post("/year/{year}/suggestions", yearHandlers.SuggestionsFromYear)
	})

	logger.Info("Server started",
		zap.String("host", "localhost"),
		zap.Int("port", 8080),
	)

	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
