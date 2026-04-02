package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newLevelCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "level <steamid-or-vanity>",
		Short: "Get a player's Steam level",
		Long:  "Fetch the Steam level for a player.",
		Example: `  steam-web-pp-cli level gabelogannewell
  steam-web-pp-cli level 76561198006409530`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetSteamLevel/v1")
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

			data, err := c.Get("/IPlayerService/GetSteamLevel/v1", map[string]string{
				"key":     key,
				"steamid": steamID,
			})
			if err != nil {
				return classifyAPIError(err)
			}
			resp, err := extractResponse(data)
			if err != nil {
				return err
			}
			var respObj map[string]any
			if json.Unmarshal(resp, &respObj) == nil {
				out, _ := json.Marshal(respObj)
				return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
			}
			return printOutputWithFlags(cmd.OutOrStdout(), resp, flags)
		},
	}
	return cmd
}
