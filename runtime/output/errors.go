package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Exit codes for generated CLIs. Agents can branch on $? without parsing stderr.
const (
	ExitOK         = 0 // Success
	ExitErr        = 1 // Generic / unknown error
	ExitAuth       = 2 // Authentication failed (401/403)
	ExitNotFound   = 3 // Resource not found (404)
	ExitValidation = 4 // Validation / bad request (422/400)
	ExitRateLimit  = 5 // Rate limited (429)
	ExitServer     = 6 // Server error (5xx)
	ExitNetwork    = 7 // Network / connection error
	ExitSchema     = 8 // Response schema drift detected
)

// CLIError is the structured error emitted to stderr in agent mode and
// for all HTTP error responses. It is always valid JSON.
type CLIError struct {
	Error      bool   `json:"error"`
	Code       string `json:"code"`
	Status     int    `json:"status"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Details    any    `json:"details,omitempty"`
	ExitCode   int    `json:"exit_code"`
}

// HTTPError maps an HTTP status code to a CLIError.
// body is the raw response body, included as Details for 422 responses.
func HTTPError(status int, body []byte) *CLIError {
	e := &CLIError{Error: true, Status: status}
	switch {
	case status == 401:
		e.Code = "auth_failed"
		e.Message = "Authentication failed"
		e.Suggestion = "Set the API key via environment variable or run '<cli> configure'"
		e.ExitCode = ExitAuth
	case status == 403:
		e.Code = "forbidden"
		e.Message = "Access denied"
		e.Suggestion = "Check that your API key has the required permissions"
		e.ExitCode = ExitAuth
	case status == 404:
		e.Code = "not_found"
		e.Message = "Resource not found"
		e.Suggestion = "Verify the ID or path is correct"
		e.ExitCode = ExitNotFound
	case status == 400 || status == 422:
		e.Code = "validation_error"
		e.Message = "Request validation failed"
		if len(body) > 0 {
			var parsed any
			if json.Unmarshal(body, &parsed) == nil {
				e.Details = parsed
			} else {
				e.Details = string(body)
			}
		}
		e.ExitCode = ExitValidation
	case status == 429:
		e.Code = "rate_limited"
		e.Message = "Rate limit exceeded"
		e.Suggestion = "Wait before retrying or reduce request frequency"
		e.ExitCode = ExitRateLimit
	case status >= 500:
		e.Code = "server_error"
		e.Message = fmt.Sprintf("Server error (%d)", status)
		e.Suggestion = "Retry the request or check the API status page"
		e.ExitCode = ExitServer
	default:
		e.Code = "request_failed"
		e.Message = fmt.Sprintf("Request failed with status %d", status)
		e.ExitCode = ExitErr
	}
	return e
}

// NetworkError wraps a Go error (DNS failure, connection refused, timeout, etc.)
// as a CLIError. Used by generated commands for non-HTTP failures.
func NetworkError(err error) *CLIError {
	return &CLIError{
		Error:      true,
		Code:       "network_error",
		Status:     0,
		Message:    err.Error(),
		Suggestion: "Check your network connection and the API base URL",
		ExitCode:   ExitNetwork,
	}
}

// Write serialises the error as JSON to w (typically os.Stderr).
func (e *CLIError) Write(w io.Writer) {
	data, _ := json.Marshal(e)
	fmt.Fprintln(w, string(data))
}

// ExitCodeOrDefault returns the exit code for this error, falling back to ExitErr.
func (e *CLIError) ExitCodeOrDefault() int {
	if e.ExitCode != 0 {
		return e.ExitCode
	}
	return ExitErr
}

// ExitError wraps a CLIError and implements the Go error interface so that generated
// RunE functions can return it instead of calling os.Exit directly. This lets
// deferred cleanup (e.g. telemetry flushing) run before the process exits.
//
// Execute() in the generated root.go inspects this type via errors.As and calls
// os.Exit with the embedded exit code.
type ExitError struct {
	CLI *CLIError
}

// NewExitError constructs an ExitError from a CLIError.
func NewExitError(cli *CLIError) *ExitError {
	return &ExitError{CLI: cli}
}

// Error implements the error interface. The message matches the CLI error's message
// so Cobra's default error printing (if not silenced) shows something useful.
func (e *ExitError) Error() string {
	if e.CLI != nil {
		return e.CLI.Message
	}
	return "error"
}

// ExitCode returns the process exit code that the CLI should use.
func (e *ExitError) ExitCode() int {
	if e.CLI != nil {
		return e.CLI.ExitCodeOrDefault()
	}
	return ExitErr
}
