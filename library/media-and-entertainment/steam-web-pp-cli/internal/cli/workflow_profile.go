package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/store"
)

func newWorkflowProfileCmd(flags *rootFlags) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "profile <steamid>",
		Short: "One-command player summary: persona name, profile URL, level, and game count",
		Long: `Show a consolidated player profile from locally synced data.
Combines player summary (ISteamUser), steam level (IPlayerService), and
owned games count into a single view. Requires a prior sync.`,
		Example: `  # Show player profile
  steam-web-pp-cli workflow profile 76561198012345678

  # Show as JSON
  steam-web-pp-cli workflow profile 76561198012345678 --json`,
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

			profile := map[string]any{
				"steamid": steamID,
			}

			// Look for player summary from isteam-user resource type
			playerFound := false
			items, err := db.List("isteam-user", 0)
			if err == nil {
				for _, raw := range items {
					var obj map[string]any
					if json.Unmarshal(raw, &obj) != nil {
						continue
					}
					if matchesSteamID(obj, steamID) {
						playerFound = true
						copyFields(obj, profile, "personaname", "profileurl", "avatar",
							"personastate", "lastlogoff", "timecreated", "communityvisibilitystate")
						break
					}
				}
			}

			// Look for steam level and game count from iplayer-service
			levelItems, err := db.List("iplayer-service", 0)
			if err == nil {
				for _, raw := range levelItems {
					var obj map[string]any
					if json.Unmarshal(raw, &obj) != nil {
						continue
					}
					if matchesSteamID(obj, steamID) {
						copyFields(obj, profile, "player_level", "player_xp", "game_count")
					}
				}
			}

			if !playerFound {
				return fmt.Errorf("no player data found for %s in local store.\nhint: run 'steam-web-pp-cli sync' to populate local data", steamID)
			}

			prov := localProvenance(db, "isteam-user", "transcendence_command")
			printProvenance(cmd, 1, prov)

			data, err := json.Marshal(profile)
			if err != nil {
				return fmt.Errorf("marshaling profile: %w", err)
			}

			return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(data), flags)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/steam-web-pp-cli/data.db)")

	return cmd
}

// matchesSteamID checks if a record matches the given Steam ID by comparing
// common field names used across Steam API responses.
func matchesSteamID(obj map[string]any, steamID string) bool {
	for _, key := range []string{"steamid", "steamID", "SteamId", "steam_id"} {
		if v, ok := obj[key]; ok {
			return fmt.Sprintf("%v", v) == steamID
		}
	}
	return false
}

// copyFields copies specified fields from src to dst if they exist.
func copyFields(src, dst map[string]any, fields ...string) {
	for _, f := range fields {
		if v, ok := src[f]; ok {
			dst[f] = v
		}
	}
}
