package services

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

func GetTopTracks(
	ctx context.Context,
	client *spotify.Client,
	timeRange spotify.Range,
) ([]spotify.FullTrack, error) {
	var allTracks []spotify.FullTrack
	limit := 50
	offset := 0

	for {
		page, err := client.CurrentUsersTopTracks(
			ctx,
			spotify.Timerange(timeRange),
			spotify.Limit(limit),
			spotify.Offset(offset),
		)
		if err != nil {
			return nil, fmt.Errorf("error fetching top tracks: %w", err)
		}

		allTracks = append(allTracks, page.Tracks...)

		if len(page.Tracks) < limit {
			break
		}
		offset += limit
	}

	return allTracks, nil
}
