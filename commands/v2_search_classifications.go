package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rishimantri795/CLICreator/runtime/httpclient"
	"github.com/rishimantri795/CLICreator/runtime/output"
	"github.com/spf13/cobra"
)

var v2SearchClassificationsCmd = &cobra.Command{
	Use:   "search-classifications",
	Short: "Classification Search",
	RunE:  withTelemetry(runV2SearchClassifications),
}

var v2SearchClassificationsFlags struct {
	sort                   string
	keyword                string
	id                     string
	source                 string
	includeTest            string
	page                   string
	size                   string
	locale                 string
	includeLicensedContent string
	includeSpellcheck      string
}

func init() {
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.sort, "sort", "", "Sorting order of the search result")
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.keyword, "keyword", "", "Keyword to search on")
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.id, "id", "", "Filter entities by its id")
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.source, "source", "", "Filter entities by its source name")
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.includeTest, "include-test", "", "True if you want to have entities flag as test in the response. Only, if you only wanted test entities")
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.page, "page", "", "Page number")
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.size, "size", "", "Page size of the response")
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.locale, "locale", "", "The locale in ISO code format. Multiple comma-separated values can be provided. When omitting the country part of the code (e.g. only 'en' or 'fr') then the first matching locale is used. When using a '*' it matches all locales. '*' can only be used at the end (e.g. 'en-us,en,*') ")
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.includeLicensedContent, "include-licensed-content", "", "Yes if you want to display licensed content")
	v2SearchClassificationsCmd.Flags().StringVar(&v2SearchClassificationsFlags.includeSpellcheck, "include-spellcheck", "", "yes, to include spell check suggestions in the response.")

	v2Cmd.AddCommand(v2SearchClassificationsCmd)
}

func runV2SearchClassifications(cmd *cobra.Command, args []string) error {
	// --schema: print full input/output type contract without making any network call.
	if rootFlags.schema {
		type flagSchema struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			Required    bool   `json:"required"`
			Location    string `json:"location"`
			Description string `json:"description,omitempty"`
		}
		var flags []flagSchema
		flags = append(flags, flagSchema{
			Name:        "sort",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Sorting order of the search result",
		})
		flags = append(flags, flagSchema{
			Name:        "keyword",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Keyword to search on",
		})
		flags = append(flags, flagSchema{
			Name:        "id",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter entities by its id",
		})
		flags = append(flags, flagSchema{
			Name:        "source",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter entities by its source name",
		})
		flags = append(flags, flagSchema{
			Name:        "include-test",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "True if you want to have entities flag as test in the response. Only, if you only wanted test entities",
		})
		flags = append(flags, flagSchema{
			Name:        "page",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Page number",
		})
		flags = append(flags, flagSchema{
			Name:        "size",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Page size of the response",
		})
		flags = append(flags, flagSchema{
			Name:        "locale",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "The locale in ISO code format. Multiple comma-separated values can be provided. When omitting the country part of the code (e.g. only 'en' or 'fr') then the first matching locale is used. When using a '*' it matches all locales. '*' can only be used at the end (e.g. 'en-us,en,*') ",
		})
		flags = append(flags, flagSchema{
			Name:        "include-licensed-content",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Yes if you want to display licensed content",
		})
		flags = append(flags, flagSchema{
			Name:        "include-spellcheck",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "yes, to include spell check suggestions in the response.",
		})

		type responseSchema struct {
			Status      string `json:"status"`
			ContentType string `json:"content_type,omitempty"`
			Description string `json:"description,omitempty"`
		}
		var responses []responseSchema
		responses = append(responses, responseSchema{
			Status:      "200",
			ContentType: "*/*",
			Description: "successful operation",
		})

		schema := map[string]any{
			"command":     "search-classifications",
			"description": "Classification Search",
			"http": map[string]any{
				"method": "GET",
				"path":   "/discovery/v2/classifications",
			},
			"input": map[string]any{
				"flags":         flags,
				"body_flag":     false,
				"body_required": false,
			},
			"output": map[string]any{
				"responses": responses,
			},
			"semantics": map[string]any{
				"safe":         true,
				"idempotent":   true,
				"reversible":   true,
				"side_effects": []string{},
				"impact":       "low",
			},
			"requires_auth": true,
		}
		data, _ := json.MarshalIndent(schema, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	cfg, err := rootConfig()
	if err != nil {
		e := output.NetworkError(err)
		e.Write(os.Stderr)
		return output.NewExitError(e)
	}

	client := httpclient.New(cfg.BaseURL, cfg.AuthProvider())
	client.Debug = rootFlags.debug
	client.DryRun = rootFlags.dryRun
	if rootFlags.noRetries {
		client.RetryConfig.MaxRetries = 0
	}

	// Build path params
	pathParams := map[string]string{}

	req := &httpclient.Request{
		Method:      "GET",
		Path:        httpclient.SubstitutePath("/discovery/v2/classifications", pathParams),
		QueryParams: map[string]string{},
		ArrayParams: map[string][]string{},
		Headers:     map[string]string{},
	}

	// Query parameters
	if cmd.Flags().Changed("sort") {
		req.QueryParams["sort"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.sort)
	}
	if cmd.Flags().Changed("keyword") {
		req.QueryParams["keyword"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.keyword)
	}
	if cmd.Flags().Changed("id") {
		req.QueryParams["id"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.id)
	}
	if cmd.Flags().Changed("source") {
		req.QueryParams["source"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.source)
	}
	if cmd.Flags().Changed("include-test") {
		req.QueryParams["includeTest"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.includeTest)
	}
	if cmd.Flags().Changed("page") {
		req.QueryParams["page"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.page)
	}
	if cmd.Flags().Changed("size") {
		req.QueryParams["size"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.size)
	}
	if cmd.Flags().Changed("locale") {
		req.QueryParams["locale"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.locale)
	}
	if cmd.Flags().Changed("include-licensed-content") {
		req.QueryParams["includeLicensedContent"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.includeLicensedContent)
	}
	if cmd.Flags().Changed("include-spellcheck") {
		req.QueryParams["includeSpellcheck"] = fmt.Sprintf("%v", v2SearchClassificationsFlags.includeSpellcheck)
	}

	// Header parameters

	resp, err := client.Do(req)
	if err != nil {
		e := output.NetworkError(err)
		e.Write(os.Stderr)
		return output.NewExitError(e)
	}

	if resp.StatusCode >= 400 {
		e := output.HTTPError(resp.StatusCode, resp.Body)
		e.Write(os.Stderr)
		return output.NewExitError(e)
	}

	if rootFlags.jq != "" {
		return output.JQFilter(os.Stdout, resp.Body, rootFlags.jq)
	}
	return output.Print(os.Stdout, resp.Body, output.Format(cfg.OutputFormat))
}
