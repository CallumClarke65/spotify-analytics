package yearHandlers

import (
	"context"
	"net/http"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/zmb3/spotify/v2"
)

// SongsOnPlaylistsFromYearRequestBody godoc
// @Description Request body for fetching tracks from playlists filtered by year
// @name SongsOnPlaylistsFromYearRequestBody
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

// SongsOnPlaylistsFromYearHandler godoc
// @Summary Get tracks from user playlists filtered by year
// @Description Returns tracks from all playlists, excluding ones with ignored substrings. Optionally saves results if SaveObject=true.
// @Tags year
// @Accept json
// @Produce json
// @Param year path int true "Year to filter by"
// @Param body body SongsOnPlaylistsFromYearRequestBody true "Request body"
// @Success 200 {array} services.TrackInfo
// @Failure 400 {string} string "Invalid year or JSON body"
// @Failure 500 {string} string "Failed to fetch tracks"
// @Security ApiKeyAuth
// @Router /year/{year}/songsFromPlaylists [post]
func SongsOnPlaylistsFromYearHandler(w http.ResponseWriter, r *http.Request) {
	SongsOnPlaylistsFromYear(w, r)
}
