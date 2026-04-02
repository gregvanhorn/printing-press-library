package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newBansCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bans <steamid-or-vanity>...",
		Short: "Check VAC and community bans for players",
		Long:  "Fetch ban status for one or more players. Accepts SteamID64s or vanity names.",
		Example: `  steam-web-pp-cli bans 76561198006409530
  steam-web-pp-cli bans gabelogannewell
  steam-web-pp-cli bans 76561198006409530 76561197960287930`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUser/GetPlayerBans/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}

			key, err := steamAPIKey(c)
			if err != nil {
				return err
			}

			// Resolve all inputs to SteamID64s
			var ids []string
			for _, arg := range args {
				sid, err := resolveSteamID(c, arg)
				if err != nil {
					return err
				}
				ids = append(ids, sid)
			}

			var csvIDs string
			for i, id := range ids {
				if i > 0 {
					csvIDs += ","
				}
				csvIDs += id
			}

			data, err := c.Get("/ISteamUser/GetPlayerBans/v1", map[string]string{
				"key":      key,
				"steamids": csvIDs,
			})
			if err != nil {
				return classifyAPIError(err)
			}
			// Bans uses {"players":[]} not {"response":{}}
			players, err := extractPlayers(data)
			if err != nil {
				return err
			}
			// If single player, unwrap array
			if len(ids) == 1 {
				var arr []json.RawMessage
				if json.Unmarshal(players, &arr) == nil && len(arr) == 1 {
					return printOutputWithFlags(cmd.OutOrStdout(), arr[0], flags)
				}
			}
			return printOutputWithFlags(cmd.OutOrStdout(), players, flags)
		},
	}
	return cmd
}
