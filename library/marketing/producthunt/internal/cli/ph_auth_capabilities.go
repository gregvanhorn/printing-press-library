package cli

import (
	"encoding/json"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/config"
)

const productHuntOAuthAppsURL = "https://www.producthunt.com/v2/oauth/applications"

type authCapabilityStatus struct {
	Mode                 string            `json:"mode"`
	ConfigPath           string            `json:"config_path,omitempty"`
	GraphQLConfigured    bool              `json:"graphql_configured"`
	GraphQLValidation    string            `json:"graphql_validation"`
	CanBackfill          bool              `json:"can_backfill"`
	CanEnrichSearch      bool              `json:"can_enrich_search"`
	AnonymousHistoryMode string            `json:"anonymous_history_mode"`
	SetupCommand         string            `json:"setup_command,omitempty"`
	SetupURL             string            `json:"setup_url,omitempty"`
	EnvVars              []string          `json:"env_vars,omitempty"`
	Unlocked             []string          `json:"unlocked,omitempty"`
	Notes                []string          `json:"notes,omitempty"`
	Instructions         map[string]string `json:"instructions,omitempty"`
}

type authImprovementHint struct {
	ImprovesWithAuth bool   `json:"improves_with_auth"`
	Command          string `json:"command"`
	Reason           string `json:"reason"`
	SetupURL         string `json:"setup_url"`
}

func buildAuthStatus(cfg *config.Config) authCapabilityStatus {
	status := authCapabilityStatus{
		Mode:                 "atom_only",
		GraphQLValidation:    "not_configured",
		AnonymousHistoryMode: "public Atom /feed snapshots accumulate history from now forward; Atom cannot retroactively backfill past windows",
		SetupCommand:         "producthunt-pp-cli auth setup",
		SetupURL:             productHuntOAuthAppsURL,
		EnvVars: []string{
			"PRODUCTHUNT_CLIENT_ID",
			"PRODUCTHUNT_CLIENT_SECRET",
			"PRODUCTHUNT_DEVELOPER_TOKEN",
			"PRODUCTHUNT_GRAPHQL_TOKEN",
			"PRODUCTHUNT_API_TOKEN",
		},
		Notes: []string{
			"Anonymous commands keep working without auth.",
			"GraphQL auth unlocks immediate historical backfill and search enrichment.",
		},
	}
	if cfg != nil {
		status.ConfigPath = cfg.Path
		if cfg.HasGraphQLToken() {
			status.Mode = cfg.GraphQLAuthMode()
			status.GraphQLConfigured = true
			status.GraphQLValidation = "configured_not_validated"
			status.CanBackfill = true
			status.CanEnrichSearch = true
			status.Unlocked = []string{"backfill", "search --enrich"}
			status.SetupCommand = ""
			status.SetupURL = ""
			status.Notes = append(status.Notes, "Configured GraphQL credentials are not live-validated by status/doctor; run backfill --dry-run to test API access.")
		} else if cfg.AccessToken != "" {
			status.Mode = cfg.GraphQLAuthMode()
			status.GraphQLValidation = "unsupported_auth_type"
			status.Notes = append(status.Notes, "A token is present, but auth_type is unsupported; use auth setup or auth set-token to configure a supported Product Hunt GraphQL token.")
		}
	}
	return status
}

func authSetupPayload(cfg *config.Config) authCapabilityStatus {
	status := buildAuthStatus(cfg)
	if status.SetupCommand == "" {
		status.SetupCommand = "producthunt-pp-cli auth setup"
	}
	if status.SetupURL == "" {
		status.SetupURL = productHuntOAuthAppsURL
	}
	status.Instructions = map[string]string{
		"create_app_url": productHuntOAuthAppsURL,
		"name":           "producthunt-pp-cli",
		"redirect_uri":   "https://localhost/callback",
		"interactive":    "producthunt-pp-cli auth register",
		"agent_oauth":    "PRODUCTHUNT_CLIENT_ID=... PRODUCTHUNT_CLIENT_SECRET=... producthunt-pp-cli auth register --no-input",
		"agent_token":    "PRODUCTHUNT_DEVELOPER_TOKEN=... producthunt-pp-cli auth set-token --token-env PRODUCTHUNT_DEVELOPER_TOKEN",
	}
	return status
}

func graphQLAuthHint(reason string) authImprovementHint {
	return authImprovementHint{
		ImprovesWithAuth: true,
		Command:          "producthunt-pp-cli auth setup",
		Reason:           reason,
		SetupURL:         productHuntOAuthAppsURL,
	}
}

func withMeta(body []byte, meta map[string]any) []byte {
	if len(meta) == 0 {
		return body
	}
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		existing, _ := obj["_meta"].(map[string]any)
		if existing == nil {
			existing = map[string]any{}
		}
		for k, v := range meta {
			existing[k] = v
		}
		obj["_meta"] = existing
		out, err := json.Marshal(obj)
		if err == nil {
			return out
		}
		return body
	}
	var arr []any
	if err := json.Unmarshal(body, &arr); err == nil {
		out, err := json.Marshal(map[string]any{
			"results": arr,
			"_meta":   meta,
		})
		if err == nil {
			return out
		}
	}
	return body
}
