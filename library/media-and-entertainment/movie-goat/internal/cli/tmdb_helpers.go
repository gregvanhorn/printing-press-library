package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/movie-goat/internal/client"
)

// tmdbSearchResult represents a single result from TMDb search endpoints.
type tmdbSearchResult struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Name          string  `json:"name"`
	ReleaseDate   string  `json:"release_date"`
	FirstAirDate  string  `json:"first_air_date"`
	VoteAverage   float64 `json:"vote_average"`
	Overview      string  `json:"overview"`
	MediaType     string  `json:"media_type"`
	Popularity    float64 `json:"popularity"`
	PosterPath    string  `json:"poster_path"`
	OriginalTitle string  `json:"original_title"`
	OriginalName  string  `json:"original_name"`
	GenreIDs      []int   `json:"genre_ids"`
}

// DisplayTitle returns the appropriate title for either movies or TV shows.
func (r *tmdbSearchResult) DisplayTitle() string {
	if r.Title != "" {
		return r.Title
	}
	return r.Name
}

// Year returns the year from the release date or first air date.
func (r *tmdbSearchResult) Year() string {
	d := r.ReleaseDate
	if d == "" {
		d = r.FirstAirDate
	}
	if len(d) >= 4 {
		return d[:4]
	}
	return ""
}

// tmdbSearchResponse represents the envelope from TMDb search/discover endpoints.
type tmdbSearchResponse struct {
	Page         int                `json:"page"`
	Results      []tmdbSearchResult `json:"results"`
	TotalPages   int                `json:"total_pages"`
	TotalResults int                `json:"total_results"`
}

