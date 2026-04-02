package cli

import (
	"encoding/json"
	"fmt"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newPlayerCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "player <steamid-or-vanity>",
		Short: "Get player profile summary",
		Long:  "Fetch a Steam player's profile summary by SteamID64 or vanity name.",
		Example: `  steam-web-pp-cli player 76561198006409530
  steam-web-pp-cli player gabelogannewell`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUser/GetPlayerSummaries/v2 ?steamids=<resolved>")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}
			return runPlayer(cmd, c, flags, args[0])
		},
	}
	return cmd
}

func runPlayer(cmd *cobra.Command, c *client.Client, flags *rootFlags, input string) error {
	steamID, err := resolveSteamID(c, input)
	if err != nil {
		return err
	}
	key, err := steamAPIKey(c)
	if err != nil {
		return err
	}
	data, err := c.Get("/ISteamUser/GetPlayerSummaries/v2", map[string]string{
		"key":      key,
		"steamids": steamID,
	})
	if err != nil {
		return classifyAPIError(err)
	}
	resp, err := extractResponse(data)
	if err != nil {
		return err
	}
	// Extract players array from response
	var respObj struct {
		Players json.RawMessage `json:"players"`
	}
	if json.Unmarshal(resp, &respObj) == nil && respObj.Players != nil {
		var players []json.RawMessage
		if json.Unmarshal(respObj.Players, &players) == nil && len(players) == 1 {
			return printOutputWithFlags(cmd.OutOrStdout(), players[0], flags)
		}
		return printOutputWithFlags(cmd.OutOrStdout(), respObj.Players, flags)
	}
	return printOutputWithFlags(cmd.OutOrStdout(), resp, flags)
}
