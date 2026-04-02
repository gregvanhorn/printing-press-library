package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newBacklogCmd(flags *rootFlags) *cobra.Command {
	var minPlaytime float64

	cmd := &cobra.Command{
		Use:   "backlog <steamid-or-vanity>",
		Short: "Show unplayed or barely-played games (the shame pile)",
		Long: `List games with zero or minimal playtime — the backlog of shame.
Use --min-playtime to adjust the threshold (in hours).`,
		Example: `  steam-web-pp-cli backlog gabelogannewell
  steam-web-pp-cli backlog 76561198006409530 --min-playtime 1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetOwnedGames/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}
			return runBacklog(cmd, c, flags, args[0], minPlaytime)
		},
	}
	cmd.Flags().Float64Var(&minPlaytime, "min-playtime", 0, "Include games with less than this many hours (default: 0 = unplayed only)")
	return cmd
}

func runBacklog(cmd *cobra.Command, c *client.Client, flags *rootFlags, input string, minPlaytime float64) error {
	games, err := fetchOwnedGames(c, input)
	if err != nil {
		return err
	}

	thresholdMinutes := minPlaytime * 60

	type backlogGame struct {
		AppID      int     `json:"appid"`
		Name       string  `json:"name"`
		PlaytimeHr float64 `json:"playtime_hours"`
	}

	var results []backlogGame
	for _, g := range games {
		playtime, _ := g["playtime_forever"].(float64)
		if playtime <= thresholdMinutes {
			appID, _ := g["appid"].(float64)
			name, _ := g["name"].(string)
			results = append(results, backlogGame{
				AppID:      int(appID),
				Name:       name,
				PlaytimeHr: playtime / 60.0,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	out, _ := json.Marshal(results)
	summary := map[string]any{
		"total_owned":  len(games),
		"backlog_size": len(results),
		"backlog_pct":  0.0,
		"games":        json.RawMessage(out),
	}
	if len(games) > 0 {
		summary["backlog_pct"] = float64(len(results)) / float64(len(games)) * 100
	}
	summaryOut, _ := json.Marshal(summary)
	return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(summaryOut), flags)
}
