// Package httpclient provides the generic HTTP client used by all generated CLIs.
// It handles auth injection, retries, dry-run mode, and debug logging.
package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// AuthProvider injects credentials into an outgoing request.
type AuthProvider interface {
	Apply(req *http.Request)
}

// BearerAuth injects an Authorization: Bearer <token> header.
type BearerAuth struct {
	Token string
}

func (b BearerAuth) Apply(req *http.Request) {
	if b.Token != "" {
		req.Header.Set("Authorization", "Bearer "+b.Token)
	}
}

// APIKeyAuth injects an API key into a header or query parameter.
type APIKeyAuth struct {
	Key      string
	Name     string // Header or query param name e.g. "X-Api-Key"
	Location string // "header" or "query"
}

func (a APIKeyAuth) Apply(req *http.Request) {
	if a.Key == "" {
		return
	}
	switch a.Location {
	case "query":
		q := req.URL.Query()
		q.Set(a.Name, a.Key)
		req.URL.RawQuery = q.Encode()
	default: // "header"
		req.Header.Set(a.Name, a.Key)
	}
}

// RetryConfig controls retry behaviour on 429 and 5xx responses.
type RetryConfig struct {
	MaxRetries     int           // Maximum number of retry attempts (default: 3)
	MaxElapsedTime time.Duration // Stop retrying after this total duration (default: 30s)
	InitialBackoff time.Duration // Starting backoff duration (default: 500ms)
	MaxBackoff     time.Duration // Cap on backoff duration (default: 10s)
}

func defaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		MaxElapsedTime: 30 * time.Second,
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
	}
}

// Client is the generic HTTP client shared by all generated CLI commands.
type Client struct {
	BaseURL     string
	Auth        AuthProvider
	Headers     map[string]string // Extra headers added to every request
	Timeout     time.Duration
	RetryConfig RetryConfig
	Debug      bool
	DryRun     bool
	HTTPClient *http.Client // injectable for testing
}

// New returns a Client with sensible defaults.
func New(baseURL string, auth AuthProvider) *Client {
	return &Client{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		Auth:        auth,
		Headers:     make(map[string]string),
		Timeout:     30 * time.Second,
		RetryConfig: defaultRetryConfig(),
	}
}

// Request represents a generic API call built from CLI flags.
type Request struct {
	Method      string
	Path        string              // Path with params already substituted, e.g. /users/123
	QueryParams map[string]string   // Single-value query params
	ArrayParams map[string][]string // Multi-value query params (emitted as ?tag=a&tag=b)
	Headers     map[string]string
	Body        any    // Will be JSON-serialized; nil means no body
	ContentType string // Defaults to "application/json" when Body is non-nil
}

// Response wraps the HTTP response with parsed metadata.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	IsJSON     bool
}

// Do executes the request, handling retries and dry-run mode.
func (c *Client) Do(req *Request) (*Response, error) {
	if c.DryRun {
		return c.dryRun(req)
	}

	start := time.Now()
	var lastResp *Response

	for attempt := 0; attempt <= c.RetryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.backoffDuration(attempt)
			if time.Since(start)+backoff > c.RetryConfig.MaxElapsedTime {
				break
			}
			time.Sleep(backoff)
		}

		resp, retryable, err := c.doOnce(req)
		if err != nil {
			if !retryable {
				return nil, err
			}
			continue
		}

		if !retryable {
			return resp, nil
		}

		lastResp = resp

		// Respect Retry-After header if present
		if ra := resp.Headers.Get("Retry-After"); ra != "" {
			if secs, parseErr := strconv.Atoi(ra); parseErr == nil {
				wait := time.Duration(secs) * time.Second
				if time.Since(start)+wait > c.RetryConfig.MaxElapsedTime {
					return resp, nil
				}
				time.Sleep(wait)
				attempt-- // don't consume a retry slot for this wait
				continue
			}
		}
	}

	if lastResp != nil {
		return lastResp, nil
	}
	return nil, fmt.Errorf("max retries exceeded")
}

// doOnce performs a single HTTP request. Returns (response, shouldRetry, error).
func (c *Client) doOnce(req *Request) (*Response, bool, error) {
	httpReq, err := c.buildHTTPRequest(req)
	if err != nil {
		return nil, false, fmt.Errorf("building request: %w", err)
	}

	if c.Debug {
		c.logRequest(httpReq, req.Body)
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: c.Timeout}
	}

	reqStart := time.Now()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		// Network errors are retryable
		return nil, true, fmt.Errorf("executing request: %w", err)
	}
	defer httpResp.Body.Close()
	latency := time.Since(reqStart)

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("reading response body: %w", err)
	}

	ct := httpResp.Header.Get("Content-Type")
	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       body,
		IsJSON:     strings.Contains(ct, "application/json"),
	}

	if c.Debug {
		c.logResponse(resp, latency)
	}

	// Retry on 429 or any 5xx
	shouldRetry := httpResp.StatusCode == 429 || httpResp.StatusCode >= 500
	return resp, shouldRetry, nil
}

