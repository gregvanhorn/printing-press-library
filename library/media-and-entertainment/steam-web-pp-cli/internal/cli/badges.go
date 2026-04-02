package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newBadgesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "badges <steamid-or-vanity>",
		Short: "List a player's badges with XP",
		Long:  "Fetch all badges owned by a player, including level and XP data.",
		Example: `  steam-web-pp-cli badges 76561198006409530
  steam-web-pp-cli badges gabelogannewell --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetBadges/v1")
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

			data, err := c.Get("/IPlayerService/GetBadges/v1", map[string]string{
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
			var respObj struct {
				Badges json.RawMessage `json:"badges"`
			}
			if json.Unmarshal(resp, &respObj) == nil && respObj.Badges != nil {
				return printOutputWithFlags(cmd.OutOrStdout(), respObj.Badges, flags)
			}
			return printOutputWithFlags(cmd.OutOrStdout(), resp, flags)
		},
	}
	return cmd
}
