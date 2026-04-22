// Copyright 2026 pejman-pour-moezzi. Licensed under Apache-2.0. See LICENSE.
// Hand-authored: unit tests for pure-logic helpers in the decisions package.

package decisions

import (
	"math"
	"testing"
	"time"
)

func TestCalcFollowUpMetaBaseline(t *testing.T) {
	now := time.Now().UTC()
	got := CalcFollowUp("meta_ads", 0.05)
	// Meta base window is 18 days.
	want := now.AddDate(0, 0, 18)
	if absDur(got.Sub(want)) > 2*time.Second {
		t.Fatalf("meta baseline follow-up: got %v, want ~%v", got, want)
	}
}

func TestCalcFollowUpGoogleBaselineStaysShort(t *testing.T) {
	now := time.Now().UTC()
	got := CalcFollowUp("google_ads", 0.05)
	// Google base window is 7 days when change is small.
	want := now.AddDate(0, 0, 7)
	if absDur(got.Sub(want)) > 2*time.Second {
		t.Fatalf("google baseline follow-up: got %v, want ~%v", got, want)
	}
}

func TestCalcFollowUpLargeChangeBumpsGoogleToFourteen(t *testing.T) {
	now := time.Now().UTC()
	got := CalcFollowUp("google_ads", 0.25)
	// >20% change on Google bumps the base from 7 to 14 days.
	want := now.AddDate(0, 0, 14)
	if absDur(got.Sub(want)) > 2*time.Second {
		t.Fatalf("google large-change follow-up: got %v, want ~%v", got, want)
	}
}

func TestCalcFollowUpLargeChangeLeavesMetaUnchanged(t *testing.T) {
	now := time.Now().UTC()
	got := CalcFollowUp("meta_ads", 0.50)
	// Meta base (18) already exceeds the 14-day floor, so no bump.
	want := now.AddDate(0, 0, 18)
	if absDur(got.Sub(want)) > 2*time.Second {
		t.Fatalf("meta large-change follow-up: got %v, want ~%v", got, want)
	}
}

func TestCalcFollowUpNegativeChangeBumpsViaAbs(t *testing.T) {
	now := time.Now().UTC()
	got := CalcFollowUp("google_ads", -0.30)
	// -30% on Google: |changePct| >= 0.20 AND base(7) < 14, so base becomes 14.
	want := now.AddDate(0, 0, 14)
	if absDur(got.Sub(want)) > 2*time.Second {
		t.Fatalf("google negative-change follow-up: got %v, want ~%v", got, want)
	}
}

func TestAbsFloat(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{0, 0},
		{1.5, 1.5},
		{-1.5, 1.5},
		{math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64},
		{-math.MaxFloat64, math.MaxFloat64},
	}
	for _, tc := range cases {
		if got := absFloat(tc.in); got != tc.want {
			t.Errorf("absFloat(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestEntryTypeAndActionConstants(t *testing.T) {
	// Guard against accidental value churn — these strings are serialized
	// into the decisions table and into JSON output for agent consumption.
	if string(EntryBudgetDecision) != "budget_decision" {
		t.Errorf("EntryBudgetDecision = %q, want %q", EntryBudgetDecision, "budget_decision")
	}
	if string(EntryStatusChange) != "status_change" {
		t.Errorf("EntryStatusChange = %q, want %q", EntryStatusChange, "status_change")
	}
	if string(EntryAnalysis) != "decision_analysis" {
		t.Errorf("EntryAnalysis = %q, want %q", EntryAnalysis, "decision_analysis")
	}
	if string(ActionIncrease) != "INCREASE" {
		t.Errorf("ActionIncrease = %q", ActionIncrease)
	}
	if string(ActionDecrease) != "DECREASE" {
		t.Errorf("ActionDecrease = %q", ActionDecrease)
	}
	if string(ActionHold) != "HOLD" {
		t.Errorf("ActionHold = %q", ActionHold)
	}
	if string(ActionFlag) != "FLAG" {
		t.Errorf("ActionFlag = %q", ActionFlag)
	}
	if string(OutcomeSuccess) != "success" {
		t.Errorf("OutcomeSuccess = %q", OutcomeSuccess)
	}
	if string(OutcomePartial) != "partial" {
		t.Errorf("OutcomePartial = %q", OutcomePartial)
	}
	if string(OutcomeFailure) != "failure" {
		t.Errorf("OutcomeFailure = %q", OutcomeFailure)
	}
	if string(OutcomeInconclusive) != "inconclusive" {
		t.Errorf("OutcomeInconclusive = %q", OutcomeInconclusive)
	}
}

func TestNullHelpers(t *testing.T) {
	if got := nullIfEmpty(""); got != nil {
		t.Errorf("nullIfEmpty(\"\") = %v, want nil", got)
	}
	if got := nullIfEmpty("hi"); got != "hi" {
		t.Errorf("nullIfEmpty(\"hi\") = %v, want \"hi\"", got)
	}
	if got := nullIfZero(0); got != nil {
		t.Errorf("nullIfZero(0) = %v, want nil", got)
	}
	if got := nullIfZero(0.1); got != 0.1 {
		t.Errorf("nullIfZero(0.1) = %v, want 0.1", got)
	}
	if got := nullIfEmptyRaw(nil); got != nil {
		t.Errorf("nullIfEmptyRaw(nil) = %v, want nil", got)
	}
	if got := nullIfEmptyRaw([]byte(`{"k":"v"}`)); got != `{"k":"v"}` {
		t.Errorf("nullIfEmptyRaw(json) = %v", got)
	}
}

func absDur(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
