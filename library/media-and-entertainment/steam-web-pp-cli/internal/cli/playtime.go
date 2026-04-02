package cli

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newPlaytimeCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "playtime <steamid-or-vanity>",
		Short: "Analyze playtime distribution across library",
		Long: `Generate playtime statistics for a player's game library:
total hours, median, mean, top games, and distribution buckets.`,
		Example: `  steam-web-pp-cli playtime gabelogannewell
  steam-web-pp-cli playtime 76561198006409530 --json`,
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
			return runPlaytime(cmd, c, flags, args[0])
		},
	}
	return cmd
}

func runPlaytime(cmd *cobra.Command, c *client.Client, flags *rootFlags, input string) error {
	games, err := fetchOwnedGames(c, input)
	if err != nil {
		return err
	}

	if len(games) == 0 {
		result := map[string]any{"total_games": 0, "message": "no games found"}
		out, _ := json.Marshal(result)
		return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
	}

	// Extract playtimes in hours
	type gameTime struct {
		Name  string  `json:"name"`
		Hours float64 `json:"hours"`
	}

	var playtimes []float64
	var gameTimes []gameTime
	var totalMinutes float64

	for _, g := range games {
		pt, _ := g["playtime_forever"].(float64)
		name, _ := g["name"].(string)
		hours := pt / 60.0
		playtimes = append(playtimes, hours)
		totalMinutes += pt
		gameTimes = append(gameTimes, gameTime{Name: name, Hours: math.Round(hours*10) / 10})
	}

	sort.Float64s(playtimes)
	sort.Slice(gameTimes, func(i, j int) bool {
		return gameTimes[i].Hours > gameTimes[j].Hours
	})

	totalHours := totalMinutes / 60.0
	mean := totalHours / float64(len(games))
	var median float64
	n := len(playtimes)
	if n%2 == 0 {
		median = (playtimes[n/2-1] + playtimes[n/2]) / 2
	} else {
		median = playtimes[n/2]
	}

	// Distribution buckets
	buckets := map[string]int{
		"0h (unplayed)": 0,
		"0-1h":          0,
		"1-10h":         0,
		"10-100h":       0,
		"100-1000h":     0,
		"1000h+":        0,
	}
	for _, h := range playtimes {
		switch {
		case h == 0:
			buckets["0h (unplayed)"]++
		case h < 1:
			buckets["0-1h"]++
		case h < 10:
			buckets["1-10h"]++
		case h < 100:
			buckets["10-100h"]++
		case h < 1000:
			buckets["100-1000h"]++
		default:
			buckets["1000h+"]++
		}
	}

	// Top 10 games
	topN := 10
	if topN > len(gameTimes) {
		topN = len(gameTimes)
	}

	result := map[string]any{
		"total_games":    len(games),
		"total_hours":    math.Round(totalHours*10) / 10,
		"mean_hours":     math.Round(mean*10) / 10,
		"median_hours":   math.Round(median*10) / 10,
		"top_games":      gameTimes[:topN],
		"distribution":   buckets,
		"unplayed_count": buckets["0h (unplayed)"],
		"unplayed_pct":   math.Round(float64(buckets["0h (unplayed)"])/float64(len(games))*1000) / 10,
	}
	out, _ := json.Marshal(result)
	return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
}
