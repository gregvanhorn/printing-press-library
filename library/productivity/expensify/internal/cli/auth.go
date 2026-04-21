// Copyright 2026 matt-van-horn. Licensed under Apache-2.0.
// Expensify authentication: session login (headed browser), set-token, set-keys,
// status, logout.

package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mvanhorn/printing-press-library/library/productivity/expensify/internal/client"
	"github.com/mvanhorn/printing-press-library/library/productivity/expensify/internal/config"
	"github.com/mvanhorn/printing-press-library/library/productivity/expensify/internal/credentials"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newAuthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Expensify session and partner credentials",
	}

	cmd.AddCommand(newAuthStatusCmd(flags))
	cmd.AddCommand(newAuthLoginCmd(flags))
	cmd.AddCommand(newAuthSetTokenCmd(flags))
	cmd.AddCommand(newAuthSetKeysCmd(flags))
	cmd.AddCommand(newAuthStoreCredentialsCmd(flags))
	cmd.AddCommand(newAuthLogoutCmd(flags))

	return cmd
}

func newAuthStatusCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show authentication status",
		Example: "  expensify-pp-cli auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			w := cmd.OutOrStdout()

			// Compose the email + headless-credentials lines once so both the
			// authenticated and not-authenticated branches show them.
			emailLine := "  Email: not configured"
			if cfg.ExpensifyEmail != "" {
				emailLine = fmt.Sprintf("  Email: %s", cfg.ExpensifyEmail)
			}
			credsLine := "  Headless credentials: not configured"
			if cfg.ExpensifyEmail != "" && credentials.Has(cfg.ExpensifyEmail) {
				credsLine = "  Headless credentials: configured"
			}

			if !cfg.HasSessionAuth() && !cfg.HasPartnerAuth() {
				fmt.Fprintln(w, red("Not authenticated"))
				fmt.Fprintln(w, emailLine)
				fmt.Fprintln(w, credsLine)
				fmt.Fprintln(w, "")
				fmt.Fprintln(w, "Log in with a headed browser:")
				fmt.Fprintln(w, "  expensify-pp-cli auth login")
				fmt.Fprintln(w, "Or set manually:")
				fmt.Fprintln(w, "  export EXPENSIFY_AUTH_TOKEN=<token>")
				fmt.Fprintln(w, "  export EXPENSIFY_PARTNER_USER_ID=<id>")
				fmt.Fprintln(w, "  export EXPENSIFY_PARTNER_USER_SECRET=<secret>")
				fmt.Fprintln(w, "Or store headless credentials for automatic token minting:")
				fmt.Fprintln(w, "  expensify-pp-cli auth store-credentials --email <you> --password <pw>")
				return authErr(fmt.Errorf("no credentials configured"))
			}
			fmt.Fprintln(w, green("Authenticated"))
			if cfg.HasSessionAuth() {
				fmt.Fprintf(w, "  Session token: present (%d chars)\n", len(cfg.ExpensifyAuthToken))
			}
			if cfg.HasPartnerAuth() {
				fmt.Fprintln(w, "  Partner credentials: set")
			}
			fmt.Fprintln(w, emailLine)
			fmt.Fprintln(w, credsLine)
			if cfg.AuthSource != "" {
				fmt.Fprintf(w, "  Source: %s\n", cfg.AuthSource)
			}
			fmt.Fprintf(w, "  Config: %s\n", cfg.Path)
			return nil
		},
	}
}

