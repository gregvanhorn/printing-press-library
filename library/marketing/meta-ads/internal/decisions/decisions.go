// Package decisions implements the budget-decision audit log.
// Entries are append-only. Three entry types: budget_decision, status_change, decision_analysis.
package decisions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EntryType distinguishes log entry categories.
type EntryType string

const (
	EntryBudgetDecision EntryType = "budget_decision"
	EntryStatusChange   EntryType = "status_change"
	EntryAnalysis       EntryType = "decision_analysis"
)

// Action is the decision outcome on a budget_decision row.
type Action string

const (
	ActionIncrease Action = "INCREASE"
	ActionDecrease Action = "DECREASE"
	ActionHold     Action = "HOLD"
	ActionFlag     Action = "FLAG"
)

// Outcome is the analysis verdict when a decision has been reviewed after its follow-up window.
type Outcome string

const (
	OutcomeSuccess      Outcome = "success"
	OutcomePartial      Outcome = "partial"
	OutcomeFailure      Outcome = "failure"
	OutcomeInconclusive Outcome = "inconclusive"
)

// Entry is a single row in the log. Nullable fields are only set for the matching EntryType.
type Entry struct {
	LogID           string          `json:"log_id"`
	EntryType       EntryType       `json:"entry_type"`
	Timestamp       time.Time       `json:"timestamp"`
	Platform        string          `json:"platform"`
	CampaignID      string          `json:"campaign_id,omitempty"`
	CampaignName    string          `json:"campaign_name,omitempty"`
	Action          Action          `json:"action,omitempty"`
	OldBudget       float64         `json:"old_budget,omitempty"`
	NewBudget       float64         `json:"new_budget,omitempty"`
	ChangePct       float64         `json:"change_pct,omitempty"`
	WasApplied      bool            `json:"was_applied"`
	Confidence      string          `json:"confidence,omitempty"`
	Strategy        string          `json:"strategy,omitempty"`
	Reasoning       json.RawMessage `json:"reasoning,omitempty"`
	ValidationKPIs  json.RawMessage `json:"validation_kpis,omitempty"`
	ExpectedOutcome string          `json:"expected_outcome,omitempty"`
	FollowUpDate    *time.Time      `json:"follow_up_date,omitempty"`
	Warnings        json.RawMessage `json:"warnings,omitempty"`
	MetricsSnapshot json.RawMessage `json:"metrics_snapshot,omitempty"`
	SessionID       string          `json:"session_id,omitempty"`
	// status_change fields
	OldStatus string `json:"old_status,omitempty"`
	NewStatus string `json:"new_status,omitempty"`
	Reason    string `json:"reason,omitempty"`
	// decision_analysis fields
	OriginalLogID   string          `json:"original_log_id,omitempty"`
	AnalysisOutcome Outcome         `json:"outcome,omitempty"`
	OutcomeSummary  string          `json:"outcome_summary,omitempty"`
	KPIResults      json.RawMessage `json:"kpi_results,omitempty"`
	Observation     string          `json:"observation,omitempty"`
	Hypothesis      string          `json:"hypothesis,omitempty"`
}

// Log is the SQLite-backed decision log.
type Log struct {
	db *sql.DB
}

