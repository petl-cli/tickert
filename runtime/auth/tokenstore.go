// Package auth provides OAuth 2.0 support for generated CLIs.
package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Token represents a stored OAuth 2.0 token.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// IsExpired returns true if the access token has expired (with a 30-second buffer).
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt.Add(-30 * time.Second))
}

// TokenStore manages reading and writing OAuth tokens to disk.
type TokenStore struct {
	// Dir is the directory where the token file is stored,
	// typically ~/.config/<cliName>/
	Dir string
}

func (s *TokenStore) path() string {
	return filepath.Join(s.Dir, "token.json")
}

// Load reads the stored token from disk. Returns nil if no token exists.
func (s *TokenStore) Load() *Token {
	data, err := os.ReadFile(s.path())
	if err != nil {
		return nil
	}
	var tok Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil
	}
	return &tok
}

// Save writes a token to disk with 0600 permissions.
func (s *TokenStore) Save(tok *Token) error {
	if err := os.MkdirAll(s.Dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(), data, 0600)
}

// Delete removes the stored token file.
func (s *TokenStore) Delete() error {
	err := os.Remove(s.path())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
