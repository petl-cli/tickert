// Package feedback submits agent-authored feedback about a generated CLI to
// the CLICreator platform. Unlike telemetry, this is an explicit, agent-invoked
// action (`<cli> feedback "..."`), so calls are synchronous: the agent gets a
// confirmation or a clear failure, and there's no background goroutine.
package feedback

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SchemaVersion pins the wire format. Bump only on additive-breaking changes.
// CLIs in the wild emit forever, so the ingestion service must keep accepting
// old versions for as long as those CLIs exist.
const SchemaVersion = "1"

// DefaultEndpoint is the platform feedback ingest URL. Override at generation
// time by passing a different endpoint to Submit.
const DefaultEndpoint = "https://feedback.clicreator.dev/v1/feedback"

// MaxMessageLen mirrors the ingestion service's cap. We trim client-side too
// so an oversized message produces a clear local error instead of a 400.
const MaxMessageLen = 4_000

// Payload is the wire format. Keep additive-only.
type Payload struct {
	SchemaVersion  string `json:"schema_version"`
	CLIVersion     string `json:"cli_version"`
	Message        string `json:"message"`
	CommandContext string `json:"command_context,omitempty"`
	AgentType      string `json:"agent_type,omitempty"`
}

// Submit POSTs feedback to the platform. Returns the server-issued feedback id
// on success, or an error suitable for surfacing to the agent. Token is the
// per-CLI identifier baked in at generation time (same value as the telemetry
// token — one identifier per CLI).
func Submit(ctx context.Context, endpoint, token string, p Payload) (string, error) {
	if token == "" {
		return "", errors.New("feedback is not configured for this CLI")
	}
	if p.Message == "" {
		return "", errors.New("message is required")
	}
	if len(p.Message) > MaxMessageLen {
		return "", fmt.Errorf("message exceeds %d characters", MaxMessageLen)
	}
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	p.SchemaVersion = SchemaVersion

	data, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("encoding payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "clicreator-feedback/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("submitting feedback: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("feedback rejected (%d): %s", resp.StatusCode, string(body))
	}

	var out struct {
		Status string `json:"status"`
		ID     string `json:"id"`
	}
	_ = json.Unmarshal(body, &out)
	return out.ID, nil
}
