package yearHandlers

import (
	"context"
	"net/http"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/zmb3/spotify/v2"
)

// LikedSongsBody godoc
// @Description Body for fetching liked songs
// @name LikedSongsBody
type LikedSongsBody struct {
	SaveObject bool `json:"saveObject"`
}

func (b LikedSongsBody) GetSaveObject() bool {
	return b.SaveObject
}

var LikedSongsFromYear = BaseYearHandler(func(ctx context.Context, client *spotify.Client, body LikedSongsBody) ([]spotify.FullTrack, error) {
	return services.GetAllUserSavedTracks(ctx, client)
})

// LikedSongsFromYearHandler godoc
// @Summary Get liked songs from a specific year
// @Description Returns a list of liked tracks filtered by year. Optionally saves the results if SaveObject=true.
// @Tags year
// @Accept json
// @Produce json
// @Param year path int true "Year to filter by"
// @Param body body LikedSongsBody true "Request body"
// @Success 200 {array} services.TrackInfo
// @Failure 400 {string} string "Invalid year or JSON body"
// @Failure 500 {string} string "Failed to fetch tracks"
// @Security ApiKeyAuth
// @Router /year/{year}/likedSongs [post]
func LikedSongsFromYearHandler(w http.ResponseWriter, r *http.Request) {
	LikedSongsFromYear(w, r)
}