// Schema creates the decisions table + FTS index if absent. Idempotent.
func Schema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS decisions (
			log_id TEXT PRIMARY KEY,
			entry_type TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			platform TEXT NOT NULL,
			campaign_id TEXT,
			campaign_name TEXT,
			action TEXT,
			old_budget REAL,
			new_budget REAL,
			change_pct REAL,
			was_applied INTEGER NOT NULL DEFAULT 0,
			confidence TEXT,
			strategy TEXT,
			reasoning TEXT,
			validation_kpis TEXT,
			expected_outcome TEXT,
			follow_up_date TEXT,
			warnings TEXT,
			metrics_snapshot TEXT,
			session_id TEXT,
			old_status TEXT,
			new_status TEXT,
			reason TEXT,
			original_log_id TEXT,
			analysis_outcome TEXT,
			outcome_summary TEXT,
			kpi_results TEXT,
			observation TEXT,
			hypothesis TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_decisions_campaign ON decisions(campaign_id)`,
		`CREATE INDEX IF NOT EXISTS idx_decisions_followup ON decisions(follow_up_date) WHERE was_applied = 1 AND entry_type = 'budget_decision'`,
		`CREATE INDEX IF NOT EXISTS idx_decisions_original ON decisions(original_log_id) WHERE original_log_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_decisions_timestamp ON decisions(timestamp DESC)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS decisions_fts USING fts5(log_id UNINDEXED, reasoning, observation, hypothesis, campaign_name, reason, content='decisions', content_rowid='rowid')`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("decisions schema: %w", err)
		}
	}
	return nil
}

// Open binds a Log to an open DB. Call Schema once per DB before using.
func Open(db *sql.DB) *Log {
	return &Log{db: db}
}

// Append writes a new entry. Generates LogID if empty; sets Timestamp to now if zero.
func (l *Log) Append(e *Entry) error {
	if e.LogID == "" {
		e.LogID = uuid.NewString()
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	if e.EntryType == "" {
		e.EntryType = EntryBudgetDecision
	}

	appliedInt := 0
	if e.WasApplied {
		appliedInt = 1
	}
	var followUp *string
	if e.FollowUpDate != nil {
		s := e.FollowUpDate.Format(time.RFC3339)
		followUp = &s
	}

	_, err := l.db.Exec(`INSERT INTO decisions (
		log_id, entry_type, timestamp, platform, campaign_id, campaign_name,
		action, old_budget, new_budget, change_pct, was_applied, confidence, strategy,
		reasoning, validation_kpis, expected_outcome, follow_up_date, warnings, metrics_snapshot, session_id,
		old_status, new_status, reason,
		original_log_id, analysis_outcome, outcome_summary, kpi_results, observation, hypothesis
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		e.LogID, e.EntryType, e.Timestamp.Format(time.RFC3339), e.Platform,
		nullIfEmpty(e.CampaignID), nullIfEmpty(e.CampaignName),
		nullIfEmpty(string(e.Action)), nullIfZero(e.OldBudget), nullIfZero(e.NewBudget), nullIfZero(e.ChangePct),
		appliedInt, nullIfEmpty(e.Confidence), nullIfEmpty(e.Strategy),
		nullIfEmptyRaw(e.Reasoning), nullIfEmptyRaw(e.ValidationKPIs), nullIfEmpty(e.ExpectedOutcome),
		followUp, nullIfEmptyRaw(e.Warnings), nullIfEmptyRaw(e.MetricsSnapshot), nullIfEmpty(e.SessionID),
		nullIfEmpty(e.OldStatus), nullIfEmpty(e.NewStatus), nullIfEmpty(e.Reason),
		nullIfEmpty(e.OriginalLogID), nullIfEmpty(string(e.AnalysisOutcome)), nullIfEmpty(e.OutcomeSummary),
		nullIfEmptyRaw(e.KPIResults), nullIfEmpty(e.Observation), nullIfEmpty(e.Hypothesis),
	)
	if err != nil {
		return err
	}

	// FTS index update — write searchable text even for rows with empty fields (FTS handles NULL gracefully)
	_, err = l.db.Exec(`INSERT INTO decisions_fts (rowid, log_id, reasoning, observation, hypothesis, campaign_name, reason)
		SELECT rowid, log_id, IFNULL(reasoning, ''), IFNULL(observation, ''), IFNULL(hypothesis, ''), IFNULL(campaign_name, ''), IFNULL(reason, '')
		FROM decisions WHERE log_id = ?`, e.LogID)
	return err
}

