package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newGroupsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups <steamid-or-vanity>",
		Short: "List a player's Steam groups",
		Long:  "Fetch all Steam groups a player belongs to.",
		Example: `  steam-web-pp-cli groups gabelogannewell
  steam-web-pp-cli groups 76561198006409530 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUser/GetUserGroupList/v1")
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

			data, err := c.Get("/ISteamUser/GetUserGroupList/v1", map[string]string{
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
				Groups json.RawMessage `json:"groups"`
			}
			if json.Unmarshal(resp, &respObj) == nil && respObj.Groups != nil {
				return printOutputWithFlags(cmd.OutOrStdout(), respObj.Groups, flags)
			}
			return printOutputWithFlags(cmd.OutOrStdout(), resp, flags)
		},
	}
	return cmd
}
