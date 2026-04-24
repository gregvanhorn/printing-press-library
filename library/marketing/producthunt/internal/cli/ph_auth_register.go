package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/config"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/phgraphql"
)

// Filed under the ph_ prefix so the CLI Printing Press generator does not
// reclaim this file on regeneration. See AGENTS.md for the convention.

// newAuthRegisterCmd is the top-level "auth register" subcommand that walks
// the user through registering a PH OAuth app and persisting an app-level
// access token via client_credentials exchange.
func newAuthRegisterCmd(flags *rootFlags) *cobra.Command {
	var clientIDFlag, clientSecretFlag, clientIDEnvFlag, clientSecretEnvFlag string
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a Product Hunt OAuth app and store the access token",
		Long: `Register a Product Hunt OAuth application and exchange its credentials for an
app-level access token via the client_credentials grant.

	Register the app once at https://www.producthunt.com/v2/oauth/applications:

	  Name: producthunt-pp-cli (or any recognizable local name)
	  Redirect URI: https://localhost/callback

	Then paste the API Key as client_id and API Secret as client_secret below.
	The CLI exchanges those for an access token and saves it (along with the
	client id and secret for later refreshes) to the TOML config file with 0600
	permissions.

Atom-runtime commands (sync, recent, today, list, search, etc.) do NOT use
this token — they hit the public /feed endpoint and need no auth. This
command unlocks Tier 2 (search --enrich) and Tier 3 (backfill) features.

	Interactive mode reads the client ID and secret from stdin. Agent/CI mode can
	read PRODUCTHUNT_CLIENT_ID and PRODUCTHUNT_CLIENT_SECRET, or explicit env var
	names via --client-id-env and --client-secret-env. Passing secrets as flags is
	supported but less safe because shell history may retain them.`,
		Example: `  # Interactive (recommended): paste values when prompted
  producthunt-pp-cli auth register

  # Agent-friendly: read credentials from environment
  PRODUCTHUNT_CLIENT_ID=... PRODUCTHUNT_CLIENT_SECRET=... producthunt-pp-cli auth register --no-input

  # Explicit env var names
  producthunt-pp-cli auth register --no-input --client-id-env PH_ID --client-secret-env PH_SECRET`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthRegister(cmd, flags, authRegisterInputs{
				ClientID:        clientIDFlag,
				ClientSecret:    clientSecretFlag,
				ClientIDEnv:     clientIDEnvFlag,
				ClientSecretEnv: clientSecretEnvFlag,
			})
		},
	}
	cmd.Flags().StringVar(&clientIDFlag, "client-id", "", "PH OAuth app client ID (non-interactive mode)")
	cmd.Flags().StringVar(&clientSecretFlag, "client-secret", "", "PH OAuth app client secret (non-interactive mode)")
	cmd.Flags().StringVar(&clientIDEnvFlag, "client-id-env", "PRODUCTHUNT_CLIENT_ID", "Environment variable containing PH OAuth app client ID")
	cmd.Flags().StringVar(&clientSecretEnvFlag, "client-secret-env", "PRODUCTHUNT_CLIENT_SECRET", "Environment variable containing PH OAuth app client secret")
	return cmd
}

type authRegisterInputs struct {
	ClientID        string
	ClientSecret    string
	ClientIDEnv     string
	ClientSecretEnv string
}

var exchangeClientCredentialsFunc = exchangeClientCredentials

