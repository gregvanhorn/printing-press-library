// Hand-written: funding, funding-trend, and funding --who commands.
// SEC EDGAR Form D extraction is the killer feature.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mvanhorn/printing-press-library/library/developer-tools/company-goat/internal/source/sec"
	"github.com/mvanhorn/printing-press-library/library/developer-tools/company-goat/internal/source/yc"
	"github.com/spf13/cobra"
)

// fundingResult is the JSON shape for `funding <co>`.
type fundingResult struct {
	Domain   string          `json:"domain,omitempty"`
	Query    string          `json:"query,omitempty"`
	Filings  []fundingFiling `json:"form_d_filings"`
	YCEntry  *yc.Company     `json:"yc_entry,omitempty"`
	Coverage string          `json:"coverage_note,omitempty"`
}

type fundingFiling struct {
	FilingDate     string            `json:"filing_date"`
	Accession      string            `json:"accession"`
	EntityName     string            `json:"entity_name"`
	State          string            `json:"state_of_inc,omitempty"`
	IndustryGroup  string            `json:"industry_group,omitempty"`
	OfferingAmount int64             `json:"offering_amount,omitempty"` // -1 means "Indefinite"
	AmountSold     int64             `json:"amount_sold,omitempty"`
	Exemptions     []string          `json:"exemptions_claimed,omitempty"`
	RelatedPersons []sec.FormDPerson `json:"related_persons,omitempty"`
}

func newFundingCmd(flags *rootFlags) *cobra.Command {
	var t targetFlags
	var who string
	var maxFilings int
	var sinceYear int

	cmd := &cobra.Command{
		Use:   "funding <co>",
		Short: "SEC EDGAR Form D filings + YC batch lookup. The killer feature for US private fundraising.",
		Long: `funding fetches every Form D filing the SEC has for a company name, parses the structured XML, and reports offering amount, filing date, exemption claimed, and related persons.

Form D is filed by US private companies raising capital under Reg D (506(b) or 506(c)). The data is free and public — Crunchbase Pro charges thousands/year for what's essentially a wrapper around this same source.

With --who <person>, lists every Form D filing where the named person appears as a related party (officer, director, promoter). Useful for mapping serial founders.

Exit codes:
  0  at least one filing found (or candidate list rendered)
  2  ambiguous — rerun with --pick or --domain
  4  no candidates found
  5  no filings found for resolved company`,
		Example: strings.Trim(`
  company-goat-pp-cli funding anthropic
  company-goat-pp-cli funding stripe --json
  company-goat-pp-cli funding --domain anthropic.com --max 3
  company-goat-pp-cli funding --who "Patrick Collison" --json
  company-goat-pp-cli funding ramp --since 2020
`, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			if who == "" && t.Domain == "" && len(args) == 0 {
				return cmd.Help()
			}
			if maxFilings <= 0 {
				maxFilings = 5
			}

			secCli := sec.NewClient(getContactEmail(flags))

			// --who path: show every Form D filing for a named person.
			if who != "" {
				return runFundingWho(cmd, flags, secCli, who, maxFilings, sinceYear)
			}

			// Standard path: resolve company → search Form D.
			domain, err := runResolveOrExit(cmd, flags, args, t)
			if err != nil {
				return err
			}

			// Use the domain stem (e.g. "anthropic" from "anthropic.com") as
			// the EFTS query. This matches issuer-name keyword indexing
			// well; if it misses, the user can pass --domain with a more
			// specific query.
			stem := strings.SplitN(domain, ".", 2)[0]

			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			filings, err := secCli.SearchAndFetchAll(ctx, stem, maxFilings)
			if err != nil {
				return fmt.Errorf("sec edgar: %w", err)
			}
			if sinceYear > 0 {
				filings = filterByYear(filings, sinceYear)
			}
			ycCli := yc.NewClient()
			ycCtx, ycCancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer ycCancel()
			ycEntry, _ := ycCli.FindByDomain(ycCtx, domain)

			out := buildFundingResult(domain, filings, ycEntry)
			if len(out.Filings) == 0 && ycEntry == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "no Form D filings found for %q (looked up by stem %q at SEC EDGAR)\n", domain, stem)
				fmt.Fprintf(cmd.OutOrStdout(), "Coverage note: Form D is US-only. Non-US companies and pre-priced-round startups won't appear.\n")
				os.Exit(5)
			}
			renderFunding(cmd, flags, out)
			return nil
		},
	}
	addTargetFlags(cmd, &t)
	cmd.Flags().StringVar(&who, "who", "", "Show all Form D filings naming this person (e.g. \"Patrick Collison\")")
	cmd.Flags().IntVar(&maxFilings, "max", 5, "Maximum filings to fetch and parse")
	cmd.Flags().IntVar(&sinceYear, "since", 0, "Filter to filings on or after this year")
	return cmd
}

