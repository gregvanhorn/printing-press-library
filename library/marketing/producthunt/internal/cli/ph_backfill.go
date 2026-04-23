package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/config"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/phgraphql"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/store"
)

// Pagination and budget defaults for backfill. Pulled out as constants so
// tests and docs can refer to them without hard-coding.
const (
	BackfillPageSize           = 20 // PH GraphQL posts edges per page — small & polite
	BackfillBudgetHardStopPct  = 10 // abort when budget remaining drops below this %
	BackfillBudgetSoftBrakePct = 25 // slow pagination to 300ms when below this %
	BackfillBudgetEasePct      = 50 // no sleep above this %
	BackfillSleepSoftMs        = 300
	BackfillSleepEaseMs        = 100
	BackfillUserAgentBase      = "producthunt-pp-cli"

	// Complexity cost estimate per page. PH doesn't publish an exact figure,
	// so we use a conservative approximation: 20 posts × 4 points = ~80.
	// Total used is read from X-Rate-Limit-Remaining after each response.
	BackfillEstimatedComplexityPerPage = 4
)

// newBackfillCmd is the top-level `backfill` command. It accepts window
// flags directly; `backfill resume` is a subcommand for continuing
// interrupted runs.
//
// Shape:
//
//	producthunt-pp-cli backfill [--days N | --from DATE --to DATE] [--dry-run]
//	producthunt-pp-cli backfill resume [--window-id ID]
func newBackfillCmd(flags *rootFlags) *cobra.Command {
	var (
		days     int
		fromDate string
		toDate   string
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "backfill",
		Short: "Seed the local store with 30 days of posts via GraphQL",
		Long: `Paginate the Product Hunt GraphQL posts query over a time window and upsert
every result into the local SQLite store. One-shot, resumable, budget-aware.

Use --dry-run to preview estimated cost before burning API budget; use
--days to set a window size (default 30) or --from/--to for explicit
dates. Run 'backfill resume' to continue an interrupted run from the
last saved cursor.

	Requires Product Hunt GraphQL auth: run 'auth setup' first. Atom-runtime
	commands (sync, list, search) do NOT use this path and need no auth.`,
		Example: `  # Preview the 30-day backfill cost
  producthunt-pp-cli backfill --days 30 --dry-run

  # Run a 30-day backfill
  producthunt-pp-cli backfill --days 30

  # Specific window
  producthunt-pp-cli backfill --from 2026-03-01 --to 2026-04-01`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackfill(cmd, flags, backfillOpts{
				Days:   days,
				From:   fromDate,
				To:     toDate,
				DryRun: dryRun || flags.dryRun,
			})
		},
	}
	cmd.Flags().IntVar(&days, "days", 30, "Rolling window in days ending today (ignored when --from/--to are set)")
	cmd.Flags().StringVar(&fromDate, "from", "", "Window start (YYYY-MM-DD). Requires --to.")
	cmd.Flags().StringVar(&toDate, "to", "", "Window end (YYYY-MM-DD). Requires --from.")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Estimate cost and exit without hitting GraphQL")

	cmd.AddCommand(newBackfillResumeCmd(flags))
	return cmd
}

// backfillOpts captures the parameters of one backfill invocation.
type backfillOpts struct {
	Days   int
	From   string
	To     string
	DryRun bool
}

// windowID produces a deterministic identifier for a (from, to) pair so
// repeating the same window is idempotent at the store layer.
func windowID(from, to string) string {
	sum := sha256.Sum256([]byte(from + "|" + to))
	return "w-" + hex.EncodeToString(sum[:8])
}

