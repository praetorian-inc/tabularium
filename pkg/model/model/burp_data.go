package model

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
)

const BURP_COURIER_SOURCE = "burp-courier"

// BurpHTTPData represents the top-level structure of Burp Suite HTTP export data
type BurpHTTPData struct {
	Metadata BurpMetadata       `json:"metadata"`
	Data     map[string]Webpage `json:"data"` // Changed to map messageID to Webpage directly
}

// BurpRawRequestData represents the raw HTTP request structure from Burp Suite JSON for unmarshaling
type BurpRawRequestData struct {
	Body      string              `json:"body" desc:"The body of the request." example:"This is the body of the request."`
	MessageID int                 `json:"messageId" desc:"The message ID of the request." example:"123"`
	InScope   bool                `json:"inScope" desc:"Is the request in scope?" example:"true"`
	Method    string              `json:"method" desc:"The method of the request." example:"GET"`
	Path      string              `json:"path" desc:"The path of the request." example:"/api/login"`
	URL       string              `json:"url" desc:"The URL of the request." example:"https://example.com/api/login"`
	Headers   []map[string]string `json:"headers" desc:"The headers of the request."`
}

// BurpRawResponseData represents the raw HTTP response structure from Burp Suite JSON for unmarshaling
type BurpRawResponseData struct {
	Body       string              `json:"body" desc:"The body of the response." example:"This is the body of the response."`
	MessageID  int                 `json:"messageId" desc:"The message ID of the response." example:"123"`
	InScope    bool                `json:"inScope" desc:"Is the response in scope?" example:"true"`
	Method     string              `json:"method" desc:"The method of the response." example:"GET"`
	Path       string              `json:"path" desc:"The path of the response." example:"/api/login"`
	Headers    []map[string]string `json:"headers" desc:"The headers of the response."`
	StatusCode int                 `json:"statusCode" desc:"The status code of the response." example:"200"`
}

// BurpHTTPEntryRaw is used for JSON unmarshaling of the raw Burp data
type BurpHTTPEntryRaw struct {
	OriginalRequest                      *BurpRawRequestData  `json:"originalRequest"`
	OriginalResponse                     *BurpRawResponseData `json:"originalResponse"`
	ModifiedRequest                      *BurpRawRequestData  `json:"modifiedRequest"`
	ModifiedResponse                     *BurpRawResponseData `json:"modifiedResponse"`
	WasRequestIntercepted                bool                 `json:"wasRequestIntercepted" desc:"Was the request intercepted?" example:"true"`
	WasResponseIntercepted               bool                 `json:"wasResponseIntercepted" desc:"Was the response intercepted?" example:"true"`
	WasRequestModified                   bool                 `json:"wasRequestModified" desc:"Was the request modified?" example:"true"`
	WasResponseModified                  bool                 `json:"wasResponseModified" desc:"Was the response modified?" example:"true"`
	WasModifiedRequestBodyBase64Encoded  bool                 `json:"wasModifiedRequestBodyBase64Encoded" desc:"Was the request body base64 encoded?" example:"true"`
	WasModifiedResponseBodyBase64Encoded bool                 `json:"wasModifiedResponseBodyBase64Encoded" desc:"Was the response body base64 encoded?" example:"true"`
	WasRequestBodyBase64Encoded          bool                 `json:"wasRequestBodyBase64Encoded" desc:"Was the request body base64 encoded?" example:"true"`
	WasResponseBodyBase64Encoded         bool                 `json:"wasResponseBodyBase64Encoded" desc:"Was the response body base64 encoded?" example:"true"`
	ToolSource                           string               `json:"toolSource" desc:"The tool source of the entry." example:"Repeater"`
}

