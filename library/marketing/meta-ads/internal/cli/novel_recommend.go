// Hand-authored novel-feature commands: recommend, apply, verify, history,
// capacity, decision-review, and honest stubs for features that depend on
// additional local state this iteration doesn't yet persist (fatigue, pace,
// learning, alerts, overlap, rollup).
//
// Not generated.

package cli

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/mvanhorn/printing-press-library/library/marketing/meta-ads/internal/attribution"
	"github.com/mvanhorn/printing-press-library/library/marketing/meta-ads/internal/client"
	"github.com/mvanhorn/printing-press-library/library/marketing/meta-ads/internal/decisions"
	"github.com/mvanhorn/printing-press-library/library/marketing/meta-ads/internal/store"

	"github.com/spf13/cobra"
)

// ---------- strategy configs (ported from magoosh-founder recommendations.py) ----------

type strategyConfig struct {
	roasThreshold       float64 // diff_pct to classify above/near/below vs roas_target
	capacityThreshold   float64
	saturationThreshold float64
	increasePct         float64 // default magnitude on INCREASE
	decreasePct         float64 // default magnitude on DECREASE
	maxChangePct        float64 // cap per decision
}

var strategyConfigs = map[string]strategyConfig{
	"conservative": {0.25, 0.25, 0.10, 0.10, 0.10, 0.15},
	"moderate":     {0.20, 0.20, 0.10, 0.15, 0.15, 0.25},
	"aggressive":   {0.15, 0.15, 0.10, 0.25, 0.20, 0.35},
}

// ---------- recommend ----------

// Recommendation is one campaign's decision envelope — fields match the shape
// Phase 5 dogfood expects and that agents narrow with --select.
type Recommendation struct {
	Platform       string            `json:"platform"`
	CampaignID     string            `json:"campaign_id"`
	CampaignName   string            `json:"campaign_name"`
	Action         decisions.Action  `json:"action"`
	CurrentBudget  float64           `json:"current_budget"`
	ProposedBudget float64           `json:"proposed_budget"`
	ChangePct      float64           `json:"change_pct"`
	Reasoning      map[string]string `json:"reasoning"`
	Confidence     string            `json:"confidence"`
	Warnings       []string          `json:"warnings,omitempty"`
	Issues         []string          `json:"issues,omitempty"`
	Metrics        map[string]any    `json:"metrics"`
}

func newRecommendCmd(flags *rootFlags) *cobra.Command {
	var days int
	var strategy string
	var roasTarget float64

	cmd := &cobra.Command{
		Use:   "recommend",
		Short: "Budget autopilot: ROAS + frequency + utilization decisioning",
		Long: `Run the recommendations engine over locally-synced insights.

For each campaign with spend in the last N days, classify the action:
  INCREASE — good ROAS AND capacity headroom
  DECREASE — poor ROAS (below target by strategy threshold)
  HOLD     — near target, or above target but saturated (no headroom)
  FLAG     — data quality problem, learning phase, or zero conversions

Strategy presets control thresholds and change magnitudes:
  conservative  — ROAS diff 25%, max change 15%
  moderate      — ROAS diff 20%, max change 25%  (default)
  aggressive    — ROAS diff 15%, max change 35%

The output is the full reasoning envelope so agents can decide which
recommendations to apply via 'apply --from-recommendations'.`,
		Example: `  # Default: last 14 days, moderate strategy, requires sync first
  meta-ads-pp-cli recommend

  # Aggressive scaling against a specific ROAS target, JSON for agent consumption
  meta-ads-pp-cli recommend --days 14 --strategy aggressive --roas-target 3.0 --agent

  # Narrow the output to the fields you need
  meta-ads-pp-cli recommend --json --select recommendations.campaign_name,recommendations.action,recommendations.proposed_budget,recommendations.confidence`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := strategyConfigs[strategy]; !ok {
				return usageErr(fmt.Errorf("invalid --strategy %q: must be conservative, moderate, or aggressive", strategy))
			}
			emptyResult := map[string]any{
				"recommendations":   []any{},
				"campaigns_held":    []any{},
				"campaigns_flagged": []any{},
				"summary":           "No insights in local store — run 'sync' and fetch campaign insights first.",
				"strategy":          strategy,
				"days":              days,
				"roas_target":       roasTarget,
			}
			db, err := openStoreForRead("meta-ads-pp-cli")
			if err != nil {
				return fmt.Errorf("opening local store: %w", err)
			}
			if db == nil {
				// Empty store: exit 0 with a clearly-empty envelope so agents
				// can detect the "nothing to do" case without a bare error.
				if flags.asJSON {
					return flags.printJSON(cmd, emptyResult)
				}
				fmt.Fprintln(cmd.OutOrStdout(), emptyResult["summary"])
				return nil
			}
			defer db.Close()

			insights, err := loadInsightsFromStore(db)
			if err != nil {
				return fmt.Errorf("reading insights: %w", err)
			}
			campaignsMap, err := loadCampaignsMap(db)
			if err != nil {
				return fmt.Errorf("reading campaigns: %w", err)
			}
			if len(insights) == 0 {
				if flags.asJSON {
					return flags.printJSON(cmd, emptyResult)
				}
				fmt.Fprintln(cmd.OutOrStdout(), emptyResult["summary"])
				return nil
			}

			cfg := strategyConfigs[strategy]
			result := struct {
				Recommendations  []Recommendation `json:"recommendations"`
				CampaignsHeld    []Recommendation `json:"campaigns_held"`
				CampaignsFlagged []Recommendation `json:"campaigns_flagged"`
				Summary          string           `json:"summary"`
				Strategy         string           `json:"strategy"`
				Days             int              `json:"days"`
				RoasTarget       float64          `json:"roas_target"`
			}{Strategy: strategy, Days: days, RoasTarget: roasTarget}

			for _, ins := range insights {
				campID, _ := ins["campaign_id"].(string)
				if campID == "" {
					continue
				}
				campName, _ := ins["campaign_name"].(string)
				if campName == "" {
					if c, ok := campaignsMap[campID]; ok {
						if n, _ := c["name"].(string); n != "" {
							campName = n
						}
					}
				}
				spend := parseFloat(ins["spend"])
				if spend <= 0 {
					continue
				}
				freq := parseFloat(ins["frequency"])
				haveFreq := ins["frequency"] != nil

				_, convValue, roasOverride, _ := attribution.ExtractPurchaseMetrics(ins)
				var roas float64
				if roasOverride > 0 && spend > 0 {
					roas = roasOverride
				} else if spend > 0 {
					roas = convValue / spend
				}

				// Pull current daily budget (cents) from campaign record
				var currentBudget float64
				if c, ok := campaignsMap[campID]; ok {
					if b := parseFloat(c["daily_budget"]); b > 0 {
						currentBudget = b / 100.0
					}
				}
				utilization := 0.0
				if currentBudget > 0 && days > 0 {
					utilization = spend / (currentBudget * float64(days))
				}

				cap := attribution.CapacitySignal(freq, haveFreq, utilization, "", 0, 0)

				rec := evaluateCampaign(cfg, roas, roasTarget, utilization, cap, currentBudget, roasOverride, convValue, spend)
				rec.Platform = "meta_ads"
				rec.CampaignID = campID
				rec.CampaignName = campName
				rec.Metrics = map[string]any{
					"spend":              spend,
					"roas":               roundFloat(roas, 3),
					"conversion_value":   roundFloat(convValue, 2),
					"frequency":          roundFloat(freq, 3),
					"budget_utilization": roundFloat(utilization, 3),
					"current_budget":     currentBudget,
				}

				switch rec.Action {
				case decisions.ActionIncrease, decisions.ActionDecrease:
					result.Recommendations = append(result.Recommendations, rec)
				case decisions.ActionHold:
					result.CampaignsHeld = append(result.CampaignsHeld, rec)
				case decisions.ActionFlag:
					result.CampaignsFlagged = append(result.CampaignsFlagged, rec)
				}
			}

			result.Summary = summarizeRecs(result.Recommendations, len(result.CampaignsHeld), len(result.CampaignsFlagged))

			if flags.asJSON {
				return flags.printJSON(cmd, result)
			}
			printRecsHuman(cmd, &result)
			return nil
		},
	}
	cmd.Flags().IntVar(&days, "days", 14, "Lookback window in days")
	cmd.Flags().StringVar(&strategy, "strategy", "moderate", "Decisioning strategy: conservative, moderate, aggressive")
	cmd.Flags().Float64Var(&roasTarget, "roas-target", 2.0, "Target ROAS; recommendations classify relative to this")
	return cmd
}

