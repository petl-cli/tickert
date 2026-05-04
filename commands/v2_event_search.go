package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rishimantri795/CLICreator/runtime/httpclient"
	"github.com/rishimantri795/CLICreator/runtime/output"
	"github.com/spf13/cobra"
)

var v2EventSearchCmd = &cobra.Command{
	Use:   "event-search",
	Short: "Event Search",
	RunE:  withTelemetry(runV2EventSearch),
}

var v2EventSearchFlags struct {
	sort                   string
	startDateTime          string
	endDateTime            string
	onsaleStartDateTime    string
	onsaleOnStartDate      string
	onsaleOnAfterStartDate string
	onsaleEndDateTime      string
	city                   string
	countryCode            string
	stateCode              string
	postalCode             string
	venueId                string
	attractionId           string
	segmentId              string
	segmentName            string
	classificationName     []string
	classificationId       []string
	marketId               string
	promoterId             string
	dmaId                  string
	includeTba             string
	includeTbd             string
	clientVisibility       string
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
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.sort, "sort", "", "Sorting order of the search result. Allowable values : 'name,asc', 'name,desc', 'date,asc', 'date,desc', 'relevance,asc', 'relevance,desc', 'distance,asc', 'name,date,asc', 'name,date,desc', 'date,name,asc', 'date,name,desc','onsaleStartDate,asc', 'id,asc'")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.startDateTime, "start-date-time", "", "Filter events with a start date after this date")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.endDateTime, "end-date-time", "", "Filter events with a start date before this date")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.onsaleStartDateTime, "onsale-start-date-time", "", "Filter events with onsale start date after this date")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.onsaleOnStartDate, "onsale-on-start-date", "", "Filter events with onsale start date on this date")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.onsaleOnAfterStartDate, "onsale-on-after-start-date", "", "Filter events with onsale range within this date")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.onsaleEndDateTime, "onsale-end-date-time", "", "Filter events with onsale end date before this date")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.city, "city", "", "Filter events by city")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.countryCode, "country-code", "", "Filter events by country code")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.stateCode, "state-code", "", "Filter events by state code")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.postalCode, "postal-code", "", "Filter events by postal code / zipcode")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.venueId, "venue-id", "", "Filter events by venue id")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.attractionId, "attraction-id", "", "Filter events by attraction id")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.segmentId, "segment-id", "", "Filter events by segment id")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.segmentName, "segment-name", "", "Filter events by segment name")
	v2EventSearchCmd.Flags().StringSliceVar(&v2EventSearchFlags.classificationName, "classification-name", nil, "Filter events by classification name: name of any segment, genre, sub-genre, type, sub-type")
	v2EventSearchCmd.Flags().StringSliceVar(&v2EventSearchFlags.classificationId, "classification-id", nil, "Filter events by classification id: id of any segment, genre, sub-genre, type, sub-type")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.marketId, "market-id", "", "Filter events by market id")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.promoterId, "promoter-id", "", "Filter events by promoter id")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.dmaId, "dma-id", "", "Filter events by dma id")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.includeTba, "include-tba", "", "True, to include events with date to be announce (TBA)")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.includeTbd, "include-tbd", "", "True, to include event with a date to be defined (TBD)")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.clientVisibility, "client-visibility", "", "Filter events by clientName")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.latlong, "latlong", "", "Filter events by latitude and longitude, this filter is deprecated and maybe removed in a future release, please use geoPoint instead")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.radius, "radius", "", "Radius of the area in which we want to search for events.")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.unit, "unit", "", "Unit of the radius")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.geoPoint, "geo-point", "", "filter events by geoHash")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.keyword, "keyword", "", "Keyword to search on")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.id, "id", "", "Filter entities by its id")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.source, "source", "", "Filter entities by its source name")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.includeTest, "include-test", "", "True if you want to have entities flag as test in the response. Only, if you only wanted test entities")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.page, "page", "", "Page number")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.size, "size", "", "Page size of the response")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.locale, "locale", "", "The locale in ISO code format. Multiple comma-separated values can be provided. When omitting the country part of the code (e.g. only 'en' or 'fr') then the first matching locale is used. When using a '*' it matches all locales. '*' can only be used at the end (e.g. 'en-us,en,*') ")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.includeLicensedContent, "include-licensed-content", "", "Yes if you want to display licensed content")
	v2EventSearchCmd.Flags().StringVar(&v2EventSearchFlags.includeSpellcheck, "include-spellcheck", "", "yes, to include spell check suggestions in the response.")

	v2Cmd.AddCommand(v2EventSearchCmd)
}