func createWebpage(burpEntry *BurpHTTPEntryRaw) (*Webpage, error) {
	parsedURL, err := url.Parse(burpEntry.OriginalRequest.URL)
	if err != nil {
		return nil, err
	}
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	parentWebApplication := NewWebApplication(baseURL, baseURL)

	var webpageRequests []WebpageRequest

	// Convert original request if present
	if burpEntry.OriginalRequest != nil {
		originalRequest, err := convertRawRequestToWebpageRequest(burpEntry.OriginalRequest, burpEntry.OriginalResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to convert original request: %w", err)
		}
		// Set Burp-specific metadata for original request
		originalRequest.WasIntercepted = burpEntry.WasRequestIntercepted
		originalRequest.WasModified = false

		// Set response interception/modification status if response exists
		if originalRequest.Response != nil {
			originalRequest.Response.WasIntercepted = burpEntry.WasResponseIntercepted
			originalRequest.Response.WasModified = false
		}

		webpageRequests = append(webpageRequests, originalRequest)
	}

	// Convert modified request if present
	if burpEntry.ModifiedRequest != nil {
		modifiedRequest, err := convertRawRequestToWebpageRequest(burpEntry.ModifiedRequest, burpEntry.ModifiedResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to convert modified request: %w", err)
		}
		// Set Burp-specific metadata for modified request
		modifiedRequest.WasIntercepted = burpEntry.WasRequestIntercepted
		modifiedRequest.WasModified = burpEntry.WasRequestModified // Use actual modification status

		// Set response interception/modification status if response exists
		if modifiedRequest.Response != nil {
			modifiedRequest.Response.WasIntercepted = burpEntry.WasResponseIntercepted
			modifiedRequest.Response.WasModified = burpEntry.WasResponseModified // Use actual modification status
		}

		webpageRequests = append(webpageRequests, modifiedRequest)
	}
	if len(webpageRequests) == 0 {
		slog.Error("wtfoo no requests found", "burpEntry", burpEntry)
	}
	// Create webpage with requests
	webpage := NewWebpage(*parsedURL, &parentWebApplication, WithRequests(webpageRequests...))

	// Set source to indicate this came from Burp
	webpage.Source = []string{BURP_COURIER_SOURCE}

	// Add Burp-specific metadata
	if webpage.Metadata == nil {
		webpage.Metadata = make(map[string]any)
	}
	webpage.Metadata["tool_source"] = burpEntry.ToolSource
	webpage.Metadata["was_request_intercepted"] = burpEntry.WasRequestIntercepted
	webpage.Metadata["was_response_intercepted"] = burpEntry.WasResponseIntercepted
	webpage.Metadata["was_request_modified"] = burpEntry.WasRequestModified
	webpage.Metadata["was_response_modified"] = burpEntry.WasResponseModified

	return &webpage, nil
}

func (e *BurpHTTPData) UnmarshalJSON(data []byte) error {
	// First unmarshal the raw structure to get metadata and data separately
	var rawData struct {
		Metadata BurpMetadata                `json:"metadata"`
		Data     map[string]BurpHTTPEntryRaw `json:"data"`
	}

	if err := json.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("failed to unmarshal raw burp data: %w", err)
	}

	// Set metadata
	e.Metadata = rawData.Metadata

	// Initialize the Data map
	e.Data = make(map[string]Webpage)

	// Convert each BurpHTTPEntryRaw to a Webpage
	for messageID, burpEntry := range rawData.Data {
		// Skip entries without original request
		if burpEntry.OriginalRequest == nil {
			continue
		}

		// Create webpage from the burp entry
		webpage, err := createWebpage(&burpEntry)
		if err != nil {
			return fmt.Errorf("failed to create webpage for message ID %s: %w", messageID, err)
		}

		// Store the webpage in the data map
		e.Data[messageID] = *webpage
	}

	return nil
}

// ToWebpageRequests converts BurpHTTPData to a slice of WebpageRequest objects
func (b *BurpHTTPData) ToWebpageRequests() ([]WebpageRequest, error) {
	var requests []WebpageRequest

	for _, webpage := range b.Data {
		// Extract all requests from the webpage
		requests = append(requests, webpage.Requests...)
	}

	return requests, nil
}

// ExtractBaseURLs returns a list of unique base URLs from the Burp data
func (b *BurpHTTPData) ExtractBaseURLs() []string {
	urlSet := make(map[string]struct{})

	for _, webpage := range b.Data {
		baseURL := ExtractBaseURL(webpage.URL)
		if baseURL != "" {
			urlSet[baseURL] = struct{}{}
		}
	}

	var urls []string
	for url := range urlSet {
		urls = append(urls, url)
	}

	return urls
}