// tmdbMovieDetail represents a detailed movie response from TMDb /movie/{id}.
type tmdbMovieDetail struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	Runtime     int     `json:"runtime"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
	Budget      int64   `json:"budget"`
	Revenue     int64   `json:"revenue"`
	Popularity  float64 `json:"popularity"`
	Tagline     string  `json:"tagline"`
	Status      string  `json:"status"`
	ImdbID      string  `json:"imdb_id"`
	PosterPath  string  `json:"poster_path"`
	Genres      []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	ProductionCompanies []struct {
		Name string `json:"name"`
	} `json:"production_companies"`
	ExternalIDs     json.RawMessage     `json:"external_ids"`
	Credits         *tmdbCredits        `json:"credits"`
	WatchProviders  json.RawMessage     `json:"watch/providers"`
	Videos          json.RawMessage     `json:"videos"`
	Recommendations *tmdbSearchResponse `json:"recommendations"`
}

// tmdbCredits represents the credits response.
type tmdbCredits struct {
	Cast []tmdbCastMember `json:"cast"`
	Crew []tmdbCrewMember `json:"crew"`
}

// tmdbCastMember represents a single cast member.
type tmdbCastMember struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Character   string  `json:"character"`
	Order       int     `json:"order"`
	Popularity  float64 `json:"popularity"`
	ProfilePath string  `json:"profile_path"`
}

// tmdbCrewMember represents a single crew member.
type tmdbCrewMember struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Job         string  `json:"job"`
	Department  string  `json:"department"`
	Popularity  float64 `json:"popularity"`
	ProfilePath string  `json:"profile_path"`
}

// tmdbPersonDetail represents a detailed person response.
type tmdbPersonDetail struct {
	ID              int                  `json:"id"`
	Name            string               `json:"name"`
	Biography       string               `json:"biography"`
	Birthday        string               `json:"birthday"`
	Deathday        string               `json:"deathday"`
	PlaceOfBirth    string               `json:"place_of_birth"`
	ProfilePath     string               `json:"profile_path"`
	KnownFor        string               `json:"known_for_department"`
	Popularity      float64              `json:"popularity"`
	CombinedCredits *tmdbCombinedCredits `json:"combined_credits"`
}

// tmdbCombinedCredits contains cast and crew credits across movies and TV.
type tmdbCombinedCredits struct {
	Cast []tmdbCombinedCreditEntry `json:"cast"`
	Crew []tmdbCombinedCreditEntry `json:"crew"`
}

// tmdbCombinedCreditEntry represents one credit.
type tmdbCombinedCreditEntry struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Name         string  `json:"name"`
	Character    string  `json:"character"`
	Job          string  `json:"job"`
	Department   string  `json:"department"`
	MediaType    string  `json:"media_type"`
	ReleaseDate  string  `json:"release_date"`
	FirstAirDate string  `json:"first_air_date"`
	VoteAverage  float64 `json:"vote_average"`
	VoteCount    int     `json:"vote_count"`
	Popularity   float64 `json:"popularity"`
	EpisodeCount int     `json:"episode_count"`
	PosterPath   string  `json:"poster_path"`
	Overview     string  `json:"overview"`
}

// DisplayTitle returns the appropriate title.
func (e *tmdbCombinedCreditEntry) DisplayTitle() string {
	if e.Title != "" {
		return e.Title
	}
	return e.Name
}

// Year returns the year from the release/air date.
func (e *tmdbCombinedCreditEntry) Year() string {
	d := e.ReleaseDate
	if d == "" {
		d = e.FirstAirDate
	}
	if len(d) >= 4 {
		return d[:4]
	}
	return ""
}

// tmdbWatchProviders is the watch/providers response structure.
type tmdbWatchProviders struct {
	Results map[string]tmdbRegionProviders `json:"results"`
}

// tmdbRegionProviders contains providers for one region.
type tmdbRegionProviders struct {
	Link     string         `json:"link"`
	Flatrate []tmdbProvider `json:"flatrate"`
	Rent     []tmdbProvider `json:"rent"`
	Buy      []tmdbProvider `json:"buy"`
	Free     []tmdbProvider `json:"free"`
	Ads      []tmdbProvider `json:"ads"`
}

// tmdbProvider represents a single streaming/rental provider.
type tmdbProvider struct {
	ProviderID      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
	LogoPath        string `json:"logo_path"`
	DisplayPriority int    `json:"display_priority"`
}

// searchMovieByTitle searches TMDb for a movie by title and returns the top result's ID.
// Returns 0 and an error if no results found.
func searchMovieByTitle(c *client.Client, title string) (int, string, error) {
	data, err := c.Get("/search/movie", map[string]string{"query": title})
	if err != nil {
		return 0, "", fmt.Errorf("searching for %q: %w", title, err)
	}
	var resp tmdbSearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, "", fmt.Errorf("parsing search results: %w", err)
	}
	if len(resp.Results) == 0 {
		return 0, "", fmt.Errorf("no movies found for %q", title)
	}
	r := resp.Results[0]
	return r.ID, r.DisplayTitle(), nil
}

// searchPersonByName searches TMDb for a person by name and returns the top result.
func searchPersonByName(c *client.Client, name string) (*tmdbSearchResult, error) {
	data, err := c.Get("/search/person", map[string]string{"query": name})
	if err != nil {
		return nil, fmt.Errorf("searching for person %q: %w", name, err)
	}
	var resp tmdbSearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing search results: %w", err)
	}
	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no person found for %q", name)
	}
	return &resp.Results[0], nil
}

// resolveMovieID resolves a string argument to a TMDb movie ID.
// If the argument is numeric, returns it directly. Otherwise searches by title.
func resolveMovieID(c *client.Client, arg string) (int, string, error) {
	if id, err := strconv.Atoi(arg); err == nil {
		return id, "", nil
	}
	return searchMovieByTitle(c, arg)
}

// getMovieDetail fetches full movie details from TMDb.
func getMovieDetail(c *client.Client, movieID int, appendToResponse string) (*tmdbMovieDetail, json.RawMessage, error) {
	path := fmt.Sprintf("/movie/%d", movieID)
	params := map[string]string{}
	if appendToResponse != "" {
		params["append_to_response"] = appendToResponse
	}
	data, err := c.Get(path, params)
	if err != nil {
		return nil, nil, err
	}
	var detail tmdbMovieDetail
	if err := json.Unmarshal(data, &detail); err != nil {
		return nil, data, fmt.Errorf("parsing movie detail: %w", err)
	}
	return &detail, data, nil
}

// genreNames returns a comma-separated string of genre names.
func genreNames(detail *tmdbMovieDetail) string {
	names := make([]string, 0, len(detail.Genres))
	for _, g := range detail.Genres {
		names = append(names, g.Name)
	}
	return strings.Join(names, ", ")
}

// formatMoney formats a number as a dollar amount (e.g. $150,000,000).
func formatMoney(amount int64) string {
	if amount == 0 {
		return "N/A"
	}
	s := fmt.Sprintf("%d", amount)
	// Insert commas
	n := len(s)
	if n <= 3 {
		return "$" + s
	}
	var result strings.Builder
	result.WriteString("$")
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}
