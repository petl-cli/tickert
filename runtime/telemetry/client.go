package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// DefaultEndpoint is the CLICreator platform telemetry ingest URL.
const DefaultEndpoint = "https://telemetry.clicreator.dev/v1/events"

// flushTimeout is the maximum duration Flush() will wait for in-flight events.
// Kept short so telemetry never meaningfully delays process exit.
const flushTimeout = 3 * time.Second

// Client is the telemetry sink interface. Implementations must be safe for
// concurrent use and must never block the caller.
type Client interface {
	// Track queues an event for delivery. Must not block the caller.
	Track(event Event)
	// Flush blocks until all queued events are delivered or flushTimeout elapses.
	// Call via defer in Execute() so events are not lost when the process exits.
	Flush()
}

// NoopClient discards all events. Zero allocations, zero overhead.
// Used when the telemetry token is empty or the user has opted out.
type NoopClient struct{}

func (NoopClient) Track(Event) {}
func (NoopClient) Flush()      {}

// New returns the appropriate Client for the given token.
//
//   - Empty token → NoopClient (telemetry disabled at generation time).
//   - noTelemetryEnv set to any non-empty value → NoopClient (user opt-out).
//   - Otherwise → AsyncClient that ships events to endpoint.
//
// Pass endpoint="" to use DefaultEndpoint.
// Pass noTelemetryEnv="" to skip the env-var opt-out check.
func New(token, endpoint, noTelemetryEnv string) Client {
	if token == "" {
		return NoopClient{}
	}
	if noTelemetryEnv != "" && os.Getenv(noTelemetryEnv) != "" {
		return NoopClient{}
	}
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	return &AsyncClient{
		Token:    token,
		Endpoint: endpoint,
		httpCli:  &http.Client{Timeout: 5 * time.Second},
	}
}

// AsyncClient delivers events to the CLICreator platform in background goroutines.
// Track never blocks. Flush waits up to flushTimeout for in-flight sends to complete,
// ensuring events are not silently dropped on process exit.
type AsyncClient struct {
	Token    string
	Endpoint string
	httpCli  *http.Client
	wg       sync.WaitGroup
}

// Track fires a goroutine to deliver the event. The WaitGroup is incremented
// before the goroutine starts, ensuring Flush() sees it even if called immediately.
func (c *AsyncClient) Track(evt Event) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.send(evt)
	}()
}

// Flush waits up to flushTimeout for all in-flight events to complete.
// After the deadline, remaining in-flight sends are abandoned — CLI exit wins.
func (c *AsyncClient) Flush() {
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(flushTimeout):
	}
}

// send serialises and POSTs one event. All errors are silently swallowed —
// telemetry failures must never surface to the CLI user.
func (c *AsyncClient) send(evt Event) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("User-Agent", "clicreator/1.0")

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // drain body to allow TCP keep-alive reuse
}
