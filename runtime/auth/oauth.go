package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// LoginConfig holds the parameters needed to run an OAuth 2.0 Authorization Code + PKCE flow.
type LoginConfig struct {
	AuthorizeURL string
	TokenURL     string
	ClientID     string
	Scopes       []string
}

// tokenResponse is the JSON body returned by the token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"`
}

// Login runs the Authorization Code + PKCE flow:
//  1. Start a localhost callback server on a random port
//  2. Open the user's browser to the authorization URL
//  3. Wait for the callback with the authorization code (2 minute timeout)
//  4. Exchange the code for tokens
//
// Returns the obtained token or an error.
func Login(cfg LoginConfig) (*Token, error) {
	// Generate PKCE code verifier and challenge
	verifier, challenge, err := generatePKCE()
	if err != nil {
		return nil, fmt.Errorf("generating PKCE: %w", err)
	}

	// Generate random state parameter
	state, err := randomString(32)
	if err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}

	// Start local callback server on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// Channel to receive the authorization code
	type callbackResult struct {
		code string
		err  error
	}
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			desc := r.URL.Query().Get("error_description")
			resultCh <- callbackResult{err: fmt.Errorf("authorization error: %s — %s", errParam, desc)}
			fmt.Fprintf(w, "<html><body><h2>Authorization failed</h2><p>%s: %s</p><p>You can close this window.</p></body></html>", errParam, desc)
			return
		}

		if r.URL.Query().Get("state") != state {
			resultCh <- callbackResult{err: fmt.Errorf("state mismatch — possible CSRF attack")}
			fmt.Fprint(w, "<html><body><h2>Error: state mismatch</h2><p>You can close this window.</p></body></html>")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			resultCh <- callbackResult{err: fmt.Errorf("no authorization code in callback")}
			fmt.Fprint(w, "<html><body><h2>Error: no code received</h2><p>You can close this window.</p></body></html>")
			return
		}

		resultCh <- callbackResult{code: code}
		fmt.Fprint(w, "<html><body><h2>Login successful!</h2><p>You can close this window and return to the terminal.</p></body></html>")
	})

	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	// Build the authorization URL
	authURL, err := buildAuthorizeURL(cfg.AuthorizeURL, cfg.ClientID, redirectURI, cfg.Scopes, state, challenge)
	if err != nil {
		return nil, fmt.Errorf("building authorize URL: %w", err)
	}

	// Try to open the browser; if it fails, print the URL for manual use
	if err := openBrowser(authURL); err != nil {
		fmt.Println("Could not open browser automatically.")
		fmt.Println("Open this URL in your browser to log in:")
		fmt.Println()
		fmt.Println("  " + authURL)
		fmt.Println()
	} else {
		fmt.Println("Opening browser for login...")
		fmt.Println("If the browser didn't open, visit:")
		fmt.Println()
		fmt.Println("  " + authURL)
		fmt.Println()
	}

	fmt.Println("Waiting for authorization (2 minute timeout)...")

	// Wait for callback or timeout
	select {
	case result := <-resultCh:
		if result.err != nil {
			return nil, result.err
		}
		// Exchange code for token
		return exchangeCode(cfg.TokenURL, cfg.ClientID, result.code, redirectURI, verifier)
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("login timed out — no callback received within 2 minutes")
	}
}

// RefreshAccessToken exchanges a refresh token for a new access token.
// Returns nil if the refresh fails (caller should treat as expired session).
func RefreshAccessToken(tokenURL, clientID, refreshToken string) *Token {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {refreshToken},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil
	}

	expiresIn := tr.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600 // default 1 hour
	}

	return &Token{
		AccessToken:  tr.AccessToken,
		RefreshToken: firstNonEmpty(tr.RefreshToken, refreshToken), // keep old refresh token if new one not issued
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
}

func exchangeCode(tokenURL, clientID, code, redirectURI, verifier string) (*Token, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, body)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	expiresIn := tr.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}

	return &Token{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}, nil
}

func buildAuthorizeURL(baseURL, clientID, redirectURI string, scopes []string, state, challenge string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	if len(scopes) > 0 {
		q.Set("scope", strings.Join(scopes, " "))
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// generatePKCE creates a code_verifier and its S256 code_challenge.
func generatePKCE() (verifier, challenge string, err error) {
	// code_verifier: 32 random bytes → base64url (43 characters)
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(buf)

	// code_challenge: SHA-256(verifier) → base64url
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge, nil
}

func randomString(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	return cmd.Start()
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
