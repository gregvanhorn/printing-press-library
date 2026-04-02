package cli

import (
	"encoding/json"
	"fmt"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newOverlapCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overlap <steamid-or-vanity>...",
		Short: "Find games owned by all given players",
		Long: `Find games that every listed player owns — great for picking a game
to play together. Accepts 2 or more SteamID64s or vanity names.`,
		Example: `  steam-web-pp-cli overlap gabelogannewell 76561198006409530
  steam-web-pp-cli overlap player1 player2 player3 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintf(cmd.ErrOrStderr(), "GET /IPlayerService/GetOwnedGames/v1 (x%d)\n", len(args))
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}
			return runOverlap(cmd, c, flags, args)
		},
	}
	return cmd
}

func runOverlap(cmd *cobra.Command, c *client.Client, flags *rootFlags, inputs []string) error {
	// Fetch all players' games
	var allGamesLists [][]map[string]any
	for _, input := range inputs {
		games, err := fetchOwnedGames(c, input)
		if err != nil {
			return fmt.Errorf("player %q: %w", input, err)
		}
		allGamesLists = append(allGamesLists, games)
	}

	if len(allGamesLists) == 0 {
		out, _ := json.Marshal([]any{})
		return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
	}

	// Start with the first player's games as the candidate set
	overlap := map[int]map[string]any{}
	for _, g := range allGamesLists[0] {
		appID, _ := g["appid"].(float64)
		overlap[int(appID)] = g
	}

	// Intersect with each subsequent player
	for i := 1; i < len(allGamesLists); i++ {
		playerApps := map[int]bool{}
		for _, g := range allGamesLists[i] {
			appID, _ := g["appid"].(float64)
			playerApps[int(appID)] = true
		}
		for id := range overlap {
			if !playerApps[id] {
				delete(overlap, id)
			}
		}
	}

	type overlapGame struct {
		AppID int    `json:"appid"`
		Name  string `json:"name"`
	}

	var results []overlapGame
	for _, g := range overlap {
		appID, _ := g["appid"].(float64)
		name, _ := g["name"].(string)
		results = append(results, overlapGame{
			AppID: int(appID),
			Name:  name,
		})
	}

	result := map[string]any{
		"players":       len(inputs),
		"overlap_count": len(results),
		"games":         results,
	}
	out, _ := json.Marshal(result)
	return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
}
