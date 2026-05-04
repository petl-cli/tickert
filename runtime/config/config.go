// Package config resolves credentials and runtime settings for generated CLIs.
//
// Precedence (highest to lowest):
//  1. CLI flags (passed in directly at call time)
//  2. Environment variables (<PREFIX>_API_KEY, <PREFIX>_BEARER_TOKEN, etc.)
//  3. Config file (~/.config/<cliName>/config.yaml)
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rishimantri795/CLICreator/runtime/auth"
	"github.com/rishimantri795/CLICreator/runtime/httpclient"
)

// Config holds all resolved runtime settings for a generated CLI.
type Config struct {
	BaseURL      string
	OutputFormat string
	// Credentials — at most one will be non-empty
	BearerToken string
	APIKey      string
	APIKeyName  string // The header/param name for the API key (e.g. "X-Api-Key")
	APIKeyIn    string // "header" or "query"

	// OAuth2 fields — populated by generated code when the spec uses oauth2
	OAuthTokenURL string // Token endpoint for refresh
	OAuthClientID string // Client ID for refresh
	CLIName       string // CLI name for locating token store directory
}

// Loader resolves configuration from flags → env vars → config file.
type Loader struct {
	CLIName      string // e.g. "myapi" — used for config file path
	EnvVarPrefix string // e.g. "MYAPI" — uppercased
	DefaultURL   string // From the generated CLI's baked-in base URL
}

// Load resolves all settings. flagOverrides contains values from CLI flags
// (empty string means "not provided via flag, fall through to env/file").
func (l *Loader) Load(flagOverrides Config) (*Config, error) {
	prefix := strings.ToUpper(l.EnvVarPrefix)

	// Start with defaults
	resolved := Config{
		BaseURL:      l.DefaultURL,
		OutputFormat: "json",
	}

	// Layer 3: config file
	if fileConf, err := l.loadFile(); err == nil {
		if fileConf.BaseURL != "" {
			resolved.BaseURL = fileConf.BaseURL
		}
		if fileConf.OutputFormat != "" {
			resolved.OutputFormat = fileConf.OutputFormat
		}
		if fileConf.BearerToken != "" {
			resolved.BearerToken = fileConf.BearerToken
		}
		if fileConf.APIKey != "" {
			resolved.APIKey = fileConf.APIKey
		}
	}

	// Layer 2: environment variables
	if v := os.Getenv(prefix + "_BASE_URL"); v != "" {
		resolved.BaseURL = v
	}
	if v := os.Getenv(prefix + "_OUTPUT_FORMAT"); v != "" {
		resolved.OutputFormat = v
	}
	if v := os.Getenv(prefix + "_BEARER_TOKEN"); v != "" {
		resolved.BearerToken = v
	}
	if v := os.Getenv(prefix + "_API_KEY"); v != "" {
		resolved.APIKey = v
	}

	// Layer 2 (continued): OAuth client ID from env
	if v := os.Getenv(prefix + "_OAUTH_CLIENT_ID"); v != "" {
		resolved.OAuthClientID = v
	}

	// Layer 1: CLI flags (highest priority)
	if flagOverrides.BaseURL != "" {
		resolved.BaseURL = flagOverrides.BaseURL
	}
	if flagOverrides.OutputFormat != "" {
		resolved.OutputFormat = flagOverrides.OutputFormat
	}
	if flagOverrides.BearerToken != "" {
		resolved.BearerToken = flagOverrides.BearerToken
	}
	if flagOverrides.APIKey != "" {
		resolved.APIKey = flagOverrides.APIKey
	}
	if flagOverrides.OAuthClientID != "" {
		resolved.OAuthClientID = flagOverrides.OAuthClientID
	}

	// Copy through static metadata (baked from spec, not user-configurable)
	resolved.APIKeyName = flagOverrides.APIKeyName
	resolved.APIKeyIn = flagOverrides.APIKeyIn
	resolved.OAuthTokenURL = flagOverrides.OAuthTokenURL
	resolved.CLIName = l.CLIName

	return &resolved, nil
}

// AuthProvider builds the appropriate httpclient.AuthProvider from resolved config.
// Priority: explicit flags (bearer/apiKey) > OAuth token store > nil.
// Returns nil if no credentials are configured (for public APIs).
func (c *Config) AuthProvider() httpclient.AuthProvider {
	// Explicit bearer token flag/env takes highest priority
	if c.BearerToken != "" {
		return httpclient.BearerAuth{Token: c.BearerToken}
	}
	// Explicit API key flag/env
	if c.APIKey != "" {
		name := c.APIKeyName
		if name == "" {
			name = "X-Api-Key"
		}
		loc := c.APIKeyIn
		if loc == "" {
			loc = "header"
		}
		return httpclient.APIKeyAuth{Key: c.APIKey, Name: name, Location: loc}
	}
	// OAuth token store — check if a token exists on disk
	if c.OAuthTokenURL != "" {
		store := c.oauthTokenStore()
		if store != nil && store.Load() != nil {
			return &auth.OAuth2Auth{
				TokenStore: store,
				TokenURL:   c.OAuthTokenURL,
				ClientID:   c.OAuthClientID,
			}
		}
	}
	return nil
}

// OAuthTokenStore returns the token store for this CLI, or nil if not configured.
func (c *Config) OAuthTokenStore() *auth.TokenStore {
	return c.oauthTokenStore()
}

func (c *Config) oauthTokenStore() *auth.TokenStore {
	if c.CLIName == "" {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return &auth.TokenStore{
		Dir: filepath.Join(home, ".config", c.CLIName),
	}
}

// configFile is the minimal structure parsed from the YAML config file.
// We use a simple key=value approach to avoid requiring a YAML dependency in the runtime.
type configFile struct {
	BaseURL      string
	OutputFormat string
	BearerToken  string
	APIKey       string
}

// loadFile reads ~/.config/<cliName>/config.yaml and parses it.
// Returns an empty config (no error) if the file doesn't exist.
func (l *Loader) loadFile() (*configFile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return &configFile{}, nil
	}
	path := filepath.Join(home, ".config", l.CLIName, "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		// File not found is normal
		return &configFile{}, nil
	}

	cfg := &configFile{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		switch k {
		case "base_url":
			cfg.BaseURL = v
		case "output_format":
			cfg.OutputFormat = v
		case "bearer_token":
			cfg.BearerToken = v
		case "api_key":
			cfg.APIKey = v
		}
	}
	return cfg, nil
}

// ConfigFilePath returns the path to the config file for this CLI.
// Useful for the `configure` command to tell users where settings are stored.
func (l *Loader) ConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", l.CLIName, "config.yaml"), nil
}
