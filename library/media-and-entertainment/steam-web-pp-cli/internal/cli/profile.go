package cli

import (
	"encoding/json"
	"fmt"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newProfileCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile <steamid-or-vanity>",
		Short: "Full player profile with level, badges, games, and recent activity",
		Long: `Aggregate a player's profile data in one call: summary, level,
badge count, game count, and recent games. Handles private profiles gracefully.`,
		Example: `  steam-web-pp-cli profile gabelogannewell
  steam-web-pp-cli profile 76561198006409530 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUser/GetPlayerSummaries/v2")
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetSteamLevel/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetOwnedGames/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetBadges/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetRecentlyPlayedGames/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}
			return runProfile(cmd, c, flags, args[0])
		},
	}
	return cmd
}

func runProfile(cmd *cobra.Command, c *client.Client, flags *rootFlags, input string) error {
	steamID, err := resolveSteamID(c, input)
	if err != nil {
		return err
	}
	key, err := steamAPIKey(c)
	if err != nil {
		return err
	}

	profile := map[string]any{
		"steamid": steamID,
	}

	// Player summary
	summaryData, err := c.Get("/ISteamUser/GetPlayerSummaries/v2", map[string]string{
		"key":      key,
		"steamids": steamID,
	})
	if err == nil {
		resp, _ := extractResponse(summaryData)
		var sObj struct {
			Players []map[string]any `json:"players"`
		}
		if json.Unmarshal(resp, &sObj) == nil && len(sObj.Players) > 0 {
			p := sObj.Players[0]
			profile["personaname"] = p["personaname"]
			profile["profileurl"] = p["profileurl"]
			profile["avatar"] = p["avatarfull"]
			profile["personastate"] = p["personastate"]
			profile["timecreated"] = p["timecreated"]
			profile["communityvisibilitystate"] = p["communityvisibilitystate"]

			// Check if private
			vis, _ := p["communityvisibilitystate"].(float64)
			if vis != 3 {
				profile["private"] = true
				out, _ := json.Marshal(profile)
				return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
			}
		}
	}

	// Steam level
	levelData, err := c.Get("/IPlayerService/GetSteamLevel/v1", map[string]string{
		"key":     key,
		"steamid": steamID,
	})
	if err == nil {
		resp, _ := extractResponse(levelData)
		var levelObj struct {
			PlayerLevel int `json:"player_level"`
		}
		if json.Unmarshal(resp, &levelObj) == nil {
			profile["steam_level"] = levelObj.PlayerLevel
		}
	}

	// Game count
	gamesData, err := c.Get("/IPlayerService/GetOwnedGames/v1", map[string]string{
		"key":                       key,
		"steamid":                   steamID,
		"include_played_free_games": "true",
	})
	if err == nil {
		resp, _ := extractResponse(gamesData)
		var gObj struct {
			GameCount int `json:"game_count"`
		}
		if json.Unmarshal(resp, &gObj) == nil {
			profile["game_count"] = gObj.GameCount
		}
	}

	// Badges
	badgesData, err := c.Get("/IPlayerService/GetBadges/v1", map[string]string{
		"key":     key,
		"steamid": steamID,
	})
	if err == nil {
		resp, _ := extractResponse(badgesData)
		var bObj struct {
			Badges      json.RawMessage `json:"badges"`
			PlayerXP    int             `json:"player_xp"`
			PlayerLevel int             `json:"player_level"`
			PlayerXPLow int             `json:"player_xp_needed_to_level_up"`
		}
		if json.Unmarshal(resp, &bObj) == nil {
			var badges []any
			if json.Unmarshal(bObj.Badges, &badges) == nil {
				profile["badge_count"] = len(badges)
			}
			profile["player_xp"] = bObj.PlayerXP
		}
	}

	// Recent games
	recentData, err := c.Get("/IPlayerService/GetRecentlyPlayedGames/v1", map[string]string{
		"key":     key,
		"steamid": steamID,
		"count":   "5",
	})
	if err == nil {
		resp, _ := extractResponse(recentData)
		var rObj struct {
			Games []map[string]any `json:"games"`
		}
		if json.Unmarshal(resp, &rObj) == nil {
			var recentNames []string
			for _, g := range rObj.Games {
				if n, ok := g["name"].(string); ok {
					recentNames = append(recentNames, n)
				}
			}
			profile["recent_games"] = recentNames
		}
	}

	out, _ := json.Marshal(profile)
	return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
}
