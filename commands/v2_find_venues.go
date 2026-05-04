package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rishimantri795/CLICreator/runtime/httpclient"
	"github.com/rishimantri795/CLICreator/runtime/output"
	"github.com/spf13/cobra"
)

var v2FindVenuesCmd = &cobra.Command{
	Use:   "find-venues",
	Short: "Venue Search",
	RunE:  withTelemetry(runV2FindVenues),
}

var v2FindVenuesFlags struct {
	sort                   string
	stateCode              string
	countryCode            string
	latlong                string
	radius                 string
	unit                   string
	geoPoint               string
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
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.sort, "sort", "", "Sorting order of the search result. Allowable Values: 'name,asc', 'name,desc', 'relevance,asc', 'relevance,desc', 'distance,asc', 'distance,desc'")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.stateCode, "state-code", "", "Filter venues by state / province code")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.countryCode, "country-code", "", "Filter venues by country code")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.latlong, "latlong", "", "Filter events by latitude and longitude, this filter is deprecated and maybe removed in a future release, please use geoPoint instead")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.radius, "radius", "", "Radius of the area in which we want to search for events.")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.unit, "unit", "", "Unit of the radius")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.geoPoint, "geo-point", "", "filter events by geoHash")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.keyword, "keyword", "", "Keyword to search on")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.id, "id", "", "Filter entities by its id")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.source, "source", "", "Filter entities by its source name")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.includeTest, "include-test", "", "True if you want to have entities flag as test in the response. Only, if you only wanted test entities")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.page, "page", "", "Page number")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.size, "size", "", "Page size of the response")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.locale, "locale", "", "The locale in ISO code format. Multiple comma-separated values can be provided. When omitting the country part of the code (e.g. only 'en' or 'fr') then the first matching locale is used. When using a '*' it matches all locales. '*' can only be used at the end (e.g. 'en-us,en,*') ")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.includeLicensedContent, "include-licensed-content", "", "Yes if you want to display licensed content")
	v2FindVenuesCmd.Flags().StringVar(&v2FindVenuesFlags.includeSpellcheck, "include-spellcheck", "", "yes, to include spell check suggestions in the response.")

	v2Cmd.AddCommand(v2FindVenuesCmd)
}

func runV2FindVenues(cmd *cobra.Command, args []string) error {
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
			Description: "Sorting order of the search result. Allowable Values: 'name,asc', 'name,desc', 'relevance,asc', 'relevance,desc', 'distance,asc', 'distance,desc'",
		})
		flags = append(flags, flagSchema{
			Name:        "state-code",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter venues by state / province code",
		})
		flags = append(flags, flagSchema{
			Name:        "country-code",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter venues by country code",
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
			Name:        "geo-point",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "filter events by geoHash",
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
			"command":     "find-venues",
			"description": "Venue Search",
			"http": map[string]any{
				"method": "GET",
				"path":   "/discovery/v2/venues",
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
		Path:        httpclient.SubstitutePath("/discovery/v2/venues", pathParams),
		QueryParams: map[string]string{},
		ArrayParams: map[string][]string{},
		Headers:     map[string]string{},
	}

	// Query parameters
	if cmd.Flags().Changed("sort") {
		req.QueryParams["sort"] = fmt.Sprintf("%v", v2FindVenuesFlags.sort)
	}
	if cmd.Flags().Changed("state-code") {
		req.QueryParams["stateCode"] = fmt.Sprintf("%v", v2FindVenuesFlags.stateCode)
	}
	if cmd.Flags().Changed("country-code") {
		req.QueryParams["countryCode"] = fmt.Sprintf("%v", v2FindVenuesFlags.countryCode)
	}
	if cmd.Flags().Changed("latlong") {
		req.QueryParams["latlong"] = fmt.Sprintf("%v", v2FindVenuesFlags.latlong)
	}
	if cmd.Flags().Changed("radius") {
		req.QueryParams["radius"] = fmt.Sprintf("%v", v2FindVenuesFlags.radius)
	}
	if cmd.Flags().Changed("unit") {
		req.QueryParams["unit"] = fmt.Sprintf("%v", v2FindVenuesFlags.unit)
	}
	if cmd.Flags().Changed("geo-point") {
		req.QueryParams["geoPoint"] = fmt.Sprintf("%v", v2FindVenuesFlags.geoPoint)
	}
	if cmd.Flags().Changed("keyword") {
		req.QueryParams["keyword"] = fmt.Sprintf("%v", v2FindVenuesFlags.keyword)
	}
	if cmd.Flags().Changed("id") {
		req.QueryParams["id"] = fmt.Sprintf("%v", v2FindVenuesFlags.id)
	}
	if cmd.Flags().Changed("source") {
		req.QueryParams["source"] = fmt.Sprintf("%v", v2FindVenuesFlags.source)
	}
	if cmd.Flags().Changed("include-test") {
		req.QueryParams["includeTest"] = fmt.Sprintf("%v", v2FindVenuesFlags.includeTest)
	}
	if cmd.Flags().Changed("page") {
		req.QueryParams["page"] = fmt.Sprintf("%v", v2FindVenuesFlags.page)
	}
	if cmd.Flags().Changed("size") {
		req.QueryParams["size"] = fmt.Sprintf("%v", v2FindVenuesFlags.size)
	}
	if cmd.Flags().Changed("locale") {
		req.QueryParams["locale"] = fmt.Sprintf("%v", v2FindVenuesFlags.locale)
	}
	if cmd.Flags().Changed("include-licensed-content") {
		req.QueryParams["includeLicensedContent"] = fmt.Sprintf("%v", v2FindVenuesFlags.includeLicensedContent)
	}
	if cmd.Flags().Changed("include-spellcheck") {
		req.QueryParams["includeSpellcheck"] = fmt.Sprintf("%v", v2FindVenuesFlags.includeSpellcheck)
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