// resolveWindow computes the (from, to) ISO8601 DateTime strings for a
// backfill run from the CLI flags. Returns (from, to, windowID, error).
func resolveWindow(opts backfillOpts) (string, string, string, error) {
	hasRange := opts.From != "" || opts.To != ""
	if hasRange && (opts.From == "" || opts.To == "") {
		return "", "", "", fmt.Errorf("--from and --to must be used together")
	}
	if hasRange && opts.Days != 30 {
		// --days uses its default when flag not set. Reject explicit combos.
		return "", "", "", fmt.Errorf("--days cannot be combined with --from/--to")
	}

	var fromT, toT time.Time
	if hasRange {
		// Dates are interpreted as start-of-day UTC.
		var err error
		fromT, err = time.Parse("2006-01-02", opts.From)
		if err != nil {
			return "", "", "", fmt.Errorf("--from %q: %w", opts.From, err)
		}
		toT, err = time.Parse("2006-01-02", opts.To)
		if err != nil {
			return "", "", "", fmt.Errorf("--to %q: %w", opts.To, err)
		}
		if toT.Before(fromT) {
			return "", "", "", fmt.Errorf("--to must be on or after --from")
		}
	} else {
		if opts.Days <= 0 {
			return "", "", "", fmt.Errorf("--days must be positive")
		}
		toT = time.Now().UTC()
		fromT = toT.AddDate(0, 0, -opts.Days)
	}

	from := fromT.UTC().Format(time.RFC3339)
	to := toT.UTC().Format(time.RFC3339)
	return from, to, windowID(from, to), nil
}

// runBackfill is the body of the backfill command. Extracted so the
// `resume` command can share the core loop.
func runBackfill(cmd *cobra.Command, flags *rootFlags, opts backfillOpts) error {
	from, to, winID, err := resolveWindow(opts)
	if err != nil {
		return usageErr(err)
	}

	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return configErr(err)
	}

	if opts.DryRun {
		return emitDryRunEstimate(cmd, flags, from, to, opts)
	}

	if !cfg.HasGraphQLToken() {
		return emitMissingGraphQLAuth(cmd, flags, "backfill needs Product Hunt GraphQL auth because the anonymous Atom feed cannot retroactively fetch historical windows")
	}

	dbPath := defaultDBPath("producthunt-pp-cli")
	db, err := store.Open(dbPath)
	if err != nil {
		return configErr(fmt.Errorf("open store: %w", err))
	}
	defer db.Close()
	if err := store.EnsurePHTables(db); err != nil {
		return configErr(err)
	}

	// Fetch or create the state row so we know where to start.
	state, err := db.GetBackfillState(winID)
	if err != nil {
		return configErr(err)
	}
	if state == nil {
		state = &store.BackfillState{
			WindowID:     winID,
			PostedAfter:  from,
			PostedBefore: to,
		}
		if err := db.UpsertBackfillState(*state); err != nil {
			return configErr(err)
		}
	}
	if state.IsComplete() {
		return emitAlreadyComplete(cmd, flags, state)
	}

	client := phgraphql.NewClient(cfg.AccessToken, userAgent())
	return executeBackfillLoop(cmd, flags, db, client, state)
}

