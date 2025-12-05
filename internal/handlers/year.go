package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

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

func SongsOnPlaylistsFromYear(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	client := spotifyauth.ClientFromContext(r.Context())

	var allPlaylists []spotify.SimplePlaylist
	allPlaylists, err = GetAllUserPlaylists(r.Context(), client)

	if err != nil {
		http.Error(w, "Failed to fetch playlists", http.StatusInternalServerError)
		return
	}

	trackMap := make(map[string]TrackInfo)
	for _, p := range allPlaylists {
		playlistItems, err := GetAllPlaylistTracks(r.Context(), client, p)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func GetAllUserPlaylists(ctx context.Context, client *spotify.Client) ([]spotify.SimplePlaylist, error) {
	var allPlaylists []spotify.SimplePlaylist

	page, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, err
	}
	allPlaylists = append(allPlaylists, page.Playlists...)

	for {
		err := client.NextPage(ctx, page)
		if err != nil {
			if err == spotify.ErrNoMorePages {
				break
			}
			zap.L().Warn("Failed to fetch next page of playlists", zap.Error(err))
			break
		}
		allPlaylists = append(allPlaylists, page.Playlists...)
	}

	zap.L().Info("Fetched all user playlists", zap.Int("count", len(allPlaylists)))
	return allPlaylists, nil
}

func GetAllPlaylistTracks(ctx context.Context, client *spotify.Client, playlist spotify.SimplePlaylist) ([]spotify.PlaylistItem, error) {
	var allTracks []spotify.PlaylistItem

	page, err := client.GetPlaylistItems(ctx, playlist.ID)
	if err != nil {
		return nil, err
	}
	allTracks = append(allTracks, page.Items...)

	for {
		err := client.NextPage(ctx, page)
		if err != nil {
			if err == spotify.ErrNoMorePages {
				break
			}
			zap.L().Warn("Failed to fetch next page of tracks", zap.Error(err))
			break
		}
		allTracks = append(allTracks, page.Items...)
	}

	zap.L().Info(
		"Fetched all tracks from playlist",
		zap.String("playlist_id", string(playlist.ID)),
		zap.Int("count", len(allTracks)),
	)
	return allTracks, nil
}
