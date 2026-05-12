// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/client"
	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/predicate"
)

type predicateNode struct {
	KNumber      string           `json:"k_number"`
	Applicant    string           `json:"applicant,omitempty"`
	DeviceName   string           `json:"device_name,omitempty"`
	DecisionDate string           `json:"decision_date,omitempty"`
	ProductCode  string           `json:"product_code,omitempty"`
	Primary      bool             `json:"primary,omitempty"`
	Recalled     bool             `json:"recalled,omitempty"`
	RecallCount  int              `json:"recall_count,omitempty"`
	Depth        int              `json:"depth"`
	Source       string           `json:"summary_pdf,omitempty"`
	Note         string           `json:"note,omitempty"`
	Predicates   []*predicateNode `json:"predicates,omitempty"`
}

func newPredicateChainCmd(flags *rootFlags) *cobra.Command {
	var depth int
	cmd := &cobra.Command{
		Use:         "predicate-chain <K-number>",
		Short:       "Recursively walk the predicate ancestry of a 510(k) clearance",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return nil
			}
			kNumber := args[0]
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			home, _ := os.UserHomeDir()
			cacheDir := filepath.Join(home, ".fda-devices-pp-cli", "predicate-cache")

			visited := map[string]*predicateNode{}
			root, err := walkPredicates(ctx, c, kNumber, 0, depth, cacheDir, visited, "")
			if err != nil {
				return err
			}
			raw, err := json.Marshal(root)
			if err != nil {
				return err
			}
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
	cmd.Flags().IntVar(&depth, "depth", 3, "Max predicate chain depth (0 = subject only)")
	return cmd
}

func walkPredicates(ctx context.Context, c *client.Client, kNumber string, curDepth, maxDepth int, cacheDir string, visited map[string]*predicateNode, primary string) (*predicateNode, error) {
	if existing, ok := visited[kNumber]; ok {
		// Cycle: return a shallow reference to the already-built node.
		return &predicateNode{KNumber: kNumber, Depth: curDepth, Note: "cycle: already visited at depth " + fmt.Sprint(existing.Depth)}, nil
	}

	node := &predicateNode{KNumber: kNumber, Depth: curDepth}
	visited[kNumber] = node
	if primary == kNumber {
		node.Primary = true
	}

	rec, err := openfda.One(ctx, c, path510k, openfda.Query{
		Search: fmt.Sprintf("k_number:%q", kNumber),
		Limit:  1,
	})
	if err == nil && rec != nil {
		var f struct {
			Applicant    string `json:"applicant"`
			DeviceName   string `json:"device_name"`
			DecisionDate string `json:"decision_date"`
			ProductCode  string `json:"product_code"`
		}
		_ = json.Unmarshal(rec, &f)
		node.Applicant = f.Applicant
		node.DeviceName = f.DeviceName
		node.DecisionDate = f.DecisionDate
		node.ProductCode = f.ProductCode
	}

	recEnv, _ := openfda.Run(ctx, c, "/device/recall.json", openfda.Query{
		Search: fmt.Sprintf("k_numbers:%q", kNumber),
		Limit:  1,
	})
	if recEnv != nil {
		if t, ok := recEnv.Meta.Results["total"].(float64); ok && t > 0 {
			node.Recalled = true
			node.RecallCount = int(t)
		}
	}

	if curDepth >= maxDepth {
		return node, nil
	}

	pred, err := predicate.Fetch(ctx, kNumber, cacheDir)
	if err != nil {
		if predicate.ErrNotFound(err) {
			node.Note = "no FDA summary PDF published for this K-number"
		} else {
			node.Note = "predicate fetch failed: " + err.Error()
		}
		return node, nil
	}
	node.Source = pred.Source
	if pred.Note != "" {
		node.Note = pred.Note
	}
	for _, child := range pred.Predicates {
		childNode, _ := walkPredicates(ctx, c, child, curDepth+1, maxDepth, cacheDir, visited, pred.Primary)
		node.Predicates = append(node.Predicates, childNode)
	}
	return node, nil
}