// executeBackfillLoop paginates GraphQL until done, budget runs low, or an
// error. Shared by `backfill` and `backfill resume`.
func executeBackfillLoop(cmd *cobra.Command, flags *rootFlags, db *store.Store, client *phgraphql.Client, state *store.BackfillState) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Minute)
	defer cancel()

	start := time.Now()
	cursor := state.Cursor
	pages := state.PagesCompleted
	upserted := state.PostsUpserted

	for {
		// Budget check BEFORE the next call.
		pct := client.Budget().PercentRemaining()
		if client.Budget().Known() && pct < float64(BackfillBudgetHardStopPct)/100.0 {
			return persistAndExitRateLimited(cmd, flags, db, state, cursor, pages, upserted, client.Budget())
		}
		if client.Budget().Known() && pct < float64(BackfillBudgetSoftBrakePct)/100.0 {
			time.Sleep(time.Duration(BackfillSleepSoftMs) * time.Millisecond)
		} else if !client.Budget().Known() || pct < float64(BackfillBudgetEasePct)/100.0 {
			time.Sleep(time.Duration(BackfillSleepEaseMs) * time.Millisecond)
		}

		page, err := fetchBackfillPage(ctx, client, state.PostedAfter, state.PostedBefore, cursor)
		if err != nil {
			return handleBackfillError(cmd, flags, db, state, cursor, pages, upserted, err, client.Budget())
		}

		// Upsert everything returned.
		tx, err := db.DB().Begin()
		if err != nil {
			return apiErr(err)
		}
		rollback := true
		defer func() {
			if rollback {
				_ = tx.Rollback()
			}
		}()
		for _, edge := range page.Edges {
			p := postNodeToStore(edge.Node)
			if p.Slug == "" {
				continue
			}
			if err := store.UpsertPost(tx, p); err != nil {
				return apiErr(err)
			}
			upserted++
		}
		if err := tx.Commit(); err != nil {
			return apiErr(err)
		}
		rollback = false

		pages++
		cursor = page.PageInfo.EndCursor

		// Update cursor after every page so a crash doesn't re-burn budget
		// on pages we've already paid for.
		state.Cursor = cursor
		state.PagesCompleted = pages
		state.PostsUpserted = upserted
		state.LastRunAt = time.Now().UTC()
		state.LastError = ""
		if err := db.UpsertBackfillState(*state); err != nil {
			return configErr(err)
		}

		if !page.PageInfo.HasNextPage {
			break
		}
	}

	// Mark complete.
	state.CompletedAt = time.Now().UTC()
	state.Cursor = ""
	if err := db.UpsertBackfillState(*state); err != nil {
		return configErr(err)
	}
	return emitRunSummary(cmd, flags, state, time.Since(start), client.Budget(), false)
}

// fetchBackfillPage issues one GraphQL call and decodes the `posts` field.
func fetchBackfillPage(ctx context.Context, client *phgraphql.Client, postedAfter, postedBefore, after string) (*phgraphql.PostsPage, error) {
	vars := map[string]any{
		"first":        BackfillPageSize,
		"postedAfter":  postedAfter,
		"postedBefore": postedBefore,
	}
	if after != "" {
		vars["after"] = after
	}
	resp, err := client.Execute(ctx, phgraphql.BackfillPostsQuery, vars)
	if err != nil {
		return nil, err
	}
	if resp.HasErrors() {
		return nil, fmt.Errorf("graphql error: %s", resp.ErrorMessage())
	}
	var envelope struct {
		Posts phgraphql.PostsPage `json:"posts"`
	}
	if err := json.Unmarshal(resp.Data, &envelope); err != nil {
		return nil, fmt.Errorf("decode posts: %w", err)
	}
	return &envelope.Posts, nil
}

// postNodeToStore adapts a GraphQL PostNode into the shape the store
// upsert expects. createdAt from GraphQL is the PH launch time; we map it
// to published_at so Atom-runtime commands (list --since 30d, etc) see
// the same semantic.
func postNodeToStore(n phgraphql.PostNode) store.Post {
	// PH GraphQL ids are opaque strings; posts table stores int64 post_id.
	// Parse as int when possible (current PH ids are numeric).
	var postID int64
	if v, err := strconv.ParseInt(n.ID, 10, 64); err == nil {
		postID = v
	}
	author := strings.TrimSpace(n.User.Name)
	if author == "" {
		author = strings.TrimSpace(n.User.Username)
	}
	discussion := fmt.Sprintf("https://www.producthunt.com/products/%s", n.Slug)
	if n.URL != "" {
		discussion = n.URL
	}
	external := n.Website
	return store.Post{
		PostID:        postID,
		Slug:          n.Slug,
		Title:         n.Name,
		Tagline:       n.Tagline,
		Author:        author,
		DiscussionURL: discussion,
		ExternalURL:   external,
		PublishedAt:   n.CreatedAt,
		UpdatedAt:     n.CreatedAt,
	}
}

