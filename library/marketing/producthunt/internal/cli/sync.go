// Sync pulls the public Product Hunt Atom feed at /feed, parses every entry,
// and writes the entries plus a ranked snapshot into the local SQLite store.
//
// Runs without auth. /feed is the only surface the runtime uses — HTML routes
// are CF-gated and are not fetched here. Each sync adds exactly one snapshot
// row; commands like trend/watch/makers read the accumulated snapshots.

package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/atom"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/store"
	"github.com/spf13/cobra"
)

func newSyncCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Fetch /feed and persist a ranked snapshot to the local SQLite store",
		Long: `Pull the public Atom feed at producthunt.com/feed, parse the 50 entries,
and write them into the local store as a ranked snapshot. Incremental and
idempotent — rerunning upserts posts, increments seen_count, and records
a new snapshot row.

Run this on a schedule (cron, launchd, systemd timer) to build up the
history that powers 'trend', 'watch', 'makers', and 'calendar'.

No authentication required.`,
		Example: `  # One-shot sync
  producthunt-pp-cli sync

  # Show what would happen without writing
  producthunt-pp-cli sync --dry-run

  # Use a non-default store path (e.g., a test DB)
  producthunt-pp-cli sync --db /tmp/ph.db --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("producthunt-pp-cli")
			}
			start := time.Now()

			body, err := fetchFeedBody(flags.timeout)
			if err != nil {
				return apiErr(err)
			}
			feed, err := atom.Parse(body)
			if err != nil {
				return apiErr(fmt.Errorf("parse feed: %w", err))
			}

			result := map[string]any{
				"source":             FeedEndpoint,
				"feed_title":         feed.Title,
				"entries_in_feed":    len(feed.Entries),
				"store_path":         dbPath,
				"elapsed_fetch_secs": time.Since(start).Seconds(),
			}

			if dryRun || flags.dryRun {
				result["dry_run"] = true
				out, _ := json.Marshal(result)
				return printOutputWithFlags(cmd.OutOrStdout(), out, flags)
			}

			db, err := store.Open(dbPath)
			if err != nil {
				return configErr(fmt.Errorf("open store: %w", err))
			}
			defer db.Close()
			if err := store.EnsurePHTables(db); err != nil {
				return configErr(err)
			}

			tx, err := db.DB().Begin()
			if err != nil {
				return configErr(fmt.Errorf("begin tx: %w", err))
			}
			rollback := true
			defer func() {
				if rollback {
					_ = tx.Rollback()
				}
			}()

			snapID, err := store.RecordSnapshot(tx, time.Now(), len(feed.Entries), "feed")
			if err != nil {
				return apiErr(err)
			}

			var upserts, skipped int
			for i, e := range feed.Entries {
				if e.Slug == "" {
					skipped++
					continue
				}
				p := store.Post{
					PostID:        e.PostID,
					Slug:          e.Slug,
					Title:         e.Title,
					Tagline:       e.Tagline,
					Author:        e.Author,
					DiscussionURL: e.DiscussionURL,
					ExternalURL:   e.ExternalURL,
					PublishedAt:   e.Published,
					UpdatedAt:     e.Updated,
				}
				if err := store.UpsertPost(tx, p); err != nil {
					return apiErr(err)
				}
				rank := i + 1
				if err := store.RecordSnapshotEntry(tx, snapID, e.PostID, rank, e.ExternalURL); err != nil {
					return apiErr(err)
				}
				upserts++
			}

			if err := tx.Commit(); err != nil {
				return apiErr(fmt.Errorf("commit sync tx: %w", err))
			}
			rollback = false

			totalPosts, _ := db.PostCount()
			totalSnapshots, _ := db.SnapshotCount()

			result["snapshot_id"] = snapID
			result["posts_upserted"] = upserts
			result["entries_skipped"] = skipped
			result["total_posts_in_store"] = totalPosts
			result["total_snapshots_in_store"] = totalSnapshots
			result["elapsed_total_secs"] = time.Since(start).Seconds()

			out, _ := json.Marshal(result)
			return printOutputWithFlags(cmd.OutOrStdout(), out, flags)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Path to local SQLite store (default: ~/.local/share/producthunt-pp-cli/data.db)")
	cmd.Flags().BoolVar(&dryRun, "dry-run-feed", false, "Fetch and parse feed but do not write anything")
	return cmd
}
