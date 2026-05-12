// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

type watchSubscription struct {
	ID          string `json:"id"`
	ProductCode string `json:"product_code,omitempty"`
	Applicant   string `json:"applicant,omitempty"`
	Notify      string `json:"notify,omitempty"`
	LastSeen    string `json:"last_seen,omitempty"` // YYYYMMDD cursor
	CreatedAt   string `json:"created_at"`
}

type watchStore struct {
	Subscriptions []watchSubscription `json:"subscriptions"`
}

func watchesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "fda-devices-pp-cli", "watches.json")
}

func loadWatches() (*watchStore, error) {
	p := watchesPath()
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &watchStore{}, nil
		}
		return nil, err
	}
	var ws watchStore
	if err := json.Unmarshal(b, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

func saveWatches(ws *watchStore) error {
	p := watchesPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(ws, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

func newWatchCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "watch",
		Short:       "Subscribe to product codes / applicants; pipe diffs to a sink",
		Annotations: readOnlyAnnotations,
	}
	cmd.AddCommand(newWatchNewCmd(flags), newWatchListCmd(flags), newWatchRunCmd(flags), newWatchTestCmd(flags))
	return cmd
}

func newWatchNewCmd(flags *rootFlags) *cobra.Command {
	var (
		productCode string
		applicant   string
		notify      string
	)
	cmd := &cobra.Command{
		Use:         "new",
		Short:       "Add a new watch subscription",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if productCode == "" && applicant == "" {
				return usageErr(fmt.Errorf("--product-code or --applicant is required"))
			}
			if dryRunOK(flags) {
				return nil
			}
			ws, err := loadWatches()
			if err != nil {
				return err
			}
			sub := watchSubscription{
				ID:          fmt.Sprintf("w%d", time.Now().UnixNano()),
				ProductCode: productCode,
				Applicant:   applicant,
				Notify:      notify,
				CreatedAt:   time.Now().UTC().Format(time.RFC3339),
			}
			ws.Subscriptions = append(ws.Subscriptions, sub)
			if err := saveWatches(ws); err != nil {
				return err
			}
			raw, _ := json.Marshal(sub)
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
	cmd.Flags().StringVar(&productCode, "product-code", "", "3-letter FDA product code")
	cmd.Flags().StringVar(&applicant, "applicant", "", "Applicant company name")
	cmd.Flags().StringVar(&notify, "notify", "", "Sink: file:<path> (slack:/webhook: planned)")
	return cmd
}

func newWatchListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "list",
		Short:       "List all watch subscriptions",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			ws, err := loadWatches()
			if err != nil {
				return err
			}
			raw, _ := json.Marshal(ws.Subscriptions)
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
}

func newWatchRunCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "run",
		Short:       "Run all watches, emitting new records to each sink",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			ws, err := loadWatches()
			if err != nil {
				return err
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			type runResult struct {
				ID       string            `json:"id"`
				NewCount int               `json:"new_count"`
				Records  []json.RawMessage `json:"records"`
			}
			results := []runResult{}
			for i := range ws.Subscriptions {
				sub := &ws.Subscriptions[i]
				clauses := []string{}
				if sub.ProductCode != "" {
					clauses = append(clauses, fmt.Sprintf("product_code:%q", sub.ProductCode))
				}
				if sub.Applicant != "" {
					clauses = append(clauses, fmt.Sprintf("applicant:%q", sub.Applicant))
				}
				cursor := sub.LastSeen
				if cursor == "" {
					cursor = time.Now().UTC().AddDate(0, 0, -30).Format("20060102")
				}
				clauses = append(clauses, fmt.Sprintf("decision_date:[%s TO 99991231]", cursor))
				env, err := openfda.Run(ctx, c, path510k, openfda.Query{
					Search: joinAND(clauses...),
					Sort:   "decision_date:desc",
					Limit:  100,
				})
				if err != nil {
					continue
				}
				if sub.Notify != "" {
					if err := deliverWatch(sub.Notify, env.Results); err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "warning: deliver to %s failed: %v\n", sub.Notify, err)
					}
				}
				// Update cursor to today
				sub.LastSeen = time.Now().UTC().Format("20060102")
				results = append(results, runResult{ID: sub.ID, NewCount: len(env.Results), Records: env.Results})
			}
			_ = saveWatches(ws)
			raw, _ := json.Marshal(results)
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
}

func newWatchTestCmd(flags *rootFlags) *cobra.Command {
	var productCode string
	cmd := &cobra.Command{
		Use:         "test",
		Short:       "Dry-run a watch query without persisting",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if productCode == "" {
				return usageErr(fmt.Errorf("--product-code is required"))
			}
			if dryRunOK(flags) {
				return nil
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			env, err := openfda.Run(ctx, c, path510k, openfda.Query{
				Search: fmt.Sprintf("product_code:%q", productCode),
				Sort:   "decision_date:desc",
				Limit:  5,
			})
			if err != nil {
				return classifyAPIError(err, flags)
			}
			raw, _ := json.Marshal(env.Results)
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
	cmd.Flags().StringVar(&productCode, "product-code", "", "3-letter FDA product code")
	return cmd
}

func deliverWatch(spec string, records []json.RawMessage) error {
	switch {
	case strings.HasPrefix(spec, "file:"):
		path := strings.TrimPrefix(spec, "file:")
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		for _, r := range records {
			if _, err := f.Write(append(r, '\n')); err != nil {
				return err
			}
		}
		return nil
	case strings.HasPrefix(spec, "slack:"), strings.HasPrefix(spec, "webhook:"):
		return fmt.Errorf("notify scheme %q is planned but not implemented in this build", spec)
	default:
		return fmt.Errorf("unknown notify scheme %q (expected file:<path>)", spec)
	}
}