func runFundingWho(cmd *cobra.Command, flags *rootFlags, secCli *sec.Client, who string, maxFilings, sinceYear int) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
	defer cancel()

	// EFTS supports phrase search — use the person's name in quotes.
	filings, err := secCli.SearchAndFetchAll(ctx, who, maxFilings*2)
	if err != nil {
		return fmt.Errorf("sec edgar: %w", err)
	}
	// Filter to filings actually naming this person in relatedPersons.
	wantLower := strings.ToLower(who)
	matched := filings[:0]
	for _, fd := range filings {
		hit := false
		for _, p := range fd.RelatedPersons {
			if strings.Contains(strings.ToLower(p.Name), wantLower) {
				hit = true
				break
			}
		}
		if hit {
			matched = append(matched, fd)
		}
	}
	if sinceYear > 0 {
		matched = filterByYear(matched, sinceYear)
	}

	if len(matched) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "no Form D filings found naming %q\n", who)
		os.Exit(5)
	}

	out := struct {
		Person   string          `json:"person"`
		Filings  []fundingFiling `json:"form_d_filings"`
		Coverage string          `json:"coverage_note"`
	}{
		Person:   who,
		Filings:  fundingFilingsFromSEC(matched),
		Coverage: "Form D is US-only. Filings count: " + fmt.Sprintf("%d", len(matched)),
	}
	w := cmd.OutOrStdout()
	asJSON := flags.asJSON || !isTerminal(w)
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}
	fmt.Fprintf(w, "Form D filings naming %q:\n\n", who)
	for _, f := range out.Filings {
		fmt.Fprintf(w, "  %s  %-40s  %s\n", f.FilingDate, f.EntityName, formatAmount(f.OfferingAmount))
	}
	return nil
}

func filterByYear(in []sec.FormD, year int) []sec.FormD {
	var out []sec.FormD
	prefix := fmt.Sprintf("%04d", year)
	for _, f := range in {
		if f.FilingDate >= prefix {
			out = append(out, f)
		}
	}
	return out
}

func buildFundingResult(domain string, filings []sec.FormD, ycEntry *yc.Company) fundingResult {
	r := fundingResult{
		Domain:   domain,
		Filings:  fundingFilingsFromSEC(filings),
		YCEntry:  ycEntry,
		Coverage: "Form D is US-only. Non-US companies and pre-priced-round startups won't appear.",
	}
	return r
}

func fundingFilingsFromSEC(in []sec.FormD) []fundingFiling {
	out := make([]fundingFiling, 0, len(in))
	for _, fd := range in {
		out = append(out, fundingFiling{
			FilingDate:     fd.FilingDate,
			Accession:      fd.Accession,
			EntityName:     fd.EntityName,
			State:          fd.State,
			IndustryGroup:  fd.IndustryGroup,
			OfferingAmount: fd.OfferingAmount,
			AmountSold:     fd.AmountSold,
			Exemptions:     fd.ExemptionsClaimed,
			RelatedPersons: fd.RelatedPersons,
		})
	}
	// Sort by filing date descending.
	sort.SliceStable(out, func(i, j int) bool { return out[i].FilingDate > out[j].FilingDate })
	return out
}

func renderFunding(cmd *cobra.Command, flags *rootFlags, r fundingResult) {
	w := cmd.OutOrStdout()
	asJSON := flags.asJSON || !isTerminal(w)
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(r)
		return
	}
	fmt.Fprintf(w, "Domain: %s\n", r.Domain)
	if r.YCEntry != nil {
		fmt.Fprintf(w, "YC: %s (batch %s, status %s)\n", r.YCEntry.Name, r.YCEntry.Batch, r.YCEntry.Status)
	}
	if len(r.Filings) == 0 {
		fmt.Fprintf(w, "Form D: no filings found\n")
		return
	}
	fmt.Fprintf(w, "\nForm D filings (%d):\n", len(r.Filings))
	for _, f := range r.Filings {
		fmt.Fprintf(w, "  %s  %-40s  %s  exempt:%v  state:%s  industry:%s\n",
			f.FilingDate, fundingTruncate(f.EntityName, 40), formatAmount(f.OfferingAmount),
			f.Exemptions, f.State, f.IndustryGroup)
	}
	fmt.Fprintf(w, "\n%s\n", r.Coverage)
}