// buildHTTPRequest constructs a *http.Request from our Request type.
func (c *Client) buildHTTPRequest(req *Request) (*http.Request, error) {
	fullURL := c.BaseURL + req.Path

	u, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", fullURL, err)
	}

	// Build query string
	q := u.Query()
	for k, v := range req.QueryParams {
		q.Set(k, v)
	}
	for k, vals := range req.ArrayParams {
		for _, v := range vals {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	// Serialize body
	var bodyReader io.Reader
	if req.Body != nil {
		data, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("serializing body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	httpReq, err := http.NewRequest(req.Method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	// Content-Type
	if req.Body != nil {
		ct := req.ContentType
		if ct == "" {
			ct = "application/json"
		}
		httpReq.Header.Set("Content-Type", ct)
	}

	// Client-level headers
	for k, v := range c.Headers {
		httpReq.Header.Set(k, v)
	}
	// Per-request headers override client-level
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Inject auth credentials
	if c.Auth != nil {
		c.Auth.Apply(httpReq)
	}

	return httpReq, nil
}

// dryRun returns a synthetic response describing what would be sent without
// making any network call. Sensitive headers are redacted.
func (c *Client) dryRun(req *Request) (*Response, error) {
	httpReq, err := c.buildHTTPRequest(req)
	if err != nil {
		return nil, fmt.Errorf("building request (dry-run): %w", err)
	}

	redactedHeaders := make(map[string]string)
	for k := range httpReq.Header {
		redactedHeaders[k] = redactHeader(k, httpReq.Header.Get(k))
	}

	payload := map[string]any{
		"dry_run": true,
		"request": map[string]any{
			"method":  httpReq.Method,
			"url":     httpReq.URL.String(),
			"headers": redactedHeaders,
			"body":    req.Body,
		},
	}

	data, _ := json.MarshalIndent(payload, "", "  ")
	return &Response{
		StatusCode: 0,
		Headers:    make(http.Header),
		Body:       data,
		IsJSON:     true,
	}, nil
}

// backoffDuration returns exponential backoff with jitter for the given attempt (1-based).
func (c *Client) backoffDuration(attempt int) time.Duration {
	backoff := c.RetryConfig.InitialBackoff * (1 << uint(attempt-1))
	if backoff > c.RetryConfig.MaxBackoff {
		backoff = c.RetryConfig.MaxBackoff
	}
	// Add up to 25% jitter
	jitter := time.Duration(rand.Int63n(int64(backoff) / 4))
	return backoff + jitter
}

func (c *Client) logRequest(req *http.Request, body any) {
	fmt.Fprintf(os.Stderr, "[DEBUG] --> %s %s\n", req.Method, req.URL.String())
	for k := range req.Header {
		fmt.Fprintf(os.Stderr, "[DEBUG]     %s: %s\n", k, redactHeader(k, req.Header.Get(k)))
	}
	if body != nil {
		data, _ := json.MarshalIndent(body, "            ", "  ")
		fmt.Fprintf(os.Stderr, "[DEBUG]     body: %s\n", data)
	}
}

func (c *Client) logResponse(resp *Response, latency time.Duration) {
	fmt.Fprintf(os.Stderr, "[DEBUG] <-- %d (%dms)\n", resp.StatusCode, latency.Milliseconds())
	if resp.IsJSON && len(resp.Body) > 0 {
		fmt.Fprintf(os.Stderr, "[DEBUG]     body: %s\n", resp.Body)
	}
}

// sensitiveHeaderWords are substrings that mark a header name as sensitive.
var sensitiveHeaderWords = []string{"authorization", "key", "secret", "token"}

// redactHeader returns "[REDACTED]" for headers whose names suggest credentials.
func redactHeader(name, value string) string {
	lower := strings.ToLower(name)
	for _, word := range sensitiveHeaderWords {
		if strings.Contains(lower, word) {
			return "[REDACTED]"
		}
	}
	return value
}

// SubstitutePath replaces {param} placeholders with URL-encoded values.
// e.g. SubstitutePath("/users/{id}", map[string]string{"id": "a b"}) → "/users/a%20b"
func SubstitutePath(pathTemplate string, params map[string]string) string {
	result := pathTemplate
	for k, v := range params {
		result = strings.ReplaceAll(result, "{"+k+"}", url.PathEscape(v))
	}
	return result
}