// handleBackfillError decides how to react to a fetch failure mid-loop.
// Rate limits → save cursor, exit 3 with resume guidance. Other errors →
// save cursor with LastError set, exit non-zero.
func handleBackfillError(cmd *cobra.Command, flags *rootFlags, db *store.Store, state *store.BackfillState, cursor string, pages, upserted int, err error, budget phgraphql.Budget) error {
	state.Cursor = cursor
	state.PagesCompleted = pages
	state.PostsUpserted = upserted
	state.LastRunAt = time.Now().UTC()
	state.LastError = err.Error()
	_ = db.UpsertBackfillState(*state)

	switch e := err.(type) {
	case *phgraphql.RateLimitedError:
		_ = e
		return persistAndExitRateLimited(cmd, flags, db, state, cursor, pages, upserted, budget)
	case *phgraphql.AuthError:
		return authErr(fmt.Errorf("auth failed: %w (re-run `auth register`)", err))
	}
	return apiErr(err)
}

func emitMissingGraphQLAuth(cmd *cobra.Command, flags *rootFlags, reason string) error {
	if flags.asJSON || flags.agent {
		result := map[string]any{
			"error":                 "missing_graphql_auth",
			"can_run_anonymously":   false,
			"anonymous_alternative": "Run `producthunt-pp-cli sync` on a schedule to accumulate Atom history from now forward.",
			"auth_hint":             graphQLAuthHint(reason),
		}
		raw, _ := json.Marshal(result)
		_ = printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
	}
	return authErr(fmt.Errorf("%s. Run `producthunt-pp-cli auth setup` for guided setup, or configure PRODUCTHUNT_DEVELOPER_TOKEN", reason))
}

// persistAndExitRateLimited saves the cursor and exits with a typed
// rateLimitErr (exit code 3) and a message the user can act on.
func persistAndExitRateLimited(cmd *cobra.Command, flags *rootFlags, db *store.Store, state *store.BackfillState, cursor string, pages, upserted int, budget phgraphql.Budget) error {
	state.Cursor = cursor
	state.PagesCompleted = pages
	state.PostsUpserted = upserted
	state.LastRunAt = time.Now().UTC()
	state.LastError = "rate limited"
	_ = db.UpsertBackfillState(*state)

	resetDelta := time.Duration(0)
	if !budget.Reset.IsZero() {
		resetDelta = time.Until(budget.Reset).Round(time.Second)
	}

	w := cmd.OutOrStdout()
	if flags.asJSON {
		result := map[string]any{
			"rate_limited":   true,
			"posts_upserted": upserted,
			"pages":          pages,
			"resume_hint":    fmt.Sprintf("producthunt-pp-cli backfill resume --window-id %s", state.WindowID),
		}
		if resetDelta > 0 {
			result["retry_after_secs"] = int(resetDelta.Seconds())
		}
		raw, _ := json.Marshal(result)
		_ = printOutputWithFlags(w, raw, flags)
	} else {
		fmt.Fprintln(w, yellow("Rate limit hit — progress saved"))
		fmt.Fprintf(w, "  Posts upserted so far: %d across %d pages\n", upserted, pages)
		if resetDelta > 0 {
			fmt.Fprintf(w, "  Retry in: %s\n", resetDelta)
		}
		fmt.Fprintf(w, "  Resume with: producthunt-pp-cli backfill resume --window-id %s\n", state.WindowID)
	}
	return rateLimitErr(fmt.Errorf("backfill suspended at %d pages / %d posts; cursor saved", pages, upserted))
}

