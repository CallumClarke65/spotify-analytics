package yearHandlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/zmb3/spotify/v2"
)

type SuggestionsFromYearRequestBody struct {
	SaveObject bool `json:"saveObject"`
}

func (b SuggestionsFromYearRequestBody) GetSaveObject() bool {
	return b.SaveObject
}

var SuggestionsFromYear = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	year, _ := strconv.Atoi(yearStr)

	BaseYearHandler(func(ctx context.Context, client *spotify.Client, body SuggestionsFromYearRequestBody) ([]spotify.FullTrack, error) {
		return services.GetSuggestedTracksFromYear(ctx, client, year)
	})(w, r)
})