func evaluateCampaign(cfg strategyConfig, roas, target, utilization float64, cap attribution.Capacity, currentBudget, roasOverride, convValue, spend float64) Recommendation {
	rec := Recommendation{Reasoning: map[string]string{}, Confidence: "medium"}

	// Data-quality gates — emit FLAG early.
	if target <= 0 {
		rec.Action = decisions.ActionFlag
		rec.Issues = append(rec.Issues, "no ROAS target provided")
		rec.Reasoning["conclusion"] = "Need a --roas-target to classify ROAS"
		rec.Confidence = "unknown"
		return rec
	}
	if currentBudget <= 0 {
		rec.Action = decisions.ActionFlag
		rec.Issues = append(rec.Issues, "campaign has no daily_budget set locally — sync may be stale")
		rec.Reasoning["conclusion"] = "Cannot compute utilization without a current daily budget"
		rec.Confidence = "low"
		return rec
	}
	if convValue == 0 && roasOverride == 0 {
		rec.Action = decisions.ActionFlag
		rec.Issues = append(rec.Issues, "zero conversions with spend>0")
		rec.Reasoning["conclusion"] = "Zero attributed value — check pixel, attribution window, or purchase action dedup"
		rec.Confidence = "medium"
		return rec
	}

	diffPct := (roas - target) / target
	var efficiency string
	switch {
	case diffPct > cfg.roasThreshold:
		efficiency = "above"
	case diffPct < -cfg.roasThreshold:
		efficiency = "below"
	default:
		efficiency = "near"
	}
	rec.Reasoning["efficiency"] = fmt.Sprintf("ROAS %.2f vs target %.2f (%+.0f%%) — %s", roas, target, diffPct*100, efficiency)
	rec.Reasoning["utilization"] = fmt.Sprintf("budget utilization %.0f%%", utilization*100)
	rec.Reasoning["capacity"] = cap.Details

	// Meta-specific decision tree — utilization gates capacity.
	switch {
	case utilization < 0.70 && efficiency == "below":
		rec.Action = decisions.ActionDecrease
		rec.Confidence = "high"
		rec.Reasoning["conclusion"] = "Underutilized AND below target — scale budget down"
	case utilization < 0.70:
		rec.Action = decisions.ActionHold
		rec.Confidence = "high"
		rec.Reasoning["conclusion"] = "Underutilized but ROAS is near/above target — hold and watch pacing"
	case efficiency == "above" && cap.HasHeadroom:
		rec.Action = decisions.ActionIncrease
		rec.Confidence = cap.Confidence
		rec.Reasoning["conclusion"] = "Above target AND capacity headroom — scale budget up"
	case efficiency == "above" && !cap.HasHeadroom:
		rec.Action = decisions.ActionHold
		rec.Confidence = "medium"
		rec.Reasoning["conclusion"] = "Above target but saturated audience — more budget won't spend, expand audience instead"
	case efficiency == "near":
		rec.Action = decisions.ActionHold
		rec.Confidence = "medium"
		rec.Reasoning["conclusion"] = "Within ±" + strconv.FormatFloat(cfg.roasThreshold*100, 'f', 0, 64) + "% of target — hold"
	case efficiency == "below" && cap.HasHeadroom:
		rec.Action = decisions.ActionHold
		rec.Confidence = "medium"
		rec.Reasoning["conclusion"] = "Below target — improve efficiency (audience/creative) before scaling"
	case efficiency == "below" && !cap.HasHeadroom:
		rec.Action = decisions.ActionDecrease
		rec.Confidence = "medium"
		rec.Reasoning["conclusion"] = "Below target AND saturated — scale down"
	default:
		rec.Action = decisions.ActionHold
		rec.Confidence = "low"
		rec.Reasoning["conclusion"] = "Uncategorized — hold pending more data"
	}

	rec.CurrentBudget = currentBudget
	switch rec.Action {
	case decisions.ActionIncrease:
		scale := 1 + cfg.increasePct
		if scale-1 > cfg.maxChangePct {
			scale = 1 + cfg.maxChangePct
		}
		rec.ProposedBudget = roundDollar(currentBudget * scale)
	case decisions.ActionDecrease:
		scale := 1 - cfg.decreasePct
		floor := 1 - cfg.maxChangePct
		if scale < floor {
			scale = floor
		}
		rec.ProposedBudget = roundDollar(currentBudget * scale)
	default:
		rec.ProposedBudget = currentBudget
	}
	if currentBudget > 0 {
		rec.ChangePct = (rec.ProposedBudget - currentBudget) / currentBudget
	}
	if math.Abs(rec.ChangePct) > 0.20 {
		rec.Warnings = append(rec.Warnings, "change >20% may trigger Meta Smart Bidding learning period (7-14 days)")
	}
	return rec
}

