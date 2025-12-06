package yearHandlers

import (
	"context"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/zmb3/spotify/v2"
)

type LikedSongsBody struct {
	SaveObject bool `json:"saveObject"`
}

func (b LikedSongsBody) GetSaveObject() bool {
	return b.SaveObject
}

var LikedSongsFromYear = BaseYearHandler(func(ctx context.Context, client *spotify.Client, body LikedSongsBody) ([]spotify.FullTrack, error) {
	return services.GetAllUserSavedTracks(ctx, client)
})
