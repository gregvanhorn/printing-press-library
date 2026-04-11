package omdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://www.omdbapi.com/"

// Rating represents a single rating source (e.g. Rotten Tomatoes, Metacritic).
type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

// Result holds the full OMDb API response for a single title.
type Result struct {
	Title      string   `json:"Title"`
	Year       string   `json:"Year"`
	Rated      string   `json:"Rated"`
	Released   string   `json:"Released"`
	Runtime    string   `json:"Runtime"`
	Genre      string   `json:"Genre"`
	Director   string   `json:"Director"`
	Writer     string   `json:"Writer"`
	Actors     string   `json:"Actors"`
	Plot       string   `json:"Plot"`
	Language   string   `json:"Language"`
	Country    string   `json:"Country"`
	Awards     string   `json:"Awards"`
	Poster     string   `json:"Poster"`
	Ratings    []Rating `json:"Ratings"`
	Metascore  string   `json:"Metascore"`
	ImdbRating string   `json:"imdbRating"`
	ImdbVotes  string   `json:"imdbVotes"`
	BoxOffice  string   `json:"BoxOffice"`
	Response   string   `json:"Response"`
	Error      string   `json:"Error"`
}

// Fetch retrieves OMDb data for the given IMDb ID. Returns nil without error
// if apiKey is empty (graceful degradation when OMDB_API_KEY is not set).
func Fetch(imdbID string, apiKey string) (*Result, error) {
	if apiKey == "" {
		return nil, nil
	}
	if imdbID == "" {
		return nil, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s?i=%s&apikey=%s&plot=full", baseURL, imdbID, apiKey)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("omdb request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading omdb response: %w", err)
	}

	var result Result
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing omdb response: %w", err)
	}

	if result.Response == "False" {
		// OMDb returned an error (e.g. "Movie not found!")
		return nil, nil
	}

	return &result, nil
}

// RatingBySource returns the value for a specific rating source, or empty string if not found.
func (r *Result) RatingBySource(source string) string {
	for _, rating := range r.Ratings {
		if rating.Source == source {
			return rating.Value
		}
	}
	return ""
}
