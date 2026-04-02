package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newCompletionistCmd(flags *rootFlags) *cobra.Command {
	var limit int
	var minPct float64

	cmd := &cobra.Command{
		Use:   "completionist <steamid-or-vanity>",
		Short: "Show achievement completion percentage across all games",
		Long: `Calculate achievement completion percentage for each game a player owns.
Fetches achievements per game and ranks by completion rate.
Games that return 403 (private profile) are skipped gracefully.`,
		Example: `  steam-web-pp-cli completionist gabelogannewell
  steam-web-pp-cli completionist 76561198006409530 --limit 20 --min-pct 50`,
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
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUserStats/GetPlayerAchievements/v1 (per game)")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}
			return runCompletionist(cmd, c, flags, args[0], limit, minPct)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Max games to check (0 = all)")
	cmd.Flags().Float64Var(&minPct, "min-pct", 0, "Only show games above this completion percentage")
	return cmd
}

func runCompletionist(cmd *cobra.Command, c *client.Client, flags *rootFlags, input string, limit int, minPct float64) error {
	games, err := fetchOwnedGames(c, input)
	if err != nil {
		return err
	}

	steamID, err := resolveSteamID(c, input)
	if err != nil {
		return err
	}
	key, err := steamAPIKey(c)
	if err != nil {
		return err
	}

	// Sort by playtime descending to check most-played first
	sort.Slice(games, func(i, j int) bool {
		pi, _ := games[i]["playtime_forever"].(float64)
		pj, _ := games[j]["playtime_forever"].(float64)
		return pi > pj
	})

	if limit > 0 && limit < len(games) {
		games = games[:limit]
	}

	type gameCompletion struct {
		AppID      int     `json:"appid"`
		Name       string  `json:"name"`
		Total      int     `json:"total_achievements"`
		Achieved   int     `json:"achieved"`
		Percent    float64 `json:"percent"`
		PlaytimeHr float64 `json:"playtime_hours"`
	}

	var results []gameCompletion
	for _, g := range games {
		appID, _ := g["appid"].(float64)
		name, _ := g["name"].(string)
		playtime, _ := g["playtime_forever"].(float64)

		appIDStr := fmt.Sprintf("%d", int(appID))
		data, err := c.Get("/ISteamUserStats/GetPlayerAchievements/v1", map[string]string{
			"key":     key,
			"steamid": steamID,
			"appid":   appIDStr,
		})
		if err != nil {
			// 403 = private, 400 = no achievements — skip
			continue
		}

		var ps struct {
			Playerstats struct {
				Achievements []struct {
					Achieved int `json:"achieved"`
				} `json:"achievements"`
			} `json:"playerstats"`
		}
		if json.Unmarshal(data, &ps) != nil || len(ps.Playerstats.Achievements) == 0 {
			continue
		}

		total := len(ps.Playerstats.Achievements)
		achieved := 0
		for _, a := range ps.Playerstats.Achievements {
			if a.Achieved == 1 {
				achieved++
			}
		}
		pct := float64(achieved) / float64(total) * 100

		if pct >= minPct {
			results = append(results, gameCompletion{
				AppID:      int(appID),
				Name:       name,
				Total:      total,
				Achieved:   achieved,
				Percent:    pct,
				PlaytimeHr: playtime / 60.0,
			})
		}
	}

	// Sort by completion percentage descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Percent > results[j].Percent
	})

	out, _ := json.Marshal(results)
	return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
}
