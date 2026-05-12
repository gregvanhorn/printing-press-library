// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

// Package predicate extracts predicate K-number citations from FDA 510(k)
// Summary PDFs. openFDA does not expose predicate relationships as structured
// data, so the only public source is the FDA's SE letter / 510(k) summary
// PDF hosted at accessdata.fda.gov.
package predicate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const userAgent = "Mozilla/5.0 (compatible; fda-devices-pp-cli; +https://open.fda.gov)"

var (
	kNumberRx = regexp.MustCompile(`\bK(\d{6})\b`)
	// Heading after the predicate listing — varies by submitter. We stop the
	// predicate-section scan when any of these appear so we don't pull
	// K-numbers cited later in the body as comparative references.
	endHeadings = regexp.MustCompile(`(?i)(INDICATIONS FOR USE|DEVICE DESCRIPTION|INTENDED USE|TECHNOLOGICAL CHARACTERISTICS|PERFORMANCE DATA|SUBSTANTIAL EQUIVALENCE|CONCLUSION)`)
	// All-caps only — lowercase "predicate device(s)" appears in SE-letter
	// boilerplate and would otherwise consume the section start before the
	// real heading inside the 510(k) Summary.
	predHeading = regexp.MustCompile(`PREDICATE\s+DEVICE`)
)

// Result is the parsed predicate set for one K-number.
type Result struct {
	KNumber    string   `json:"k_number"`
	Primary    string   `json:"primary,omitempty"`
	Predicates []string `json:"predicates"`
	Source     string   `json:"source"` // PDF URL
	FetchedAt  string   `json:"fetched_at"`
	Note       string   `json:"note,omitempty"`
}

// PDFURL returns the canonical FDA accessdata URL for a 510(k) summary PDF.
// K-number format is K + 2-digit year + 4-digit sequence; the path uses /pdfYY/.
func PDFURL(kNumber string) string {
	kNumber = strings.ToUpper(strings.TrimSpace(kNumber))
	if len(kNumber) < 3 || kNumber[0] != 'K' {
		return ""
	}
	return fmt.Sprintf("https://www.accessdata.fda.gov/cdrh_docs/pdf%s/%s.pdf", kNumber[1:3], kNumber)
}

// Fetch downloads the summary PDF, extracts text via pdftotext, and parses out
// predicate K-numbers. Results are cached under cacheDir.
func Fetch(ctx context.Context, kNumber, cacheDir string) (*Result, error) {
	kNumber = strings.ToUpper(strings.TrimSpace(kNumber))
	if !kNumberRx.MatchString(kNumber) {
		return nil, fmt.Errorf("invalid K-number: %q", kNumber)
	}

	if cacheDir != "" {
		if r := readCache(cacheDir, kNumber); r != nil {
			return r, nil
		}
	}

	url := PDFURL(kNumber)
	pdfBytes, err := downloadPDF(ctx, url)
	if err != nil {
		return nil, err
	}

	text, err := pdfToText(pdfBytes)
	if err != nil {
		return nil, err
	}

	res := parsePredicates(kNumber, text)
	res.Source = url
	res.FetchedAt = time.Now().UTC().Format(time.RFC3339)

	if cacheDir != "" {
		_ = writeCache(cacheDir, kNumber, res)
	}
	return res, nil
}

func downloadPDF(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/pdf,*/*")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20)) // 50 MB cap
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(body, []byte("%PDF")) {
		return nil, fmt.Errorf("response from %s is not a PDF", url)
	}
	return body, nil
}

var errNotFound = errors.New("510(k) summary PDF not found")

// ErrNotFound reports whether err signals a missing summary PDF.
func ErrNotFound(err error) bool { return errors.Is(err, errNotFound) }

// pdfToText shells out to pdftotext (poppler). If the binary is missing we
// return a clear error so callers can fall back gracefully.
func pdfToText(pdf []byte) (string, error) {
	bin, err := exec.LookPath("pdftotext")
	if err != nil {
		return "", fmt.Errorf("pdftotext binary not found on PATH (install poppler-utils)")
	}
	cmd := exec.Command(bin, "-layout", "-", "-")
	cmd.Stdin = bytes.NewReader(pdf)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext: %w: %s", err, stderr.String())
	}
	return out.String(), nil
}

// parsePredicates extracts predicate K-numbers from the PREDICATE DEVICES
// section. Excludes the subject K-number itself and de-duplicates.
func parsePredicates(subject, text string) *Result {
	r := &Result{KNumber: subject, Predicates: []string{}}

	predIdx := predHeading.FindStringIndex(text)
	if predIdx == nil {
		r.Note = "no PREDICATE DEVICES section found in summary"
		return r
	}
	// End the predicate section at the next major heading (case-insensitive,
	// since trailing prose may be mixed-case).
	tail := text[predIdx[1]:]
	endIdx := endHeadings.FindStringIndex(tail)
	end := len(tail)
	if endIdx != nil {
		end = endIdx[0]
	}
	section := tail[:end]

	// First K-number after the literal "Primary Predicate" is the primary.
	if pi := regexp.MustCompile(`(?is)Primary\s+Predicate`).FindStringIndex(section); pi != nil {
		after := section[pi[1]:]
		if m := kNumberRx.FindString(after); m != "" && m != subject {
			r.Primary = m
		}
	}

	seen := map[string]bool{subject: true}
	for _, m := range kNumberRx.FindAllString(section, -1) {
		if seen[m] {
			continue
		}
		seen[m] = true
		r.Predicates = append(r.Predicates, m)
	}
	if r.Primary == "" && len(r.Predicates) > 0 {
		r.Primary = r.Predicates[0]
	}
	if len(r.Predicates) == 0 {
		r.Note = "PREDICATE DEVICES section found but no K-numbers parsed"
	}
	return r
}

func cachePath(dir, kNumber string) string {
	return filepath.Join(dir, kNumber+".json")
}

func readCache(dir, kNumber string) *Result {
	b, err := os.ReadFile(cachePath(dir, kNumber))
	if err != nil {
		return nil
	}
	var r Result
	if err := json.Unmarshal(b, &r); err != nil {
		return nil
	}
	return &r
}

func writeCache(dir, kNumber string, r *Result) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(dir, kNumber), b, 0o644)
}
