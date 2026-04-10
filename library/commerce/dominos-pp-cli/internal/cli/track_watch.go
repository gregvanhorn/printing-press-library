package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var trackStages = []string{"Order Placed", "Prep", "Bake", "Quality Check", "Out for Delivery", "Delivered"}

func newTrackCmd(flags *rootFlags) *cobra.Command {
	var phone string
	var watch bool
	var interval, timeout time.Duration
	cmd := &cobra.Command{
		Use:   "track --phone <phone> [--watch]",
		Short: "Track an order and optionally watch for status updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if !watch {
				return renderTrackOnce(cmd, flags, c, phone)
			}
			return watchTrack(cmd, flags, c, phone, interval, timeout)
		},
	}
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number associated with the order")
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll the tracker until delivered")
	cmd.Flags().DurationVar(&interval, "interval", 30*time.Second, "Polling interval")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Hour, "Maximum watch duration")
	_ = cmd.MarkFlagRequired("phone")
	return cmd
}

func renderTrackOnce(cmd *cobra.Command, flags *rootFlags, c interface {
	Get(string, map[string]string) (json.RawMessage, error)
}, phone string) error {
	data, snap, err := fetchTrack(c, phone)
	if err != nil {
		return err
	}
	if flags.asJSON {
		return flags.printJSON(cmd, snap)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", progressBar(snap.Stage), snap.Stage)
	if snap.Status != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", snap.Status)
	}
	if snap.ETA != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "ETA: %s\n", snap.ETA)
	}
	if snap.Stage == "" {
		return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
	}
	return nil
}

func watchTrack(cmd *cobra.Command, flags *rootFlags, c interface {
	Get(string, map[string]string) (json.RawMessage, error)
}, phone string, interval, timeout time.Duration) error {
	deadline, ticker, last := time.Now().Add(timeout), time.NewTicker(interval), ""
	defer ticker.Stop()
	for {
		_, snap, err := fetchTrack(c, phone)
		if err != nil {
			return err
		}
		line := progressBar(snap.Stage) + " " + firstNonEmpty(snap.Stage, snap.Status)
		if snap.ETA != "" {
			line += " ETA " + snap.ETA
		}
		if line != last {
			last = line
			if flags.asJSON {
				if err := flags.printJSON(cmd, snap); err != nil {
					return err
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
		}
		if snap.Stage == "Delivered" {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("tracking watch timed out after %s", timeout)
		}
		<-ticker.C
	}
}

type trackSnapshot struct{ Phone, Stage, Status, ETA string }

func fetchTrack(c interface {
	Get(string, map[string]string) (json.RawMessage, error)
}, phone string) (json.RawMessage, trackSnapshot, error) {
	data, err := c.Get("/power/tracker", map[string]string{"Phone": phone})
	if err != nil {
		return nil, trackSnapshot{}, classifyAPIError(err)
	}
	snap := trackSnapshot{Phone: phone}
	for _, m := range findMaps(data) {
		s := firstNonEmpty(pickS(m, "status", "Status", "orderStatus", "stage"), pickS(m, "description", "StatusDescription"))
		if stage := normalizeStage(s + " " + pickS(m, "deliveryStatus", "serviceMethod")); stage != "" {
			snap.Stage = stage
			snap.Status = s
			snap.ETA = firstNonEmpty(pickS(m, "eta", "estimatedDeliveryTime", "estimatedTime"), snap.ETA)
			break
		}
	}
	return data, snap, nil
}

func normalizeStage(v string) string {
	s := strings.ToLower(v)
	switch {
	case strings.Contains(s, "deliver"):
		return "Delivered"
	case strings.Contains(s, "out for delivery"), strings.Contains(s, "driver"):
		return "Out for Delivery"
	case strings.Contains(s, "quality"):
		return "Quality Check"
	case strings.Contains(s, "bake"), strings.Contains(s, "oven"):
		return "Bake"
	case strings.Contains(s, "prep"), strings.Contains(s, "make"), strings.Contains(s, "makeline"):
		return "Prep"
	case strings.Contains(s, "placed"), strings.Contains(s, "received"), strings.Contains(s, "order"):
		return "Order Placed"
	default:
		return ""
	}
}

func progressBar(stage string) string {
	idx := 0
	for i, s := range trackStages {
		if s == stage {
			idx = i + 1
			break
		}
	}
	return "[" + strings.Repeat("#", idx) + strings.Repeat("-", len(trackStages)-idx) + "]"
}
