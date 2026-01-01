package yearHandlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/zmb3/spotify/v2"
)

// SuggestionsFromYearRequestBody godoc
// @Description Body for fetching suggested tracks from a year
// @name SuggestionsFromYearRequestBody
type SuggestionsFromYearRequestBody struct {
	SaveObject bool `json:"saveObject"`
}

func (b SuggestionsFromYearRequestBody) GetSaveObject() bool {
	return b.SaveObject
}

// SuggestionsFromYearHandler godoc
// @Summary Get suggested tracks from a specific year
// @Description Returns a list of suggested tracks for the given year. Optionally saves results if SaveObject=true.
// @Tags year
// @Accept json
// @Produce json
// @Param year path int true "Year to get suggestions for"
// @Param body body SuggestionsFromYearRequestBody true "Request body"
// @Success 200 {array} services.TrackInfo
// @Failure 400 {string} string "Invalid year or JSON body"
// @Failure 500 {string} string "Failed to fetch tracks"
// @Security ApiKeyAuth
// @Router /year/{year}/suggestions [post]
func SuggestionsFromYearHandler(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	// Wrap the generic handler
	BaseYearHandler(func(ctx context.Context, client *spotify.Client, body SuggestionsFromYearRequestBody) ([]spotify.FullTrack, error) {
		return services.GetSuggestedTracksFromYear(ctx, client, year)
	})(w, r)
}
