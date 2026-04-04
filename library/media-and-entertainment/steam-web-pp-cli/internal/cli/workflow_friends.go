package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/store"
)

func newWorkflowFriendsCmd(flags *rootFlags) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "friends <steamid>",
		Short: "List friends with persona names and friendship dates",
		Long: `Show all friends for a player from locally synced data. Displays each friend's
Steam ID, persona name (if synced), relationship type, and when the friendship
was established.`,
		Example: `  # List friends for a player
  steam-web-pp-cli workflow friends 76561198012345678

  # As JSON
  steam-web-pp-cli workflow friends 76561198012345678 --json`,
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

			items, err := db.List("isteam-user", 0)
			if err != nil {
				return fmt.Errorf("querying store: %w", err)
			}
			if len(items) == 0 {
				return fmt.Errorf("no friend data in local store.\nhint: run 'steam-web-pp-cli sync' to populate local data")
			}

			type friendInfo struct {
				SteamID      string `json:"steamid"`
				PersonaName  string `json:"persona_name,omitempty"`
				Relationship string `json:"relationship,omitempty"`
				FriendSince  string `json:"friend_since,omitempty"`
			}

			var friends []friendInfo

			// Look for friend list records — these have "steamid" and "relationship" fields
			for _, raw := range items {
				var obj map[string]any
				if json.Unmarshal(raw, &obj) != nil {
					continue
				}
				// Friend records have a "relationship" field and a "friend_since" timestamp
				if _, hasRelationship := obj["relationship"]; !hasRelationship {
					continue
				}

				f := friendInfo{}
				if v, ok := obj["steamid"].(string); ok {
					f.SteamID = v
				} else if v, ok := obj["steamid"].(float64); ok {
					f.SteamID = fmt.Sprintf("%.0f", v)
				}
				if v, ok := obj["relationship"].(string); ok {
					f.Relationship = v
				}
				if v, ok := obj["friend_since"].(float64); ok {
					t := time.Unix(int64(v), 0)
					f.FriendSince = t.Format("2006-01-02")
				}
				friends = append(friends, f)
			}

			// Enrich friend records with persona names from player summaries
			personaMap := buildPersonaMap(items)
			for i := range friends {
				if name, ok := personaMap[friends[i].SteamID]; ok {
					friends[i].PersonaName = name
				}
			}

			if len(friends) == 0 {
				return fmt.Errorf("no friends found for player %s in local store.\nhint: run 'steam-web-pp-cli isteam-user get-friend-list --steamid %s' then 'steam-web-pp-cli sync'", steamID, steamID)
			}

			// Sort by friend_since date (newest first)
			sort.Slice(friends, func(i, j int) bool {
				return friends[i].FriendSince > friends[j].FriendSince
			})

			prov := localProvenance(db, "isteam-user", "transcendence_command")
			printProvenance(cmd, len(friends), prov)

			data, err := json.Marshal(friends)
			if err != nil {
				return fmt.Errorf("marshaling friends: %w", err)
			}

			return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(data), flags)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/steam-web-pp-cli/data.db)")

	return cmd
}

// buildPersonaMap creates a steamid -> persona name lookup from player summary records.
func buildPersonaMap(items []json.RawMessage) map[string]string {
	m := make(map[string]string)
	for _, raw := range items {
		var obj map[string]any
		if json.Unmarshal(raw, &obj) != nil {
			continue
		}
		name, hasName := obj["personaname"].(string)
		if !hasName {
			continue
		}
		var sid string
		if v, ok := obj["steamid"].(string); ok {
			sid = v
		} else if v, ok := obj["steamid"].(float64); ok {
			sid = fmt.Sprintf("%.0f", v)
		}
		if sid != "" {
			m[sid] = name
		}
	}
	return m
}
