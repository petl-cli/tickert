package output

import (
	"os"
)

var knownAgentEnvVars = []string{
	"CLAUDE_CODE",
	"CURSOR_SESSION_ID",
	"CODEX",
	"AIDER",
	"CLINE",
	"WINDSURF_SESSION",
	"GITHUB_COPILOT",
	"AMAZON_Q_SESSION",
	"GEMINI_CODE_ASSIST",
	"CODY",
}

// DetectAgentMode returns true if any known agent environment variable is set,
// or if the --agent-mode flag was explicitly passed.
func DetectAgentMode(explicitFlag bool) bool {
	if explicitFlag {
		return true
	}
	for _, env := range knownAgentEnvVars {
		if os.Getenv(env) != "" {
			return true
		}
	}
	return false
}

// StdoutIsPipe returns true when stdout is connected to a pipe rather than a
// terminal. In this case we default to compact output so piped commands work
// without the user needing to set -o compact manually.
func StdoutIsPipe() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

// DefaultFormat returns the appropriate default format based on context.
// Agent mode or piped stdout → compact. Otherwise → json.
func DefaultFormat(agentMode bool) Format {
	if agentMode || StdoutIsPipe() {
		return FormatCompact
	}
	return FormatJSON
}