func formatAmount(amt int64) string {
	if amt == -1 {
		return "$Indefinite"
	}
	if amt == 0 {
		return "$0"
	}
	switch {
	case amt >= 1_000_000_000:
		return fmt.Sprintf("$%.1fB", float64(amt)/1_000_000_000)
	case amt >= 1_000_000:
		return fmt.Sprintf("$%.1fM", float64(amt)/1_000_000)
	case amt >= 1_000:
		return fmt.Sprintf("$%.0fK", float64(amt)/1_000)
	default:
		return fmt.Sprintf("$%d", amt)
	}
}

// fundingTruncate is a local helper. The generated helpers.go already has
// a truncate(...) but its semantics differ; we use this variant only inside
// company_funding.go (and other novel commands) for consistency.
func fundingTruncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

func newFundingTrendCmd(flags *rootFlags) *cobra.Command {
	var t targetFlags
	var sinceYear int
	var maxFilings int

	cmd := &cobra.Command{
		Use:   "funding-trend <co>",
		Short: "Time series of Form D filings showing fundraising cadence over years.",
		Long: `funding-trend renders a year-by-year count of Form D filings for a company. Useful for spotting fundraising gaps or a startup that quietly stopped raising.

Output bins by filing year and shows offering amount totals per year.`,
		Example: strings.Trim(`
  company-goat-pp-cli funding-trend stripe
  company-goat-pp-cli funding-trend anthropic --since 2020 --json
`, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			if t.Domain == "" && len(args) == 0 {
				return cmd.Help()
			}
			if maxFilings <= 0 {
				maxFilings = 25
			}

			domain, err := runResolveOrExit(cmd, flags, args, t)
			if err != nil {
				return err
			}
			stem := strings.SplitN(domain, ".", 2)[0]

			secCli := sec.NewClient(getContactEmail(flags))
			ctx, cancel := context.WithTimeout(cmd.Context(), 90*time.Second)
			defer cancel()

			filings, err := secCli.SearchAndFetchAll(ctx, stem, maxFilings)
			if err != nil {
				return fmt.Errorf("sec edgar: %w", err)
			}
			if sinceYear > 0 {
				filings = filterByYear(filings, sinceYear)
			}

			type yearBucket struct {
				Year         int   `json:"year"`
				FilingCount  int   `json:"filing_count"`
				TotalOffered int64 `json:"total_offered_usd"`
			}
			buckets := map[int]*yearBucket{}
			for _, f := range filings {
				if len(f.FilingDate) < 4 {
					continue
				}
				yr := 0
				_, err := fmt.Sscanf(f.FilingDate[:4], "%d", &yr)
				if err != nil {
					continue
				}
				b, ok := buckets[yr]
				if !ok {
					b = &yearBucket{Year: yr}
					buckets[yr] = b
				}
				b.FilingCount++
				if f.OfferingAmount > 0 {
					b.TotalOffered += f.OfferingAmount
				}
			}
			years := make([]int, 0, len(buckets))
			for y := range buckets {
				years = append(years, y)
			}
			sort.Ints(years)
			out := make([]yearBucket, 0, len(years))
			for _, y := range years {
				out = append(out, *buckets[y])
			}

			w := cmd.OutOrStdout()
			asJSON := flags.asJSON || !isTerminal(w)
			if asJSON {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{
					"domain":        domain,
					"buckets":       out,
					"total_filings": len(filings),
				})
			}
			if len(out) == 0 {
				fmt.Fprintf(w, "no Form D filings found for %q\n", domain)
				return nil
			}
			fmt.Fprintf(w, "Form D fundraising trend for %s:\n\n", domain)
			fmt.Fprintf(w, "  YEAR  FILINGS  TOTAL OFFERED\n")
			for _, b := range out {
				fmt.Fprintf(w, "  %d  %5d    %s\n", b.Year, b.FilingCount, formatAmount(b.TotalOffered))
			}
			return nil
		},
	}
	addTargetFlags(cmd, &t)
	cmd.Flags().IntVar(&sinceYear, "since", 0, "Only include filings on or after this year")
	cmd.Flags().IntVar(&maxFilings, "max", 25, "Maximum filings to fetch")
	return cmd
}

// getContactEmail reads the SEC fair-access contact email from
// COMPANY_PP_CONTACT_EMAIL. Empty falls back to the generic User-Agent.
func getContactEmail(flags *rootFlags) string {
	return strings.TrimSpace(os.Getenv("COMPANY_PP_CONTACT_EMAIL"))
}