func newAuthLoginCmd(flags *rootFlags) *cobra.Command {
	var sessionName string
	var fallbackToken string
	var headless bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in via a headed browser and capture the Expensify session authToken",
		Long: `Launches a headed Chromium browser via 'agent-browser' to https://new.expensify.com/,
waits for you to finish logging in, then reads the authToken cookie and persists it.

If agent-browser is not installed, falls back to asking for the token on stdin.

With --headless, posts stored email+password (see 'auth store-credentials') to
Expensify's /Authenticate endpoint and mints a session token without a browser.
Accounts with 2FA fall back with a clear message.`,
		Example: "  expensify-pp-cli auth login\n  expensify-pp-cli auth login --headless",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			w := cmd.OutOrStdout()

			// --headless short-circuits the headed-browser path and mints a
			// token directly from stored credentials. Everything below this
			// branch preserves the existing headed flow unchanged.
			if headless {
				c := client.New(cfg, flags.timeout, 0)
				return doHeadlessLogin(cmd, cfg, flags, c.Authenticate)
			}

			// Explicit --token overrides and short-circuits the browser flow.
			if fallbackToken != "" {
				if err := cfg.SaveSessionToken(strings.TrimSpace(fallbackToken), ""); err != nil {
					return configErr(err)
				}
				fmt.Fprintf(w, "Token saved to %s\n", cfg.Path)
				return nil
			}

			if _, err := exec.LookPath("agent-browser"); err != nil {
				return fallbackPromptForToken(cmd, cfg, flags, "agent-browser not found on PATH")
			}

			fmt.Fprintf(w, "Opening https://new.expensify.com/ in session %q ...\n", sessionName)
			openCmd := exec.Command("agent-browser", "--session", sessionName, "--headed", "open", "https://new.expensify.com/")
			openCmd.Stdout = os.Stderr
			openCmd.Stderr = os.Stderr
			if err := openCmd.Run(); err != nil {
				return fallbackPromptForToken(cmd, cfg, flags, fmt.Sprintf("failed to open browser: %v", err))
			}

			fmt.Fprintln(w, "Log in to Expensify in the opened browser window.")
			fmt.Fprintln(w, "This CLI will poll for the authToken cookie every 5 seconds (2 min timeout).")

			deadline := time.Now().Add(2 * time.Minute)
			for time.Now().Before(deadline) {
				token, email, err := captureTokenViaAgentBrowser(sessionName)
				if err == nil && token != "" {
					if err := cfg.SaveSessionToken(token, email); err != nil {
						return configErr(err)
					}
					if email != "" {
						fmt.Fprintf(w, "Captured session token (%d chars) for %s. Saved to %s\n", len(token), email, cfg.Path)
					} else {
						fmt.Fprintf(w, "Captured session token (%d chars). Saved to %s\n", len(token), cfg.Path)
					}
					return nil
				}
				time.Sleep(5 * time.Second)
			}
			return fallbackPromptForToken(cmd, cfg, flags, "login poll timed out after 2 minutes")
		},
	}
	cmd.Flags().StringVar(&sessionName, "session", "expensify-pp-login", "Named agent-browser session to use")
	cmd.Flags().StringVar(&fallbackToken, "token", "", "Provide the authToken directly instead of opening a browser")
	cmd.Flags().BoolVar(&headless, "headless", false, "Mint a session token via /Authenticate using stored email+password (no browser)")
	return cmd
}

// authenticator is the signature of Client.Authenticate. Injecting it into
// doHeadlessLogin makes the CLI-layer integration thin and testable without
// spinning up an httptest server for every case.
type authenticator func(email, password string) (*client.AuthenticateResult, error)

// doHeadlessLogin implements `auth login --headless`:
//
//  1. Require cfg.ExpensifyEmail (populated by `auth store-credentials`).
//  2. Fetch the password from the OS keychain.
//  3. Call the injected authenticator.
//  4. Translate typed errors into actionable exit codes.
//  5. Persist token + email + accountID + LastLoginAt on success.
//
// The authenticator is passed in so unit tests can swap in a spy; in production
// it's wired to the live Client.Authenticate.
func doHeadlessLogin(cmd *cobra.Command, cfg *config.Config, flags *rootFlags, auth authenticator) error {
	w := cmd.OutOrStdout()

	email := cfg.ExpensifyEmail
	if email == "" {
		return usageErr(fmt.Errorf("no email configured. Run `expensify-pp-cli auth store-credentials` first"))
	}

	password, err := credentials.Get(email)
	if err != nil {
		if errors.Is(err, credentials.ErrNotFound) {
			return usageErr(fmt.Errorf("no credentials stored for %s. Run `auth store-credentials` to save them", email))
		}
		return configErr(fmt.Errorf("reading keychain: %w", err))
	}

	result, err := auth(email, password)
	if err != nil {
		switch {
		case errors.Is(err, client.ErrTwoFactorRequired):
			fmt.Fprintln(w, "This account requires 2FA. Use `expensify-pp-cli auth login` (without --headless) to complete the 2FA prompt in a browser.")
			return usageErr(err)
		case errors.Is(err, client.ErrInvalidCredentials):
			return authErr(fmt.Errorf("email or password rejected by Expensify. Re-run `auth store-credentials` with fresh credentials, or log in via browser to reset"))
		default:
			return apiErr(err)
		}
	}

	// SaveSession writes token + email + accountID + LastLoginAt in one shot.
	if err := cfg.SaveSession(result.AuthToken, pickEmail(result.Email, email), result.AccountID); err != nil {
		return configErr(err)
	}

	if flags.asJSON {
		payload := map[string]any{
			"email":      pickEmail(result.Email, email),
			"account_id": result.AccountID,
			"source":     "headless",
		}
		if !result.ExpiresAt.IsZero() {
			payload["expires_at"] = result.ExpiresAt.Format(time.RFC3339)
		}
		return json.NewEncoder(w).Encode(payload)
	}

	fmt.Fprintln(w, "Session token minted via headless login. Valid for ~2-3h. Run `expensify-pp-cli doctor` to verify.")
	return nil
}

