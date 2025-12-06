package services

import (
	"context"
	"strings"

	"github.com/zmb3/spotify/v2"
	"go.uber.org/zap"
)

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

func GetFilteredUserPlaylists(
	ctx context.Context,
	client *spotify.Client,
	ignoredPlaylistNameSubstrings []string,
) ([]spotify.SimplePlaylist, error) {

	lowerIgnored := make([]string, len(ignoredPlaylistNameSubstrings))
	for i, s := range ignoredPlaylistNameSubstrings {
		lowerIgnored[i] = strings.ToLower(s)
	}
	zap.L().Debug("Ignoring playlists with substrings", zap.Strings("substrings", lowerIgnored))

	allPlaylists, err := GetAllUserPlaylists(ctx, client)
	if err != nil {
		return nil, err
	}

	var filtered []spotify.SimplePlaylist

	for _, p := range allPlaylists {

		skip := false
		for _, sub := range lowerIgnored {
			if strings.Contains(strings.ToLower(p.Name), sub) {
				skip = true
				break
			}
		}

		if skip {
			zap.L().Debug("Skipping playlist", zap.String("name", p.Name))
			continue
		}

		filtered = append(filtered, p)
	}

	return filtered, nil
}

func GetAllPlaylistTracks(
	ctx context.Context,
	client *spotify.Client,
	playlist spotify.SimplePlaylist,
) ([]spotify.FullTrack, error) {

	var allTracks []spotify.FullTrack

	page, err := client.GetPlaylistItems(ctx, playlist.ID)
	if err != nil {
		return nil, err
	}

	// Extract FullTrack from this page
	for _, item := range page.Items {
		allTracks = append(allTracks, *item.Track.Track)
	}

	// Fetch remaining pages
	for {
		err := client.NextPage(ctx, page)
		if err != nil {
			if err == spotify.ErrNoMorePages {
				break
			}
			zap.L().Warn("Failed to fetch next page of tracks", zap.Error(err))
			break
		}

		for _, item := range page.Items {
			allTracks = append(allTracks, *item.Track.Track)
		}
	}

	zap.L().Info(
		"Fetched all tracks from playlist",
		zap.String("playlist_name", playlist.Name),
		zap.Int("count", len(allTracks)),
	)

	return allTracks, nil
}
