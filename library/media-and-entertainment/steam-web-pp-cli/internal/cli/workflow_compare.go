package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/store"
)

func newWorkflowCompareCmd(flags *rootFlags) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "compare <steamid1> <steamid2>",
		Short: "Compare two players: common games, unique games, playtime differences",
		Long: `Compare two players' game libraries from locally synced data. Shows games
they have in common, games unique to each player, and total playtime difference.
Both players must have been synced locally first.`,
		Example: `  # Compare two players
  steam-web-pp-cli workflow compare 76561198012345678 76561198087654321

  # As JSON
  steam-web-pp-cli workflow compare 76561198012345678 76561198087654321 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id1, id2 := args[0], args[1]

			if dbPath == "" {
				dbPath = defaultDBPath("steam-web-pp-cli")
			}
			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}
			defer db.Close()

			items, err := db.List("iplayer-service", 0)
			if err != nil {
				return fmt.Errorf("querying store: %w", err)
			}
			if len(items) == 0 {
				return fmt.Errorf("no game data in local store.\nhint: run 'steam-web-pp-cli sync' to populate local data")
			}

			// Build game maps for each player
			type gameRecord struct {
				AppID    string  `json:"appid"`
				Name     string  `json:"name"`
				Playtime float64 `json:"playtime_forever"`
			}

			games1 := collectGames(items, id1)
			games2 := collectGames(items, id2)

			if len(games1) == 0 {
				return fmt.Errorf("no games found for player %s in local store.\nhint: run 'steam-web-pp-cli sync' to populate local data", id1)
			}
			if len(games2) == 0 {
				return fmt.Errorf("no games found for player %s in local store.\nhint: run 'steam-web-pp-cli sync' to populate local data", id2)
			}

			// Find common and unique games
			type comparedGame struct {
				AppID     string  `json:"appid"`
				Name      string  `json:"name"`
				Playtime1 float64 `json:"playtime_player1"`
				Playtime2 float64 `json:"playtime_player2"`
			}

			var commonGames []comparedGame
			var uniqueTo1 []gameRecord
			var uniqueTo2 []gameRecord

			map2 := make(map[string]gameRecord)
			for _, g := range games2 {
				map2[g.AppID] = g
			}
			map1 := make(map[string]gameRecord)
			for _, g := range games1 {
				map1[g.AppID] = g
			}

			for _, g := range games1 {
				if g2, ok := map2[g.AppID]; ok {
					commonGames = append(commonGames, comparedGame{
						AppID:     g.AppID,
						Name:      g.Name,
						Playtime1: g.Playtime,
						Playtime2: g2.Playtime,
					})
				} else {
					uniqueTo1 = append(uniqueTo1, g)
				}
			}
			for _, g := range games2 {
				if _, ok := map1[g.AppID]; !ok {
					uniqueTo2 = append(uniqueTo2, g)
				}
			}

			// Sort common games by combined playtime descending
			sort.Slice(commonGames, func(i, j int) bool {
				return (commonGames[i].Playtime1 + commonGames[i].Playtime2) >
					(commonGames[j].Playtime1 + commonGames[j].Playtime2)
			})

			// Calculate total playtime
			var total1, total2 float64
			for _, g := range games1 {
				total1 += g.Playtime
			}
			for _, g := range games2 {
				total2 += g.Playtime
			}

			result := map[string]any{
				"player1":                 id1,
				"player2":                 id2,
				"player1_total_games":     len(games1),
				"player2_total_games":     len(games2),
				"common_games_count":      len(commonGames),
				"unique_to_player1":       len(uniqueTo1),
				"unique_to_player2":       len(uniqueTo2),
				"player1_total_hours":     fmt.Sprintf("%.1f", total1/60.0),
				"player2_total_hours":     fmt.Sprintf("%.1f", total2/60.0),
				"common_games":            commonGames,
				"unique_to_player1_games": uniqueTo1,
				"unique_to_player2_games": uniqueTo2,
			}

			prov := localProvenance(db, "iplayer-service", "transcendence_command")
			printProvenance(cmd, len(commonGames), prov)

			data, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("marshaling comparison: %w", err)
			}

			return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(data), flags)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/steam-web-pp-cli/data.db)")

	return cmd
}

// collectGames extracts game records for a given steam ID from store items.
func collectGames(items []json.RawMessage, steamID string) []struct {
	AppID    string  `json:"appid"`
	Name     string  `json:"name"`
	Playtime float64 `json:"playtime_forever"`
} {
	type gameRecord struct {
		AppID    string  `json:"appid"`
		Name     string  `json:"name"`
		Playtime float64 `json:"playtime_forever"`
	}

	var games []struct {
		AppID    string  `json:"appid"`
		Name     string  `json:"name"`
		Playtime float64 `json:"playtime_forever"`
	}
	for _, raw := range items {
		var obj map[string]any
		if json.Unmarshal(raw, &obj) != nil {
			continue
		}
		if _, hasAppID := obj["appid"]; !hasAppID {
			continue
		}
		// If the record has a steamid, verify it matches
		if sid, ok := obj["steamid"]; ok {
			if fmt.Sprintf("%v", sid) != steamID {
				continue
			}
		}
		g := struct {
			AppID    string  `json:"appid"`
			Name     string  `json:"name"`
			Playtime float64 `json:"playtime_forever"`
		}{
			AppID: fmt.Sprintf("%v", obj["appid"]),
		}
		if v, ok := obj["name"].(string); ok {
			g.Name = v
		}
		if v, ok := obj["playtime_forever"].(float64); ok {
			g.Playtime = v
		}
		games = append(games, g)
	}
	return games
}