func runV2EventSearch(cmd *cobra.Command, args []string) error {
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
			Description: "Sorting order of the search result. Allowable values : 'name,asc', 'name,desc', 'date,asc', 'date,desc', 'relevance,asc', 'relevance,desc', 'distance,asc', 'name,date,asc', 'name,date,desc', 'date,name,asc', 'date,name,desc','onsaleStartDate,asc', 'id,asc'",
		})
		flags = append(flags, flagSchema{
			Name:        "start-date-time",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events with a start date after this date",
		})
		flags = append(flags, flagSchema{
			Name:        "end-date-time",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events with a start date before this date",
		})
		flags = append(flags, flagSchema{
			Name:        "onsale-start-date-time",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events with onsale start date after this date",
		})
		flags = append(flags, flagSchema{
			Name:        "onsale-on-start-date",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events with onsale start date on this date",
		})
		flags = append(flags, flagSchema{
			Name:        "onsale-on-after-start-date",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events with onsale range within this date",
		})
		flags = append(flags, flagSchema{
			Name:        "onsale-end-date-time",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events with onsale end date before this date",
		})
		flags = append(flags, flagSchema{
			Name:        "city",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by city",
		})
		flags = append(flags, flagSchema{
			Name:        "country-code",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by country code",
		})
		flags = append(flags, flagSchema{
			Name:        "state-code",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by state code",
		})
		flags = append(flags, flagSchema{
			Name:        "postal-code",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by postal code / zipcode",
		})
		flags = append(flags, flagSchema{
			Name:        "venue-id",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by venue id",
		})
		flags = append(flags, flagSchema{
			Name:        "attraction-id",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by attraction id",
		})
		flags = append(flags, flagSchema{
			Name:        "segment-id",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by segment id",
		})
		flags = append(flags, flagSchema{
			Name:        "segment-name",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by segment name",
		})
		flags = append(flags, flagSchema{
			Name:        "classification-name",
			Type:        "array",
			Required:    false,
			Location:    "query",
			Description: "Filter events by classification name: name of any segment, genre, sub-genre, type, sub-type",
		})
		flags = append(flags, flagSchema{
			Name:        "classification-id",
			Type:        "array",
			Required:    false,
			Location:    "query",
			Description: "Filter events by classification id: id of any segment, genre, sub-genre, type, sub-type",
		})
		flags = append(flags, flagSchema{
			Name:        "market-id",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by market id",
		})
		flags = append(flags, flagSchema{
			Name:        "promoter-id",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by promoter id",
		})
		flags = append(flags, flagSchema{
			Name:        "dma-id",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by dma id",
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
			Name:        "client-visibility",
			Type:        "string",
			Required:    false,
			Location:    "query",
			Description: "Filter events by clientName",
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
			"command":     "event-search",
			"description": "Event Search",
			"http": map[string]any{
				"method": "GET",
				"path":   "/discovery/v2/events",
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
		Path:        httpclient.SubstitutePath("/discovery/v2/events", pathParams),
		QueryParams: map[string]string{},
		ArrayParams: map[string][]string{},
		Headers:     map[string]string{},
	}

	// Query parameters
	if cmd.Flags().Changed("sort") {
		req.QueryParams["sort"] = fmt.Sprintf("%v", v2EventSearchFlags.sort)
	}
	if cmd.Flags().Changed("start-date-time") {
		req.QueryParams["startDateTime"] = fmt.Sprintf("%v", v2EventSearchFlags.startDateTime)
	}
	if cmd.Flags().Changed("end-date-time") {
		req.QueryParams["endDateTime"] = fmt.Sprintf("%v", v2EventSearchFlags.endDateTime)
	}
	if cmd.Flags().Changed("onsale-start-date-time") {
		req.QueryParams["onsaleStartDateTime"] = fmt.Sprintf("%v", v2EventSearchFlags.onsaleStartDateTime)
	}
	if cmd.Flags().Changed("onsale-on-start-date") {
		req.QueryParams["onsaleOnStartDate"] = fmt.Sprintf("%v", v2EventSearchFlags.onsaleOnStartDate)
	}
	if cmd.Flags().Changed("onsale-on-after-start-date") {
		req.QueryParams["onsaleOnAfterStartDate"] = fmt.Sprintf("%v", v2EventSearchFlags.onsaleOnAfterStartDate)
	}
	if cmd.Flags().Changed("onsale-end-date-time") {
		req.QueryParams["onsaleEndDateTime"] = fmt.Sprintf("%v", v2EventSearchFlags.onsaleEndDateTime)
	}
	if cmd.Flags().Changed("city") {
		req.QueryParams["city"] = fmt.Sprintf("%v", v2EventSearchFlags.city)
	}
	if cmd.Flags().Changed("country-code") {
		req.QueryParams["countryCode"] = fmt.Sprintf("%v", v2EventSearchFlags.countryCode)
	}
	if cmd.Flags().Changed("state-code") {
		req.QueryParams["stateCode"] = fmt.Sprintf("%v", v2EventSearchFlags.stateCode)
	}
	if cmd.Flags().Changed("postal-code") {
		req.QueryParams["postalCode"] = fmt.Sprintf("%v", v2EventSearchFlags.postalCode)
	}
	if cmd.Flags().Changed("venue-id") {
		req.QueryParams["venueId"] = fmt.Sprintf("%v", v2EventSearchFlags.venueId)
	}
	if cmd.Flags().Changed("attraction-id") {
		req.QueryParams["attractionId"] = fmt.Sprintf("%v", v2EventSearchFlags.attractionId)
	}
	if cmd.Flags().Changed("segment-id") {
		req.QueryParams["segmentId"] = fmt.Sprintf("%v", v2EventSearchFlags.segmentId)
	}
	if cmd.Flags().Changed("segment-name") {
		req.QueryParams["segmentName"] = fmt.Sprintf("%v", v2EventSearchFlags.segmentName)
	}
	if cmd.Flags().Changed("classification-name") {
		req.ArrayParams["classificationName"] = v2EventSearchFlags.classificationName
	}
	if cmd.Flags().Changed("classification-id") {
		req.ArrayParams["classificationId"] = v2EventSearchFlags.classificationId
	}
	if cmd.Flags().Changed("market-id") {
		req.QueryParams["marketId"] = fmt.Sprintf("%v", v2EventSearchFlags.marketId)
	}
	if cmd.Flags().Changed("promoter-id") {
		req.QueryParams["promoterId"] = fmt.Sprintf("%v", v2EventSearchFlags.promoterId)
	}
	if cmd.Flags().Changed("dma-id") {
		req.QueryParams["dmaId"] = fmt.Sprintf("%v", v2EventSearchFlags.dmaId)
	}
	if cmd.Flags().Changed("include-tba") {
		req.QueryParams["includeTBA"] = fmt.Sprintf("%v", v2EventSearchFlags.includeTba)
	}
	if cmd.Flags().Changed("include-tbd") {
		req.QueryParams["includeTBD"] = fmt.Sprintf("%v", v2EventSearchFlags.includeTbd)
	}
	if cmd.Flags().Changed("client-visibility") {
		req.QueryParams["clientVisibility"] = fmt.Sprintf("%v", v2EventSearchFlags.clientVisibility)
	}
	if cmd.Flags().Changed("latlong") {
		req.QueryParams["latlong"] = fmt.Sprintf("%v", v2EventSearchFlags.latlong)
	}
	if cmd.Flags().Changed("radius") {
		req.QueryParams["radius"] = fmt.Sprintf("%v", v2EventSearchFlags.radius)
	}
	if cmd.Flags().Changed("unit") {
		req.QueryParams["unit"] = fmt.Sprintf("%v", v2EventSearchFlags.unit)
	}
	if cmd.Flags().Changed("geo-point") {
		req.QueryParams["geoPoint"] = fmt.Sprintf("%v", v2EventSearchFlags.geoPoint)
	}
	if cmd.Flags().Changed("keyword") {
		req.QueryParams["keyword"] = fmt.Sprintf("%v", v2EventSearchFlags.keyword)
	}
	if cmd.Flags().Changed("id") {
		req.QueryParams["id"] = fmt.Sprintf("%v", v2EventSearchFlags.id)
	}
	if cmd.Flags().Changed("source") {
		req.QueryParams["source"] = fmt.Sprintf("%v", v2EventSearchFlags.source)
	}
	if cmd.Flags().Changed("include-test") {
		req.QueryParams["includeTest"] = fmt.Sprintf("%v", v2EventSearchFlags.includeTest)
	}
	if cmd.Flags().Changed("page") {
		req.QueryParams["page"] = fmt.Sprintf("%v", v2EventSearchFlags.page)
	}
	if cmd.Flags().Changed("size") {
		req.QueryParams["size"] = fmt.Sprintf("%v", v2EventSearchFlags.size)
	}
	if cmd.Flags().Changed("locale") {
		req.QueryParams["locale"] = fmt.Sprintf("%v", v2EventSearchFlags.locale)
	}
	if cmd.Flags().Changed("include-licensed-content") {
		req.QueryParams["includeLicensedContent"] = fmt.Sprintf("%v", v2EventSearchFlags.includeLicensedContent)
	}
	if cmd.Flags().Changed("include-spellcheck") {
		req.QueryParams["includeSpellcheck"] = fmt.Sprintf("%v", v2EventSearchFlags.includeSpellcheck)
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
