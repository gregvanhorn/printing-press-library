package cli

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newWorkflowCampaignCmd(flags *rootFlags) *cobra.Command {
	var tag string
	var domain string
	var prefix string

	cmd := &cobra.Command{
		Use:   "campaign [file]",
		Short: "Bulk-create tagged short links from a URL list for a campaign",
		Long: `Create multiple short links at once from a file of destination URLs.
All links are tagged and optionally assigned a custom domain.
Accepts a text file (one URL per line) or CSV (url column).`,
		Example: `  # Create campaign links from a URL list
  dub-pp-cli workflow campaign urls.txt --tag spring-sale --domain go.acme.com

  # Dry run to preview what would be created
  dub-pp-cli workflow campaign urls.txt --tag q2-launch --dry-run

  # From CSV with url column
  dub-pp-cli workflow campaign campaign.csv --tag promo --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			urls, err := readURLFile(filePath)
			if err != nil {
				return fmt.Errorf("reading URL file: %w", err)
			}
			if len(urls) == 0 {
				return fmt.Errorf("no URLs found in %s", filePath)
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Creating %d links with tag=%q domain=%q\n", len(urls), tag, domain)

			if flags.dryRun {
				type preview struct {
					URL    string `json:"url"`
					Tag    string `json:"tag,omitempty"`
					Domain string `json:"domain,omitempty"`
					Key    string `json:"key,omitempty"`
				}
				var previews []preview
				for _, u := range urls {
					p := preview{URL: u, Tag: tag, Domain: domain}
					if prefix != "" {
						p.Key = fmt.Sprintf("%s-%d", prefix, len(previews)+1)
					}
					previews = append(previews, p)
				}
				if flags.asJSON {
					enc := json.NewEncoder(cmd.OutOrStdout())
					enc.SetIndent("", "  ")
					return enc.Encode(previews)
				}
				for _, p := range previews {
					fmt.Fprintf(cmd.OutOrStdout(), "[DRY RUN] Would create: %s → tag=%s domain=%s\n", p.URL, p.Tag, p.Domain)
				}
				return nil
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			// Build bulk create payload
			type linkReq struct {
				URL    string `json:"url"`
				TagIDs []any  `json:"tagIds,omitempty"`
				Domain string `json:"domain,omitempty"`
				Key    string `json:"key,omitempty"`
			}

			var links []linkReq
			for i, u := range urls {
				lr := linkReq{URL: u}
				if domain != "" {
					lr.Domain = domain
				}
				if prefix != "" {
					lr.Key = fmt.Sprintf("%s-%d", prefix, i+1)
				}
				links = append(links, lr)
			}

			// Use bulk create endpoint
			body, err := json.Marshal(links)
			if err != nil {
				return err
			}

			resp, _, err := c.Post("/links/bulk", body)
			if err != nil {
				return fmt.Errorf("bulk create: %w", err)
			}

			if tag != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Note: Tag must be applied individually or pre-created. Bulk endpoint creates links; tagging is per-link.\n")
			}

			if flags.asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(json.RawMessage(resp))
			}

			var created []map[string]any
			if err := json.Unmarshal(resp, &created); err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), string(resp))
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created %d links:\n", len(created))
			for _, link := range created {
				shortLink := fmt.Sprintf("%s/%s", strVal(link, "domain"), strVal(link, "key"))
				fmt.Fprintf(cmd.OutOrStdout(), "  %s → %s\n", shortLink, strVal(link, "url"))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Tag to apply to all created links")
	cmd.Flags().StringVar(&domain, "domain", "", "Custom domain for the short links")
	cmd.Flags().StringVar(&prefix, "prefix", "", "Key prefix for short link slugs (e.g. 'spring' → spring-1, spring-2)")

	return cmd
}

func readURLFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Try CSV first
	if strings.HasSuffix(strings.ToLower(path), ".csv") {
		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("parsing CSV: %w", err)
		}
		if len(records) == 0 {
			return nil, nil
		}

		// Find url column
		header := records[0]
		urlIdx := -1
		for i, h := range header {
			if strings.EqualFold(strings.TrimSpace(h), "url") {
				urlIdx = i
				break
			}
		}
		if urlIdx == -1 {
			urlIdx = 0 // Default to first column
		}

		var urls []string
		for _, row := range records[1:] {
			if urlIdx < len(row) {
				u := strings.TrimSpace(row[urlIdx])
				if u != "" && (strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")) {
					urls = append(urls, u)
				}
			}
		}
		return urls, nil
	}

	// Plain text, one URL per line
	var urls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		u := strings.TrimSpace(scanner.Text())
		if u != "" && !strings.HasPrefix(u, "#") && (strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")) {
			urls = append(urls, u)
		}
	}
	return urls, scanner.Err()
}
