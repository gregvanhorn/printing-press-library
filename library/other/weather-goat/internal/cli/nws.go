package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const nwsBaseURL = "https://api.weather.gov"

// openMeteoGet performs a GET to any Open-Meteo subdomain endpoint
// (air-quality, archive, marine, etc.) that isn't on the main forecast base URL.
func openMeteoGet(fullURL string, params map[string]string) (json.RawMessage, error) {
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	q := req.URL.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	req.URL.RawQuery = q.Encode()

	resp, err := nwsHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return json.RawMessage(body), nil
}

var nwsHTTPClient = &http.Client{Timeout: 15 * time.Second}

// nwsGet performs a GET request against the NWS API with the required User-Agent header.
func nwsGet(path string) (json.RawMessage, error) {
	url := nwsBaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating NWS request: %w", err)
	}
	req.Header.Set("User-Agent", "weather-goat-pp-cli/1.0.0 (github.com/mvanhorn/cli-printing-press)")
	req.Header.Set("Accept", "application/geo+json")

	resp, err := nwsHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("NWS request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading NWS response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("NWS API returned HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	return json.RawMessage(body), nil
}

// nwsAlerts fetches active weather alerts for a given lat/lon from the NWS API.
// Returns a slice of alert feature properties.
func nwsAlerts(lat, lon float64) ([]map[string]any, error) {
	path := fmt.Sprintf("/alerts/active?point=%.4f,%.4f", lat, lon)
	data, err := nwsGet(path)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Features []struct {
			Properties map[string]any `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing NWS alerts: %w", err)
	}

	var alerts []map[string]any
	for _, f := range resp.Features {
		alerts = append(alerts, f.Properties)
	}
	return alerts, nil
}

// nwsAlertsByState fetches active weather alerts for a US state code.
func nwsAlertsByState(state string) ([]map[string]any, error) {
	path := fmt.Sprintf("/alerts/active?area=%s", state)
	data, err := nwsGet(path)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Features []struct {
			Properties map[string]any `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing NWS alerts: %w", err)
	}

	var alerts []map[string]any
	for _, f := range resp.Features {
		alerts = append(alerts, f.Properties)
	}
	return alerts, nil
}
