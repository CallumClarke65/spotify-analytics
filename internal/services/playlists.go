package services

import (
	"context"

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
		zap.String("playlist_name", string(playlist.Name)),
		zap.Int("count", len(allTracks)),
	)
	return allTracks, nil
}
