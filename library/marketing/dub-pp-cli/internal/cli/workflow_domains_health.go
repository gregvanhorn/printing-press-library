package cli

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/dub-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newWorkflowDomainsHealthCmd(flags *rootFlags) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "domains-health",
		Short: "Show domain health: DNS status, link count, and click performance",
		Long: `Checks each domain's DNS resolution and aggregates link/click stats.
Requires a prior sync of domains and links.`,
		Example: `  # Domain health dashboard
  dub-pp-cli workflow domains-health

  # As JSON
  dub-pp-cli workflow domains-health --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("dub-pp-cli")
			}
			s, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}
			defer s.Close()

			domains, err := s.List("domains", 500)
			if err != nil {
				return fmt.Errorf("listing domains: %w", err)
			}
			links, err := s.List("links", 5000)
			if err != nil {
				return fmt.Errorf("listing links: %w", err)
			}

			// Aggregate link stats per domain
			type domainStat struct {
				Slug     string `json:"slug"`
				Verified bool   `json:"verified"`
				DNS      string `json:"dns"`
				Links    int    `json:"links"`
				Clicks   int    `json:"clicks"`
			}

			domainMap := make(map[string]*domainStat)
			for _, d := range domains {
				var obj map[string]any
				if err := json.Unmarshal(d, &obj); err != nil {
					continue
				}
				slug := strVal(obj, "slug")
				verified := false
				if v, ok := obj["verified"]; ok {
					if b, ok := v.(bool); ok {
						verified = b
					}
				}
				domainMap[slug] = &domainStat{Slug: slug, Verified: verified}
			}

			for _, item := range links {
				var obj map[string]any
				if err := json.Unmarshal(item, &obj); err != nil {
					continue
				}
				domain := strVal(obj, "domain")
				ds, ok := domainMap[domain]
				if !ok {
					ds = &domainStat{Slug: domain}
					domainMap[domain] = ds
				}
				ds.Links++
				ds.Clicks += intVal(obj, "clicks")
			}

			// DNS checks
			for _, ds := range domainMap {
				_, err := net.DialTimeout("tcp", ds.Slug+":443", 3*time.Second)
				if err != nil {
					ds.DNS = "FAIL"
				} else {
					ds.DNS = "OK"
				}
			}

			var sorted []*domainStat
			for _, ds := range domainMap {
				sorted = append(sorted, ds)
			}
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Clicks > sorted[j].Clicks
			})

			if flags.asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(sorted)
			}

			if len(sorted) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No domains found. Run 'sync' first.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-8s %-5s %6s %8s\n", "Domain", "Verified", "DNS", "Links", "Clicks")
			fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-8s %-5s %6s %8s\n", "------------------------------", "--------", "-----", "------", "--------")
			for _, ds := range sorted {
				verified := "no"
				if ds.Verified {
					verified = "yes"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-8s %-5s %6d %8d\n", ds.Slug, verified, ds.DNS, ds.Links, ds.Clicks)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")

	return cmd
}