// ---------- apply ----------

func newApplyCmd(flags *rootFlags) *cobra.Command {
	var fromRecommendations string
	var confirmation string

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Batch-apply budget changes with validation KPIs and follow-up log",
		Long: `Apply a batch of budget changes produced by 'recommend'.

Each change is logged to the local decision audit log with:
  - validation KPIs (what a successful outcome looks like for this action)
  - a follow_up_date (default 18 days for Meta; 14 if change > 20%)
  - the full reasoning envelope from the recommendation

--confirmation CONFIRM is required — there is no prompt. For CI/agents, pass
explicitly. Set --dry-run to preview the API calls without sending.`,
		Example: `  # Preview without applying
  meta-ads-pp-cli recommend --json > preview.json
  meta-ads-pp-cli apply --from-recommendations preview.json --confirmation CONFIRM --dry-run

  # Apply for real
  meta-ads-pp-cli apply --from-recommendations preview.json --confirmation CONFIRM`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// No inputs at all: show help so agents see the required flags
			// before hitting a bare error.
			if fromRecommendations == "" && confirmation == "" {
				return cmd.Help()
			}
			if confirmation != "CONFIRM" {
				return usageErr(fmt.Errorf("--confirmation CONFIRM is required to apply"))
			}
			if fromRecommendations == "" {
				return usageErr(fmt.Errorf("--from-recommendations <path> is required (use 'recommend --json > preview.json' first)"))
			}

			raw, err := os.ReadFile(fromRecommendations)
			if err != nil {
				return fmt.Errorf("reading %s: %w", fromRecommendations, err)
			}
			var preview struct {
				Recommendations []Recommendation `json:"recommendations"`
				Strategy        string           `json:"strategy"`
			}
			if err := json.Unmarshal(raw, &preview); err != nil {
				return fmt.Errorf("parsing recommendations file: %w", err)
			}
			if len(preview.Recommendations) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No recommendations to apply.")
				return nil
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			db, err := openStoreForRead("meta-ads-pp-cli")
			if err != nil {
				return fmt.Errorf("opening local store: %w", err)
			}
			if db != nil {
				defer db.Close()
				if err := decisions.Schema(db.DB()); err != nil {
					return fmt.Errorf("preparing decisions table: %w", err)
				}
			}

			var applied, failed int
			results := make([]map[string]any, 0, len(preview.Recommendations))
			for _, rec := range preview.Recommendations {
				path := "/" + rec.CampaignID
				body := map[string]any{"daily_budget": int64(rec.ProposedBudget * 100)}
				var success bool
				var errMsg string

				_, statusCode, postErr := c.Post(path, body)
				if postErr != nil {
					errMsg = postErr.Error()
				} else {
					success = statusCode >= 200 && statusCode < 300
					if !success {
						errMsg = fmt.Sprintf("HTTP %d", statusCode)
					}
				}
				if flags.dryRun {
					success = true // dry-run doesn't actually hit the API
				}

				if success {
					applied++
				} else {
					failed++
				}

				// Log the decision regardless — audit the attempt.
				if db != nil && !flags.dryRun {
					log := decisions.Open(db.DB())
					reasoningJSON, _ := json.Marshal(rec.Reasoning)
					metricsJSON, _ := json.Marshal(rec.Metrics)
					warningsJSON, _ := json.Marshal(rec.Warnings)
					kpis := buildValidationKPIs(rec.Action, rec.ChangePct)
					kpisJSON, _ := json.Marshal(kpis)
					followUp := decisions.CalcFollowUp("meta_ads", rec.ChangePct)
					_ = log.Append(&decisions.Entry{
						EntryType:       decisions.EntryBudgetDecision,
						Platform:        "meta_ads",
						CampaignID:      rec.CampaignID,
						CampaignName:    rec.CampaignName,
						Action:          rec.Action,
						OldBudget:       rec.CurrentBudget,
						NewBudget:       rec.ProposedBudget,
						ChangePct:       rec.ChangePct,
						WasApplied:      success,
						Confidence:      rec.Confidence,
						Strategy:        preview.Strategy,
						Reasoning:       reasoningJSON,
						ValidationKPIs:  kpisJSON,
						ExpectedOutcome: rec.Reasoning["conclusion"],
						FollowUpDate:    &followUp,
						Warnings:        warningsJSON,
						MetricsSnapshot: metricsJSON,
					})
				}

				results = append(results, map[string]any{
					"campaign_id":   rec.CampaignID,
					"campaign_name": rec.CampaignName,
					"action":        rec.Action,
					"old_budget":    rec.CurrentBudget,
					"new_budget":    rec.ProposedBudget,
					"success":       success,
					"error":         errMsg,
					"dry_run":       flags.dryRun,
				})
			}

			out := map[string]any{
				"applied": applied,
				"failed":  failed,
				"dry_run": flags.dryRun,
				"results": results,
			}
			if flags.asJSON {
				return flags.printJSON(cmd, out)
			}
			for _, r := range results {
				status := green("OK")
				if !r["success"].(bool) {
					status = red("FAIL")
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s — %s: $%.2f → $%.2f (%s)\n",
					status, r["campaign_name"], r["action"], r["old_budget"], r["new_budget"], r["error"])
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d applied, %d failed (dry_run=%v)\n", applied, failed, flags.dryRun)
			return nil
		},
	}
	cmd.Flags().StringVar(&fromRecommendations, "from-recommendations", "", "Path to JSON file produced by 'recommend --json'")
	cmd.Flags().StringVar(&confirmation, "confirmation", "", "Must be exactly CONFIRM to apply")
	return cmd
}

