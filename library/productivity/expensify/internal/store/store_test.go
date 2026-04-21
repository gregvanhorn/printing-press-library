// Copyright 2026 matt-van-horn. Licensed under Apache-2.0.

package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "store.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestUpsertPerson_Insert verifies a new Person row lands in the people table
// and Get by accountID returns the same values.
func TestUpsertPerson_Insert(t *testing.T) {
	s := openTestStore(t)
	p := Person{
		AccountID:   20647491,
		DisplayName: "Myk Melez",
		Login:       "myk@example.com",
		Avatar:      "https://example.com/a.png",
	}
	if err := s.UpsertPerson(p); err != nil {
		t.Fatalf("UpsertPerson: %v", err)
	}
	got, err := s.GetPersonByAccountID(20647491)
	if err != nil {
		t.Fatalf("GetPersonByAccountID: %v", err)
	}
	if got == nil {
		t.Fatal("GetPersonByAccountID returned nil, want row")
	}
	if got.AccountID != p.AccountID || got.DisplayName != p.DisplayName || got.Login != p.Login || got.Avatar != p.Avatar {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, p)
	}
	if got.SyncedAt == "" {
		t.Fatalf("SyncedAt = empty, want a timestamp")
	}
}

// TestUpsertPerson_Update verifies a second upsert with the same accountID
// overwrites the previous displayName.
func TestUpsertPerson_Update(t *testing.T) {
	s := openTestStore(t)
	if err := s.UpsertPerson(Person{AccountID: 42, DisplayName: "Old Name", Login: "x@example.com"}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := s.UpsertPerson(Person{AccountID: 42, DisplayName: "New Name", Login: "x@example.com"}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	got, err := s.GetPersonByAccountID(42)
	if err != nil {
		t.Fatalf("GetPersonByAccountID: %v", err)
	}
	if got == nil || got.DisplayName != "New Name" {
		t.Fatalf("DisplayName = %+v, want %q after update", got, "New Name")
	}
}

// TestGetPersonByAccountID_NotFound verifies an unknown ID returns sql.ErrNoRows.
func TestGetPersonByAccountID_NotFound(t *testing.T) {
	s := openTestStore(t)
	got, err := s.GetPersonByAccountID(999)
	if got != nil {
		t.Fatalf("GetPersonByAccountID returned %+v, want nil", got)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

// TestGetPersonByLogin verifies case-insensitive login lookup.
func TestGetPersonByLogin(t *testing.T) {
	s := openTestStore(t)
	if err := s.UpsertPerson(Person{AccountID: 7, DisplayName: "Myk", Login: "Myk@Example.COM"}); err != nil {
		t.Fatalf("UpsertPerson: %v", err)
	}
	got, err := s.GetPersonByLogin("myk@example.com")
	if err != nil {
		t.Fatalf("GetPersonByLogin: %v", err)
	}
	if got == nil {
		t.Fatal("GetPersonByLogin returned nil, want a row")
	}
	if got.AccountID != 7 {
		t.Fatalf("AccountID = %d, want 7", got.AccountID)
	}
}

// TestGetPersonByLogin_NotFound verifies sql.ErrNoRows on miss.
func TestGetPersonByLogin_NotFound(t *testing.T) {
	s := openTestStore(t)
	got, err := s.GetPersonByLogin("nobody@example.com")
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

// TestUpsertPerson_EmptyDisplayName verifies an entry with only a login still
// upserts cleanly; display name stays empty.
func TestUpsertPerson_EmptyDisplayName(t *testing.T) {
	s := openTestStore(t)
	if err := s.UpsertPerson(Person{AccountID: 100, DisplayName: "", Login: "bare@example.com"}); err != nil {
		t.Fatalf("UpsertPerson: %v", err)
	}
	got, err := s.GetPersonByAccountID(100)
	if err != nil {
		t.Fatalf("GetPersonByAccountID: %v", err)
	}
	if got == nil {
		t.Fatal("nil row, want one")
	}
	if got.DisplayName != "" || got.Login != "bare@example.com" {
		t.Fatalf("got %+v, want DisplayName=\"\" Login=\"bare@example.com\"", got)
	}
}

// TestListPeople verifies ListPeople returns all upserted rows ordered by
// display name.
func TestListPeople(t *testing.T) {
	s := openTestStore(t)
	_ = s.UpsertPerson(Person{AccountID: 1, DisplayName: "Zed", Login: "z@example.com"})
	_ = s.UpsertPerson(Person{AccountID: 2, DisplayName: "Alice", Login: "a@example.com"})
	people, err := s.ListPeople()
	if err != nil {
		t.Fatalf("ListPeople: %v", err)
	}
	if len(people) != 2 {
		t.Fatalf("len(people) = %d, want 2", len(people))
	}
	if people[0].DisplayName != "Alice" || people[1].DisplayName != "Zed" {
		t.Fatalf("order = [%q, %q], want [Alice, Zed]", people[0].DisplayName, people[1].DisplayName)
	}
}

// upsertDamageReport is a small helper that cuts down test boilerplate by
// building a Report with the common fields set and calling UpsertReport.
func upsertDamageReport(t *testing.T, s *Store, id string, stateNum int64, total int64, created, policy string) {
	t.Helper()
	r := Report{
		ReportID:    id,
		PolicyID:    policy,
		Title:       "report " + id,
		Status:      "",
		Total:       total,
		Currency:    "USD",
		Created:     created,
		LastUpdated: created,
		StateNum:    stateNum,
	}
	if err := s.UpsertReport(r); err != nil {
		t.Fatalf("UpsertReport(%s): %v", id, err)
	}
}

// TestDamage_BucketsByStateNum verifies each stateNum maps to the documented
// bucket: 0→Expensed, 1→Pending, 3→Approved, 4→Paid.
func TestDamage_BucketsByStateNum(t *testing.T) {
	s := openTestStore(t)
	month := "2026-04"
	created := "2026-04-15"
	upsertDamageReport(t, s, "r0", 0, 1000, created, "")
	upsertDamageReport(t, s, "r1", 1, 2000, created, "")
	upsertDamageReport(t, s, "r3", 3, 3000, created, "")
	upsertDamageReport(t, s, "r4", 4, 4000, created, "")

	bd, err := s.Damage(month, "")
	if err != nil {
		t.Fatalf("Damage: %v", err)
	}
	if bd.Expensed != 1000 || bd.ExpensedCount != 1 {
		t.Errorf("Expensed = ($%d,%d), want ($1000,1)", bd.Expensed, bd.ExpensedCount)
	}
	if bd.PendingApproval != 2000 || bd.PendingCount != 1 {
		t.Errorf("Pending = ($%d,%d), want ($2000,1)", bd.PendingApproval, bd.PendingCount)
	}
	if bd.Approved != 3000 || bd.ApprovedCount != 1 {
		t.Errorf("Approved = ($%d,%d), want ($3000,1)", bd.Approved, bd.ApprovedCount)
	}
	if bd.Paid != 4000 || bd.PaidCount != 1 {
		t.Errorf("Paid = ($%d,%d), want ($4000,1)", bd.Paid, bd.PaidCount)
	}
}

// TestDamage_NoReportsForMonth verifies that an empty store and a store with
// reports outside the target month both return all-zero buckets without error.
func TestDamage_NoReportsForMonth(t *testing.T) {
	s := openTestStore(t)

	// Empty store.
	bd, err := s.Damage("2026-04", "")
	if err != nil {
		t.Fatalf("Damage(empty): %v", err)
	}
	if bd.Expensed != 0 || bd.ExpensedCount != 0 ||
		bd.PendingApproval != 0 || bd.PendingCount != 0 ||
		bd.Approved != 0 || bd.ApprovedCount != 0 ||
		bd.Paid != 0 || bd.PaidCount != 0 ||
		bd.MissingReceipts != 0 {
		t.Fatalf("empty store: buckets = %+v, want all zero", bd)
	}

	// Report outside the target month.
	upsertDamageReport(t, s, "r-feb", 0, 5000, "2026-02-10", "")
	bd, err = s.Damage("2026-04", "")
	if err != nil {
		t.Fatalf("Damage(other month): %v", err)
	}
	if bd.Expensed != 0 || bd.ExpensedCount != 0 {
		t.Fatalf("other-month report leaked: %+v", bd)
	}
}

// TestDamage_RawJsonFallback verifies that a row with state_num=NULL but a
// raw_json containing stateNum=3 buckets to Approved via the fallback parse.
func TestDamage_RawJsonFallback(t *testing.T) {
	s := openTestStore(t)
	raw := `{"stateNum":3,"total":-5000,"currency":"USD","created":"2026-04-15"}`
	// Direct INSERT to force state_num=NULL while still carrying raw_json.
	_, err := s.DB.Exec(`INSERT INTO reports
		(report_id, policy_id, title, status, total, currency, created, last_updated, expense_count, state_num, raw_json, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?)`,
		"r-null", "", "legacy row", "", int64(3000), "USD", "2026-04-15", "2026-04-15", 0, raw, "2026-04-15T00:00:00Z")
	if err != nil {
		t.Fatalf("direct insert: %v", err)
	}

	bd, err := s.Damage("2026-04", "")
	if err != nil {
		t.Fatalf("Damage: %v", err)
	}
	if bd.Approved != 3000 || bd.ApprovedCount != 1 {
		t.Fatalf("Approved = ($%d,%d), want ($3000,1); full: %+v", bd.Approved, bd.ApprovedCount, bd)
	}
}

// TestDamage_State5_BucketsToPaid verifies that BILLING (state 5) is close
// enough to reimbursed for user-facing totals.
func TestDamage_State5_BucketsToPaid(t *testing.T) {
	s := openTestStore(t)
	upsertDamageReport(t, s, "r5", 5, 7500, "2026-04-10", "")
	bd, err := s.Damage("2026-04", "")
	if err != nil {
		t.Fatalf("Damage: %v", err)
	}
	if bd.Paid != 7500 || bd.PaidCount != 1 {
		t.Fatalf("Paid = ($%d,%d), want ($7500,1); full: %+v", bd.Paid, bd.PaidCount, bd)
	}
}

// TestDamage_State6_BucketsToPaid verifies stateNum=6 maps to Paid.
func TestDamage_State6_BucketsToPaid(t *testing.T) {
	s := openTestStore(t)
	upsertDamageReport(t, s, "r6", 6, 9900, "2026-04-10", "")
	bd, err := s.Damage("2026-04", "")
	if err != nil {
		t.Fatalf("Damage: %v", err)
	}
	if bd.Paid != 9900 || bd.PaidCount != 1 {
		t.Fatalf("Paid = ($%d,%d), want ($9900,1); full: %+v", bd.Paid, bd.PaidCount, bd)
	}
}

// TestDamage_InvalidRawJson verifies that a NULL state_num + junk raw_json
// falls through safely into the Expensed bucket with no error.
func TestDamage_InvalidRawJson(t *testing.T) {
	s := openTestStore(t)
	_, err := s.DB.Exec(`INSERT INTO reports
		(report_id, policy_id, title, status, total, currency, created, last_updated, expense_count, state_num, raw_json, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?)`,
		"r-bad", "", "bad json", "", int64(1200), "USD", "2026-04-03", "2026-04-03", 0, "{not-valid-json", "2026-04-03T00:00:00Z")
	if err != nil {
		t.Fatalf("direct insert: %v", err)
	}

	bd, err := s.Damage("2026-04", "")
	if err != nil {
		t.Fatalf("Damage: %v", err)
	}
	if bd.Expensed != 1200 || bd.ExpensedCount != 1 {
		t.Fatalf("Expensed = ($%d,%d), want ($1200,1); full: %+v", bd.Expensed, bd.ExpensedCount, bd)
	}
}

// TestDamage_PolicyIDFilter verifies that a policy filter narrows the result
// to reports under that policy only.
func TestDamage_PolicyIDFilter(t *testing.T) {
	s := openTestStore(t)
	upsertDamageReport(t, s, "r-a", 0, 1111, "2026-04-12", "POLICY_A")
	upsertDamageReport(t, s, "r-b", 0, 2222, "2026-04-12", "POLICY_B")

	bd, err := s.Damage("2026-04", "POLICY_A")
	if err != nil {
		t.Fatalf("Damage: %v", err)
	}
	if bd.Expensed != 1111 || bd.ExpensedCount != 1 {
		t.Fatalf("Expensed (A only) = ($%d,%d), want ($1111,1); full: %+v", bd.Expensed, bd.ExpensedCount, bd)
	}
}
