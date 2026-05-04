// Package output provides response formatters for generated CLIs.
// Format is selected via the --output-format / -o flag.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/tidwall/gjson"
)

// Format enumerates the supported output formats.
type Format string

const (
	FormatJSON    Format = "json"
	FormatCompact Format = "compact"
	FormatTable   Format = "table"
	FormatPretty  Format = "pretty"
	FormatYAML    Format = "yaml"
	FormatRaw     Format = "raw"
)

// Print writes the response body to w in the requested format.
// body must be valid JSON (except FormatRaw which passes bytes through as-is).
func Print(w io.Writer, body []byte, format Format) error {
	if len(body) == 0 {
		return nil
	}

	// Raw passes bytes through untouched — no JSON parsing needed.
	if format == FormatRaw {
		_, err := w.Write(body)
		return err
	}

	// Parse the JSON once; determine if it is an array or an object.
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		// Not JSON — write as-is.
		_, err = fmt.Fprintln(w, string(body))
		return err
	}

	switch format {
	case FormatCompact:
		return printCompact(w, raw)
	case FormatTable:
		return printTable(w, raw)
	case FormatPretty:
		return printPretty(w, raw)
	case FormatYAML:
		return printYAML(w, raw, 0)
	default: // FormatJSON
		return printJSON(w, raw)
	}
}

// JQFilter extracts fields from body using GJSON query syntax and writes the
// result to w. Each matched value is printed on its own line.
//
// Common patterns:
//
//	"id"                     → top-level field
//	"user.email"             → nested field
//	"items.#.id"             → all id fields from an array
//	"items.0.name"           → first array element's name
//	"items.#(active==true)"  → first item where active is true
//	"items.#(active==true)#" → all items where active is true
func JQFilter(w io.Writer, body []byte, query string) error {
	if len(body) == 0 || query == "" {
		return nil
	}
	result := gjson.GetBytes(body, query)
	if !result.Exists() {
		return nil
	}
	// Array results: print each element on its own line (JSONL-style)
	if result.IsArray() {
		for _, item := range result.Array() {
			if item.Type == gjson.String {
				_, err := fmt.Fprintln(w, item.String())
				if err != nil {
					return err
				}
			} else {
				_, err := fmt.Fprintln(w, item.Raw)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	// Scalar or object: print as-is
	if result.Type == gjson.String {
		_, err := fmt.Fprintln(w, result.String())
		return err
	}
	_, err := fmt.Fprintln(w, result.Raw)
	return err
}

// printJSON pretty-prints the entire payload as JSON.
func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// printCompact emits JSONL: one JSON object per line for arrays,
// one line for single objects. Null fields are stripped.
func printCompact(w io.Writer, v any) error {
	switch typed := v.(type) {
	case []any:
		for _, item := range typed {
			line, err := json.Marshal(stripNulls(item))
			if err != nil {
				return err
			}
			fmt.Fprintln(w, string(line))
		}
	default:
		line, err := json.Marshal(stripNulls(v))
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(line))
	}
	return nil
}

// printTable renders an array of objects as a tab-separated table.
// Columns are derived from the keys of the first object, sorted.
// Non-array input falls back to printJSON.
func printTable(w io.Writer, v any) error {
	arr, ok := v.([]any)
	if !ok || len(arr) == 0 {
		return printJSON(w, v)
	}

	// Collect column names from the first item
	first, ok := arr[0].(map[string]any)
	if !ok {
		return printJSON(w, v)
	}
	cols := make([]string, 0, len(first))
	for k := range first {
		cols = append(cols, k)
	}
	sort.Strings(cols)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Header row
	fmt.Fprintln(tw, strings.Join(cols, "\t"))
	fmt.Fprintln(tw, strings.Repeat("-\t", len(cols)))

	for _, item := range arr {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		vals := make([]string, len(cols))
		for i, col := range cols {
			vals[i] = truncate(fmt.Sprintf("%v", row[col]), 60)
		}
		fmt.Fprintln(tw, strings.Join(vals, "\t"))
	}
	return tw.Flush()
}

// printPretty is a human-friendly format. Currently delegates to printJSON with
// a header. A richer implementation can add colours and section headings later.
func printPretty(w io.Writer, v any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	_, err := fmt.Fprint(w, buf.String())
	return err
}

// stripNulls recursively removes keys with null values from maps.
func stripNulls(v any) any {
	switch typed := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for k, val := range typed {
			if val == nil {
				continue
			}
			out[k] = stripNulls(val)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = stripNulls(item)
		}
		return out
	default:
		return v
	}
}

// printYAML renders a parsed JSON value as YAML without external dependencies.
// It handles maps, slices, strings, numbers, bools, and nil.
func printYAML(w io.Writer, v any, indent int) error {
	prefix := strings.Repeat("  ", indent)
	switch typed := v.(type) {
	case map[string]any:
		// Sort keys for deterministic output
		keys := make([]string, 0, len(typed))
		for k := range typed {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			val := typed[k]
			switch child := val.(type) {
			case map[string]any:
				fmt.Fprintf(w, "%s%s:\n", prefix, k)
				if err := printYAML(w, child, indent+1); err != nil {
					return err
				}
			case []any:
				fmt.Fprintf(w, "%s%s:\n", prefix, k)
				for _, item := range child {
					switch iv := item.(type) {
					case map[string]any:
						fmt.Fprintf(w, "%s-\n", prefix+"  ")
						if err := printYAML(w, iv, indent+2); err != nil {
							return err
						}
					default:
						fmt.Fprintf(w, "%s- %s\n", prefix+"  ", yamlScalar(iv))
					}
				}
			default:
				fmt.Fprintf(w, "%s%s: %s\n", prefix, k, yamlScalar(val))
			}
		}
	case []any:
		for _, item := range typed {
			switch iv := item.(type) {
			case map[string]any:
				fmt.Fprintf(w, "%s-\n", prefix)
				if err := printYAML(w, iv, indent+1); err != nil {
					return err
				}
			default:
				fmt.Fprintf(w, "%s- %s\n", prefix, yamlScalar(iv))
			}
		}
	default:
		fmt.Fprintf(w, "%s%s\n", prefix, yamlScalar(v))
	}
	return nil
}

// yamlScalar formats a scalar value for YAML output.
func yamlScalar(v any) string {
	if v == nil {
		return "null"
	}
	switch typed := v.(type) {
	case string:
		// Quote strings that contain special YAML characters
		if strings.ContainsAny(typed, ":#{}[]|>&*!,") || typed == "" ||
			typed == "true" || typed == "false" || typed == "null" {
			return fmt.Sprintf("%q", typed)
		}
		return typed
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// truncate shortens s to max runes, appending "…" if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
