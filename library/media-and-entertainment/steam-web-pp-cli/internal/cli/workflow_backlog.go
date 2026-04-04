package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/store"
)

func newWorkflowBacklogCmd(flags *rootFlags) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "backlog <steamid>",
		Short: "Find owned games with zero playtime (the shame pile)",
		Long: `Show all games owned by a player that have never been played (zero total
playtime). Sorted alphabetically by game name. Useful for discovering your
backlog or gift-worthy titles.`,
		Example: `  # Show unplayed games
  steam-web-pp-cli workflow backlog 76561198012345678

  # As JSON
  steam-web-pp-cli workflow backlog 76561198012345678 --json`,
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

			type backlogGame struct {
				AppID any    `json:"appid"`
				Name  string `json:"name"`
			}

			var backlog []backlogGame

			for _, raw := range items {
				var obj map[string]any
				if json.Unmarshal(raw, &obj) != nil {
					continue
				}
				// Must be a game record (has appid)
				if _, hasAppID := obj["appid"]; !hasAppID {
					continue
				}
				// If the record has a steamid, check it matches
				if sid, ok := obj["steamid"]; ok {
					if fmt.Sprintf("%v", sid) != steamID {
						continue
					}
				}
				// Check for zero playtime
				playtime, _ := obj["playtime_forever"].(float64)
				if playtime > 0 {
					continue
				}

				g := backlogGame{
					AppID: obj["appid"],
				}
				if v, ok := obj["name"].(string); ok {
					g.Name = v
				}
				backlog = append(backlog, g)
			}

			if len(backlog) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No unplayed games found. Either the player has played everything, or game data hasn't been synced yet.")
				fmt.Fprintf(cmd.ErrOrStderr(), "hint: run 'steam-web-pp-cli sync' to populate local data\n")
				return nil
			}

			// Sort alphabetically by name
			sort.Slice(backlog, func(i, j int) bool {
				return backlog[i].Name < backlog[j].Name
			})

			prov := localProvenance(db, "iplayer-service", "transcendence_command")
			printProvenance(cmd, len(backlog), prov)

			data, err := json.Marshal(backlog)
			if err != nil {
				return fmt.Errorf("marshaling backlog: %w", err)
			}

			return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(data), flags)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/steam-web-pp-cli/data.db)")

	return cmd
}
