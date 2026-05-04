package httpclient_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rishimantri795/CLICreator/runtime/httpclient"
)

// newTestClient creates a client pointed at the given test server.
func newTestClient(server *httptest.Server, auth httpclient.AuthProvider) *httpclient.Client {
	c := httpclient.New(server.URL, auth)
	c.RetryConfig = httpclient.RetryConfig{
		MaxRetries:     2,
		MaxElapsedTime: 5 * time.Second,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     50 * time.Millisecond,
	}
	return c
}

// --- SubstitutePath ---

func TestSubstitutePath(t *testing.T) {
	cases := []struct {
		template string
		params   map[string]string
		want     string
	}{
		{"/users/{id}", map[string]string{"id": "123"}, "/users/123"},
		{"/users/{id}/posts/{postId}", map[string]string{"id": "1", "postId": "2"}, "/users/1/posts/2"},
		{"/items/{name}", map[string]string{"name": "hello world"}, "/items/hello%20world"},
		{"/static", map[string]string{}, "/static"},
	}
	for _, tc := range cases {
		got := httpclient.SubstitutePath(tc.template, tc.params)
		if got != tc.want {
			t.Errorf("SubstitutePath(%q, %v) = %q, want %q", tc.template, tc.params, got, tc.want)
		}
	}
}

// --- Basic request execution ---

func TestDo_GET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/pets" {
			t.Errorf("expected /pets, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":1}]`))
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	resp, err := c.Do(&httpclient.Request{Method: "GET", Path: "/pets"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !resp.IsJSON {
		t.Error("expected IsJSON=true")
	}
}

func TestDo_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", q.Get("limit"))
		}
		tags := q["tag"]
		if len(tags) != 2 || tags[0] != "a" || tags[1] != "b" {
			t.Errorf("expected tag=[a,b], got %v", tags)
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	_, err := c.Do(&httpclient.Request{
		Method:      "GET",
		Path:        "/pets",
		QueryParams: map[string]string{"limit": "10"},
		ArrayParams: map[string][]string{"tag": {"a", "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDo_POST_JSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if body["name"] != "Fido" {
			t.Errorf("expected name=Fido, got %v", body["name"])
		}
		w.WriteHeader(201)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	_, err := c.Do(&httpclient.Request{
		Method: "POST",
		Path:   "/pets",
		Body:   map[string]any{"name": "Fido"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// --- Auth injection ---

func TestDo_BearerAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer mytoken" {
			t.Errorf("expected 'Bearer mytoken', got %q", auth)
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(srv, httpclient.BearerAuth{Token: "mytoken"})
	_, err := c.Do(&httpclient.Request{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDo_APIKeyHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "secret123" {
			t.Errorf("expected X-Api-Key=secret123, got %q", r.Header.Get("X-Api-Key"))
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(srv, httpclient.APIKeyAuth{Key: "secret123", Name: "X-Api-Key", Location: "header"})
	_, err := c.Do(&httpclient.Request{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDo_APIKeyQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api_key") != "secret123" {
			t.Errorf("expected api_key=secret123, got %q", r.URL.Query().Get("api_key"))
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(srv, httpclient.APIKeyAuth{Key: "secret123", Name: "api_key", Location: "query"})
	_, err := c.Do(&httpclient.Request{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatal(err)
	}
}

// --- Retry logic ---

func TestDo_Retries429(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	resp, err := c.Do(&httpclient.Request{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after retries, got %d", resp.StatusCode)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestDo_Retries5xx(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	resp, err := c.Do(&httpclient.Request{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
}

func TestDo_NoRetryOn4xx(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(404)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	resp, err := c.Do(&httpclient.Request{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
	if attempts != 1 {
		t.Errorf("expected exactly 1 attempt for 404, got %d", attempts)
	}
}

// --- Dry-run ---

func TestDo_DryRun(t *testing.T) {
	// Server should never be hit in dry-run mode
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server was called during dry-run")
	}))
	defer srv.Close()

	c := newTestClient(srv, httpclient.BearerAuth{Token: "supersecret"})
	c.DryRun = true

	resp, err := c.Do(&httpclient.Request{
		Method: "POST",
		Path:   "/pets",
		Body:   map[string]any{"name": "Fido"},
	})
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		t.Fatalf("dry-run response is not valid JSON: %v", err)
	}

	if result["dry_run"] != true {
		t.Error("expected dry_run=true")
	}

	reqMap, ok := result["request"].(map[string]any)
	if !ok {
		t.Fatal("expected request object in dry-run response")
	}
	if reqMap["method"] != "POST" {
		t.Errorf("expected method=POST, got %v", reqMap["method"])
	}

	// Auth header must be redacted
	headers, ok := reqMap["headers"].(map[string]any)
	if !ok {
		t.Fatal("expected headers in dry-run response")
	}
	for k, v := range headers {
		if strings.EqualFold(k, "authorization") && v != "[REDACTED]" {
			t.Errorf("Authorization header not redacted, got %v", v)
		}
	}
}
