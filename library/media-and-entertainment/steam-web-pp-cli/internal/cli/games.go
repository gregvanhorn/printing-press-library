package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newGamesCmd(flags *rootFlags) *cobra.Command {
	var sortBy string
	var limit int

	cmd := &cobra.Command{
		Use:   "games <steamid-or-vanity>",
		Short: "List owned games with playtime",
		Long:  "Fetch all games owned by a player, with sorting and limit options.",
		Example: `  steam-web-pp-cli games 76561198006409530
  steam-web-pp-cli games gabelogannewell --sort playtime --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /IPlayerService/GetOwnedGames/v1 ?include_appinfo=true")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}
			return runGames(cmd, c, flags, args[0], sortBy, limit)
		},
	}
	cmd.Flags().StringVar(&sortBy, "sort", "name", "Sort by: name, playtime, appid")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of games returned (0 = all)")
	return cmd
}

func runGames(cmd *cobra.Command, c *client.Client, flags *rootFlags, input, sortBy string, limit int) error {
	games, err := fetchOwnedGames(c, input)
	if err != nil {
		return err
	}

	// Sort
	sort.Slice(games, func(i, j int) bool {
		switch sortBy {
		case "playtime":
			pi, _ := games[i]["playtime_forever"].(float64)
			pj, _ := games[j]["playtime_forever"].(float64)
			return pi > pj
		case "appid":
			ai, _ := games[i]["appid"].(float64)
			aj, _ := games[j]["appid"].(float64)
			return ai < aj
		default: // name
			ni, _ := games[i]["name"].(string)
			nj, _ := games[j]["name"].(string)
			return ni < nj
		}
	})

	// Limit
	if limit > 0 && limit < len(games) {
		games = games[:limit]
	}

	out, _ := json.Marshal(games)
	return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
}

// fetchOwnedGames retrieves the owned games list for a player.
// Shared by games, compare, overlap, and backlog commands.
func fetchOwnedGames(c *client.Client, input string) ([]map[string]any, error) {
	steamID, err := resolveSteamID(c, input)
	if err != nil {
		return nil, err
	}
	key, err := steamAPIKey(c)
	if err != nil {
		return nil, err
	}
	data, err := c.Get("/IPlayerService/GetOwnedGames/v1", map[string]string{
		"key":                       key,
		"steamid":                   steamID,
		"include_appinfo":           "true",
		"include_played_free_games": "true",
	})
	if err != nil {
		return nil, classifyAPIError(err)
	}
	resp, err := extractResponse(data)
	if err != nil {
		return nil, err
	}
	var respObj struct {
		Games []map[string]any `json:"games"`
	}
	if err := json.Unmarshal(resp, &respObj); err != nil {
		return nil, fmt.Errorf("parsing owned games: %w", err)
	}
	return respObj.Games, nil
}
