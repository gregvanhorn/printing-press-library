// Hand-written: launches and mentions commands. Hacker News Algolia.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mvanhorn/printing-press-library/library/developer-tools/company-goat/internal/source/hn"
	"github.com/spf13/cobra"
)

type launchEntry struct {
	Title     string `json:"title"`
	URL       string `json:"url,omitempty"`
	Author    string `json:"author"`
	Points    int    `json:"points"`
	Comments  int    `json:"num_comments"`
	CreatedAt string `json:"created_at"`
	StoryID   int    `json:"story_id"`
	HNURL     string `json:"hn_url"`
}

func newLaunchesCmd(flags *rootFlags) *cobra.Command {
	var t targetFlags
	var maxHits int
	var minPoints int

	cmd := &cobra.Command{
		Use:   "launches <co>",
		Short: "Show HN posts mentioning the company, sorted by points. Includes year hints to spot dead vs. active launches.",
		Long: `launches searches the Hacker News Algolia index for "Show HN" posts where the title or content mentions the resolved company. Results are sorted by points descending.

Use this to gauge launch story strength, find the canonical Show HN post for a product, or spot when a startup pivoted/relaunched.`,
		Example: strings.Trim(`
  company-goat-pp-cli launches replit
  company-goat-pp-cli launches stripe --json --max 10
  company-goat-pp-cli launches vercel --min-points 50
`, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			if t.Domain == "" && len(args) == 0 {
				return cmd.Help()
			}
			if maxHits <= 0 {
				maxHits = 20
			}
			domain, err := runResolveOrExit(cmd, flags, args, t)
			if err != nil {
				return err
			}
			stem := strings.SplitN(domain, ".", 2)[0]

			c := hn.NewClient()
			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Second)
			defer cancel()

			// Use the original args (more semantic) plus the stem fallback.
			query := strings.Join(args, " ")
			if query == "" {
				query = stem
			}
			resp, err := c.SearchShowHN(ctx, query, maxHits)
			if err != nil {
				return fmt.Errorf("hn: %w", err)
			}
			entries := make([]launchEntry, 0, len(resp.Hits))
			for _, h := range resp.Hits {
				if h.Points < minPoints {
					continue
				}
				entries = append(entries, launchEntry{
					Title:     h.Title,
					URL:       h.URL,
					Author:    h.Author,
					Points:    h.Points,
					Comments:  h.NumComments,
					CreatedAt: h.CreatedAt,
					StoryID:   h.StoryID,
					HNURL:     fmt.Sprintf("https://news.ycombinator.com/item?id=%d", h.StoryID),
				})
			}
			sort.SliceStable(entries, func(i, j int) bool { return entries[i].Points > entries[j].Points })

			out := map[string]any{
				"domain":        domain,
				"query":         query,
				"launches":      entries,
				"total_matches": resp.NbHits,
			}
			w := cmd.OutOrStdout()
			asJSON := flags.asJSON || !isTerminal(w)
			if asJSON {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}
			fmt.Fprintf(w, "Show HN posts for %q (top %d of %d total):\n\n", domain, len(entries), resp.NbHits)
			if len(entries) == 0 {
				fmt.Fprintln(w, "no Show HN posts found")
				return nil
			}
			for _, e := range entries {
				yr := ""
				if len(e.CreatedAt) >= 4 {
					yr = e.CreatedAt[:4]
				}
				fmt.Fprintf(w, "  %s  %4d↑  %3d💬  %s\n", yr, e.Points, e.Comments, fundingTruncate(e.Title, 80))
				fmt.Fprintf(w, "    %s\n", e.HNURL)
			}
			return nil
		},
	}
	addTargetFlags(cmd, &t)
	cmd.Flags().IntVar(&maxHits, "max", 20, "Maximum hits to return")
	cmd.Flags().IntVar(&minPoints, "min-points", 0, "Filter to posts with at least this many points")
	return cmd
}

func newMentionsCmd(flags *rootFlags) *cobra.Command {
	var t targetFlags
	var maxHits int

	cmd := &cobra.Command{
		Use:   "mentions <co>",
		Short: "Hacker News mention timeline: monthly histogram of mentions over time via Algolia full-text search.",
		Long: `mentions searches HN's full-text Algolia index for any story containing the resolved company name. Results are bucketed by year-month for a quick "is this still talked about?" view.

With --json, returns the raw histogram as a sorted array of {month, count} pairs.`,
		Example: strings.Trim(`
  company-goat-pp-cli mentions stripe
  company-goat-pp-cli mentions anthropic --json
`, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			if t.Domain == "" && len(args) == 0 {
				return cmd.Help()
			}
			if maxHits <= 0 {
				maxHits = 100
			}
			domain, err := runResolveOrExit(cmd, flags, args, t)
			if err != nil {
				return err
			}
			query := strings.Join(args, " ")
			if query == "" {
				query = strings.SplitN(domain, ".", 2)[0]
			}
			c := hn.NewClient()
			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Second)
			defer cancel()
			resp, err := c.SearchByDate(ctx, query, maxHits)
			if err != nil {
				return fmt.Errorf("hn: %w", err)
			}
			type bucket struct {
				Month string `json:"month"`
				Count int    `json:"count"`
			}
			counts := map[string]int{}
			for _, h := range resp.Hits {
				if len(h.CreatedAt) < 7 {
					continue
				}
				counts[h.CreatedAt[:7]]++
			}
			months := make([]string, 0, len(counts))
			for m := range counts {
				months = append(months, m)
			}
			sort.Strings(months)
			out := make([]bucket, 0, len(months))
			for _, m := range months {
				out = append(out, bucket{Month: m, Count: counts[m]})
			}

			w := cmd.OutOrStdout()
			asJSON := flags.asJSON || !isTerminal(w)
			if asJSON {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{
					"domain":         domain,
					"query":          query,
					"timeline":       out,
					"total_mentions": resp.NbHits,
					"sampled_hits":   len(resp.Hits),
				})
			}
			fmt.Fprintf(w, "HN mentions for %q (sampled %d of %d total):\n\n", domain, len(resp.Hits), resp.NbHits)
			if len(out) == 0 {
				fmt.Fprintln(w, "no mentions found")
				return nil
			}
			for _, b := range out {
				bar := strings.Repeat("█", b.Count)
				if len(bar) > 40 {
					bar = bar[:40] + "..."
				}
				fmt.Fprintf(w, "  %s  %3d  %s\n", b.Month, b.Count, bar)
			}
			return nil
		},
	}
	addTargetFlags(cmd, &t)
	cmd.Flags().IntVar(&maxHits, "max", 100, "Maximum hits to sample for the timeline (max 1000)")
	return cmd
}
