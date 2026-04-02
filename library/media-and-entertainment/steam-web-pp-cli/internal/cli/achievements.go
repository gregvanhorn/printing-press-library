package cli

import (
	"encoding/json"
	"fmt"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newAchievementsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "achievements <steamid-or-vanity> <appid>",
		Short: "List a player's achievements for a game",
		Long:  "Fetch a player's achievement list for a specific game by app ID.",
		Example: `  steam-web-pp-cli achievements 76561198006409530 440
  steam-web-pp-cli achievements gabelogannewell 730`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUserStats/GetPlayerAchievements/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}
			return runAchievements(cmd, c, flags, args[0], args[1])
		},
	}
	return cmd
}

func runAchievements(cmd *cobra.Command, c *client.Client, flags *rootFlags, input, appID string) error {
	steamID, err := resolveSteamID(c, input)
	if err != nil {
		return err
	}
	key, err := steamAPIKey(c)
	if err != nil {
		return err
	}
	data, err := c.Get("/ISteamUserStats/GetPlayerAchievements/v1", map[string]string{
		"key":     key,
		"steamid": steamID,
		"appid":   appID,
	})
	if err != nil {
		return classifyAPIError(err)
	}
	resp, err := extractResponse(data)
	if err != nil {
		return err
	}
	// Extract playerstats.achievements
	var ps struct {
		Playerstats struct {
			Achievements json.RawMessage `json:"achievements"`
		} `json:"playerstats"`
	}
	if json.Unmarshal(data, &ps) == nil && ps.Playerstats.Achievements != nil {
		return printOutputWithFlags(cmd.OutOrStdout(), ps.Playerstats.Achievements, flags)
	}
	return printOutputWithFlags(cmd.OutOrStdout(), resp, flags)
}
