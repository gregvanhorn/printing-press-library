package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newPlayersCountCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "players-count <appid>",
		Aliases: []string{"players"},
		Short:   "Get current player count for a game",
		Long:    "Fetch the number of players currently in-game for a given app ID.",
		Example: `  steam-web-pp-cli players-count 730
  steam-web-pp-cli players 440`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUserStats/GetNumberOfCurrentPlayers/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}

			data, err := c.Get("/ISteamUserStats/GetNumberOfCurrentPlayers/v1", map[string]string{
				"appid": args[0],
			})
			if err != nil {
				return classifyAPIError(err)
			}
			resp, err := extractResponse(data)
			if err != nil {
				return err
			}
			// Return the full response (contains player_count and result)
			var obj map[string]any
			if json.Unmarshal(resp, &obj) == nil {
				out, _ := json.Marshal(obj)
				return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
			}
			return printOutputWithFlags(cmd.OutOrStdout(), resp, flags)
		},
	}
	return cmd
}