func buildValidationKPIs(action decisions.Action, changePct float64) []map[string]any {
	// KPI shapes ported from magoosh-founder decision_log.py.
	switch action {
	case decisions.ActionIncrease:
		return []map[string]any{
			{"metric": "roas", "operator": ">=", "target_multiplier": 0.85, "description": "ROAS stays within 85% of target"},
			{"metric": "spend_increase", "operator": ">=", "target_ratio": 0.50, "description": "Spend absorbs at least 50% of budget increase"},
			{"metric": "conversion_value_growth", "operator": ">=", "target_ratio": 0.70, "description": "Conversion value grows at least 70% of budget %"},
			{"metric": "lost_is_budget", "operator": "<", "description": "Capacity signal improves"},
		}
	case decisions.ActionDecrease:
		return []map[string]any{
			{"metric": "roas_gap", "operator": "<=", "target_ratio": 0.90, "description": "ROAS closes at least 10% of gap to target"},
			{"metric": "conversion_drop", "operator": "<=", "target_multiplier": 1.5, "description": "Conversion drop ≤ 1.5× budget decrease"},
			{"metric": "cpa", "operator": "<=", "target_ratio": 0.95, "description": "CPA improves by ≥5%"},
		}
	case decisions.ActionHold:
		return []map[string]any{
			{"metric": "roas", "operator": "stable", "target_ratio": 0.20, "description": "ROAS stays within ±20%"},
			{"metric": "spend", "operator": "stable", "target_ratio": 0.25, "description": "Spend variance < 25%"},
		}
	case decisions.ActionFlag:
		return []map[string]any{
			{"metric": "issues_resolved", "operator": "=", "target": true, "description": "Flagged issues are resolved before re-evaluation"},
		}
	}
	return nil
}

// ---------- verify roas ----------

func newVerifyCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Data-integrity checks against local insights",
	}
	cmd.AddCommand(newVerifyRoasCmd(flags))
	return cmd
}

func newVerifyRoasCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "roas",
		Short: "Detect omni_purchase vs purchase double-counting in local insights",
		Long: `Meta sometimes returns both 'omni_purchase' and 'purchase' for the same
campaign. Dashboards that sum across action_types overstate ROAS.

This command reads locally-synced insights, computes the deduplicated
attributed value using the PurchaseActionPriority list
('omni_purchase' wins over 'purchase' wins over pixel events), and
reports campaigns where the legacy summed value differs by more than
the delta threshold.`,
		Example: `  meta-ads-pp-cli verify roas
  meta-ads-pp-cli verify roas --json --select rows.campaign_name,rows.roas_deduplicated,rows.roas_legacy,rows.legacy_delta_pct`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openStoreForRead("meta-ads-pp-cli")
			if err != nil {
				return fmt.Errorf("opening local store: %w", err)
			}
			if db == nil {
				return usageErr(fmt.Errorf("local store is empty — run 'sync' and fetch insights first"))
			}
			defer db.Close()

			insights, err := loadInsightsFromStore(db)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(insights))
			for _, ins := range insights {
				_, dedupValue, roasOverride, selected := attribution.ExtractPurchaseMetrics(ins)
				spend := parseFloat(ins["spend"])
				if spend <= 0 {
					continue
				}
				legacyValue := attribution.LegacyCombinedValue(ins)
				roasDedup := 0.0
				if roasOverride > 0 {
					roasDedup = roasOverride
				} else {
					roasDedup = dedupValue / spend
				}
				roasLegacy := 0.0
				if spend > 0 {
					roasLegacy = legacyValue / spend
				}
				var deltaPct float64
				if dedupValue > 0 {
					deltaPct = (legacyValue - dedupValue) / dedupValue
				}
				rows = append(rows, map[string]any{
					"campaign_id":        ins["campaign_id"],
					"campaign_name":      ins["campaign_name"],
					"selected_action":    selected,
					"spend":              roundFloat(spend, 2),
					"value_deduplicated": roundFloat(dedupValue, 2),
					"value_legacy":       roundFloat(legacyValue, 2),
					"roas_deduplicated":  roundFloat(roasDedup, 3),
					"roas_legacy":        roundFloat(roasLegacy, 3),
					"legacy_delta_pct":   roundFloat(deltaPct, 3),
					"double_count_risk":  math.Abs(deltaPct) > 0.01,
				})
			}
			out := map[string]any{
				"rows":          rows,
				"total_rows":    len(rows),
				"flagged_count": countWhere(rows, func(r map[string]any) bool { return r["double_count_risk"].(bool) }),
			}
			if flags.asJSON {
				return flags.printJSON(cmd, out)
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No insights to verify. Run 'sync' and fetch insights first.")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-40s  %-14s  %10s  %10s  %10s\n", "CAMPAIGN", "SELECTED", "ROAS_DEDUP", "ROAS_LEGACY", "DELTA_%")
			for _, r := range rows {
				name := truncate(fmt.Sprint(r["campaign_name"]), 40)
				flag := ""
				if r["double_count_risk"].(bool) {
					flag = red(" ⚠")
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-40s  %-14s  %10.3f  %10.3f  %+9.1f%%%s\n",
					name, truncate(fmt.Sprint(r["selected_action"]), 14),
					r["roas_deduplicated"], r["roas_legacy"],
					parseFloat(r["legacy_delta_pct"])*100, flag)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d rows, %d flagged (|delta| > 1%%)\n", out["total_rows"], out["flagged_count"])
			return nil
		},
	}
}

// ---------- capacity ----------