// emitDryRunEstimate prints what the backfill would do without calling
// GraphQL. Estimate is deterministic: ~50 launches/day × window_days ÷
// BackfillPageSize pages × estimated complexity/page.
func emitDryRunEstimate(cmd *cobra.Command, flags *rootFlags, from, to string, opts backfillOpts) error {
	fromT, _ := time.Parse(time.RFC3339, from)
	toT, _ := time.Parse(time.RFC3339, to)
	windowDays := int(toT.Sub(fromT).Hours() / 24.0)
	if windowDays < 1 {
		windowDays = 1
	}
	estPosts := windowDays * 50
	estPages := (estPosts + BackfillPageSize - 1) / BackfillPageSize
	estComplexity := estPages * BackfillEstimatedComplexityPerPage

	w := cmd.OutOrStdout()
	if flags.asJSON {
		result := map[string]any{
			"dry_run":               true,
			"window":                map[string]string{"from": from, "to": to},
			"window_days":           windowDays,
			"estimated_posts":       estPosts,
			"estimated_pages":       estPages,
			"estimated_complexity":  estComplexity,
			"budget_limit":          6250,
			"percent_of_one_window": fmt.Sprintf("%.1f%%", float64(estComplexity)/6250.0*100.0),
		}
		raw, _ := json.Marshal(result)
		return printOutputWithFlags(w, raw, flags)
	}
	fmt.Fprintln(w, bold("Dry run — estimates only"))
	fmt.Fprintf(w, "  Window:              %s → %s (%d days)\n", from, to, windowDays)
	fmt.Fprintf(w, "  Estimated posts:     ~%d\n", estPosts)
	fmt.Fprintf(w, "  Estimated pages:     ~%d (at %d posts/page)\n", estPages, BackfillPageSize)
	fmt.Fprintf(w, "  Estimated complexity:~%d points (of 6,250/15min budget)\n", estComplexity)
	fmt.Fprintf(w, "  Fraction of budget:  %.1f%%\n", float64(estComplexity)/6250.0*100.0)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Run without --dry-run to execute.")
	_ = opts
	return nil
}

// emitRunSummary prints the final summary when backfill finishes cleanly.
func emitRunSummary(cmd *cobra.Command, flags *rootFlags, state *store.BackfillState, elapsed time.Duration, budget phgraphql.Budget, resumed bool) error {
	w := cmd.OutOrStdout()
	if flags.asJSON {
		result := map[string]any{
			"completed":      true,
			"window_id":      state.WindowID,
			"posts_upserted": state.PostsUpserted,
			"pages":          state.PagesCompleted,
			"elapsed_secs":   elapsed.Seconds(),
			"resumed":        resumed,
		}
		if budget.Known() {
			result["budget_remaining_pct"] = budget.PercentRemaining()
		}
		raw, _ := json.Marshal(result)
		return printOutputWithFlags(w, raw, flags)
	}
	tag := "Backfill complete"
	if resumed {
		tag = "Resume complete"
	}
	fmt.Fprintln(w, green(tag))
	fmt.Fprintf(w, "  Posts upserted: %d\n", state.PostsUpserted)
	fmt.Fprintf(w, "  Pages:          %d\n", state.PagesCompleted)
	fmt.Fprintf(w, "  Elapsed:        %s\n", elapsed.Round(time.Second))
	if budget.Known() {
		fmt.Fprintf(w, "  Budget remaining: %d of %d (%.0f%%)\n", budget.Remaining, budget.Limit, budget.PercentRemaining()*100.0)
	}
	return nil
}

// emitAlreadyComplete prints "nothing to do" when a completed window is
// re-invoked. Exit 0 — repeating a complete window is not an error.
func emitAlreadyComplete(cmd *cobra.Command, flags *rootFlags, state *store.BackfillState) error {
	w := cmd.OutOrStdout()
	if flags.asJSON {
		result := map[string]any{
			"completed":      true,
			"window_id":      state.WindowID,
			"posts_upserted": state.PostsUpserted,
			"already_done":   true,
		}
		raw, _ := json.Marshal(result)
		return printOutputWithFlags(w, raw, flags)
	}
	fmt.Fprintln(w, green("Backfill already complete for this window"))
	fmt.Fprintf(w, "  Posts upserted: %d\n", state.PostsUpserted)
	fmt.Fprintf(w, "  Window: %s → %s\n", state.PostedAfter, state.PostedBefore)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "To re-fetch: delete the row in ph_backfill_state and retry.")
	return nil
}

