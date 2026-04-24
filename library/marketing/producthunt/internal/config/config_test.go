package config

import (
	"path/filepath"
	"testing"
)

func TestSaveGraphQLTokenUnlocksGraphQLFeatures(t *testing.T) {
	cfg := &Config{Path: filepath.Join(t.TempDir(), "config.toml")}
	if err := cfg.SaveGraphQLToken("dev_tok"); err != nil {
		t.Fatalf("SaveGraphQLToken: %v", err)
	}
	reloaded, err := Load(cfg.Path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reloaded.HasGraphQLToken() {
		t.Fatalf("developer token should unlock GraphQL features")
	}
	if reloaded.GraphQLAuthMode() != "developer_token" {
		t.Fatalf("GraphQLAuthMode = %q, want developer_token", reloaded.GraphQLAuthMode())
	}
}

func TestLoadDeveloperTokenFromEnv(t *testing.T) {
	t.Setenv("PRODUCTHUNT_DEVELOPER_TOKEN", "env_tok")
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AccessToken != "env_tok" {
		t.Fatalf("AccessToken = %q, want env_tok", cfg.AccessToken)
	}
	if !cfg.HasGraphQLToken() {
		t.Fatalf("env token should unlock GraphQL features")
	}
	if cfg.AuthSource != "env" {
		t.Fatalf("AuthSource = %q, want env", cfg.AuthSource)
	}
}

func TestHasGraphQLTokenRejectsUnknownAuthType(t *testing.T) {
	cfg := &Config{AuthType: "mystery", AccessToken: "tok"}
	if cfg.HasGraphQLToken() {
		t.Fatalf("unknown auth_type must not unlock GraphQL")
	}
	if got := cfg.GraphQLAuthMode(); got != "unsupported_auth_type" {
		t.Fatalf("GraphQLAuthMode = %q, want unsupported_auth_type", got)
	}
}

func TestLoadGraphQLTokenEnvPrecedence(t *testing.T) {
	t.Setenv("PRODUCTHUNT_GRAPHQL_TOKEN", "graphql_tok")
	t.Setenv("PRODUCTHUNT_DEVELOPER_TOKEN", "developer_tok")
	t.Setenv("PRODUCTHUNT_API_TOKEN", "api_tok")

	cfg, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AccessToken != "graphql_tok" {
		t.Fatalf("AccessToken = %q, want PRODUCTHUNT_GRAPHQL_TOKEN precedence", cfg.AccessToken)
	}
	if cfg.AuthType != "developer_token" || cfg.AuthSource != "env" {
		t.Fatalf("auth source/type = %q/%q", cfg.AuthSource, cfg.AuthType)
	}
}

func TestLoadGraphQLTokenEnvAliases(t *testing.T) {
	cases := []struct {
		name  string
		token string
	}{
		{"PRODUCTHUNT_GRAPHQL_TOKEN", "graphql_tok"},
		{"PRODUCTHUNT_DEVELOPER_TOKEN", "developer_tok"},
		{"PRODUCTHUNT_API_TOKEN", "api_tok"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, name := range GraphQLTokenEnvVars() {
				t.Setenv(name, "")
			}
			t.Setenv(tc.name, tc.token)
			cfg, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.AccessToken != tc.token {
				t.Fatalf("AccessToken = %q, want %q", cfg.AccessToken, tc.token)
			}
			if !cfg.HasGraphQLToken() {
				t.Fatalf("%s should unlock GraphQL", tc.name)
			}
		})
	}
}
