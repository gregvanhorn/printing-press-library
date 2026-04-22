// Hand-authored: Meta OAuth login command.
// Not generated — fills the gap between the generator's auth set-token and
// the Meta-specific long-lived (60-day) token exchange flow.

package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/meta-ads/internal/config"

	"github.com/spf13/cobra"
)

// Register login subcommand. Invoked from init() below; added to the auth group.
func newAuthLoginCmd(flags *rootFlags) *cobra.Command {
	var appID, appSecret, shortToken string
	var noExtend bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Exchange a Meta short-lived token for a 60-day long-lived one",
		Long: `Meta OAuth login with long-lived token exchange.

Meta's Marketing API does not support the OAuth device-code flow, so this
command uses a manual short-lived-token paste flow that avoids juggling
self-signed HTTPS certificates:

  1. Open https://developers.facebook.com/tools/explorer/ in a browser
  2. Pick your app, set scopes to: ads_read, ads_management, business_management
  3. Click "Generate Access Token"
  4. Copy the token and paste it below (or pass via --short-token)
  5. This command exchanges it for a 60-day long-lived token.

Required env vars (or flags):  META_APP_ID, META_APP_SECRET
The resulting long-lived token is saved to your config file and reused by all
subsequent commands. Tokens are silently extended within 7 days of expiry on
any command that makes an authenticated request.`,
		Example: `  # Interactive (prompts for the short-lived token)
  meta-ads-pp-cli auth login

  # Non-interactive (good for agents and CI)
  meta-ads-pp-cli auth login --short-token EAAB... --json

  # Extension is automatic; skip it with --no-extend (use only when diagnosing)
  meta-ads-pp-cli auth login --short-token EAAB... --no-extend`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}

			if appID == "" {
				appID = os.Getenv("META_APP_ID")
			}
			if appSecret == "" {
				appSecret = os.Getenv("META_APP_SECRET")
			}
			if appID == "" || appSecret == "" {
				return usageErr(fmt.Errorf("META_APP_ID and META_APP_SECRET must be set (env or --app-id/--app-secret flags)"))
			}

			token := strings.TrimSpace(shortToken)
			if token == "" {
				if flags.noInput {
					return usageErr(fmt.Errorf("--short-token required when --no-input is set"))
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Paste your short-lived Meta access token (from Graph API Explorer):")
				fmt.Fprint(cmd.OutOrStdout(), "> ")
				r := bufio.NewReader(os.Stdin)
				line, _ := r.ReadString('\n')
				token = strings.TrimSpace(line)
			}
			if token == "" {
				return usageErr(fmt.Errorf("no token provided"))
			}

			resp, expiresIn, err := exchangeForLongLived(cfg.BaseURL, appID, appSecret, token, noExtend)
			if err != nil {
				return authErr(err)
			}

			expiry := time.Now().Add(time.Duration(expiresIn) * time.Second)
			if err := cfg.SaveTokens(appID, appSecret, resp, "", expiry); err != nil {
				return configErr(fmt.Errorf("saving token: %w", err))
			}

			if flags.asJSON {
				out := map[string]any{
					"status":      "authenticated",
					"config_path": cfg.Path,
					"expires_at":  expiry.UTC().Format(time.RFC3339),
					"expires_in":  expiresIn,
					"extended":    !noExtend,
				}
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(b))
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s — long-lived token saved.\n", green("Authenticated"))
			fmt.Fprintf(cmd.OutOrStdout(), "  Expires: %s (%d days)\n", expiry.Format("2006-01-02 15:04 UTC"), expiresIn/86400)
			fmt.Fprintf(cmd.OutOrStdout(), "  Config:  %s\n", cfg.Path)
			fmt.Fprintln(cmd.OutOrStdout(), "  Tokens are extended automatically when within 7 days of expiry.")
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "Meta app ID (or set META_APP_ID env var)")
	cmd.Flags().StringVar(&appSecret, "app-secret", "", "Meta app secret (or set META_APP_SECRET env var)")
	cmd.Flags().StringVar(&shortToken, "short-token", "", "Short-lived Meta access token (if unset, prompts unless --no-input)")
	cmd.Flags().BoolVar(&noExtend, "no-extend", false, "Skip the long-lived exchange and store the short-lived token as-is (diagnostic only)")
	return cmd
}

// exchangeForLongLived trades a short-lived token for a 60-day one via the Graph API.
// If noExtend is true, returns the input token with the conventional 60d fallback expiry.
func exchangeForLongLived(baseURL, appID, appSecret, shortToken string, noExtend bool) (string, int, error) {
	defaultExpiry := 60 * 24 * 60 * 60 // 60 days in seconds
	if noExtend {
		return shortToken, defaultExpiry, nil
	}

	// Graph API endpoint lives at the root of graph.facebook.com, not the versioned path.
	// Our config.BaseURL is https://graph.facebook.com/vXX.0 — strip the version path.
	host := baseURL
	if i := strings.Index(baseURL[len("https://"):], "/"); i > 0 {
		host = "https://" + baseURL[len("https://"):len("https://")+i]
	}
	endpoint := host + "/oauth/access_token"

	q := url.Values{}
	q.Set("grant_type", "fb_exchange_token")
	q.Set("client_id", appID)
	q.Set("client_secret", appSecret)
	q.Set("fb_exchange_token", shortToken)

	resp, err := http.Get(endpoint + "?" + q.Encode())
	if err != nil {
		return "", 0, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		// Try to extract the Graph API error message for a helpful CLI error.
		var errWrap struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &errWrap) == nil && errWrap.Error.Message != "" {
			return "", 0, fmt.Errorf("token exchange failed (code %d): %s", errWrap.Error.Code, errWrap.Error.Message)
		}
		return "", 0, fmt.Errorf("token exchange HTTP %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", 0, fmt.Errorf("parsing token response: %w (body: %s)", err, string(body))
	}
	if out.AccessToken == "" {
		return "", 0, fmt.Errorf("empty access_token in response: %s", string(body))
	}
	if out.ExpiresIn == 0 {
		out.ExpiresIn = defaultExpiry
	}
	return out.AccessToken, out.ExpiresIn, nil
}