// userAgent builds the User-Agent string for GraphQL requests, identifying
// the CLI so PH API teams can distinguish legitimate integrators from
// unidentified scraping.
func userAgent() string {
	return fmt.Sprintf("%s/%s (+github.com/mvanhorn/printing-press-library)", BackfillUserAgentBase, version)
}

// --- Resume subcommand ---

func newBackfillResumeCmd(flags *rootFlags) *cobra.Command {
	var winIDFlag string
	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume an interrupted backfill from the last saved cursor",
		Long: `Look up incomplete rows in ph_backfill_state and continue from
the saved cursor. Useful after a rate-limit, network blip, or manual
interrupt.

With no arguments: if exactly one pending row exists, resume it. If
multiple, print the list and require --window-id.

If no pending rows exist, exits 0 with 'nothing to resume'.`,
		Example: `  # Resume the single pending window
  producthunt-pp-cli backfill resume

  # Resume a specific window
  producthunt-pp-cli backfill resume --window-id w-a1b2c3d4`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackfillResume(cmd, flags, winIDFlag)
		},
	}
	cmd.Flags().StringVar(&winIDFlag, "window-id", "", "Specific window_id to resume (required when multiple pending)")
	return cmd
}

func runBackfillResume(cmd *cobra.Command, flags *rootFlags, windowIDFlag string) error {
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return configErr(err)
	}
	if !cfg.HasGraphQLToken() {
		return emitMissingGraphQLAuth(cmd, flags, "backfill resume needs the same Product Hunt GraphQL auth used by the original backfill")
	}

	dbPath := defaultDBPath("producthunt-pp-cli")
	db, err := store.Open(dbPath)
	if err != nil {
		return configErr(fmt.Errorf("open store: %w", err))
	}
	defer db.Close()
	if err := store.EnsurePHTables(db); err != nil {
		return configErr(err)
	}

	pending, err := db.PendingBackfillStates()
	if err != nil {
		return configErr(err)
	}
	if len(pending) == 0 {
		w := cmd.OutOrStdout()
		if flags.asJSON {
			raw, _ := json.Marshal(map[string]any{"resumed": false, "reason": "no_pending"})
			return printOutputWithFlags(w, raw, flags)
		}
		fmt.Fprintln(w, "Nothing to resume — no incomplete backfills.")
		return nil
	}

	var target *store.BackfillState
	if windowIDFlag != "" {
		for i := range pending {
			if pending[i].WindowID == windowIDFlag {
				target = &pending[i]
				break
			}
		}
		if target == nil {
			return usageErr(fmt.Errorf("no pending window with id %q", windowIDFlag))
		}
	} else if len(pending) > 1 {
		w := cmd.OutOrStdout()
		fmt.Fprintln(w, "Multiple pending backfills. Pass --window-id to disambiguate:")
		for _, s := range pending {
			fmt.Fprintf(w, "  %s  %s → %s (pages=%d, posts=%d)\n",
				s.WindowID, s.PostedAfter, s.PostedBefore, s.PagesCompleted, s.PostsUpserted)
		}
		return usageErr(fmt.Errorf("--window-id required"))
	} else {
		target = &pending[0]
	}

	// Schema version check (defensive): if ph_schema_version is older than
	// we expect, the cursor format may have drifted. Fail fast.
	sv, _ := db.GetPHMeta(store.PHMetaSchemaVersion)
	if sv != "" && sv != fmt.Sprintf("%d", store.PHTablesSchemaVersion) {
		return configErr(fmt.Errorf("ph schema version %s does not match binary %d; restart backfill from scratch", sv, store.PHTablesSchemaVersion))
	}

	client := phgraphql.NewClient(cfg.AccessToken, userAgent())
	return executeBackfillLoop(cmd, flags, db, client, target)
}