// ExtractBaseURL extracts the base URL from a full URL
func ExtractBaseURL(fullURL string) string {
	url_parsed, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s://%s", url_parsed.Scheme, url_parsed.Host)
}

// ParseBurpHTTPData parses JSON bytes into BurpHTTPData
func ParseBurpHTTPData(data []byte) (*BurpHTTPData, error) {
	var burpData BurpHTTPData
	if err := json.Unmarshal(data, &burpData); err != nil {
		return nil, fmt.Errorf("failed to parse Burp HTTP data: %w", err)
	}
	return &burpData, nil
}

// ExtractToolSources extracts unique tool sources from Burp data
func (b *BurpHTTPData) ExtractToolSources() []string {
	toolSources := make(map[string]struct{})

	for _, webpage := range b.Data {
		if toolSource, exists := webpage.Metadata["tool_source"]; exists {
			if toolSourceStr, ok := toolSource.(string); ok && toolSourceStr != "" {
				toolSources[toolSourceStr] = struct{}{}
			}
		}
	}

	var sources []string
	for source := range toolSources {
		sources = append(sources, source)
	}

	return sources
}

// BurpIssuesData represents the top-level structure of Burp Suite Issues export data
type BurpIssuesData struct {
	Metadata BurpMetadata              `json:"metadata"`
	Data     map[string]BurpIssueEntry `json:"data"`
}

// BurpIssueEntry represents a single security issue found by Burp Suite
type BurpIssueEntry struct {
	BaseURL                  string                    `json:"baseUrl" desc:"The base URL of the issue." example:"https://example.com"`
	CollaboratorInteractions []CollaboratorInteraction `json:"collaboratorInteractions"`
	Confidence               string                    `json:"confidence" desc:"The confidence of the issue." example:"CERTAIN"`
	Severity                 string                    `json:"severity" desc:"The severity of the issue." example:"High"`
	Requests                 []BurpRawRequestData      `json:"requests" desc:"The requests of the issue."`
	Responses                []BurpRawResponseData     `json:"responses" desc:"The responses of the issue."`
	Name                     string                    `json:"name" desc:"The name of the issue." example:"SQL Injection"`
	Detail                   string                    `json:"detail" desc:"The detail of the issue." example:"SQL Injection vulnerability detected."`

	// Processed requests and responses using existing structures
	ProcessedRequests  []WebpageRequest  `json:"-"`
	ProcessedResponses []WebpageResponse `json:"-"`
}

// CollaboratorInteraction represents interactions with Burp Collaborator
type CollaboratorInteraction struct {
	Type        string `json:"type" desc:"The type of the collaborator interaction." example:"DNS"`
	Protocol    string `json:"protocol" desc:"The protocol of the collaborator interaction." example:"UDP"`
	LookupType  string `json:"lookupType" desc:"The lookup type of the collaborator interaction." example:"A"`
	Interaction string `json:"interaction" desc:"The interaction of the collaborator interaction." example:"test.collaborator.net"`
	RawDetail   string `json:"rawDetail" desc:"The raw detail of the collaborator interaction." example:"DNS lookup details"`
}

// UnmarshalJSON custom unmarshaling for BurpIssueEntry
func (e *BurpIssueEntry) UnmarshalJSON(data []byte) error {
	type _BurpIssueEntry BurpIssueEntry
	var raw _BurpIssueEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Copy simple fields
	e.BaseURL = raw.BaseURL
	e.CollaboratorInteractions = raw.CollaboratorInteractions
	e.Confidence = raw.Confidence
	e.Severity = raw.Severity
	e.Requests = raw.Requests
	e.Responses = raw.Responses
	e.Name = raw.Name
	e.Detail = raw.Detail

	// Convert requests to WebpageRequest format
	e.ProcessedRequests = make([]WebpageRequest, len(raw.Requests))
	for i, req := range raw.Requests {
		var resp *BurpRawResponseData
		if i < len(raw.Responses) {
			resp = &raw.Responses[i]
		}
		converted, err := convertRawRequestToWebpageRequest(&req, resp)
		if err != nil {
			return fmt.Errorf("failed to convert request %d: %w", i, err)
		}
		e.ProcessedRequests[i] = converted
	}

	// Convert responses to WebpageResponse format
	e.ProcessedResponses = make([]WebpageResponse, len(raw.Responses))
	for i, resp := range raw.Responses {
		e.ProcessedResponses[i] = convertRawResponseToWebpageResponse(&resp)
	}

	return nil
}

