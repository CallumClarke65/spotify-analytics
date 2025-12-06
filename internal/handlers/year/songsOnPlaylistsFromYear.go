package yearHandlers

import (
	"context"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/zmb3/spotify/v2"
)

type SongsOnPlaylistsFromYearRequestBody struct {
	IgnoredPlaylistNameSubstrings []string `json:"ignoredPlaylistNameSubstrings"`
	SaveObject                    bool     `json:"saveObject"`
}

func (b SongsOnPlaylistsFromYearRequestBody) GetSaveObject() bool {
	return b.SaveObject
}

var SongsOnPlaylistsFromYear = BaseYearHandler(func(ctx context.Context, client *spotify.Client, body SongsOnPlaylistsFromYearRequestBody) ([]spotify.FullTrack, error) {
	playlists, _ := services.GetFilteredUserPlaylists(ctx, client, body.IgnoredPlaylistNameSubstrings)
	var all []spotify.FullTrack
	for _, p := range playlists {
		tracks, _ := services.GetAllPlaylistTracks(ctx, client, p)
		all = append(all, tracks...)
	}
	return all, nil
})
