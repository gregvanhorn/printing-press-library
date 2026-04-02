package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/internal/client"

	"github.com/spf13/cobra"
)

func newFriendsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "friends <steamid-or-vanity>",
		Short: "List friends with profile summaries",
		Long:  "Fetch a player's friend list and enrich each friend with profile data (batched in groups of 100).",
		Example: `  steam-web-pp-cli friends 76561198006409530
  steam-web-pp-cli friends gabelogannewell --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if c.DryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUser/GetFriendList/v1")
				fmt.Fprintln(cmd.ErrOrStderr(), "GET /ISteamUser/GetPlayerSummaries/v2 (batched 100)")
				fmt.Fprintln(cmd.ErrOrStderr(), "\n(dry run - no request sent)")
				return nil
			}
			return runFriends(cmd, c, flags, args[0])
		},
	}
	return cmd
}

func runFriends(cmd *cobra.Command, c *client.Client, flags *rootFlags, input string) error {
	steamID, err := resolveSteamID(c, input)
	if err != nil {
		return err
	}
	key, err := steamAPIKey(c)
	if err != nil {
		return err
	}

	// Get friend list
	data, err := c.Get("/ISteamUser/GetFriendList/v1", map[string]string{
		"key":     key,
		"steamid": steamID,
	})
	if err != nil {
		return classifyAPIError(err)
	}
	resp, err := extractResponse(data)
	if err != nil {
		return err
	}

	var friendsResp struct {
		Friends []struct {
			SteamID      string `json:"steamid"`
			Relationship string `json:"relationship"`
			FriendSince  int64  `json:"friend_since"`
		} `json:"friends"`
	}
	// Try unwrapping friendslist envelope
	var envelope struct {
		Friendslist json.RawMessage `json:"friendslist"`
	}
	inner := resp
	if json.Unmarshal(resp, &envelope) == nil && envelope.Friendslist != nil {
		inner = envelope.Friendslist
	}
	if err := json.Unmarshal(inner, &friendsResp); err != nil {
		return fmt.Errorf("parsing friend list: %w", err)
	}

	if len(friendsResp.Friends) == 0 {
		out, _ := json.Marshal([]any{})
		return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
	}

	// Batch resolve summaries in groups of 100
	var allSteamIDs []string
	for _, f := range friendsResp.Friends {
		allSteamIDs = append(allSteamIDs, f.SteamID)
	}

	var allPlayers []map[string]any
	for i := 0; i < len(allSteamIDs); i += 100 {
		end := i + 100
		if end > len(allSteamIDs) {
			end = len(allSteamIDs)
		}
		batch := allSteamIDs[i:end]

		summaryData, err := c.Get("/ISteamUser/GetPlayerSummaries/v2", map[string]string{
			"key":      key,
			"steamids": strings.Join(batch, ","),
		})
		if err != nil {
			return classifyAPIError(err)
		}
		sResp, _ := extractResponse(summaryData)
		var sObj struct {
			Players []map[string]any `json:"players"`
		}
		if json.Unmarshal(sResp, &sObj) == nil {
			allPlayers = append(allPlayers, sObj.Players...)
		}
	}

	out, _ := json.Marshal(allPlayers)
	return printOutputWithFlags(cmd.OutOrStdout(), json.RawMessage(out), flags)
}
