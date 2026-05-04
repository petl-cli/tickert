package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rishimantri795/CLICreator/runtime/httpclient"
	"github.com/rishimantri795/CLICreator/runtime/output"
	"github.com/spf13/cobra"
)

var v2GetSegmentDetailsCmd = &cobra.Command{
	Use:   "get-segment-details",
	Short: "Get Segment Details",
	RunE:  withTelemetry(runV2GetSegmentDetails),
}

var v2GetSegmentDetailsFlags struct {
	id                     string
	locale                 string
	includeLicensedContent string
}

func init() {
	v2GetSegmentDetailsCmd.Flags().StringVar(&v2GetSegmentDetailsFlags.id, "id", "", "ID of the segment")
	v2GetSegmentDetailsCmd.MarkFlagRequired("id")
	v2GetSegmentDetailsCmd.Flags().StringVar(&v2GetSegmentDetailsFlags.locale, "locale", "", "The locale in ISO code format. Multiple comma-separated values can be provided. When omitting the country part of the code (e.g. only 'en' or 'fr') then the first matching locale is used. When using a '*' it matches all locales. '*' can only be used at the end (e.g. 'en-us,en,*') ")
	v2GetSegmentDetailsCmd.Flags().StringVar(&v2GetSegmentDetailsFlags.includeLicensedContent, "include-licensed-content", "", "True if you want to display licensed content")

	v2Cmd.AddCommand(v2GetSegmentDetailsCmd)
}

func runV2GetSegmentDetails(cmd *cobra.Command, args []string) error {
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
			Name:        "id",
			Type:        "string",
			Required:    true,
			Location:    "path",
			Description: "ID of the segment",
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
			Description: "True if you want to display licensed content",
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
			"command":     "get-segment-details",
			"description": "Get Segment Details",
			"http": map[string]any{
				"method": "GET",
				"path":   "/discovery/v2/classifications/segments/{id}",
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
	pathParams["id"] = fmt.Sprintf("%v", v2GetSegmentDetailsFlags.id)

	req := &httpclient.Request{
		Method:      "GET",
		Path:        httpclient.SubstitutePath("/discovery/v2/classifications/segments/{id}", pathParams),
		QueryParams: map[string]string{},
		ArrayParams: map[string][]string{},
		Headers:     map[string]string{},
	}

	// Query parameters
	if cmd.Flags().Changed("locale") {
		req.QueryParams["locale"] = fmt.Sprintf("%v", v2GetSegmentDetailsFlags.locale)
	}
	if cmd.Flags().Changed("include-licensed-content") {
		req.QueryParams["includeLicensedContent"] = fmt.Sprintf("%v", v2GetSegmentDetailsFlags.includeLicensedContent)
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
