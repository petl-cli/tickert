package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rishimantri795/CLICreator/runtime/httpclient"
	"github.com/rishimantri795/CLICreator/runtime/output"
	"github.com/spf13/cobra"
)

var v2FindSuggestCmd = &cobra.Command{
	Use:   "find-suggest",
	Short: "Find Suggest",
	RunE:  withTelemetry(runV2FindSuggest),
}

var v2FindSuggestFlags struct {
	keyword                string
	source                 string
	latlong                string
	radius                 string
	unit                   string
	size                   string
	includeFuzzy           string
	clientVisibility       string
	countryCode            string
	includeTba             string
	includeTbd             string
	segmentId              string
	geoPoint               string
	locale                 string
	includeLicensedContent string
	includeSpellcheck      string
}

func init() {
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.keyword, "keyword", "", "Keyword to search on")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.source, "source", "", "Filter entities by its source name")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.latlong, "latlong", "", "Filter events by latitude and longitude, this filter is deprecated and maybe removed in a future release, please use geoPoint instead")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.radius, "radius", "", "Radius of the area in which we want to search for events.")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.unit, "unit", "", "Unit of the radius")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.size, "size", "", "Size of every entity returned in the response")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.includeFuzzy, "include-fuzzy", "", "yes, to include fuzzy matches in the search. This has performance impact.")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.clientVisibility, "client-visibility", "", "Filter events to clientName")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.countryCode, "country-code", "", "Filter suggestions by country code")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.includeTba, "include-tba", "", "True, to include events with date to be announce (TBA)")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.includeTbd, "include-tbd", "", "True, to include event with a date to be defined (TBD)")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.segmentId, "segment-id", "", "Filter suggestions by segment id")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.geoPoint, "geo-point", "", "filter events by geoHash")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.locale, "locale", "", "The locale in ISO code format. Multiple comma-separated values can be provided. When omitting the country part of the code (e.g. only 'en' or 'fr') then the first matching locale is used. When using a '*' it matches all locales. '*' can only be used at the end (e.g. 'en-us,en,*') ")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.includeLicensedContent, "include-licensed-content", "", "Yes if you want to display licensed content")
	v2FindSuggestCmd.Flags().StringVar(&v2FindSuggestFlags.includeSpellcheck, "include-spellcheck", "", "yes, to include spell check suggestions in the response.")

	v2Cmd.AddCommand(v2FindSuggestCmd)
}

func runV2FindSuggest(cmd *cobra.Command, args []string) error {
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
			Name:        "keyword",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Keyword to search on",
		})
		flags = append(flags, flagSchema{
			Name:        "source",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter entities by its source name",
		})
		flags = append(flags, flagSchema{
			Name:        "latlong",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by latitude and longitude, this filter is deprecated and maybe removed in a future release, please use geoPoint instead",
		})
		flags = append(flags, flagSchema{
			Name:        "radius",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Radius of the area in which we want to search for events.",
		})
		flags = append(flags, flagSchema{
			Name:        "unit",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Unit of the radius",
		})
		flags = append(flags, flagSchema{
			Name:        "size",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Size of every entity returned in the response",
		})
		flags = append(flags, flagSchema{
			Name:        "include-fuzzy",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "yes, to include fuzzy matches in the search. This has performance impact.",
		})
		flags = append(flags, flagSchema{
			Name:        "client-visibility",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events to clientName",
		})
		flags = append(flags, flagSchema{
			Name:        "country-code",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter suggestions by country code",
		})
		flags = append(flags, flagSchema{
			Name:        "include-tba",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "True, to include events with date to be announce (TBA)",
		})
		flags = append(flags, flagSchema{
			Name:        "include-tbd",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "True, to include event with a date to be defined (TBD)",
		})
		flags = append(flags, flagSchema{
			Name:        "segment-id",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter suggestions by segment id",
		})
		flags = append(flags, flagSchema{
			Name:        "geo-point",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "filter events by geoHash",
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
			ContentType: "application/hal+json; charset=utf-8",
			Description: "successful operation",
		})

		schema := map[string]any{
			"command":     "find-suggest",
			"description": "Find Suggest",
			"http": map[string]any{
				"method": "GET",
				"path":   "/discovery/v2/suggest",
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
		Path:        httpclient.SubstitutePath("/discovery/v2/suggest", pathParams),
		QueryParams: map[string]string{},
		ArrayParams: map[string][]string{},
		Headers:     map[string]string{},
	}

	// Query parameters
	if cmd.Flags().Changed("keyword") {
		req.QueryParams["keyword"] = fmt.Sprintf("%v", v2FindSuggestFlags.keyword)
	}
	if cmd.Flags().Changed("source") {
		req.QueryParams["source"] = fmt.Sprintf("%v", v2FindSuggestFlags.source)
	}
	if cmd.Flags().Changed("latlong") {
		req.QueryParams["latlong"] = fmt.Sprintf("%v", v2FindSuggestFlags.latlong)
	}
	if cmd.Flags().Changed("radius") {
		req.QueryParams["radius"] = fmt.Sprintf("%v", v2FindSuggestFlags.radius)
	}
	if cmd.Flags().Changed("unit") {
		req.QueryParams["unit"] = fmt.Sprintf("%v", v2FindSuggestFlags.unit)
	}
	if cmd.Flags().Changed("size") {
		req.QueryParams["size"] = fmt.Sprintf("%v", v2FindSuggestFlags.size)
	}
	if cmd.Flags().Changed("include-fuzzy") {
		req.QueryParams["includeFuzzy"] = fmt.Sprintf("%v", v2FindSuggestFlags.includeFuzzy)
	}
	if cmd.Flags().Changed("client-visibility") {
		req.QueryParams["clientVisibility"] = fmt.Sprintf("%v", v2FindSuggestFlags.clientVisibility)
	}
	if cmd.Flags().Changed("country-code") {
		req.QueryParams["countryCode"] = fmt.Sprintf("%v", v2FindSuggestFlags.countryCode)
	}
	if cmd.Flags().Changed("include-tba") {
		req.QueryParams["includeTBA"] = fmt.Sprintf("%v", v2FindSuggestFlags.includeTba)
	}
	if cmd.Flags().Changed("include-tbd") {
		req.QueryParams["includeTBD"] = fmt.Sprintf("%v", v2FindSuggestFlags.includeTbd)
	}
	if cmd.Flags().Changed("segment-id") {
		req.QueryParams["segmentId"] = fmt.Sprintf("%v", v2FindSuggestFlags.segmentId)
	}
	if cmd.Flags().Changed("geo-point") {
		req.QueryParams["geoPoint"] = fmt.Sprintf("%v", v2FindSuggestFlags.geoPoint)
	}
	if cmd.Flags().Changed("locale") {
		req.QueryParams["locale"] = fmt.Sprintf("%v", v2FindSuggestFlags.locale)
	}
	if cmd.Flags().Changed("include-licensed-content") {
		req.QueryParams["includeLicensedContent"] = fmt.Sprintf("%v", v2FindSuggestFlags.includeLicensedContent)
	}
	if cmd.Flags().Changed("include-spellcheck") {
		req.QueryParams["includeSpellcheck"] = fmt.Sprintf("%v", v2FindSuggestFlags.includeSpellcheck)
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
