package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/store"
)

func newWorkflowGamesCmd(flags *rootFlags) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "games <steamid>",
		Short: "List owned games sorted by playtime (descending)",
		Long: `Show all games owned by a player, sorted by total playtime in descending order.
Displays game name, playtime in hours, and last played date. Reads from local
store after sync.`,
		Example: `  # List games for a player
  steam-web-pp-cli workflow games 76561198012345678

  # As JSON
  steam-web-pp-cli workflow games 76561198012345678 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			steamID := args[0]

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

			// Collect games for this player
			type gameInfo struct {
				AppID           any     `json:"appid"`
				Name            string  `json:"name"`
				PlaytimeMinutes float64 `json:"playtime_forever_minutes"`
				PlaytimeHours   string  `json:"playtime_hours"`
				LastPlayed      any     `json:"rtime_last_played,omitempty"`
			}

			var games []gameInfo
			for _, raw := range items {
				var obj map[string]any
				if json.Unmarshal(raw, &obj) != nil {
					continue
				}
				// Filter by steam ID if present, and must have an appid (game record)
				if _, hasAppID := obj["appid"]; !hasAppID {
					continue
				}
				// If the record has a steamid, check it matches
				if sid, ok := obj["steamid"]; ok {
					if fmt.Sprintf("%v", sid) != steamID {
						continue
					}
				}

				g := gameInfo{
					AppID: obj["appid"],
				}
				if v, ok := obj["name"].(string); ok {
					g.Name = v
				}
				if v, ok := obj["playtime_forever"].(float64); ok {
					g.PlaytimeMinutes = v
					g.PlaytimeHours = fmt.Sprintf("%.1f", v/60.0)
				}
				if v, ok := obj["rtime_last_played"]; ok {
					g.LastPlayed = v
				}
				games = append(games, g)
			}

			if len(games) == 0 {
				return fmt.Errorf("no games found for player %s in local store.\nhint: run 'steam-web-pp-cli sync' to populate local data", steamID)
			}

			// Sort by playtime descending
			sort.Slice(games, func(i, j int) bool {
				return games[i].PlaytimeMinutes > games[j].PlaytimeMinutes
			})

			prov := localProvenance(db, "iplayer-service", "transcendence_command")
			printProvenance(cmd, len(games), prov)

			data, err := json.Marshal(games)
			if err != nil {
				return fmt.Errorf("marshaling games: %w", err)
			}

			return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(data), flags)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/steam-web-pp-cli/data.db)")

	return cmd
}
