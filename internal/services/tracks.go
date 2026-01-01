package services

import (
	"context"
	"strconv"

	"github.com/zmb3/spotify/v2"
	"go.uber.org/zap"
)

// TrackInfo godoc
// @Description Short track info returned by year endpoints
// @name TrackInfo
type TrackInfo struct {
	TrackID     string   `json:"track_id"`
	TrackName   string   `json:"track_name"`
	Artists     []string `json:"artists"`
	AlbumName   string   `json:"album_name"`
	ReleaseDate string   `json:"release_date"`
	Popularity  int      `json:"popularity"`
}

func FilterTracksFromYear(tracks []spotify.FullTrack, year int) []spotify.FullTrack {
	seen := make(map[string]bool)
	result := make([]spotify.FullTrack, 0)

	for _, track := range tracks {
		if track.Album.ReleaseDate == "" {
			continue
		}

		releaseYear, err := strconv.Atoi(track.Album.ReleaseDate[:4])
		if err != nil || releaseYear != year {
			continue
		}

		id := track.ID.String()

		if seen[id] {
			continue
		}
		seen[id] = true

		result = append(result, track)
	}

	return result
}

func GetShortTrackDetails(track spotify.FullTrack) TrackInfo {
	artistNames := make([]string, len(track.Artists))
	for i, artist := range track.Artists {
		artistNames[i] = artist.Name
	}

	return TrackInfo{
		TrackID:     track.ID.String(),
		TrackName:   track.Name,
		Artists:     artistNames,
		AlbumName:   track.Album.Name,
		ReleaseDate: track.Album.ReleaseDate,
		Popularity:  int(track.Popularity),
	}
}

func GetAllUserSavedTracks(ctx context.Context, client *spotify.Client) ([]spotify.FullTrack, error) {
	var allTracks []spotify.FullTrack

	page, err := client.CurrentUsersTracks(ctx)
	if err != nil {
		return nil, err
	}

	for _, track := range page.Tracks {
		allTracks = append(allTracks, track.FullTrack)
	}

	for {
		err := client.NextPage(ctx, page)
		if err != nil {
			if err == spotify.ErrNoMorePages {
				break
			}
			zap.L().Warn("Failed to fetch next page of user saved tracks", zap.Error(err))
			break
		}
		for _, track := range page.Tracks {
			allTracks = append(allTracks, track.FullTrack)
		}
	}

	zap.L().Info("Fetched all user saved tracks", zap.Int("count", len(allTracks)))

	return allTracks, nil
}