// ListByCampaign returns decisions for one campaign, newest first.
func (l *Log) ListByCampaign(campaignID string, limit int) ([]*Entry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := l.db.Query(`SELECT * FROM decisions WHERE campaign_id = ? ORDER BY timestamp DESC LIMIT ?`, campaignID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

// ListRecent returns the N most recent budget_decision entries across all campaigns.
func (l *Log) ListRecent(entryType EntryType, days int, limit int) ([]*Entry, error) {
	if limit <= 0 {
		limit = 100
	}
	since := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)
	q := `SELECT * FROM decisions WHERE timestamp >= ? `
	args := []any{since}
	if entryType != "" {
		q += `AND entry_type = ? `
		args = append(args, entryType)
	}
	q += `ORDER BY timestamp DESC LIMIT ?`
	args = append(args, limit)
	rows, err := l.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

// Search runs FTS5 over reasoning/observation/hypothesis/campaign_name/reason.
func (l *Log) Search(query string, limit int) ([]*Entry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := l.db.Query(`SELECT d.* FROM decisions d
		JOIN decisions_fts f ON d.log_id = f.log_id
		WHERE decisions_fts MATCH ?
		ORDER BY d.timestamp DESC LIMIT ?`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

// Due returns applied budget_decision entries whose follow_up_date is past AND have no analysis entry.
func (l *Log) Due(platform string) ([]*Entry, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	q := `SELECT * FROM decisions d
		WHERE d.entry_type = 'budget_decision'
		  AND d.was_applied = 1
		  AND d.follow_up_date IS NOT NULL
		  AND d.follow_up_date <= ?
		  AND NOT EXISTS (
		      SELECT 1 FROM decisions a WHERE a.original_log_id = d.log_id AND a.entry_type = 'decision_analysis'
		  )`
	args := []any{now}
	if platform != "" {
		q += ` AND d.platform = ?`
		args = append(args, platform)
	}
	q += ` ORDER BY d.follow_up_date ASC`
	rows, err := l.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

// GetByID loads one entry by log_id. Returns (nil, sql.ErrNoRows) when absent.
func (l *Log) GetByID(logID string) (*Entry, error) {
	rows, err := l.db.Query(`SELECT * FROM decisions WHERE log_id = ?`, logID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, sql.ErrNoRows
	}
	return entries[0], nil
}

// CalcFollowUp computes the follow-up date for a decision, matching magoosh-founder.
// Base: meta=18 days, google=7 days. Large change (>20%) bumps to 14 days minimum.
func CalcFollowUp(platform string, changePct float64) time.Time {
	base := 18
	if platform == "google_ads" {
		base = 7
	}
	if absFloat(changePct) >= 0.20 && base < 14 {
		base = 14
	}
	return time.Now().UTC().AddDate(0, 0, base)
}

func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func scanRows(rows *sql.Rows) ([]*Entry, error) {
	var out []*Entry
	for rows.Next() {
		e := &Entry{}
		var (
			ts, campID, campName, action, confidence, strategy           string
			reasoning, validationKPIs, warnings, metricsSnap, kpiResults string
			followUp, sessionID, oldStatus, newStatus, reason            string
			origLogID, outcome, outcomeSummary, observation, hypothesis  string
			expectedOutcome                                              string
			oldBudget, newBudget, changePct                              sql.NullFloat64
			applied                                                      int
		)
		err := rows.Scan(&e.LogID, &e.EntryType, &ts, &e.Platform,
			nullableString(&campID), nullableString(&campName),
			nullableString(&action), &oldBudget, &newBudget, &changePct,
			&applied, nullableString(&confidence), nullableString(&strategy),
			nullableString(&reasoning), nullableString(&validationKPIs), nullableString(&expectedOutcome),
			nullableString(&followUp), nullableString(&warnings), nullableString(&metricsSnap), nullableString(&sessionID),
			nullableString(&oldStatus), nullableString(&newStatus), nullableString(&reason),
			nullableString(&origLogID), nullableString(&outcome), nullableString(&outcomeSummary),
			nullableString(&kpiResults), nullableString(&observation), nullableString(&hypothesis),
		)
		if err != nil {
			return nil, err
		}
		e.Timestamp, _ = time.Parse(time.RFC3339, ts)
		e.CampaignID = campID
		e.CampaignName = campName
		e.Action = Action(action)
		if oldBudget.Valid {
			e.OldBudget = oldBudget.Float64
		}
		if newBudget.Valid {
			e.NewBudget = newBudget.Float64
		}
		if changePct.Valid {
			e.ChangePct = changePct.Float64
		}
		e.WasApplied = applied != 0
		e.Confidence = confidence
		e.Strategy = strategy
		if reasoning != "" {
			e.Reasoning = json.RawMessage(reasoning)
		}
		if validationKPIs != "" {
			e.ValidationKPIs = json.RawMessage(validationKPIs)
		}
		if warnings != "" {
			e.Warnings = json.RawMessage(warnings)
		}
		if metricsSnap != "" {
			e.MetricsSnapshot = json.RawMessage(metricsSnap)
		}
		if kpiResults != "" {
			e.KPIResults = json.RawMessage(kpiResults)
		}
		e.ExpectedOutcome = expectedOutcome
		if followUp != "" {
			if t, err := time.Parse(time.RFC3339, followUp); err == nil {
				e.FollowUpDate = &t
			}
		}
		e.SessionID = sessionID
		e.OldStatus = oldStatus
		e.NewStatus = newStatus
		e.Reason = reason
		e.OriginalLogID = origLogID
		e.AnalysisOutcome = Outcome(outcome)
		e.OutcomeSummary = outcomeSummary
		e.Observation = observation
		e.Hypothesis = hypothesis
		out = append(out, e)
	}
	return out, rows.Err()
}

// nullable scan helpers

type nullStrScanner struct{ dst *string }

func (n *nullStrScanner) Scan(src any) error {
	if src == nil {
		*n.dst = ""
		return nil
	}
	switch v := src.(type) {
	case string:
		*n.dst = v
	case []byte:
		*n.dst = string(v)
	default:
		*n.dst = fmt.Sprintf("%v", v)
	}
	return nil
}

func nullableString(dst *string) *nullStrScanner { return &nullStrScanner{dst: dst} }

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullIfEmptyRaw(r json.RawMessage) any {
	if len(r) == 0 {
		return nil
	}
	return string(r)
}

func nullIfZero(f float64) any {
	if f == 0 {
		return nil
	}
	return f
}
