package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newRecentCmd(flags *rootFlags) *cobra.Command {
	var count int

	cmd := &cobra.Command{
		Use:   "recent <steamid-or-vanity>",
		Short: "List recently played games",
		Long:  "Fetch a player's recently played games with playtime data.",
		Example: `  steam-web-pp-cli recent 76561198006409530
  steam-web-pp-cli recent gabelogannewell --count 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetRecentlyPlayedGames/v1")
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

			params := map[string]string{
				"key":     key,
				"steamid": steamID,
			}
			if count > 0 {
				params["count"] = fmt.Sprintf("%d", count)
			}
			data, err := c.Get("/IPlayerService/GetRecentlyPlayedGames/v1", params)
			if err != nil {
				return classifyAPIError(err)
			}
			resp, err := extractResponse(data)
			if err != nil {
				return err
			}
			var respObj struct {
				Games json.RawMessage `json:"games"`
			}
			if json.Unmarshal(resp, &respObj) == nil && respObj.Games != nil {
				return printOutputWithFlags(cmd.OutOrStdout(), respObj.Games, flags)
			}
			return printOutputWithFlags(cmd.OutOrStdout(), resp, flags)
		},
	}
	cmd.Flags().IntVar(&count, "count", 0, "Number of games to return (0 = all)")
	return cmd
}
