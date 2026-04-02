package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newSchemaCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema <appid>",
		Short: "Get the achievement/stat schema for a game",
		Long:  "Fetch the full schema for a game including available achievements and stats definitions.",
		Example: `  steam-web-pp-cli schema 440
  steam-web-pp-cli schema 730 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUserStats/GetSchemaForGame/v2")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}

			key, err := steamAPIKey(c)
			if err != nil {
				return err
			}

			data, err := c.Get("/ISteamUserStats/GetSchemaForGame/v2", map[string]string{
				"key":   key,
				"appid": args[0],
			})
			if err != nil {
				return classifyAPIError(err)
			}

			// Extract game schema
			var gs struct {
				Game struct {
					GameName           string          `json:"gameName"`
					GameVersion        string          `json:"gameVersion"`
					AvailableGameStats json.RawMessage `json:"availableGameStats"`
				} `json:"game"`
			}
			if json.Unmarshal(data, &gs) == nil && gs.Game.GameName != "" {
				result := map[string]any{
					"game_name":    gs.Game.GameName,
					"game_version": gs.Game.GameVersion,
				}
				if gs.Game.AvailableGameStats != nil {
					var stats struct {
						Achievements []any `json:"achievements"`
						Stats        []any `json:"stats"`
					}
					if json.Unmarshal(gs.Game.AvailableGameStats, &stats) == nil {
						result["achievement_count"] = len(stats.Achievements)
						result["stat_count"] = len(stats.Stats)
						result["achievements"] = stats.Achievements
						result["stats"] = stats.Stats
					}
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
