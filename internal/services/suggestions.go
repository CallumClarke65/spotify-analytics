package services

import (
	"context"
	"fmt"
	"sort"
	"strings"

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

			// Exclude artists with "classical" in any of their genres
			exclude := false
			for _, genre := range artist.Genres {
				if strings.Contains(strings.ToLower(genre), "classical") {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}

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
	seenTracks := make(map[string]bool)

	suggestedArtists, err := GetSuggestedArtists(ctx, client)
	if err != nil {
		return nil, err
	}

	for _, artist := range suggestedArtists {
		zap.L().Info("Getting suggested tracks from year for artist", zap.Int("year", year), zap.String("artist", artist.Name))

		query := fmt.Sprintf("year:%d artist:%s", year, artist.Name)
		sr, err := client.Search(ctx, query, spotify.SearchTypeTrack, spotify.Limit(50))
		if err != nil {
			return nil, err
		}

		tracks := sr.Tracks.Tracks
		// Sort by popularity descending
		sort.Slice(tracks, func(i, j int) bool {
			return tracks[i].Popularity > tracks[j].Popularity
		})

		artistTracks := make([]spotify.FullTrack, 0, 5)
		albumCount := make(map[string]int)  // albumID -> number of tracks included for this artist
		albumAdded := make(map[string]bool) // keep track if album has been used for diversity

		// Step 1: try to pick tracks from different albums first
		for _, track := range tracks {
			if len(artistTracks) >= 5 {
				break
			}
			trackID := track.ID.String()
			albumID := track.Album.ID.String()

			if seenTracks[trackID] || albumCount[albumID] >= 3 {
				continue
			}

			// prioritize albums not yet added
			if !albumAdded[albumID] {
				artistTracks = append(artistTracks, track)
				seenTracks[trackID] = true
				albumCount[albumID]++
				albumAdded[albumID] = true
			}
		}

		// Step 2: fill remaining slots with most popular tracks (even from albums already added) respecting 3 per album
		for _, track := range tracks {
			if len(artistTracks) >= 5 {
				break
			}
			trackID := track.ID.String()
			albumID := track.Album.ID.String()

			if seenTracks[trackID] || albumCount[albumID] >= 3 {
				continue
			}

			artistTracks = append(artistTracks, track)
			seenTracks[trackID] = true
			albumCount[albumID]++
		}

		allTrackSuggestions = append(allTrackSuggestions, artistTracks...)
	}

	return allTrackSuggestions, nil
}