// convertRawRequestToWebpageRequest converts BurpRawRequestData to WebpageRequest (standalone function)
func convertRawRequestToWebpageRequest(rawRequest *BurpRawRequestData, rawResponse *BurpRawResponseData) (WebpageRequest, error) {
	// Use the URL field directly from the Burp data
	if rawRequest.URL == "" {
		return WebpageRequest{}, fmt.Errorf("URL field is empty")
	}

	// Convert headers format from Burp's []map[string]string to map[string][]string
	headers := make(map[string][]string)
	for _, header := range rawRequest.Headers {
		for key, value := range header {
			headers[key] = append(headers[key], value)
		}
	}

	webpageRequest := WebpageRequest{
		RawURL:  rawRequest.URL,
		Method:  rawRequest.Method,
		Headers: headers,
		Body:    rawRequest.Body,
	}

	// Add response if available
	if rawResponse != nil {
		webpageResponse := convertRawResponseToWebpageResponse(rawResponse)
		webpageRequest.Response = &webpageResponse
	}

	return webpageRequest, nil
}

func convertRawResponseToWebpageResponse(raw *BurpRawResponseData) WebpageResponse {
	headers := make(map[string][]string)
	for _, header := range raw.Headers {
		for key, value := range header {
			headers[key] = append(headers[key], value)
		}
	}

	statusCode := raw.StatusCode
	if statusCode == 0 {
		statusCode = 200 // fallback
	}

	return WebpageResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       raw.Body,
	}
}

// ParseBurpIssuesData parses JSON bytes into BurpIssuesData
func ParseBurpIssuesData(data []byte) (*BurpIssuesData, error) {
	var burpData BurpIssuesData
	if err := json.Unmarshal(data, &burpData); err != nil {
		return nil, fmt.Errorf("failed to parse Burp Issues data: %w", err)
	}
	return &burpData, nil
}

// ExtractBaseURLs returns a list of unique base URLs from the Burp issues data
func (b *BurpIssuesData) ExtractBaseURLs() []string {
	urlSet := make(map[string]struct{})

	for _, issue := range b.Data {
		if issue.BaseURL != "" {
			baseURL := ExtractBaseURL(issue.BaseURL)
			if baseURL != "" {
				urlSet[baseURL] = struct{}{}
			}
		}
	}

	var urls []string
	for url := range urlSet {
		urls = append(urls, url)
	}

	return urls
}

// GetIssuesByBaseURL groups issues by their base URL
func (b *BurpIssuesData) GetIssuesByURL() map[string][]BurpIssueEntry {
	result := make(map[string][]BurpIssueEntry)

	for _, issue := range b.Data {
		var baseURL string
		if len(issue.ProcessedRequests) == 0 {
			baseURL = ExtractBaseURL(issue.BaseURL)
		} else {
			baseURL = issue.ProcessedRequests[0].RawURL
		}
		if baseURL != "" {
			result[baseURL] = append(result[baseURL], issue)
		}
	}

	return result
}

// GetRiskSeverity maps Burp Suite severity levels to Chariot risk severity levels
func (e *BurpIssueEntry) GetRiskSeverity() string {
	switch strings.ToUpper(e.Severity) {
	case "HIGH":
		return "H" // High
	case "MEDIUM":
		return "M" // Medium
	case "LOW":
		return "L" // Low
	case "INFO", "INFORMATION", "INFORMATIONAL":
		return "I" // Info
	default:
		return "I" // Default to Info if severity is unknown or empty
	}
}

// HasCollaboratorInteractions returns true if the issue has collaborator interactions
func (e *BurpIssueEntry) HasCollaboratorInteractions() bool {
	return len(e.CollaboratorInteractions) > 0
}
