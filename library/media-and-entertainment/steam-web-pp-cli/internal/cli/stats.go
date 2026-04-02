package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newStatsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats <steamid-or-vanity> <appid>",
		Short: "Get a player's in-game stats for a specific game",
		Long:  "Fetch detailed player statistics for a game (kills, deaths, wins, etc.).",
		Example: `  steam-web-pp-cli stats gabelogannewell 730
  steam-web-pp-cli stats 76561198006409530 440 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUserStats/GetUserStatsForGame/v2")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}

			steamID, err := resolveSteamID(c, args[0])
			if err != nil {
				return err
			}
			key, err := steamAPIKey(c)
			if err != nil {
				return err
			}

			data, err := c.Get("/ISteamUserStats/GetUserStatsForGame/v2", map[string]string{
				"key":     key,
				"steamid": steamID,
				"appid":   args[1],
			})
			if err != nil {
				return classifyAPIError(err)
			}

			// Extract playerstats.stats
			var ps struct {
				Playerstats struct {
					SteamID  string          `json:"steamID"`
					GameName string          `json:"gameName"`
					Stats    json.RawMessage `json:"stats"`
				} `json:"playerstats"`
			}
			if json.Unmarshal(data, &ps) == nil && ps.Playerstats.Stats != nil {
				result := map[string]any{
					"steamid":   ps.Playerstats.SteamID,
					"game_name": ps.Playerstats.GameName,
					"stats":     json.RawMessage(ps.Playerstats.Stats),
				}
				out, _ := json.Marshal(result)
				return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
			}
			resp, _ := extractResponse(data)
			return printOutputWithFlags(cmd.OutOrStdout(), resp, flags)
		},
	}
	return cmd
}