func newCapacityCmd(flags *rootFlags) *cobra.Command {
	var days int
	return &cobra.Command{
		Use:   "capacity",
		Short: "Per-campaign frequency capacity signal (headroom/confidence/details)",
		Long: `Compute the frequency-based capacity signal for each campaign with spend.

Decision order:
  no freq → unknown
  delivery flag 'spending_limited' → headroom high (constrained by budget)
  delivery flag 'paused' → no headroom low
  utilization < 70% → no headroom low (not budget-constrained)
  frequency > 3.5 → no headroom high (saturated)
  frequency ≥ 2.5 → headroom medium (approaching saturation)
  else → headroom high (healthy)`,
		Example: `  meta-ads-pp-cli capacity --days 7
  meta-ads-pp-cli capacity --agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openStoreForRead("meta-ads-pp-cli")
			if err != nil {
				return err
			}
			if db == nil {
				return usageErr(fmt.Errorf("local store is empty — run 'sync' first"))
			}
			defer db.Close()

			insights, _ := loadInsightsFromStore(db)
			campaigns, _ := loadCampaignsMap(db)
			rows := make([]map[string]any, 0, len(insights))
			for _, ins := range insights {
				campID, _ := ins["campaign_id"].(string)
				if campID == "" {
					continue
				}
				freq := parseFloat(ins["frequency"])
				spend := parseFloat(ins["spend"])
				haveFreq := ins["frequency"] != nil
				if spend <= 0 {
					continue
				}
				var currentBudget float64
				if c, ok := campaigns[campID]; ok {
					currentBudget = parseFloat(c["daily_budget"]) / 100.0
				}
				utilization := 0.0
				if currentBudget > 0 && days > 0 {
					utilization = spend / (currentBudget * float64(days))
				}
				cap := attribution.CapacitySignal(freq, haveFreq, utilization, "", 0, 0)
				rows = append(rows, map[string]any{
					"campaign_id":   campID,
					"campaign_name": ins["campaign_name"],
					"frequency":     roundFloat(freq, 3),
					"utilization":   roundFloat(utilization, 3),
					"has_headroom":  cap.HasHeadroom,
					"confidence":    cap.Confidence,
					"details":       cap.Details,
				})
			}
			out := map[string]any{"rows": rows, "days": days}
			if flags.asJSON {
				return flags.printJSON(cmd, out)
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No campaigns with insight data. Run 'sync' and fetch campaign insights first.")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-35s %8s %8s %10s %-8s  %s\n",
				"CAMPAIGN", "FREQ", "UTIL%", "HEADROOM", "CONF", "DETAILS")
			for _, r := range rows {
				headroom := red("no")
				if r["has_headroom"].(bool) {
					headroom = green("yes")
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-35s %8.2f %7.0f%% %10s %-8s  %s\n",
					truncate(fmt.Sprint(r["campaign_name"]), 35),
					r["frequency"], parseFloat(r["utilization"])*100, headroom, r["confidence"], r["details"])
			}
			return nil
		},
	}
}

// ---------- history ----------

func newHistoryCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Decision audit log: list, search, due-for-follow-up, and decision-review",
	}
	cmd.AddCommand(newHistoryListCmd(flags))
	cmd.AddCommand(newHistorySearchCmd(flags))
	cmd.AddCommand(newHistoryDueCmd(flags))
	cmd.AddCommand(newHistoryReviewCmd(flags))
	cmd.AddCommand(newHistoryGetCmd(flags))
	return cmd
}

func newHistoryListCmd(flags *rootFlags) *cobra.Command {
	var days, limit int
	var entryType string
	var platform string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent decisions",
		Example: `  meta-ads-pp-cli history list --days 30 --json
  meta-ads-pp-cli history list --entry-type status_change`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openOrInitDecisionsDB()
			if err != nil {
				return err
			}
			defer db.Close()
			log := decisions.Open(db.DB())
			entries, err := log.ListRecent(decisions.EntryType(entryType), days, limit)
			if err != nil {
				return err
			}
			return printEntries(cmd, flags, filterPlatform(entries, platform))
		},
	}
	cmd.Flags().IntVar(&days, "days", 30, "Lookback window in days")
	cmd.Flags().IntVar(&limit, "limit", 100, "Max entries to return")
	cmd.Flags().StringVar(&entryType, "entry-type", "", "Filter by entry_type (budget_decision, status_change, decision_analysis)")
	cmd.Flags().StringVar(&platform, "platform", "", "Filter by platform (e.g. meta_ads)")
	return cmd
}

func newHistorySearchCmd(flags *rootFlags) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search the decision log (FTS5)",
		Args:  cobra.MinimumNArgs(1),
		Example: `  meta-ads-pp-cli history search "saturation"
  meta-ads-pp-cli history search "freq > 3" --limit 20 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openOrInitDecisionsDB()
			if err != nil {
				return err
			}
			defer db.Close()
			log := decisions.Open(db.DB())
			entries, err := log.Search(strings.Join(args, " "), limit)
			if err != nil {
				return err
			}
			return printEntries(cmd, flags, entries)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Max entries to return")
	return cmd
}

func newHistoryDueCmd(flags *rootFlags) *cobra.Command {
	var platform string
	cmd := &cobra.Command{
		Use:     "due",
		Short:   "Applied decisions whose follow-up window has passed and have no analysis yet",
		Example: `  meta-ads-pp-cli history due --platform meta_ads --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openOrInitDecisionsDB()
			if err != nil {
				return err
			}
			defer db.Close()
			log := decisions.Open(db.DB())
			entries, err := log.Due(platform)
			if err != nil {
				return err
			}
			return printEntries(cmd, flags, entries)
		},
	}
	cmd.Flags().StringVar(&platform, "platform", "meta_ads", "Platform filter (default meta_ads; empty string = all)")
	return cmd
}

func newHistoryGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <log-id>",
		Short: "Fetch one decision log entry by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openOrInitDecisionsDB()
			if err != nil {
				return err
			}
			defer db.Close()
			log := decisions.Open(db.DB())
			e, err := log.GetByID(args[0])
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return notFoundErr(fmt.Errorf("no decision with log_id %s", args[0]))
				}
				return err
			}
			return printEntries(cmd, flags, []*decisions.Entry{e})
		},
	}
}

func newHistoryReviewCmd(flags *rootFlags) *cobra.Command {
	var outcome, summary, observation, hypothesis string
	cmd := &cobra.Command{
		Use:   "review <log-id>",
		Short: "Attach a post-mortem analysis to a past decision",
		Long: `Write a decision_analysis entry linking back to the original decision.
Use this after a follow-up review to capture what actually happened so
future recommendations can learn from past outcomes.`,
		Args: cobra.ExactArgs(1),
		Example: `  meta-ads-pp-cli history review <log-id> --outcome partial \
      --summary "ROAS lifted 8% but spend lagged" \
      --observation "Audience too narrow; spend stuck at 60% utilization" \
      --hypothesis "Expand lookalike ratio from 1% to 2%"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if outcome == "" {
				// No outcome provided: show help so the user sees the allowed
				// values before hitting a bare flag error.
				fmt.Fprintln(cmd.ErrOrStderr(), "--outcome required (success|partial|failure|inconclusive)")
				return cmd.Help()
			}
			switch decisions.Outcome(outcome) {
			case decisions.OutcomeSuccess, decisions.OutcomePartial, decisions.OutcomeFailure, decisions.OutcomeInconclusive:
			default:
				return usageErr(fmt.Errorf("invalid --outcome %q: must be success, partial, failure, or inconclusive", outcome))
			}
			db, err := openOrInitDecisionsDB()
			if err != nil {
				return err
			}
			defer db.Close()
			log := decisions.Open(db.DB())
			orig, err := log.GetByID(args[0])
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return notFoundErr(fmt.Errorf("no decision with log_id %s", args[0]))
				}
				return err
			}
			entry := &decisions.Entry{
				EntryType:       decisions.EntryAnalysis,
				Platform:        orig.Platform,
				CampaignID:      orig.CampaignID,
				CampaignName:    orig.CampaignName,
				OriginalLogID:   orig.LogID,
				AnalysisOutcome: decisions.Outcome(outcome),
				OutcomeSummary:  summary,
				Observation:     observation,
				Hypothesis:      hypothesis,
			}
			if err := log.Append(entry); err != nil {
				return err
			}
			if flags.asJSON {
				return flags.printJSON(cmd, map[string]any{
					"log_id":          entry.LogID,
					"original_log_id": orig.LogID,
					"outcome":         outcome,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s analysis attached to %s (outcome=%s)\n",
				green("OK"), orig.LogID, outcome)
			return nil
		},
	}
	cmd.Flags().StringVar(&outcome, "outcome", "", "success | partial | failure | inconclusive (required)")
	cmd.Flags().StringVar(&summary, "summary", "", "One-line outcome summary")
	cmd.Flags().StringVar(&observation, "observation", "", "What you observed")
	cmd.Flags().StringVar(&hypothesis, "hypothesis", "", "What you'll try next")
	return cmd
}

