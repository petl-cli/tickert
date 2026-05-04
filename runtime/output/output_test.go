package output_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/rishimantri795/CLICreator/runtime/output"
)

// --- DetectAgentMode ---

func TestDetectAgentMode_NoEnv(t *testing.T) {
	// Ensure none of the agent env vars are set
	vars := []string{"CLAUDE_CODE", "CURSOR_SESSION_ID", "CODEX", "AIDER", "CLINE",
		"WINDSURF_SESSION", "GITHUB_COPILOT", "AMAZON_Q_SESSION", "GEMINI_CODE_ASSIST", "CODY"}
	for _, v := range vars {
		os.Unsetenv(v)
	}
	if output.DetectAgentMode(false) {
		t.Error("expected DetectAgentMode()=false with no agent env vars set")
	}
}

func TestDetectAgentMode_ClaudeCode(t *testing.T) {
	os.Setenv("CLAUDE_CODE", "1")
	defer os.Unsetenv("CLAUDE_CODE")
	if !output.DetectAgentMode(false) {
		t.Error("expected DetectAgentMode()=true when CLAUDE_CODE is set")
	}
}

func TestDetectAgentMode_Cursor(t *testing.T) {
	os.Setenv("CURSOR_SESSION_ID", "abc")
	defer os.Unsetenv("CURSOR_SESSION_ID")
	if !output.DetectAgentMode(false) {
		t.Error("expected DetectAgentMode()=true when CURSOR_SESSION_ID is set")
	}
}

// --- HTTPError ---

func TestHTTPError_401(t *testing.T) {
	e := output.HTTPError(401, nil)
	if e.Code != "auth_failed" {
		t.Errorf("expected auth_failed, got %q", e.Code)
	}
	if e.Status != 401 {
		t.Errorf("expected status 401, got %d", e.Status)
	}
	if !e.Error {
		t.Error("expected Error=true")
	}
}

func TestHTTPError_404(t *testing.T) {
	e := output.HTTPError(404, nil)
	if e.Code != "not_found" {
		t.Errorf("expected not_found, got %q", e.Code)
	}
}

func TestHTTPError_422_IncludesBody(t *testing.T) {
	body := []byte(`{"field":"email","message":"invalid format"}`)
	e := output.HTTPError(422, body)
	if e.Code != "validation_error" {
		t.Errorf("expected validation_error, got %q", e.Code)
	}
	if e.Details == nil {
		t.Error("expected Details to be populated for 422")
	}
}

func TestHTTPError_429(t *testing.T) {
	e := output.HTTPError(429, nil)
	if e.Code != "rate_limited" {
		t.Errorf("expected rate_limited, got %q", e.Code)
	}
}

func TestHTTPError_503(t *testing.T) {
	e := output.HTTPError(503, nil)
	if e.Code != "server_error" {
		t.Errorf("expected server_error, got %q", e.Code)
	}
}

func TestHTTPError_Write_IsJSON(t *testing.T) {
	var buf bytes.Buffer
	e := output.HTTPError(404, nil)
	e.Write(&buf)
	var parsed map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &parsed); err != nil {
		t.Fatalf("CLIError.Write did not produce valid JSON: %v", err)
	}
}

// --- Formatters ---

var arrayJSON = []byte(`[{"id":1,"name":"Alice","role":null},{"id":2,"name":"Bob","role":"admin"}]`)
var objectJSON = []byte(`{"id":1,"name":"Alice"}`)

func TestPrint_JSON_Array(t *testing.T) {
	var buf bytes.Buffer
	if err := output.Print(&buf, arrayJSON, output.FormatJSON); err != nil {
		t.Fatal(err)
	}
	// Must be valid JSON
	var v any
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("JSON output is not valid JSON: %v", err)
	}
}

func TestPrint_Compact_Array_OneLinePerItem(t *testing.T) {
	var buf bytes.Buffer
	if err := output.Print(&buf, arrayJSON, output.FormatCompact); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines for 2-item array, got %d: %q", len(lines), buf.String())
	}
	// Each line must be valid JSON
	for i, line := range lines {
		var v any
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			t.Errorf("line %d is not valid JSON: %q", i, line)
		}
	}
}

