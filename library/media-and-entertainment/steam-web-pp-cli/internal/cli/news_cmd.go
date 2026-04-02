package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newNewsCmdWrapper(flags *rootFlags) *cobra.Command {
	var count int

	cmd := &cobra.Command{
		Use:   "news <appid>",
		Short: "Get recent news for a game",
		Long:  "Fetch recent news articles for a game by app ID.",
		Example: `  steam-web-pp-cli news 440
  steam-web-pp-cli news 730 --count 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamNews/GetNewsForApp/v2")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}

			params := map[string]string{
				"appid": args[0],
			}
			if count > 0 {
				params["count"] = fmt.Sprintf("%d", count)
			}
			data, err := c.Get("/ISteamNews/GetNewsForApp/v2", params)
			if err != nil {
				return classifyAPIError(err)
			}
			// Extract appnews.newsitems
			var envelope struct {
				Appnews struct {
					Newsitems json.RawMessage `json:"newsitems"`
				} `json:"appnews"`
			}
			if json.Unmarshal(data, &envelope) == nil && envelope.Appnews.Newsitems != nil {
				return printOutputWithFlags(cmd.OutOrStdout(), envelope.Appnews.Newsitems, flags)
			}
			return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
		},
	}
	cmd.Flags().IntVar(&count, "count", 10, "Number of news items to return")
	return cmd
}
