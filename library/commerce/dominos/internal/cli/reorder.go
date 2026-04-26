package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

type reorderFlags struct {
	useLast               bool
	templateName          string
	storeID               string
	serviceMethod         string
	substituteUnavailable bool
}

type reorderMenuItem struct {
	Code string
	Name string
}

func newReorderCmd(flags *rootFlags) *cobra.Command {
	opts := &reorderFlags{}
	cmd := &cobra.Command{
		Use:   "reorder",
		Short: "Replay a saved order template against today's menu",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.templateName != "" && opts.useLast {
				return fmt.Errorf("use either --last or --template, not both")
			}
			tpl, err := reorderResolveTemplate(opts)
			if err != nil {
				return err
			}
			order := templateOrderBody(tpl, "")
			if opts.storeID != "" {
				order["StoreID"] = strings.TrimSpace(opts.storeID)
			}
			if opts.serviceMethod != "" {
				order["ServiceMethod"] = strings.TrimSpace(opts.serviceMethod)
			}
			storeID := comparePricesString(order["StoreID"])
			if storeID == "" {
				return fmt.Errorf("template %q has no StoreID", tpl.Name)
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			path := replacePathParam("/power/store/{storeID}/menu", "storeID", storeID)
			menuRaw, _, err := resolveRead(c, flags, "menu", false, path, map[string]string{"lang": "en", "structured": "true"})
			if err != nil {
				return classifyAPIError(err)
			}
			menu, err := reorderParseMenu(extractResponseData(menuRaw))
			if err != nil {
				return err
			}

			products, ok := order["Products"].([]any)
			if !ok {
				return fmt.Errorf("template %q has invalid Products", tpl.Name)
			}
			nextProducts, substituted, dropped := reorderProducts(products, menu, opts.substituteUnavailable, cmd.ErrOrStderr())
			order["Products"] = nextProducts
			order = ensureOrderTaker(order).(map[string]any)

			fmt.Fprintf(cmd.ErrOrStderr(), "Reorder built from template %q — %d items, %d substituted, %d dropped\n", tpl.Name, len(nextProducts), substituted, dropped)
			return writePrettyJSON(cmd.OutOrStdout(), order)
		},
	}
	cmd.Flags().BoolVar(&opts.useLast, "last", false, "Use the most recent template")
	cmd.Flags().StringVar(&opts.templateName, "template", "", "Use a specific saved template by name")
	cmd.Flags().StringVar(&opts.storeID, "store", "", "Override store ID")
	cmd.Flags().BoolVar(&opts.substituteUnavailable, "substitute-unavailable", false, "Substitute unavailable items with the closest menu match")
	cmd.Flags().StringVar(&opts.serviceMethod, "service", "", "Override service method (Delivery or Carryout)")
	return cmd
}

func reorderResolveTemplate(opts *reorderFlags) (*orderTemplate, error) {
	if opts.templateName != "" {
		tpl, err := getTemplate(opts.templateName)
		if err != nil {
			return nil, err
		}
		if tpl == nil {
			return nil, fmt.Errorf("template %q not found", opts.templateName)
		}
		return tpl, nil
	}
	templates, err := listTemplates()
	if err != nil {
		return nil, err
	}
	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found; run 'dominos-pp-cli template save <name> ...' first")
	}
	sort.Slice(templates, func(i, j int) bool { return templates[i].CreatedAt.After(templates[j].CreatedAt) })
	return &templates[0], nil
}

func reorderParseMenu(data json.RawMessage) (map[string]any, error) {
	var menu map[string]any
	if err := json.Unmarshal(data, &menu); err != nil {
		return nil, fmt.Errorf("parse menu response: %w", err)
	}
	return menu, nil
}

