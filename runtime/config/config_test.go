package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rishimantri795/CLICreator/runtime/auth"
	"github.com/rishimantri795/CLICreator/runtime/config"
	"github.com/rishimantri795/CLICreator/runtime/httpclient"
)

func newLoader() *config.Loader {
	return &config.Loader{
		CLIName:      "testcli",
		EnvVarPrefix: "TESTCLI",
		DefaultURL:   "https://default.api.example.com",
	}
}

func clearEnv() {
	for _, k := range []string{
		"TESTCLI_BASE_URL", "TESTCLI_OUTPUT_FORMAT",
		"TESTCLI_BEARER_TOKEN", "TESTCLI_API_KEY",
		"TESTCLI_OAUTH_CLIENT_ID",
	} {
		os.Unsetenv(k)
	}
}

// --- Defaults ---

func TestLoad_Defaults(t *testing.T) {
	clearEnv()
	l := newLoader()
	cfg, err := l.Load(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://default.api.example.com" {
		t.Errorf("expected default URL, got %q", cfg.BaseURL)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected default format 'json', got %q", cfg.OutputFormat)
	}
}

// --- Environment variable layer ---

func TestLoad_EnvVar_BearerToken(t *testing.T) {
	clearEnv()
	os.Setenv("TESTCLI_BEARER_TOKEN", "envtoken")
	defer os.Unsetenv("TESTCLI_BEARER_TOKEN")

	cfg, err := newLoader().Load(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BearerToken != "envtoken" {
		t.Errorf("expected BearerToken=envtoken, got %q", cfg.BearerToken)
	}
}

func TestLoad_EnvVar_APIKey(t *testing.T) {
	clearEnv()
	os.Setenv("TESTCLI_API_KEY", "mykey")
	defer os.Unsetenv("TESTCLI_API_KEY")

	cfg, err := newLoader().Load(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "mykey" {
		t.Errorf("expected APIKey=mykey, got %q", cfg.APIKey)
	}
}

func TestLoad_EnvVar_BaseURL(t *testing.T) {
	clearEnv()
	os.Setenv("TESTCLI_BASE_URL", "https://staging.example.com")
	defer os.Unsetenv("TESTCLI_BASE_URL")

	cfg, err := newLoader().Load(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://staging.example.com" {
		t.Errorf("expected staging URL, got %q", cfg.BaseURL)
	}
}

// --- Flag layer overrides env layer ---

func TestLoad_FlagOverridesEnv(t *testing.T) {
	clearEnv()
	os.Setenv("TESTCLI_BEARER_TOKEN", "envtoken")
	defer os.Unsetenv("TESTCLI_BEARER_TOKEN")

	cfg, err := newLoader().Load(config.Config{BearerToken: "flagtoken"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BearerToken != "flagtoken" {
		t.Errorf("flag should override env: expected flagtoken, got %q", cfg.BearerToken)
	}
}

func TestLoad_FlagOverridesDefault_BaseURL(t *testing.T) {
	clearEnv()
	cfg, err := newLoader().Load(config.Config{BaseURL: "https://custom.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://custom.example.com" {
		t.Errorf("expected custom URL, got %q", cfg.BaseURL)
	}
}

// --- Config file layer ---

func TestLoad_ConfigFile(t *testing.T) {
	clearEnv()

	// Write a temp config file
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".config", "testcli")
	os.MkdirAll(cfgDir, 0755)
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	os.WriteFile(cfgPath, []byte("bearer_token: filetoken\nbase_url: https://file.example.com\n"), 0600)

	// Override home dir so the loader reads our temp file
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	cfg, err := newLoader().Load(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BearerToken != "filetoken" {
		t.Errorf("expected filetoken from config file, got %q", cfg.BearerToken)
	}
	if cfg.BaseURL != "https://file.example.com" {
		t.Errorf("expected file URL, got %q", cfg.BaseURL)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	clearEnv()

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".config", "testcli")
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("bearer_token: filetoken\n"), 0600)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	os.Setenv("TESTCLI_BEARER_TOKEN", "envtoken")
	defer os.Unsetenv("TESTCLI_BEARER_TOKEN")

	cfg, err := newLoader().Load(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BearerToken != "envtoken" {
		t.Errorf("env should override file: expected envtoken, got %q", cfg.BearerToken)
	}
}

// --- AuthProvider ---

func TestAuthProvider_Bearer(t *testing.T) {
	cfg := &config.Config{BearerToken: "tok"}
	auth := cfg.AuthProvider()
	if auth == nil {
		t.Fatal("expected non-nil AuthProvider for bearer token")
	}
	_, ok := auth.(httpclient.BearerAuth)
	if !ok {
		t.Errorf("expected BearerAuth, got %T", auth)
	}
}

func TestAuthProvider_APIKey(t *testing.T) {
	cfg := &config.Config{APIKey: "k", APIKeyName: "X-Api-Key", APIKeyIn: "header"}
	auth := cfg.AuthProvider()
	if auth == nil {
		t.Fatal("expected non-nil AuthProvider for API key")
	}
	_, ok := auth.(httpclient.APIKeyAuth)
	if !ok {
		t.Errorf("expected APIKeyAuth, got %T", auth)
	}
}

func TestAuthProvider_NoCredentials(t *testing.T) {
	cfg := &config.Config{}
	if cfg.AuthProvider() != nil {
		t.Error("expected nil AuthProvider when no credentials are set")
	}
}

// Bearer takes precedence over API key when both are set
func TestAuthProvider_BearerTakesPrecedence(t *testing.T) {
	cfg := &config.Config{BearerToken: "tok", APIKey: "k"}
	auth := cfg.AuthProvider()
	_, ok := auth.(httpclient.BearerAuth)
	if !ok {
		t.Errorf("expected BearerAuth to take precedence, got %T", auth)
	}
}

// --- OAuth2 ---

func TestLoad_EnvVar_OAuthClientID(t *testing.T) {
	clearEnv()
	os.Setenv("TESTCLI_OAUTH_CLIENT_ID", "my-client-id")
	defer os.Unsetenv("TESTCLI_OAUTH_CLIENT_ID")

	cfg, err := newLoader().Load(config.Config{
		OAuthTokenURL: "https://auth.example.com/token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OAuthClientID != "my-client-id" {
		t.Errorf("expected OAuthClientID=my-client-id, got %q", cfg.OAuthClientID)
	}
}

func TestLoad_FlagOverridesEnv_OAuthClientID(t *testing.T) {
	clearEnv()
	os.Setenv("TESTCLI_OAUTH_CLIENT_ID", "env-id")
	defer os.Unsetenv("TESTCLI_OAUTH_CLIENT_ID")

	cfg, err := newLoader().Load(config.Config{
		OAuthTokenURL: "https://auth.example.com/token",
		OAuthClientID: "flag-id",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OAuthClientID != "flag-id" {
		t.Errorf("flag should override env: expected flag-id, got %q", cfg.OAuthClientID)
	}
}

func TestLoad_OAuthTokenURL_PassedThrough(t *testing.T) {
	clearEnv()
	cfg, err := newLoader().Load(config.Config{
		OAuthTokenURL: "https://auth.example.com/token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OAuthTokenURL != "https://auth.example.com/token" {
		t.Errorf("OAuthTokenURL = %q", cfg.OAuthTokenURL)
	}
}

func TestAuthProvider_OAuth_WithToken(t *testing.T) {
	clearEnv()
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	// Write a valid token file
	tokenDir := filepath.Join(dir, ".config", "testcli")
	os.MkdirAll(tokenDir, 0700)
	os.WriteFile(filepath.Join(tokenDir, "token.json"), []byte(`{
		"access_token": "oauthtoken",
		"refresh_token": "refresh",
		"expires_at": "2099-01-01T00:00:00Z"
	}`), 0600)

	cfg := &config.Config{
		OAuthTokenURL: "https://auth.example.com/token",
		OAuthClientID: "my-client",
		CLIName:       "testcli",
	}
	provider := cfg.AuthProvider()
	if provider == nil {
		t.Fatal("expected non-nil AuthProvider for OAuth with stored token")
	}
	if _, ok := provider.(*auth.OAuth2Auth); !ok {
		t.Errorf("expected *auth.OAuth2Auth, got %T", provider)
	}
}

func TestAuthProvider_OAuth_NoToken_ReturnsNil(t *testing.T) {
	clearEnv()
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	// No token.json on disk
	cfg := &config.Config{
		OAuthTokenURL: "https://auth.example.com/token",
		OAuthClientID: "my-client",
		CLIName:       "testcli",
	}
	if cfg.AuthProvider() != nil {
		t.Error("expected nil AuthProvider when no OAuth token is stored")
	}
}

func TestAuthProvider_BearerTakesPrecedenceOverOAuth(t *testing.T) {
	clearEnv()
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	// Write a valid OAuth token
	tokenDir := filepath.Join(dir, ".config", "testcli")
	os.MkdirAll(tokenDir, 0700)
	os.WriteFile(filepath.Join(tokenDir, "token.json"), []byte(`{
		"access_token": "oauthtoken",
		"expires_at": "2099-01-01T00:00:00Z"
	}`), 0600)

	cfg := &config.Config{
		BearerToken:   "explicit-bearer",
		OAuthTokenURL: "https://auth.example.com/token",
		OAuthClientID: "my-client",
		CLIName:       "testcli",
	}
	provider := cfg.AuthProvider()
	_, ok := provider.(httpclient.BearerAuth)
	if !ok {
		t.Errorf("BearerAuth should take precedence over OAuth, got %T", provider)
	}
}
