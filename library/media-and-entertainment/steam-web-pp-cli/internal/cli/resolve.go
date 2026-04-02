package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newResolveCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resolve <vanity-name>",
		Short:   "Resolve a vanity URL to a SteamID64",
		Long:    "Convert a Steam vanity URL name to its corresponding SteamID64.",
		Example: `  steam-web-pp-cli resolve gabelogannewell`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUser/ResolveVanityURL/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}

			steamID, err := resolveSteamID(c, args[0])
			if err != nil {
				return err
			}

			result := map[string]string{
				"input":   args[0],
				"steamid": steamID,
			}
			out, _ := json.Marshal(result)
			return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
		},
	}
	return cmd
}