// pickEmail prefers the email echoed back by Expensify (canonical casing) but
// falls back to the email we sent when the response omits it.
func pickEmail(responseEmail, requestEmail string) string {
	if responseEmail != "" {
		return responseEmail
	}
	return requestEmail
}

// captureTokenViaAgentBrowser runs `agent-browser cookies get --json` and
// extracts the authToken cookie value (and associated account email if any).
func captureTokenViaAgentBrowser(session string) (string, string, error) {
	out, err := exec.Command("agent-browser", "--session", session, "cookies", "get", "--json").Output()
	if err != nil {
		return "", "", err
	}
	// Output is expected to be a JSON array. Be defensive: accept {cookies: [...]}
	// as well.
	var asArray []map[string]any
	if err := json.Unmarshal(out, &asArray); err != nil {
		var wrapper struct {
			Cookies []map[string]any `json:"cookies"`
		}
		if err := json.Unmarshal(out, &wrapper); err != nil {
			return "", "", fmt.Errorf("parsing cookies JSON: %w", err)
		}
		asArray = wrapper.Cookies
	}
	var token, email string
	for _, c := range asArray {
		name, _ := c["name"].(string)
		domain, _ := c["domain"].(string)
		if !strings.Contains(strings.ToLower(domain), "expensify.com") {
			continue
		}
		val, _ := c["value"].(string)
		switch name {
		case "authToken":
			token = val
		case "email", "userEmail", "expensify_email":
			email = val
		}
	}
	if token == "" {
		return "", "", fmt.Errorf("authToken cookie not present yet")
	}
	return token, email, nil
}

func fallbackPromptForToken(cmd *cobra.Command, cfg *config.Config, flags *rootFlags, reason string) error {
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Could not auto-capture session: %s\n", reason)
	if flags.noInput {
		return authErr(fmt.Errorf("cannot prompt for token with --no-input; rerun without --no-input or pass --token"))
	}
	fmt.Fprintln(w, "Paste your authToken (find it at https://new.expensify.com/ -> DevTools -> Application -> Cookies -> authToken):")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return authErr(fmt.Errorf("reading stdin: %w", err))
	}
	token := strings.TrimSpace(line)
	if token == "" {
		return authErr(fmt.Errorf("no token provided"))
	}
	if err := cfg.SaveSessionToken(token, ""); err != nil {
		return configErr(err)
	}
	fmt.Fprintf(w, "Token saved to %s\n", cfg.Path)
	return nil
}

func newAuthSetTokenCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "set-token <token>",
		Short:   "Save a session authToken to the config file",
		Example: "  expensify-pp-cli auth set-token <800-char-hex>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			if err := cfg.SaveSessionToken(strings.TrimSpace(args[0]), ""); err != nil {
				return configErr(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Token saved to %s\n", cfg.Path)
			return nil
		},
	}
}

