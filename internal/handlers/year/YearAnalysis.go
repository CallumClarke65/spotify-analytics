package yearHandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/CallumClarke65/spotify-analytics/internal/services"
	"github.com/CallumClarke65/spotify-analytics/internal/spotifyauth"
	"github.com/go-chi/chi/v5"
	"github.com/zmb3/spotify/v2"
)

type YearAnalysisRequestBody struct {
	IgnoredPlaylistNameSubstrings []string `json:"ignoredPlaylistNameSubstrings"`
	SaveObject                    bool     `json:"saveObject"`
	MakePlaylists                 bool     `json:"makePlaylists"`
}

func (b YearAnalysisRequestBody) GetSaveObject() bool {
	return b.SaveObject
}

func (b YearAnalysisRequestBody) GetMakePlaylists() bool {
	return b.MakePlaylists
}

func fetchTracksForYear(
	ctx context.Context,
	client *spotify.Client,
	year int,
	fetch func(ctx context.Context, client *spotify.Client) ([]spotify.FullTrack, error),
) ([]services.TrackInfo, error) {

	tracks, err := fetch(ctx, client)
	if err != nil {
		return nil, err
	}

	filtered := services.FilterTracksFromYear(tracks, year)
	result := make([]services.TrackInfo, 0, len(filtered))
	for _, t := range filtered {
		result = append(result, services.GetShortTrackDetails(t))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Popularity > result[j].Popularity
	})

	return result, nil
}

var YearAnalysis = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	var body YearAnalysisRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	client := spotifyauth.ClientFromContext(r.Context())

	onPlaylists, _ := fetchTracksForYear(r.Context(), client, year, func(ctx context.Context, client *spotify.Client) ([]spotify.FullTrack, error) {
		playlists, err := services.GetFilteredUserPlaylists(ctx, client, body.IgnoredPlaylistNameSubstrings)
		if err != nil {
			return nil, err
		}
		all := make([]spotify.FullTrack, 0)
		for _, p := range playlists {
			tracks, _ := services.GetAllPlaylistTracks(ctx, client, p)
			all = append(all, tracks...)
		}
		return all, nil
	})

	liked, _ := fetchTracksForYear(r.Context(), client, year, services.GetAllUserSavedTracks)

	seen := make(map[string]struct{})
	for _, t := range onPlaylists {
		seen[t.TrackID] = struct{}{}
	}
	for _, t := range liked {
		seen[t.TrackID] = struct{}{}
	}

	suggestionsAll, _ := fetchTracksForYear(r.Context(), client, year, func(ctx context.Context, client *spotify.Client) ([]spotify.FullTrack, error) {
		return services.GetSuggestedTracksFromYear(ctx, client, year)
	})

	suggestions := make([]services.TrackInfo, 0, len(suggestionsAll))
	for _, t := range suggestionsAll {
		if _, exists := seen[t.TrackID]; !exists {
			suggestions = append(suggestions, t)
			seen[t.TrackID] = struct{}{}
		}
	}

	if body.GetSaveObject() {
		username := spotifyauth.UserNameFromContext(r.Context())
		_ = services.WriteJsonObjectToFile(map[string]interface{}{
			"on_playlists": onPlaylists,
			"liked":        liked,
			"suggestions":  suggestions,
		}, "year_analysis_"+strconv.Itoa(year)+"_"+username)
	}

	if body.GetMakePlaylists() {
		userId := spotifyauth.UserIDFromContext(r.Context())

		// Combine liked + onPlaylists for "favourites"
		favouritesTracks := append(onPlaylists, liked...)

		// Convert to Spotify IDs
		favTrackIDs := make([]spotify.ID, 0, len(favouritesTracks))
		for _, t := range favouritesTracks {
			favTrackIDs = append(favTrackIDs, spotify.ID(t.TrackID))
		}

		sugTrackIDs := make([]spotify.ID, 0, len(suggestions))
		for _, t := range suggestions {
			sugTrackIDs = append(sugTrackIDs, spotify.ID(t.TrackID))
		}

		// Create playlists
		favPlaylist, err := client.CreatePlaylistForUser(
			r.Context(), userId,
			strconv.Itoa(year)+" - favourites",
			"Generated playlist of favourites for "+strconv.Itoa(year),
			false,
			false,
		)
		if err != nil {
			http.Error(w, "Failed to create favourites playlist: "+err.Error(), http.StatusInternalServerError)
			return
		}

		sugPlaylist, err := client.CreatePlaylistForUser(
			r.Context(), userId,
			strconv.Itoa(year)+" - suggestions",
			"Generated playlist of suggested tracks for "+strconv.Itoa(year),
			false,
			false,
		)
		if err != nil {
			http.Error(w, "Failed to create suggestions playlist: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Add tracks in batches of 100 (Spotify API limit)
		batchSize := 100
		for i := 0; i < len(favTrackIDs); i += batchSize {
			end := i + batchSize
			if end > len(favTrackIDs) {
				end = len(favTrackIDs)
			}
			client.AddTracksToPlaylist(r.Context(), favPlaylist.ID, favTrackIDs[i:end]...)
		}
		for i := 0; i < len(sugTrackIDs); i += batchSize {
			end := i + batchSize
			if end > len(sugTrackIDs) {
				end = len(sugTrackIDs)
			}
			client.AddTracksToPlaylist(r.Context(), sugPlaylist.ID, sugTrackIDs[i:end]...)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"on_playlists": onPlaylists,
		"liked":        liked,
		"suggestions":  suggestions,
	})
})