// ---------- deferred transcendence stubs (shipping-scope honest placeholders) ----------
// These were approved in Phase 1.5; they ship as stubs in this iteration because
// they require additional local state (hourly insights, creative timelines,
// audience targeting diffs) that a single /sync pass doesn't yet persist.
// Each stub documents exactly what's missing and offers the data path the user
// can use today.

func newFatigueCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "fatigue",
		Short: "Detect creative fatigue (rising CPM × rising freq × falling CTR) (stub)",
		Long: `Detect creative fatigue: ads where CPM and frequency rise while CTR falls
across a rolling window.

Approved shipping scope from the absorb manifest; ships as a structural
placeholder in this iteration because it requires multi-day insights stored
across a rolling window. The generator's /sync captures the most-recent
date-preset window only.

Today, reproduce the intent manually with 'query':
  meta-ads-pp-cli query "SELECT json_extract(data,'$.ad_name'), json_extract(data,'$.cpm'), json_extract(data,'$.frequency'), json_extract(data,'$.ctr') FROM insights WHERE json_extract(data,'$.ad_id') IS NOT NULL ORDER BY CAST(json_extract(data,'$.cpm') AS REAL) DESC"`,
		Example: "  meta-ads-pp-cli fatigue",
		RunE:    stubRun("fatigue"),
	}
}

func newPaceCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "pace",
		Short: "Budget pacing monitor: hourly burn rate to ETA-to-cap (stub)",
		Long: `Hourly spend-rate vs daily_budget with ETA-to-cap per campaign.

Approved shipping scope from the absorb manifest; ships as a structural
placeholder in this iteration because it requires hourly insights with a
minute-level time_increment. Current /sync uses day-level granularity.

Today, check current-day spend via:
  meta-ads-pp-cli campaigns insights <campaign_id> --date-preset today`,
		Example: "  meta-ads-pp-cli pace",
		RunE:    stubRun("pace"),
	}
}

func newLearningCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "learning",
		Short: "Identify campaigns in the Meta Smart Bidding learning phase (stub)",
		Long: `List campaigns detected in the Smart Bidding learning phase (recent large
budget change OR status_issues containing "learning").

Approved shipping scope from the absorb manifest; ships as a structural
placeholder in this iteration.

Today:
  meta-ads-pp-cli history list --days 14  (surfaces recent budget changes; filter change_pct > 0.20)`,
		Example: "  meta-ads-pp-cli learning --days 14",
		RunE:    stubRun("learning"),
	}
}

func newOverlapCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "overlap <adset_id_a> <adset_id_b>",
		Short: "Audience overlap analysis between two ad sets (stub)",
		Long: `Compute the targeting overlap between two ad sets on interests, geo, age,
and behaviors. Approved shipping scope; ships as a structural placeholder.

Today:
  meta-ads-pp-cli adsets get <id_a> --select targeting
  meta-ads-pp-cli adsets get <id_b> --select targeting`,
		Example: "  meta-ads-pp-cli overlap 120123 120456",
		RunE:    stubRun("overlap"),
	}
}

func newAlertsCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "alerts",
		Short: "Threshold watchers against local data (ROAS min, freq max) (stub)",
		Long: `Threshold watchers that check the local store and print offenders.
Approved shipping scope; ships as a structural placeholder.

Today, encode your thresholds in a query:
  meta-ads-pp-cli query "SELECT json_extract(data,'$.campaign_name'), CAST(json_extract(data,'$.frequency') AS REAL) AS freq FROM insights WHERE CAST(json_extract(data,'$.frequency') AS REAL) > 3.5"`,
		Example: "  meta-ads-pp-cli alerts --roas-min 1.5 --freq-max 3.5",
		RunE:    stubRun("alerts"),
	}
}

func newRollupCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "rollup",
		Short: "Aggregate spend/ROAS/conversions across multiple ad accounts (stub)",
		Long: `Aggregate metrics across multiple ad accounts synced into the same local
store. Approved shipping scope; ships as a structural placeholder.

Today, run 'insights account <id>' against each account separately.`,
		Example: "  meta-ads-pp-cli rollup --accounts act_111,act_222",
		RunE:    stubRun("rollup"),
	}
}

func stubRun(name string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		out := map[string]any{
			"status":  "stub",
			"command": name,
			"message": fmt.Sprintf("'%s' is an approved-scope stub in this iteration. See '--help' for the manual workaround.", name),
		}
		// Emit JSON when --json or --agent is set; otherwise human-friendly stderr message.
		if j, _ := cmd.Flags().GetBool("json"); j {
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}
		if a, _ := cmd.Flags().GetBool("agent"); a {
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "%s — stub in this iteration. See '%s --help' for the manual workaround.\n", yellow(name), name)
		return nil
	}
}

// ---------- helpers ----------

func openOrInitDecisionsDB() (*store.Store, error) {
	dbPath := defaultDBPath("meta-ads-pp-cli")
	db, err := store.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening store at %s: %w", dbPath, err)
	}
	if err := decisions.Schema(db.DB()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("preparing decisions schema: %w", err)
	}
	return db, nil
}

func loadInsightsFromStore(db *store.Store) ([]map[string]any, error) {
	// The generator's sync writes insights into the resources table with a "insights" type
	// OR into a per-resource insights table. Both paths write raw JSON.
	raws, err := db.List("insights", 10000)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(raws))
	for _, r := range raws {
		var m map[string]any
		if err := json.Unmarshal(r, &m); err == nil {
			out = append(out, m)
		}
	}
	return out, nil
}