func newAuthSetKeysCmd(flags *rootFlags) *cobra.Command {
	var fromEnv bool
	var partnerID, partnerSecret string
	cmd := &cobra.Command{
		Use:     "set-keys",
		Short:   "Save partnerUserID and partnerUserSecret for Integration Server calls",
		Example: "  expensify-pp-cli auth set-keys --env\n  expensify-pp-cli auth set-keys --id <id> --secret <secret>",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			if fromEnv {
				partnerID = os.Getenv("EXPENSIFY_PARTNER_USER_ID")
				partnerSecret = os.Getenv("EXPENSIFY_PARTNER_USER_SECRET")
			}
			if partnerID == "" || partnerSecret == "" {
				if flags.noInput {
					return authErr(fmt.Errorf("partner id/secret missing and --no-input set; pass --id/--secret or set env vars"))
				}
				reader := bufio.NewReader(os.Stdin)
				if partnerID == "" {
					fmt.Fprint(cmd.OutOrStdout(), "partnerUserID: ")
					s, _ := reader.ReadString('\n')
					partnerID = strings.TrimSpace(s)
				}
				if partnerSecret == "" {
					fmt.Fprint(cmd.OutOrStdout(), "partnerUserSecret: ")
					s, _ := reader.ReadString('\n')
					partnerSecret = strings.TrimSpace(s)
				}
			}
			if partnerID == "" || partnerSecret == "" {
				return authErr(fmt.Errorf("partnerUserID and partnerUserSecret are both required"))
			}
			cfg.ExpensifyPartnerUserId = partnerID
			cfg.ExpensifyPartnerUserSecret = partnerSecret
			if err := cfg.SaveTokens(cfg.ClientID, cfg.ClientSecret, cfg.AccessToken, cfg.RefreshToken, cfg.TokenExpiry); err != nil {
				return configErr(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Partner credentials saved to %s\n", cfg.Path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&fromEnv, "env", false, "Read EXPENSIFY_PARTNER_USER_ID and EXPENSIFY_PARTNER_USER_SECRET from the environment")
	cmd.Flags().StringVar(&partnerID, "id", "", "partnerUserID")
	cmd.Flags().StringVar(&partnerSecret, "secret", "", "partnerUserSecret")
	return cmd
}

func newAuthStoreCredentialsCmd(flags *rootFlags) *cobra.Command {
	var email, password string
	var fromEnv bool
	cmd := &cobra.Command{
		Use:   "store-credentials",
		Short: "Store Expensify email+password in the OS keychain for headless token minting",
		Long: `Writes the password to the OS-native keychain (macOS Keychain, Windows
Credential Manager, Linux Secret Service) under service "expensify-pp-cli" and
mirrors the email to the TOML config. The password never touches disk.

Later, 'auth login --headless' uses these to mint a session authToken via the
Expensify /Authenticate endpoint without a browser.`,
		Example: `  expensify-pp-cli auth store-credentials --email you@example.com
  expensify-pp-cli auth store-credentials --from-env
  expensify-pp-cli auth store-credentials --email you@example.com --password 'secret'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			w := cmd.OutOrStdout()

			if fromEnv {
				if email == "" {
					email = os.Getenv("EXPENSIFY_EMAIL")
				}
				if password == "" {
					password = os.Getenv("EXPENSIFY_PASSWORD")
				}
			}

			// Interactive prompts: email first, then password (silent).
			if email == "" {
				if flags.noInput {
					return usageErr(fmt.Errorf("--email required with --no-input (or pass --from-env)"))
				}
				fmt.Fprint(w, "Expensify email: ")
				reader := bufio.NewReader(os.Stdin)
				s, err := reader.ReadString('\n')
				if err != nil {
					return usageErr(fmt.Errorf("reading email: %w", err))
				}
				email = strings.TrimSpace(s)
			}
			if email == "" {
				return usageErr(fmt.Errorf("email is required"))
			}
			if password == "" {
				if flags.noInput {
					return usageErr(fmt.Errorf("--password or --from-env required with --no-input"))
				}
				fmt.Fprint(w, "Expensify password (input hidden): ")
				pw, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Fprintln(w) // newline after hidden prompt
				if err != nil {
					return usageErr(fmt.Errorf("reading password: %w", err))
				}
				password = strings.TrimSpace(string(pw))
			}
			if password == "" {
				return usageErr(fmt.Errorf("password is required"))
			}

			if err := credentials.Set(email, password); err != nil {
				if errors.Is(err, credentials.ErrEmptyEmail) || errors.Is(err, credentials.ErrEmptyPassword) {
					return usageErr(err)
				}
				return configErr(fmt.Errorf("keychain access failed — requires a graphical session or unlock (error: %w)", err))
			}
			if err := cfg.SaveEmail(email); err != nil {
				return configErr(err)
			}

			fmt.Fprintf(w, "Credentials stored for %s.\n", email)
			fmt.Fprintln(w, "Run `expensify-pp-cli auth login --headless` to mint a session token.")
			return nil
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "Expensify account email")
	cmd.Flags().StringVar(&password, "password", "", "Expensify account password (prompted if omitted and interactive)")
	cmd.Flags().BoolVar(&fromEnv, "from-env", false, "Read email/password from EXPENSIFY_EMAIL and EXPENSIFY_PASSWORD")
	return cmd
}

func newAuthLogoutCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "logout",
		Short:   "Clear stored credentials",
		Example: "  expensify-pp-cli auth logout",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			cfg.ExpensifyAuthToken = ""
			cfg.ExpensifyPartnerUserId = ""
			cfg.ExpensifyPartnerUserSecret = ""
			if err := cfg.ClearTokens(); err != nil {
				return configErr(err)
			}
			if os.Getenv("EXPENSIFY_AUTH_TOKEN") != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Config cleared. Note: EXPENSIFY_AUTH_TOKEN env var is still set.")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out. Credentials cleared.")
			return nil
		},
	}
}
