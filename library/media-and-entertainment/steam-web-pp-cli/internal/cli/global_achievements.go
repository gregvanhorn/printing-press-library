package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newGlobalAchievementsCmd(flags *rootFlags) *cobra.Command {
	var rare bool
	var limit int

	cmd := &cobra.Command{
		Use:   "global-achievements <appid>",
		Short: "Show global achievement percentages for a game",
		Long: `Fetch global achievement unlock percentages for a game.
Use --rare to sort rarest first, --limit to cap results.`,
		Example: `  steam-web-pp-cli global-achievements 730
  steam-web-pp-cli global-achievements 440 --rare --limit 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}

			data, err := c.Get("/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2", map[string]string{
				"gameid": args[0],
			})
			if err != nil {
				return classifyAPIError(err)
			}

			// Parse - uses json.RawMessage because structure is nested
			var envelope struct {
				Achievementpercentages struct {
					Achievements json.RawMessage `json:"achievements"`
				} `json:"achievementpercentages"`
			}
			if json.Unmarshal(data, &envelope) != nil || envelope.Achievementpercentages.Achievements == nil {
				return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
			}

			var achievements []struct {
				Name    string  `json:"name"`
				Percent float64 `json:"percent"`
			}
			if err := json.Unmarshal(envelope.Achievementpercentages.Achievements, &achievements); err != nil {
				return printOutputWithFlags(cmd.OutOrStdout(), envelope.Achievementpercentages.Achievements, flags)
			}

			if rare {
				sort.Slice(achievements, func(i, j int) bool {
					return achievements[i].Percent < achievements[j].Percent
				})
			}

			if limit > 0 && limit < len(achievements) {
				achievements = achievements[:limit]
			}

			out, _ := json.Marshal(achievements)
			return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
		},
	}
	cmd.Flags().BoolVar(&rare, "rare", false, "Sort by rarest first")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of results (0 = all)")
	return cmd
}
