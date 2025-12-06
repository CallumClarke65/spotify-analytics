package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zmb3/spotify/v2"
	"go.uber.org/zap"
)

func GetSuggestedArtists(
	ctx context.Context,
	client *spotify.Client,
) ([]spotify.FullArtist, error) {
	// Slice to store all unique artists
	var allSuggestedArtists []spotify.FullArtist
	seen := make(map[string]bool) // track artist IDs to avoid duplicates

	allSpotifyTimeRanges := []spotify.Range{
		spotify.ShortTermRange,
		spotify.MediumTermRange,
		spotify.LongTermRange,
	}

	for _, tr := range allSpotifyTimeRanges {
		page, err := client.CurrentUsersTopArtists(ctx, spotify.Timerange(tr), spotify.Limit(25))
		if err != nil {
			return nil, err
		}

		for _, artist := range page.Artists {
			if !seen[artist.ID.String()] {
				seen[artist.ID.String()] = true
				allSuggestedArtists = append(allSuggestedArtists, artist)
			}
		}
	}

	zap.L().Info("Found user's top artists", zap.Int("count", len(allSuggestedArtists)))
	return allSuggestedArtists, nil
}

func GetSuggestedTracksFromYear(
	ctx context.Context,
	client *spotify.Client,
	year int,
) ([]spotify.FullTrack, error) {
	var allTrackSuggestions []spotify.FullTrack
	seen := make(map[string]bool)

	suggestedArtists, err := GetSuggestedArtists(ctx, client)
	if err != nil {
		return nil, err
	}

	for _, artist := range suggestedArtists {
		zap.L().Info("Getting suggested tracks from year for artist", zap.Int("year", year), zap.String("artist", artist.Name))
		sr, err := client.Search(ctx, fmt.Sprintf("year:%s artist:%s", strconv.Itoa(year), artist.Name), spotify.SearchType(spotify.SearchTypeTrack), spotify.Limit(30))
		if err != nil {
			return nil, err
		}

		for _, track := range sr.Tracks.Tracks {
			if !seen[track.ID.String()] {
				seen[track.ID.String()] = true
				allTrackSuggestions = append(allTrackSuggestions, track)
			}
		}
	}

	return allTrackSuggestions, nil
}
