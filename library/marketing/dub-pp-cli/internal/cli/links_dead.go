package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/dub-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newLinksDeadCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var concurrency int
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "dead",
		Short: "Find links whose destination URLs return errors (404, 500, timeout)",
		Long: `Health-check every link's destination URL to find broken links.
Requires a prior sync. Checks are parallelized for speed.`,
		Example: `  # Find dead links
  dub-pp-cli links dead

  # Check with higher concurrency
  dub-pp-cli links dead --concurrency 20

  # Output as JSON for scripting
  dub-pp-cli links dead --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("dub-pp-cli")
			}
			s, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}
			defer s.Close()

			items, err := s.List("links", 5000)
			if err != nil {
				return fmt.Errorf("listing links: %w", err)
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "No links in local store. Run 'sync' first.")
				return nil
			}

			type linkCheck struct {
				ID     string `json:"id"`
				Key    string `json:"key"`
				Domain string `json:"domain"`
				URL    string `json:"url"`
				Status int    `json:"status"`
				Error  string `json:"error,omitempty"`
			}

			type linkInfo struct {
				id, key, domain, url string
			}

			var links []linkInfo
			for _, item := range items {
				var obj map[string]any
				if err := json.Unmarshal(item, &obj); err != nil {
					continue
				}
				url := strVal(obj, "url")
				if url == "" {
					continue
				}
				links = append(links, linkInfo{
					id:     strVal(obj, "id"),
					key:    strVal(obj, "key"),
					domain: strVal(obj, "domain"),
					url:    url,
				})
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Checking %d links...\n", len(links))

			client := &http.Client{
				Timeout: timeout,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					if len(via) >= 10 {
						return fmt.Errorf("too many redirects")
					}
					return nil
				},
			}

			var dead []linkCheck
			var mu sync.Mutex
			sem := make(chan struct{}, concurrency)
			var wg sync.WaitGroup

			for _, link := range links {
				wg.Add(1)
				sem <- struct{}{}
				go func(l linkInfo) {
					defer wg.Done()
					defer func() { <-sem }()

					lc := linkCheck{
						ID:     l.id,
						Key:    l.key,
						Domain: l.domain,
						URL:    l.url,
					}

					resp, err := client.Head(l.url)
					if err != nil {
						lc.Error = err.Error()
						lc.Status = 0
						mu.Lock()
						dead = append(dead, lc)
						mu.Unlock()
						return
					}
					resp.Body.Close()
					lc.Status = resp.StatusCode

					if resp.StatusCode >= 400 {
						mu.Lock()
						dead = append(dead, lc)
						mu.Unlock()
					}
				}(link)
			}
			wg.Wait()

			if flags.asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(dead)
			}

			if len(dead) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "All links are healthy!")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Found %d dead/broken links:\n\n", len(dead))
			fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-6s %s\n", "Short Link", "Status", "Destination URL")
			fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-6s %s\n", "--------------------", "------", "---")
			for _, lc := range dead {
				status := fmt.Sprintf("%d", lc.Status)
				if lc.Status == 0 {
					status = "ERR"
				}
				url := lc.URL
				if len(url) > 60 {
					url = url[:57] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-6s %s\n", lc.Key, status, url)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().IntVar(&concurrency, "concurrency", 10, "Number of parallel health checks")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Timeout per URL check")

	return cmd
}