func runAuthRegister(cmd *cobra.Command, flags *rootFlags, inputs authRegisterInputs) error {
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return configErr(err)
	}

	w := cmd.OutOrStdout()
	humanOutput := !flags.asJSON && !flags.agent
	interactiveHuman := humanOutput && !flags.noInput
	if interactiveHuman {
		fmt.Fprintln(w, bold("Register a Product Hunt OAuth application"))
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "1. Open this URL in your browser:")
		fmt.Fprintln(w, "     https://www.producthunt.com/v2/oauth/applications")
		fmt.Fprintln(w, "2. Click 'New application'.")
		fmt.Fprintln(w, "3. Fill in:")
		fmt.Fprintln(w, "     Name: producthunt-pp-cli")
		fmt.Fprintln(w, "     Redirect URI: https://localhost/callback")
		fmt.Fprintln(w, "4. Paste the API Key (client_id) and API Secret (client_secret) below.")
		fmt.Fprintln(w, "")
	}

	clientID := strings.TrimSpace(inputs.ClientID)
	clientSecret := strings.TrimSpace(inputs.ClientSecret)
	if clientID == "" && inputs.ClientIDEnv != "" {
		clientID = strings.TrimSpace(os.Getenv(inputs.ClientIDEnv))
	}
	if clientSecret == "" && inputs.ClientSecretEnv != "" {
		clientSecret = strings.TrimSpace(os.Getenv(inputs.ClientSecretEnv))
	}

	if clientID == "" {
		if flags.noInput {
			return usageErr(fmt.Errorf("--no-input set but no client ID supplied; set %s or pass --client-id", inputs.ClientIDEnv))
		}
		v, err := readLineFromStdin(cmd, "Client ID: ")
		if err != nil {
			return err
		}
		clientID = strings.TrimSpace(v)
	}
	if clientID == "" {
		return usageErr(fmt.Errorf("client_id is required"))
	}

	if clientSecret == "" {
		if flags.noInput {
			return usageErr(fmt.Errorf("--no-input set but no client secret supplied; set %s or pass --client-secret", inputs.ClientSecretEnv))
		}
		v, err := readSecretFromStdin(cmd, "Client Secret (no echo): ")
		if err != nil {
			return err
		}
		clientSecret = strings.TrimSpace(v)
	}
	if clientSecret == "" {
		return usageErr(fmt.Errorf("client_secret is required"))
	}

	if humanOutput {
		if interactiveHuman {
			fmt.Fprintln(w, "")
		}
		fmt.Fprintln(w, "Exchanging credentials for access token...")
	}

	parentCtx := cmd.Context()
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
	defer cancel()
	token, expiry, err := exchangeClientCredentialsFunc(ctx, clientID, clientSecret, "")
	if err != nil {
		return authErr(fmt.Errorf("token exchange: %w", err))
	}

	if err := cfg.SaveOAuth(clientID, clientSecret, token, expiry); err != nil {
		return configErr(fmt.Errorf("saving config: %w", err))
	}

	if flags.asJSON || flags.agent {
		payload := map[string]any{
			"status":         "authenticated",
			"config_path":    cfg.Path,
			"auth_mode":      cfg.GraphQLAuthMode(),
			"client_id":      maskMiddle(clientID, 4, 4),
			"token_redacted": true,
			"unlocked":       []string{"backfill", "search --enrich"},
			"next":           "producthunt-pp-cli backfill --days 30 --dry-run",
		}
		if !expiry.IsZero() {
			payload["token_expires"] = expiry.Format(time.RFC3339)
		}
		raw, _ := json.Marshal(payload)
		return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
	}

	fmt.Fprintln(w, green("Authenticated"))
	fmt.Fprintf(w, "  Config: %s\n", cfg.Path)
	// Show only the prefix of the client id, never the secret or the token.
	fmt.Fprintf(w, "  Client ID: %s\n", maskMiddle(clientID, 4, 4))
	if !expiry.IsZero() {
		fmt.Fprintf(w, "  Token expires: %s\n", expiry.Format(time.RFC3339))
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Next: try `producthunt-pp-cli backfill --days 30 --dry-run` to preview what the")
	fmt.Fprintln(w, "30-day backfill will cost in API budget before running it for real.")
	return nil
}

// exchangeClientCredentials runs the OAuth 2.0 client_credentials grant
// against PH and returns (access_token, expires_at, error).
//
// endpoint is the token URL; empty means "use phgraphql.DefaultTokenEndpoint".
// This parameter exists so integration tests can point the flow at a mock.
func exchangeClientCredentials(ctx context.Context, clientID, clientSecret, endpoint string) (string, time.Time, error) {
	if endpoint == "" {
		endpoint = phgraphql.DefaultTokenEndpoint
	}
	// PH accepts either form-encoded or JSON bodies on this endpoint;
	// JSON is simpler for us and easier to mock in tests.
	payload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"grant_type":    "client_credentials",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Surface PH's error body verbatim — their error messages are the
		// right guidance for "your secret is wrong" etc.
		return "", time.Time{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", time.Time{}, fmt.Errorf("decode token response: %w (body: %s)", err, strings.TrimSpace(string(respBody)))
	}
	if parsed.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("token endpoint returned no access_token (body: %s)", strings.TrimSpace(string(respBody)))
	}
	var expiry time.Time
	if parsed.ExpiresIn > 0 {
		expiry = time.Now().Add(time.Duration(parsed.ExpiresIn) * time.Second)
	}
	return parsed.AccessToken, expiry, nil
}

// readLineFromStdin prints prompt and reads a line from the command's
// configured stdin.
func readLineFromStdin(cmd *cobra.Command, prompt string) (string, error) {
	fmt.Fprint(cmd.OutOrStdout(), prompt)
	var buf []byte
	b := make([]byte, 1)
	in := cmd.InOrStdin()
	for {
		n, err := in.Read(b)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if n == 0 {
			continue
		}
		if b[0] == '\n' {
			break
		}
		if b[0] == '\r' {
			continue
		}
		buf = append(buf, b[0])
	}
	return string(buf), nil
}

// readSecretFromStdin prints prompt and reads a line without echoing. Falls
// back to plain readLineFromStdin when stdin is not a terminal (e.g. piped
// input for testing).
func readSecretFromStdin(cmd *cobra.Command, prompt string) (string, error) {
	fmt.Fprint(cmd.OutOrStdout(), prompt)
	if f, ok := cmd.InOrStdin().(interface{ Fd() uintptr }); ok {
		fd := int(f.Fd())
		if term.IsTerminal(fd) {
			b, err := term.ReadPassword(fd)
			fmt.Fprintln(cmd.OutOrStdout(), "")
			return string(b), err
		}
	}
	// Non-terminal stdin: just read a line.
	return readLineFromStdin(cmd, "")
}

// maskMiddle returns a value with the middle replaced by '*' so secrets
// can be printed for user confirmation without leaking them to logs.
func maskMiddle(s string, head, tail int) string {
	if len(s) <= head+tail {
		return strings.Repeat("*", len(s))
	}
	return s[:head] + strings.Repeat("*", len(s)-head-tail) + s[len(s)-tail:]
}

// Ensure imports are used when test-only refactors drop references.
var _ = url.Parse
