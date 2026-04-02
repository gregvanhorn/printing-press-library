package cli

import (
	"encoding/json"
	"fmt"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newCompareCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare <steamid-or-vanity-1> <steamid-or-vanity-2>",
		Short: "Compare game libraries between two players",
		Long: `Compare the game libraries of two Steam players.
Shows games unique to each player and games they share.`,
		Example: `  steam-web-pp-cli compare gabelogannewell 76561198006409530`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetOwnedGames/v1 (x2)")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}
			return runCompare(cmd, c, flags, args[0], args[1])
		},
	}
	return cmd
}

func runCompare(cmd *cobra.Command, c *client.Client, flags *rootFlags, input1, input2 string) error {
	games1, err := fetchOwnedGames(c, input1)
	if err != nil {
		return fmt.Errorf("player 1: %w", err)
	}
	games2, err := fetchOwnedGames(c, input2)
	if err != nil {
		return fmt.Errorf("player 2: %w", err)
	}

	// Build lookup maps by appid
	set1 := map[int]map[string]any{}
	for _, g := range games1 {
		appID, _ := g["appid"].(float64)
		set1[int(appID)] = g
	}
	set2 := map[int]map[string]any{}
	for _, g := range games2 {
		appID, _ := g["appid"].(float64)
		set2[int(appID)] = g
	}

	type comparedGame struct {
		AppID     int     `json:"appid"`
		Name      string  `json:"name"`
		Playtime1 float64 `json:"playtime_hours_player1"`
		Playtime2 float64 `json:"playtime_hours_player2"`
		SharedBy  string  `json:"owned_by"`
	}

	var shared, only1, only2 []comparedGame

	for id, g := range set1 {
		name, _ := g["name"].(string)
		pt1, _ := g["playtime_forever"].(float64)
		if g2, ok := set2[id]; ok {
			pt2, _ := g2["playtime_forever"].(float64)
			shared = append(shared, comparedGame{
				AppID: id, Name: name,
				Playtime1: pt1 / 60, Playtime2: pt2 / 60,
				SharedBy: "both",
			})
		} else {
			only1 = append(only1, comparedGame{
				AppID: id, Name: name,
				Playtime1: pt1 / 60,
				SharedBy:  "player1",
			})
		}
	}
	for id, g := range set2 {
		if _, ok := set1[id]; !ok {
			name, _ := g["name"].(string)
			pt2, _ := g["playtime_forever"].(float64)
			only2 = append(only2, comparedGame{
				AppID: id, Name: name,
				Playtime2: pt2 / 60,
				SharedBy:  "player2",
			})
		}
	}

	result := map[string]any{
		"player1_total": len(games1),
		"player2_total": len(games2),
		"shared_count":  len(shared),
		"shared":        shared,
		"only_player1":  only1,
		"only_player2":  only2,
	}
	out, _ := json.Marshal(result)
	return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
}
