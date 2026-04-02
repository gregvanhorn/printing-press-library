package cli

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"
)

// steamAPIKey returns the API key from config, or an error if unset.
func steamAPIKey(c *client.Client) (string, error) {
	if c.Config != nil && c.Config.APIKey != "" {
		return c.Config.APIKey, nil
	}
	return "", fmt.Errorf("Steam API key not configured.\nhint: set STEAM_API_KEY env var or add api_key to config.toml.\n      Run 'steam-web-pp-cli doctor' to check auth status")
}

// steamID64Re matches a 17-digit Steam ID64.
var steamID64Re = regexp.MustCompile(`^[0-9]{17}$`)

// isSteamID64 returns true if s looks like a 17-digit SteamID64.
func isSteamID64(s string) bool {
	return steamID64Re.MatchString(s)
}

// resolveSteamID resolves a vanity URL or SteamID64 to a SteamID64.
// If the input is already a 17-digit number, it is returned as-is.
// Otherwise it calls ResolveVanityURL.
func resolveSteamID(c *client.Client, input string) (string, error) {
	if isSteamID64(input) {
		return input, nil
	}
	key, err := steamAPIKey(c)
	if err != nil {
		return "", err
	}
	params := map[string]string{
		"key":       key,
		"vanityurl": input,
	}
	data, err := c.Get("/ISteamUser/ResolveVanityURL/v1", params)
	if err != nil {
		return "", fmt.Errorf("resolving vanity URL %q: %w", input, err)
	}
	resolved, err := extractResponse(data)
	if err != nil {
		return "", err
	}
	var result struct {
		SteamID string `json:"steamid"`
		Success int    `json:"success"`
	}
	if err := json.Unmarshal(resolved, &result); err != nil {
		return "", fmt.Errorf("parsing resolve response: %w", err)
	}
	if result.Success != 1 || result.SteamID == "" {
		return "", fmt.Errorf("could not resolve %q to a Steam ID (success=%d)", input, result.Success)
	}
	return result.SteamID, nil
}

// extractResponse unwraps the standard Steam API {"response":{...}} envelope.
// Returns the inner object, or the original data if no envelope is found.
func extractResponse(data json.RawMessage) (json.RawMessage, error) {
	var envelope struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.Response != nil {
		return envelope.Response, nil
	}
	return data, nil
}

// extractPlayers unwraps the Steam bans format {"players":[...]}.
// Returns the players array, or the original data if no players key.
func extractPlayers(data json.RawMessage) (json.RawMessage, error) {
	var envelope struct {
		Players json.RawMessage `json:"players"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.Players != nil {
		return envelope.Players, nil
	}
	return data, nil
}