func TestPrint_Compact_StripsNulls(t *testing.T) {
	var buf bytes.Buffer
	if err := output.Print(&buf, arrayJSON, output.FormatCompact); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "null") {
		t.Errorf("compact output should strip null values, got: %s", buf.String())
	}
}

func TestPrint_Compact_SingleObject_OneLine(t *testing.T) {
	var buf bytes.Buffer
	if err := output.Print(&buf, objectJSON, output.FormatCompact); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected single object to produce 1 line, got %d", len(lines))
	}
}

func TestPrint_Table_HasHeader(t *testing.T) {
	var buf bytes.Buffer
	if err := output.Print(&buf, arrayJSON, output.FormatTable); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "id") || !strings.Contains(out, "name") {
		t.Errorf("table output should contain column headers, got:\n%s", out)
	}
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Errorf("table output should contain data rows, got:\n%s", out)
	}
}

func TestPrint_Table_NonArray_FallsBackToJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := output.Print(&buf, objectJSON, output.FormatTable); err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("table fallback for non-array should produce valid JSON: %v", err)
	}
}

func TestPrint_EmptyBody(t *testing.T) {
	var buf bytes.Buffer
	if err := output.Print(&buf, nil, output.FormatJSON); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty body, got %q", buf.String())
	}
}

func TestDefaultFormat_AgentMode(t *testing.T) {
	if output.DefaultFormat(true) != output.FormatCompact {
		t.Error("expected compact format in agent mode")
	}
}

func TestDefaultFormat_Normal(t *testing.T) {
	// In the test runner stdout is always a pipe, so DefaultFormat returns compact
	// even without agent mode — that is correct behaviour (piped = compact).
	f := output.DefaultFormat(false)
	if f != output.FormatJSON && f != output.FormatCompact {
		t.Errorf("expected json or compact format in normal mode, got %s", f)
	}
}

// --- JQFilter ---

var jqObject = []byte(`{"id":"usr_123","email":"alice@example.com","active":true,"score":42}`)
var jqArray = []byte(`{"items":[{"id":"usr_1","email":"alice@example.com","active":true},{"id":"usr_2","email":"bob@example.com","active":false},{"id":"usr_3","email":"carol@example.com","active":true}],"total":3}`)

func TestJQFilter_ScalarField(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, jqObject, "id"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "usr_123" {
		t.Errorf("expected usr_123, got %q", got)
	}
}

func TestJQFilter_NestedField(t *testing.T) {
	body := []byte(`{"user":{"email":"alice@example.com","role":"admin"}}`)
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, body, "user.email"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %q", got)
	}
}

func TestJQFilter_AllFieldsFromArray(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, jqArray, "items.#.id"); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 ids, got %d: %q", len(lines), buf.String())
	}
	if lines[0] != "usr_1" || lines[1] != "usr_2" || lines[2] != "usr_3" {
		t.Errorf("unexpected ids: %v", lines)
	}
}

func TestJQFilter_ArrayIndex(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, jqArray, "items.0.email"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %q", got)
	}
}

func TestJQFilter_FilterByCondition(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, jqArray, "items.#(active==true)#.email"); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 active users, got %d: %q", len(lines), buf.String())
	}
	if lines[0] != "alice@example.com" || lines[1] != "carol@example.com" {
		t.Errorf("unexpected filtered emails: %v", lines)
	}
}

func TestJQFilter_Count(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, jqArray, "items.#"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "3" {
		t.Errorf("expected count 3, got %q", got)
	}
}

func TestJQFilter_BooleanField(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, jqObject, "active"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "true" {
		t.Errorf("expected true, got %q", got)
	}
}

func TestJQFilter_NumericField(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, jqObject, "score"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "42" {
		t.Errorf("expected 42, got %q", got)
	}
}

func TestJQFilter_MissingField_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, jqObject, "nonexistent"); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for missing field, got %q", buf.String())
	}
}

func TestJQFilter_EmptyBody_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, nil, "id"); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty body, got %q", buf.String())
	}
}

func TestJQFilter_EmptyQuery_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JQFilter(&buf, jqObject, ""); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty query, got %q", buf.String())
	}
}
