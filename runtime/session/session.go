// Package session provides a persistent, file-backed session ID for generated CLIs.
//
// A "session" is a run of activity within a 30-minute idle window — the same
// threshold used by standard analytics platforms. Any command execution resets
// the clock, so a multi-step agent workflow that spans 45 minutes of wall time
// but has no gap longer than 30 minutes counts as a single session.
//
// The session file lives at {configDir}/session_id. Two processes sharing the
// same configDir (parallel agent invocations in the same project) will share a
// session ID, which is intentional: they are part of the same working session.
package session

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	filename    = "session_id"
	idleTimeout = 30 * time.Minute
)

// GetOrCreateSessionID returns a stable UUID for the current analytics session.
//
// Rules:
//   - If {configDir}/session_id exists and its mtime is within the last 30 minutes,
//     the stored UUID is returned as-is.
//   - Otherwise a new UUID v4 is generated, written to the file, and returned.
//
// The function is intentionally best-effort: if the file cannot be read or written
// (no home directory, read-only filesystem, permission error) it returns a fresh
// in-process UUID so telemetry always has a SessionId without ever failing the CLI.
func GetOrCreateSessionID(configDir string) string {
	if configDir == "" {
		return newUUID()
	}

	path := filepath.Join(configDir, filename)

	if info, err := os.Stat(path); err == nil && time.Since(info.ModTime()) < idleTimeout {
		if data, err := os.ReadFile(path); err == nil {
			if id := strings.TrimSpace(string(data)); isValidUUID(id) {
				return id
			}
		}
	}

	id := newUUID()
	// Best-effort write; failure is silent — caller gets a valid in-process ID either way.
	_ = os.MkdirAll(configDir, 0700)
	_ = os.WriteFile(path, []byte(id+"\n"), 0600)
	return id
}

// Touch updates the mtime of the session file, extending the idle window by
// another 30 minutes from now. Call once per command invocation (success or
// error) so that multi-step workflows keep a single session across many runs.
// No-op when configDir is empty or the file does not exist.
func Touch(configDir string) {
	if configDir == "" {
		return
	}
	path := filepath.Join(configDir, filename)
	now := time.Now()
	_ = os.Chtimes(path, now, now)
}

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// isValidUUID does a lightweight check: correct length, dashes in the right
// positions, hex everywhere else. Rejects empty files, truncated writes, and
// hand-edited garbage without pulling in a UUID library.
func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}
