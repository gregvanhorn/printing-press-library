package cli

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type orderBulkRow struct {
	ItemCode string
	Qty      int
	Options  map[string]string
	TagNote  string
}

type orderBulkStore struct {
	StoreID  string
	Name     string
	Distance float64
}

type orderBulkEval struct {
	Store    orderBulkStore
	Total    float64
	Missing  []string
	Included []orderBulkRow
}

func newOrderBulkCmd(flags *rootFlags) *cobra.Command {
	var flagCSV, flagAddress, flagCity, flagService, flagOutput string
	var flagMaxStores int

	cmd := &cobra.Command{
		Use:   "order-bulk",
		Short: "Build a combined group order cart from a CSV and pick the best nearby store",
		Example: "  dominos-pp-cli order-bulk --csv group.csv --address \"123 Main St\" --city \"Seattle, WA 98101\"\n" +
			"  dominos-pp-cli order-bulk --csv group.csv --address \"123 Main St\" --city \"Seattle, WA 98101\" --service Carryout\n" +
			"  dominos-pp-cli order-bulk --csv group.csv --address \"123 Main St\" --city \"Seattle, WA 98101\" --dry-run",
		RunE: func(cmd *cobra.Command, args []string) error {
			// When invoked with no flags at all, print help so users (and verify) get a useful response.
			if !flags.dryRun && strings.TrimSpace(flagCSV) == "" && strings.TrimSpace(flagAddress) == "" && strings.TrimSpace(flagCity) == "" {
				return cmd.Help()
			}
			if flagMaxStores <= 0 {
				return fmt.Errorf("--max-stores must be greater than 0")
			}
			if flagService != "Delivery" && flagService != "Carryout" {
				return fmt.Errorf("invalid --service value %q: must be Delivery or Carryout", flagService)
			}
			if flags.dryRun {
				c, err := flags.newClient()
				if err != nil {
					return err
				}
				_, err = c.Get("/power/store-locator", map[string]string{"s": flagAddress, "c": flagCity, "type": flagService})
				if err != nil {
					return classifyAPIError(err)
				}
				return nil
			}
			for name, value := range map[string]string{"csv": flagCSV, "address": flagAddress, "city": flagCity} {
				if strings.TrimSpace(value) == "" {
					return fmt.Errorf("required flag %q not set", name)
				}
			}

			rows, err := orderBulkReadCSV(flagCSV)
			if err != nil {
				return err
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			storesRaw, err := c.Get("/power/store-locator", map[string]string{"s": flagAddress, "c": flagCity, "type": flagService})
			if err != nil {
				return classifyAPIError(err)
			}

			stores, err := orderBulkParseStores(storesRaw)
			if err != nil {
				return fmt.Errorf("parse store locator response: %w", err)
			}
			if len(stores) == 0 {
				return fmt.Errorf("no stores found for the given address")
			}
			if len(stores) > flagMaxStores {
				stores = stores[:flagMaxStores]
			}

			evals, err := orderBulkEvaluateStores(c, stores, rows)
			if err != nil {
				return err
			}
			chosen := orderBulkChooseStore(evals)
			if chosen == nil {
				return fmt.Errorf("no stores could be evaluated")
			}
			if len(chosen.Missing) > 0 {
				fmt.Fprintln(os.Stderr, "warning: no nearby store had every requested item")
				for _, eval := range evals {
					fmt.Fprintf(os.Stderr, "store %s (%s) missing: %s\n", eval.Store.StoreID, eval.Store.Name, strings.Join(eval.Missing, ", "))
				}
				fmt.Fprintf(os.Stderr, "warning: dropped missing items from store %s: %s\n", chosen.Store.StoreID, strings.Join(chosen.Missing, ", "))
			}

			cart := orderBulkBuildCart(chosen, flagService, flagAddress, flagCity)
			fmt.Fprintf(os.Stderr, "Picked store %s (%s): %d items, $%.2f\n", chosen.Store.StoreID, chosen.Store.Name, len(chosen.Included), chosen.Total)

			data, err := json.MarshalIndent(cart, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal cart JSON: %w", err)
			}
			data = append(data, '\n')
			if strings.TrimSpace(flagOutput) != "" {
				if err := os.WriteFile(flagOutput, data, 0o644); err != nil {
					return fmt.Errorf("write output file: %w", err)
				}
				return nil
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}

	cmd.Flags().StringVar(&flagCSV, "csv", "", "CSV path")
	cmd.Flags().StringVar(&flagAddress, "address", "", "Delivery address street")
	cmd.Flags().StringVar(&flagCity, "city", "", "City, state, zip")
	cmd.Flags().StringVar(&flagService, "service", "Delivery", "Service type: Delivery or Carryout")
	cmd.Flags().IntVar(&flagMaxStores, "max-stores", 5, "Max stores to compare for best")
	cmd.Flags().StringVar(&flagOutput, "output", "", "Write combined cart JSON to file instead of stdout")
	return cmd
}

func orderBulkReadCSV(path string) ([]orderBulkRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open CSV: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		buf.WriteString(scanner.Text())
		buf.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}

	r := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV requires a header row and at least one order row")
	}

	index := map[string]int{}
	for i, col := range records[0] {
		index[strings.ToLower(strings.TrimSpace(col))] = i
	}
	for _, name := range []string{"name", "item_code", "qty", "toppings", "notes"} {
		if _, ok := index[name]; !ok {
			return nil, fmt.Errorf("CSV missing required header %q", name)
		}
	}

	rows := make([]orderBulkRow, 0, len(records)-1)
	for i, rec := range records[1:] {
		get := func(col string) string {
			pos := index[col]
			if pos >= len(rec) {
				return ""
			}
			return strings.TrimSpace(rec[pos])
		}
		code := get("item_code")
		if code == "" {
			continue
		}
		qty, err := strconv.Atoi(get("qty"))
		if err != nil || qty <= 0 {
			return nil, fmt.Errorf("row %d has invalid qty %q", i+2, get("qty"))
		}
		options, ok := orderBulkParseToppings(get("toppings"))
		if !ok {
			fmt.Fprintf(os.Stderr, "warning: row %d toppings could not be parsed; omitting toppings\n", i+2)
		}
		note := strings.TrimSpace(get("name"))
		if extra := strings.TrimSpace(get("notes")); extra != "" {
			if note != "" {
				note += ": " + extra
			} else {
				note = extra
			}
		}
		rows = append(rows, orderBulkRow{ItemCode: code, Qty: qty, Options: options, TagNote: note})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("CSV did not contain any valid order rows")
	}
	return rows, nil
}

func orderBulkParseStores(data json.RawMessage) ([]orderBulkStore, error) {
	rawStores, err := comparePricesParseStores(data)
	if err != nil {
		return nil, err
	}
	rawItems := []map[string]any(nil)
	var env struct {
		Stores []map[string]any `json:"Stores"`
	}
	if json.Unmarshal(data, &env) == nil && len(env.Stores) > 0 {
		rawItems = env.Stores
	} else {
		_ = json.Unmarshal(data, &rawItems)
	}
	names := map[string]string{}
	for _, item := range rawItems {
		storeID := comparePricesString(item["StoreID"])
		if storeID == "" {
			storeID = comparePricesString(item["StoreId"])
		}
		if storeID == "" {
			storeID = comparePricesString(item["store_id"])
		}
		if storeID == "" {
			continue
		}
		name := strings.TrimSpace(comparePricesString(item["Name"]))
		if name == "" {
			name = strings.TrimSpace(comparePricesString(item["StoreName"]))
		}
		if name != "" {
			names[storeID] = name
		}
	}
	stores := make([]orderBulkStore, 0, len(rawStores))
	for _, store := range rawStores {
		name := names[store.StoreID]
		if name == "" {
			name = "Unknown"
		}
		stores = append(stores, orderBulkStore{StoreID: store.StoreID, Name: name, Distance: store.Distance})
	}
	return stores, nil
}

func orderBulkEvaluateStores(c interface {
	Get(string, map[string]string) (json.RawMessage, error)
}, stores []orderBulkStore, rows []orderBulkRow) ([]orderBulkEval, error) {
	evals := make([]orderBulkEval, 0, len(stores))
	for _, store := range stores {
		menuRaw, err := c.Get("/power/store/"+store.StoreID+"/menu", map[string]string{"lang": "en", "structured": "true"})
		if err != nil {
			return nil, fmt.Errorf("fetch menu for store %s: %w", store.StoreID, classifyAPIError(err))
		}
		prices, err := comparePricesParseMenuPrices(extractResponseData(menuRaw))
		if err != nil {
			return nil, fmt.Errorf("parse menu for store %s: %w", store.StoreID, err)
		}
		eval := orderBulkEval{Store: store, Included: make([]orderBulkRow, 0, len(rows))}
		missing := map[string]struct{}{}
		for _, row := range rows {
			price, ok := prices[row.ItemCode]
			if !ok {
				if _, seen := missing[row.ItemCode]; !seen {
					eval.Missing = append(eval.Missing, row.ItemCode)
					missing[row.ItemCode] = struct{}{}
				}
				continue
			}
			eval.Total += price * float64(row.Qty)
			eval.Included = append(eval.Included, row)
		}
		evals = append(evals, eval)
	}
	return evals, nil
}

func orderBulkChooseStore(evals []orderBulkEval) *orderBulkEval {
	sort.Slice(evals, func(i, j int) bool {
		iFull, jFull := len(evals[i].Missing) == 0, len(evals[j].Missing) == 0
		if iFull != jFull {
			return iFull
		}
		if len(evals[i].Missing) != len(evals[j].Missing) {
			return len(evals[i].Missing) < len(evals[j].Missing)
		}
		if evals[i].Total != evals[j].Total {
			return evals[i].Total < evals[j].Total
		}
		if evals[i].Store.Distance != evals[j].Store.Distance {
			return evals[i].Store.Distance < evals[j].Store.Distance
		}
		return evals[i].Store.StoreID < evals[j].Store.StoreID
	})
	if len(evals) == 0 {
		return nil
	}
	return &evals[0]
}

func orderBulkBuildCart(chosen *orderBulkEval, service, street, cityLine string) map[string]any {
	products := make([]map[string]any, 0, len(chosen.Included))
	for i, row := range chosen.Included {
		product := map[string]any{"Code": row.ItemCode, "Qty": row.Qty, "ID": i + 1}
		if len(row.Options) > 0 {
			product["Options"] = row.Options
		}
		if row.TagNote != "" {
			product["Tags"] = map[string]any{"Notes": row.TagNote}
		}
		products = append(products, product)
	}
	order := map[string]any{
		"StoreID":       chosen.Store.StoreID,
		"ServiceMethod": service,
		"Address":       orderBulkParseAddress(street, cityLine),
		"Products":      products,
	}
	return map[string]any{"Order": ensureOrderTaker(order)}
}

func orderBulkParseAddress(street, cityLine string) map[string]any {
	addr := map[string]any{"Street": strings.TrimSpace(street)}
	parts := strings.Split(cityLine, ",")
	if len(parts) >= 2 {
		addr["City"] = strings.TrimSpace(parts[0])
		stateZip := strings.Fields(strings.Join(parts[1:], " "))
		if len(stateZip) > 0 {
			addr["Region"] = stateZip[0]
		}
		if len(stateZip) > 1 {
			addr["PostalCode"] = stateZip[len(stateZip)-1]
		}
		return addr
	}
	tokens := strings.Fields(cityLine)
	if len(tokens) == 0 {
		return addr
	}
	if last := tokens[len(tokens)-1]; len(last) == 5 || (len(last) == 10 && last[5] == '-') {
		addr["PostalCode"] = last
		tokens = tokens[:len(tokens)-1]
	}
	if len(tokens) > 0 && len(tokens[len(tokens)-1]) <= 3 {
		addr["Region"] = tokens[len(tokens)-1]
		tokens = tokens[:len(tokens)-1]
	}
	addr["City"] = strings.TrimSpace(strings.Join(tokens, " "))
	return addr
}

func orderBulkParseToppings(raw string) (map[string]string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, true
	}
	options := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		fields := strings.Split(strings.TrimSpace(part), ":")
		if len(fields) != 3 {
			return nil, false
		}
		code := strings.TrimSpace(fields[0])
		portion := strings.ToLower(strings.TrimSpace(fields[1]))
		weight := strings.TrimSpace(fields[2])
		if code == "" || (portion != "full" && portion != "left" && portion != "right") {
			return nil, false
		}
		if _, err := strconv.ParseFloat(weight, 64); err != nil {
			return nil, false
		}
		switch portion {
		case "full":
			options[code] = weight + "/1"
		case "left":
			options[code] = weight + "/2"
		case "right":
			options[code] = "2/" + weight
		}
	}
	return options, true
}
