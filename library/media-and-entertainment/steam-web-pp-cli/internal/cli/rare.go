package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newRareCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rare <steamid-or-vanity> <appid>",
		Short: "Show rarest achievements a player has earned",
		Long: `Cross-reference a player's achievements with global percentages
to find their rarest unlocks. Lower percentage = rarer achievement.`,
		Example: `  steam-web-pp-cli rare gabelogannewell 440
  steam-web-pp-cli rare 76561198006409530 730 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUserStats/GetPlayerAchievements/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2")
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

			// Get player achievements
			playerData, err := c.Get("/ISteamUserStats/GetPlayerAchievements/v1", map[string]string{
				"key":     key,
				"steamid": steamID,
				"appid":   args[1],
			})
			if err != nil {
				return classifyAPIError(err)
			}

			var ps struct {
				Playerstats struct {
					Achievements []struct {
						APIName  string `json:"apiname"`
						Achieved int    `json:"achieved"`
						Name     string `json:"name"`
					} `json:"achievements"`
				} `json:"playerstats"`
			}
			if err := json.Unmarshal(playerData, &ps); err != nil {
				return fmt.Errorf("parsing player achievements: %w", err)
			}

			// Get global percentages
			globalData, err := c.Get("/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2", map[string]string{
				"gameid": args[1],
			})
			if err != nil {
				return classifyAPIError(err)
			}

			// Parse global percentages - uses json.RawMessage because the structure is nested
			var globalEnvelope struct {
				Achievementpercentages struct {
					Achievements json.RawMessage `json:"achievements"`
				} `json:"achievementpercentages"`
			}
			globalPct := map[string]float64{}
			if json.Unmarshal(globalData, &globalEnvelope) == nil && globalEnvelope.Achievementpercentages.Achievements != nil {
				var globalAchievements []struct {
					Name    string  `json:"name"`
					Percent float64 `json:"percent"`
				}
				if json.Unmarshal(globalEnvelope.Achievementpercentages.Achievements, &globalAchievements) == nil {
					for _, ga := range globalAchievements {
						globalPct[ga.Name] = ga.Percent
					}
				}
			}

			type rareAchievement struct {
				Name      string  `json:"name"`
				APIName   string  `json:"apiname"`
				GlobalPct float64 `json:"global_percent"`
				Achieved  bool    `json:"achieved"`
			}

			var results []rareAchievement
			for _, a := range ps.Playerstats.Achievements {
				if a.Achieved == 1 {
					pct := globalPct[a.APIName]
					results = append(results, rareAchievement{
						Name:      a.Name,
						APIName:   a.APIName,
						GlobalPct: pct,
						Achieved:  true,
					})
				}
			}

			sort.Slice(results, func(i, j int) bool {
				return results[i].GlobalPct < results[j].GlobalPct
			})

			out, _ := json.Marshal(results)
			return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
		},
	}
	return cmd
}
