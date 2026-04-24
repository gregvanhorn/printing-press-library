package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/config"
	"github.com/spf13/cobra"
)

func TestAgentContextUsesConfiguredConfigPath(t *testing.T) {
	for _, name := range []string{"PRODUCTHUNT_GRAPHQL_TOKEN", "PRODUCTHUNT_DEVELOPER_TOKEN", "PRODUCTHUNT_API_TOKEN"} {
		t.Setenv(name, "")
	}

	flags := newTestFlags(t)
	cfg, err := configForTest(flags)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if err := cfg.SaveGraphQLToken("tok_from_config"); err != nil {
		t.Fatalf("save token: %v", err)
	}

	root := &cobra.Command{Use: "producthunt-pp-cli", Version: "test"}
	ctx := buildAgentContext(root, flags)
	if !ctx.Auth.GraphQLConfigured {
		t.Fatalf("agent-context ignored --config auth state: %+v", ctx.Auth)
	}
	if ctx.Auth.GraphQLValidation != "configured_not_validated" {
		t.Fatalf("GraphQLValidation = %q", ctx.Auth.GraphQLValidation)
	}
}

func TestAuthSetTokenJSONDoesNotLeakToken(t *testing.T) {
	flags := newTestFlags(t)
	flags.asJSON = true
	t.Setenv("PH_TEST_PRODUCTHUNT_TOKEN", "secret_token_value")

	cmd := newAuthSetTokenCmd(flags)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--token-env", "PH_TEST_PRODUCTHUNT_TOKEN"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth set-token: %v", err)
	}

	if strings.Contains(out.String(), "secret_token_value") {
		t.Fatalf("JSON output leaked token: %s", out.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, out.String())
	}
	if payload["status"] != "saved" || payload["auth_mode"] != "developer_token" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestAuthLogoutReportsActiveEnvCredentials(t *testing.T) {
	flags := newTestFlags(t)
	flags.asJSON = true
	t.Setenv("PRODUCTHUNT_DEVELOPER_TOKEN", "still_active")

	cmd := newAuthLogoutCmd(flags)
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth logout: %v", err)
	}

	var payload struct {
		Status               string   `json:"status"`
		EnvCredentialsActive bool     `json:"env_credentials_active"`
		ActiveEnvVars        []string `json:"active_env_vars"`
		EffectiveMode        string   `json:"effective_mode"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, out.String())
	}
	if payload.Status != "cleared" || !payload.EnvCredentialsActive || payload.EffectiveMode != "developer_token_env" {
		t.Fatalf("unexpected logout payload: %+v", payload)
	}
	if len(payload.ActiveEnvVars) != 1 || payload.ActiveEnvVars[0] != "PRODUCTHUNT_DEVELOPER_TOKEN" {
		t.Fatalf("active env vars = %#v", payload.ActiveEnvVars)
	}
}

func TestAuthRegisterJSONSuccessSuppressesInstructionsAndSecrets(t *testing.T) {
	flags := newTestFlags(t)
	flags.asJSON = true
	flags.noInput = true

	oldExchange := exchangeClientCredentialsFunc
	exchangeClientCredentialsFunc = func(ctx context.Context, clientID, clientSecret, endpoint string) (string, time.Time, error) {
		return "access_token_secret", time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC), nil
	}
	defer func() { exchangeClientCredentialsFunc = oldExchange }()

	cmd := &cobra.Command{Use: "register"}
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := runAuthRegister(cmd, flags, authRegisterInputs{
		ClientID:        "client_id_value",
		ClientSecret:    "client_secret_value",
		ClientIDEnv:     "PRODUCTHUNT_CLIENT_ID",
		ClientSecretEnv: "PRODUCTHUNT_CLIENT_SECRET",
	})
	if err != nil {
		t.Fatalf("auth register: %v", err)
	}

	text := out.String()
	for _, leaked := range []string{"client_secret_value", "access_token_secret", "Open this URL", "New application"} {
		if strings.Contains(text, leaked) {
			t.Fatalf("register JSON output leaked/surfaced %q: %s", leaked, text)
		}
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, text)
	}
	if payload["status"] != "authenticated" || payload["auth_mode"] != "oauth_client_credentials" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func configForTest(flags *rootFlags) (*config.Config, error) {
	return config.Load(flags.configPath)
}