func loadCampaignsMap(db *store.Store) (map[string]map[string]any, error) {
	raws, err := db.List("campaigns", 10000)
	if err != nil {
		return nil, err
	}
	out := make(map[string]map[string]any, len(raws))
	for _, r := range raws {
		var m map[string]any
		if err := json.Unmarshal(r, &m); err == nil {
			id, _ := m["id"].(string)
			if id == "" {
				continue
			}
			out[id] = m
		}
	}
	return out, nil
}

func parseFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	}
	return 0
}

func roundFloat(v float64, places int) float64 {
	m := math.Pow10(places)
	return math.Round(v*m) / m
}

func roundDollar(v float64) float64 { return math.Round(v) }

func countWhere(rows []map[string]any, fn func(map[string]any) bool) int {
	n := 0
	for _, r := range rows {
		if fn(r) {
			n++
		}
	}
	return n
}

func summarizeRecs(recs []Recommendation, held, flagged int) string {
	inc, dec := 0, 0
	var incBudget, decBudget float64
	for _, r := range recs {
		switch r.Action {
		case decisions.ActionIncrease:
			inc++
			incBudget += r.ProposedBudget - r.CurrentBudget
		case decisions.ActionDecrease:
			dec++
			decBudget += r.CurrentBudget - r.ProposedBudget
		}
	}
	parts := []string{}
	if inc > 0 {
		parts = append(parts, fmt.Sprintf("%d increase(s) (+$%.0f/day)", inc, incBudget))
	}
	if dec > 0 {
		parts = append(parts, fmt.Sprintf("%d decrease(s) (-$%.0f/day)", dec, decBudget))
	}
	if held > 0 {
		parts = append(parts, fmt.Sprintf("%d held", held))
	}
	if flagged > 0 {
		parts = append(parts, fmt.Sprintf("%d flagged", flagged))
	}
	if len(parts) == 0 {
		return "No actionable campaigns."
	}
	return strings.Join(parts, ", ")
}

func printRecsHuman(cmd *cobra.Command, result *struct {
	Recommendations  []Recommendation `json:"recommendations"`
	CampaignsHeld    []Recommendation `json:"campaigns_held"`
	CampaignsFlagged []Recommendation `json:"campaigns_flagged"`
	Summary          string           `json:"summary"`
	Strategy         string           `json:"strategy"`
	Days             int              `json:"days"`
	RoasTarget       float64          `json:"roas_target"`
}) {
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Strategy: %s | ROAS target: %.2f | Window: %d days\n\n", bold(result.Strategy), result.RoasTarget, result.Days)
	all := append([]Recommendation{}, result.Recommendations...)
	all = append(all, result.CampaignsHeld...)
	all = append(all, result.CampaignsFlagged...)
	if len(all) == 0 {
		fmt.Fprintln(w, "No campaigns with spend in the window.")
		return
	}
	fmt.Fprintf(w, "%-35s %-10s %8s %8s %7s  %-8s  %s\n",
		"CAMPAIGN", "ACTION", "CURRENT", "PROPOSED", "Δ%", "CONF", "CONCLUSION")
	for _, r := range all {
		action := string(r.Action)
		switch r.Action {
		case decisions.ActionIncrease:
			action = green("INCREASE")
		case decisions.ActionDecrease:
			action = red("DECREASE")
		case decisions.ActionFlag:
			action = yellow("FLAG")
		}
		fmt.Fprintf(w, "%-35s %-10s %8.2f %8.2f %+6.0f%%  %-8s  %s\n",
			truncate(r.CampaignName, 35), action, r.CurrentBudget, r.ProposedBudget, r.ChangePct*100,
			r.Confidence, truncate(r.Reasoning["conclusion"], 60))
	}
	fmt.Fprintf(w, "\n%s\n", bold(result.Summary))
}

func printEntries(cmd *cobra.Command, flags *rootFlags, entries []*decisions.Entry) error {
	if flags.asJSON {
		return flags.printJSON(cmd, map[string]any{"entries": entries, "count": len(entries)})
	}
	w := cmd.OutOrStdout()
	if len(entries) == 0 {
		fmt.Fprintln(w, "No decisions.")
		return nil
	}
	fmt.Fprintf(w, "%-36s %-18s %-10s %-20s %s\n", "LOG_ID", "TYPE", "ACTION", "TIMESTAMP", "CAMPAIGN")
	for _, e := range entries {
		ts := e.Timestamp.Format("2026-01-02 15:04")
		action := string(e.Action)
		if e.EntryType == decisions.EntryAnalysis {
			action = string(e.AnalysisOutcome)
		} else if e.EntryType == decisions.EntryStatusChange {
			action = e.OldStatus + "→" + e.NewStatus
		}
		fmt.Fprintf(w, "%-36s %-18s %-10s %-20s %s\n",
			e.LogID, string(e.EntryType), action, ts, truncate(e.CampaignName, 40))
	}
	return nil
}

func filterPlatform(entries []*decisions.Entry, platform string) []*decisions.Entry {
	if platform == "" {
		return entries
	}
	out := make([]*decisions.Entry, 0, len(entries))
	for _, e := range entries {
		if e.Platform == platform {
			out = append(out, e)
		}
	}
	return out
}

// Compile-time guard: ensure client package is used (referenced via flags.newClient in apply).
var _ = (*client.Client)(nil)

// bufio used by auth_login but import must be present elsewhere; touch here to keep analyzers happy.
var _ = bufio.NewReader