func reorderProducts(products []any, menu map[string]any, substitute bool, stderr io.Writer) ([]any, int, int) {
	variants, _ := nutritionMap(menu["Variants"])
	if len(variants) == 0 {
		variants, _ = nutritionMap(menu["variants"])
	}
	candidates := reorderCandidates(menu)
	out := make([]any, 0, len(products))
	substituted := 0
	dropped := 0
	for _, raw := range products {
		product, ok := nutritionMap(raw)
		if !ok {
			out = append(out, raw)
			continue
		}
		code := comparePricesString(product["Code"])
		if code == "" {
			code = comparePricesString(product["code"])
		}
		if code == "" {
			out = append(out, product)
			continue
		}
		if _, ok := variants[code]; ok {
			out = append(out, product)
			continue
		}
		oldName := reorderProductName(product)
		if substitute {
			if match, score, ok := reorderBestCandidate(oldName, code, candidates); ok {
				product["Code"] = match.Code
				delete(product, "code")
				fmt.Fprintf(stderr, "Substituted: %s ('%s') → %s ('%s') [match score: %.2f]\n", code, oldName, match.Code, match.Name, score)
				out = append(out, product)
				substituted++
				continue
			}
		}
		fmt.Fprintf(stderr, "Dropped: %s ('%s') — no longer on menu\n", code, oldName)
		dropped++
	}
	return out, substituted, dropped
}

func reorderCandidates(menu map[string]any) []reorderMenuItem {
	items := make([]reorderMenuItem, 0)
	seen := map[string]struct{}{}
	for _, key := range []string{"Variants", "Products", "variants", "products"} {
		nodes, _ := nutritionMap(menu[key])
		for code, raw := range nodes {
			obj, ok := nutritionMap(raw)
			if !ok {
				continue
			}
			itemCode := comparePricesString(obj["Code"])
			if itemCode == "" {
				itemCode = comparePricesString(obj["code"])
			}
			if itemCode == "" {
				itemCode = code
			}
			if itemCode == "" {
				continue
			}
			if _, ok := seen[itemCode]; ok {
				continue
			}
			name := comparePricesString(obj["Name"])
			if name == "" {
				name = comparePricesString(obj["name"])
			}
			if name == "" {
				continue
			}
			items = append(items, reorderMenuItem{Code: itemCode, Name: name})
			seen[itemCode] = struct{}{}
		}
	}
	return items
}

func reorderBestCandidate(name, code string, candidates []reorderMenuItem) (reorderMenuItem, float64, bool) {
	query := strings.TrimSpace(name)
	if query == "" {
		query = code
	}
	best := reorderMenuItem{}
	bestScore := 0.0
	for _, candidate := range candidates {
		score := reorderSimilarity(query, candidate.Name)
		if strings.EqualFold(code, candidate.Code) {
			score = 1
		}
		if score > bestScore {
			best = candidate
			bestScore = score
		}
	}
	return best, bestScore, best.Code != "" && bestScore >= 0.35
}

func reorderSimilarity(a, b string) float64 {
	na := reorderNormalize(a)
	nb := reorderNormalize(b)
	if na == "" || nb == "" {
		return 0
	}
	maxLen := len(na)
	if len(nb) > maxLen {
		maxLen = len(nb)
	}
	editScore := 1 - float64(levenshteinDistance(na, nb))/float64(maxLen)
	if editScore < 0 {
		editScore = 0
	}
	at := reorderTokens(na)
	bt := reorderTokens(nb)
	if len(at) == 0 || len(bt) == 0 {
		return editScore
	}
	shared := 0
	for token := range at {
		if _, ok := bt[token]; ok {
			shared++
		}
	}
	tokenScore := float64(shared*2) / float64(len(at)+len(bt))
	return (editScore * 0.6) + (tokenScore * 0.4)
}

func reorderProductName(product map[string]any) string {
	for _, key := range []string{"Name", "name"} {
		if name := comparePricesString(product[key]); name != "" {
			return name
		}
	}
	if opts, ok := product["Options"].([]any); ok {
		parts := make([]string, 0, len(opts))
		for _, raw := range opts {
			if obj, ok := nutritionMap(raw); ok {
				if name := comparePricesString(obj["Name"]); name != "" {
					parts = append(parts, name)
				}
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, " ")
		}
	}
	return comparePricesString(product["Code"])
}

func reorderNormalize(s string) string {
	var b strings.Builder
	lastSpace := true
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func reorderTokens(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, token := range strings.Fields(s) {
		out[token] = struct{}{}
	}
	return out
}
