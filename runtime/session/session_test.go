package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rishimantri795/CLICreator/runtime/session"
)

// ---- helpers ----

func tempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func sessionFile(dir string) string {
	return filepath.Join(dir, "session_id")
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return strings.TrimSpace(string(data))
}

func setMtime(t *testing.T, path string, mt time.Time) {
	t.Helper()
	if err := os.Chtimes(path, mt, mt); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
}

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

// ---- GetOrCreateSessionID ----

func TestGetOrCreate_FirstCall_CreatesFile(t *testing.T) {
	dir := tempDir(t)

	id := session.GetOrCreateSessionID(dir)

	if !isValidUUID(id) {
		t.Errorf("expected valid UUID v4, got %q", id)
	}
	stored := readFile(t, sessionFile(dir))
	if stored != id {
		t.Errorf("file contains %q, want %q", stored, id)
	}
}

func TestGetOrCreate_RecentFile_ReusesID(t *testing.T) {
	dir := tempDir(t)

	first := session.GetOrCreateSessionID(dir)
	second := session.GetOrCreateSessionID(dir)

	if first != second {
		t.Errorf("expected same ID within idle window: first=%q second=%q", first, second)
	}
}

func TestGetOrCreate_ExpiredFile_GeneratesNewID(t *testing.T) {
	dir := tempDir(t)

	first := session.GetOrCreateSessionID(dir)

	// Back-date the file past the 30-minute idle window.
	expired := time.Now().Add(-31 * time.Minute)
	setMtime(t, sessionFile(dir), expired)

	second := session.GetOrCreateSessionID(dir)

	if first == second {
		t.Errorf("expected new ID after idle timeout, got same ID %q", first)
	}
	if !isValidUUID(second) {
		t.Errorf("new ID is not a valid UUID: %q", second)
	}
}

func TestGetOrCreate_ExactBoundary_StillFresh(t *testing.T) {
	dir := tempDir(t)

	first := session.GetOrCreateSessionID(dir)

	// 29 minutes ago — still within the window.
	setMtime(t, sessionFile(dir), time.Now().Add(-29*time.Minute))

	second := session.GetOrCreateSessionID(dir)
	if first != second {
		t.Errorf("session should survive a 29-minute gap: first=%q second=%q", first, second)
	}
}

func TestGetOrCreate_CorruptFile_GeneratesNewID(t *testing.T) {
	dir := tempDir(t)

	// Write garbage into the session file with a fresh mtime.
	path := sessionFile(dir)
	if err := os.WriteFile(path, []byte("not-a-uuid\n"), 0600); err != nil {
		t.Fatal(err)
	}

	id := session.GetOrCreateSessionID(dir)
	if !isValidUUID(id) {
		t.Errorf("expected valid UUID after corrupt file, got %q", id)
	}
}

func TestGetOrCreate_EmptyFile_GeneratesNewID(t *testing.T) {
	dir := tempDir(t)

	path := sessionFile(dir)
	if err := os.WriteFile(path, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	id := session.GetOrCreateSessionID(dir)
	if !isValidUUID(id) {
		t.Errorf("expected valid UUID after empty file, got %q", id)
	}
}

func TestGetOrCreate_EmptyDir_ReturnsValidID(t *testing.T) {
	// Empty configDir → no file I/O → in-process UUID, still valid.
	id := session.GetOrCreateSessionID("")
	if !isValidUUID(id) {
		t.Errorf("expected valid UUID for empty configDir, got %q", id)
	}
}

func TestGetOrCreate_NonexistentParent_CreatesDir(t *testing.T) {
	// configDir does not exist yet — GetOrCreate should mkdir it.
	base := tempDir(t)
	dir := filepath.Join(base, "nested", "config")

	id := session.GetOrCreateSessionID(dir)

	if !isValidUUID(id) {
		t.Errorf("expected valid UUID, got %q", id)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected directory %s to be created: %v", dir, err)
	}
}

func TestGetOrCreate_FileContainsNewline_ParsedCorrectly(t *testing.T) {
	dir := tempDir(t)

	// Simulate a file written with a trailing newline (normal case).
	expected := "a1b2c3d4-e5f6-4789-abcd-ef0123456789"
	path := sessionFile(dir)
	if err := os.WriteFile(path, []byte(expected+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	id := session.GetOrCreateSessionID(dir)
	if id != expected {
		t.Errorf("got %q, want %q", id, expected)
	}
}

func TestGetOrCreate_TwoCallsYieldDifferentIDsAfterExpiry(t *testing.T) {
	dir := tempDir(t)

	first := session.GetOrCreateSessionID(dir)
	setMtime(t, sessionFile(dir), time.Now().Add(-31*time.Minute))
	second := session.GetOrCreateSessionID(dir)
	setMtime(t, sessionFile(dir), time.Now().Add(-31*time.Minute))
	third := session.GetOrCreateSessionID(dir)

	if first == second || second == third {
		t.Errorf("each expired call should generate a fresh ID: %q %q %q", first, second, third)
	}
}

// ---- Touch ----

func TestTouch_UpdatesMtime(t *testing.T) {
	dir := tempDir(t)
	session.GetOrCreateSessionID(dir) // create the file

	// Back-date to confirm Touch changes it.
	old := time.Now().Add(-10 * time.Minute)
	setMtime(t, sessionFile(dir), old)

	before, err := os.Stat(sessionFile(dir))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond) // ensure clock ticks
	session.Touch(dir)

	after, err := os.Stat(sessionFile(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().After(before.ModTime()) {
		t.Errorf("mtime not updated: before=%v after=%v", before.ModTime(), after.ModTime())
	}
}

func TestTouch_ExtendsMakesFileReusable(t *testing.T) {
	dir := tempDir(t)

	first := session.GetOrCreateSessionID(dir)

	// Age the file past the idle window, then Touch to renew it.
	setMtime(t, sessionFile(dir), time.Now().Add(-31*time.Minute))
	session.Touch(dir)

	second := session.GetOrCreateSessionID(dir)
	if first != second {
		t.Errorf("Touch should have extended the window: first=%q second=%q", first, second)
	}
}

func TestTouch_NoopOnMissingFile(t *testing.T) {
	dir := tempDir(t)
	// File doesn't exist — Touch must not panic or create it.
	session.Touch(dir)

	if _, err := os.Stat(sessionFile(dir)); !os.IsNotExist(err) {
		t.Errorf("Touch should not create the session file, stat err: %v", err)
	}
}

func TestTouch_NoopOnEmptyDir(t *testing.T) {
	// Must not panic.
	session.Touch("")
}

// ---- ID uniqueness ----

func TestGetOrCreate_IDsAreUnique(t *testing.T) {
	// Generate 50 independent IDs from different directories and verify no collisions.
	seen := make(map[string]struct{}, 50)
	for i := 0; i < 50; i++ {
		dir := tempDir(t)
		id := session.GetOrCreateSessionID(dir)
		if _, dup := seen[id]; dup {
			t.Errorf("UUID collision at i=%d: %q", i, id)
		}
		seen[id] = struct{}{}
	}
}
