package auth

import (
	"net/http"
)

// OAuth2Auth implements httpclient.AuthProvider for OAuth 2.0.
// It loads tokens from disk, refreshes them transparently when expired,
// and injects the access token as a Bearer header.
type OAuth2Auth struct {
	TokenStore *TokenStore
	TokenURL   string
	ClientID   string
}

// Apply injects the OAuth2 Bearer token into the request.
// If the token is expired, it attempts a silent refresh.
// If refresh fails, no auth header is set (the 401 propagates to the user).
func (o *OAuth2Auth) Apply(req *http.Request) {
	tok := o.TokenStore.Load()
	if tok == nil {
		return
	}

	if tok.IsExpired() {
		if tok.RefreshToken == "" {
			return // no refresh token, can't refresh
		}
		refreshed := RefreshAccessToken(o.TokenURL, o.ClientID, tok.RefreshToken)
		if refreshed == nil {
			return // refresh failed, let 401 propagate
		}
		_ = o.TokenStore.Save(refreshed)
		tok = refreshed
	}

	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
}
