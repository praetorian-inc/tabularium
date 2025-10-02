package model

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// BurpHTTPData represents the top-level structure of Burp Suite HTTP export data
type BurpHTTPData struct {
	Metadata BurpMetadata             `json:"metadata"`
	Data     map[string]BurpHTTPEntry `json:"data"`
}

// BurpHTTPEntry represents a single HTTP request/response pair from Burp Suite
// Uses existing WebpageRequest and WebpageResponse types with Burp-specific metadata
type BurpHTTPEntry struct {
	// Use existing WebpageRequest and WebpageResponse types
	OriginalRequest  *WebpageRequest  `json:"-"`
	OriginalResponse *WebpageResponse `json:"-"`
	ModifiedRequest  *WebpageRequest  `json:"-"`
	ModifiedResponse *WebpageResponse `json:"-"`

	// Burp-specific fields stored directly in the entry
	RequestMessageID  int  `json:"-"`
	RequestInScope    bool `json:"-"`
	ResponseMessageID int  `json:"-"`
	ResponseInScope   bool `json:"-"`

	WasRequestIntercepted                bool   `json:"wasRequestIntercepted" desc:"Was the request intercepted?" example:"true"`
	WasResponseIntercepted               bool   `json:"wasResponseIntercepted" desc:"Was the response intercepted?" example:"true"`
	WasRequestModified                   bool   `json:"wasRequestModified" desc:"Was the request modified?" example:"true"`
	WasResponseModified                  bool   `json:"wasResponseModified" desc:"Was the response modified?" example:"true"`
	WasModifiedRequestBodyBase64Encoded  bool   `json:"wasModifiedRequestBodyBase64Encoded" desc:"Was the request body base64 encoded?" example:"true"`
	WasModifiedResponseBodyBase64Encoded bool   `json:"wasModifiedResponseBodyBase64Encoded" desc:"Was the response body base64 encoded?" example:"true"`
	WasRequestBodyBase64Encoded          bool   `json:"wasRequestBodyBase64Encoded" desc:"Was the request body base64 encoded?" example:"true"`
	WasResponseBodyBase64Encoded         bool   `json:"wasResponseBodyBase64Encoded" desc:"Was the response body base64 encoded?" example:"true"`
	ToolSource                           string `json:"toolSource" desc:"The tool source of the entry." example:"Repeater"`
	Note                                 string `json:"notes" desc:"The notes of the entry." example:"This is a note about the entry."`
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
	Body      string              `json:"body" desc:"The body of the response." example:"This is the body of the response."`
	MessageID int                 `json:"messageId" desc:"The message ID of the response." example:"123"`
	InScope   bool                `json:"inScope" desc:"Is the response in scope?" example:"true"`
	Method    string              `json:"method" desc:"The method of the response." example:"GET"`
	Path      string              `json:"path" desc:"The path of the response." example:"/api/login"`
	Headers   []map[string]string `json:"headers" desc:"The headers of the response."`
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

// UnmarshalJSON custom unmarshaling for BurpHTTPEntry
func (e *BurpHTTPEntry) UnmarshalJSON(data []byte) error {
	var raw BurpHTTPEntryRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Copy simple fields
	e.WasRequestIntercepted = raw.WasRequestIntercepted
	e.WasResponseIntercepted = raw.WasResponseIntercepted
	e.WasRequestModified = raw.WasRequestModified
	e.WasResponseModified = raw.WasResponseModified
	e.WasModifiedRequestBodyBase64Encoded = raw.WasModifiedRequestBodyBase64Encoded
	e.WasModifiedResponseBodyBase64Encoded = raw.WasModifiedResponseBodyBase64Encoded
	e.WasRequestBodyBase64Encoded = raw.WasRequestBodyBase64Encoded
	e.WasResponseBodyBase64Encoded = raw.WasResponseBodyBase64Encoded
	e.ToolSource = raw.ToolSource

	// Convert raw request to WebpageRequest
	if raw.OriginalRequest != nil {
		req, err := convertRawRequestToWebpageRequest(raw.OriginalRequest)
		if err != nil {
			return fmt.Errorf("failed to convert original request: %w", err)
		}
		e.OriginalRequest = &req
		e.RequestMessageID = raw.OriginalRequest.MessageID
		e.RequestInScope = raw.OriginalRequest.InScope
	}

	// Convert raw response to WebpageResponse
	if raw.OriginalResponse != nil {
		resp := convertRawResponseToWebpageResponse(raw.OriginalResponse)
		e.OriginalResponse = &resp
		e.ResponseMessageID = raw.OriginalResponse.MessageID
		e.ResponseInScope = raw.OriginalResponse.InScope
	}

	// Convert modified request if present
	if raw.ModifiedRequest != nil {
		req, err := convertRawRequestToWebpageRequest(raw.ModifiedRequest)
		if err != nil {
			return fmt.Errorf("failed to convert modified request: %w", err)
		}
		e.ModifiedRequest = &req
	}

	// Convert modified response if present
	if raw.ModifiedResponse != nil {
		resp := convertRawResponseToWebpageResponse(raw.ModifiedResponse)
		e.ModifiedResponse = &resp
	}

	return nil
}

// ToWebpageRequests converts BurpHTTPData to a slice of WebpageRequest objects
func (b *BurpHTTPData) ToWebpageRequests() ([]WebpageRequest, error) {
	var requests []WebpageRequest

	for _, entry := range b.Data {
		if entry.OriginalRequest == nil {
			continue
		}

		// The request is already converted during unmarshaling
		request := *entry.OriginalRequest

		// Add response if available
		if entry.OriginalResponse != nil {
			request.Response = entry.OriginalResponse
		}

		requests = append(requests, request)
	}

	return requests, nil
}

// ExtractBaseURLs returns a list of unique base URLs from the Burp data
func (b *BurpHTTPData) ExtractBaseURLs() []string {
	urlSet := make(map[string]struct{})

	for _, entry := range b.Data {
		if entry.OriginalRequest == nil {
			continue
		}

		baseURL := ExtractBaseURL(entry.OriginalRequest.RawURL)
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

	for _, entry := range b.Data {
		if entry.ToolSource != "" {
			toolSources[entry.ToolSource] = struct{}{}
		}
	}

	var sources []string
	for source := range toolSources {
		sources = append(sources, source)
	}

	return sources
}

// GetBurpMetadata returns Burp-specific metadata for the entry
func (e *BurpHTTPEntry) GetBurpMetadata() map[string]any {
	metadata := make(map[string]any)

	metadata["request_message_id"] = e.RequestMessageID
	metadata["request_in_scope"] = e.RequestInScope
	metadata["response_message_id"] = e.ResponseMessageID
	metadata["response_in_scope"] = e.ResponseInScope
	metadata["tool_source"] = e.ToolSource
	metadata["was_request_intercepted"] = e.WasRequestIntercepted
	metadata["was_response_intercepted"] = e.WasResponseIntercepted
	metadata["was_request_modified"] = e.WasRequestModified
	metadata["was_response_modified"] = e.WasResponseModified
	metadata["was_request_body_base64_encoded"] = e.WasRequestBodyBase64Encoded
	metadata["was_response_body_base64_encoded"] = e.WasResponseBodyBase64Encoded

	return metadata
}

// HasModifiedRequest returns true if the entry has a modified request
func (e *BurpHTTPEntry) HasModifiedRequest() bool {
	return e.ModifiedRequest != nil
}

// HasModifiedResponse returns true if the entry has a modified response
func (e *BurpHTTPEntry) HasModifiedResponse() bool {
	return e.ModifiedResponse != nil
}

// GetOriginalRequestMessageID returns the message ID of the original request
func (e *BurpHTTPEntry) GetOriginalRequestMessageID() int {
	return e.RequestMessageID
}

// GetOriginalResponseMessageID returns the message ID of the original response
func (e *BurpHTTPEntry) GetOriginalResponseMessageID() int {
	return e.ResponseMessageID
}

// IsInScope returns true if the original request was in scope
func (e *BurpHTTPEntry) IsInScope() bool {
	return e.RequestInScope
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
		converted, err := convertRawRequestToWebpageRequest(&req)
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
func convertRawRequestToWebpageRequest(raw *BurpRawRequestData) (WebpageRequest, error) {
	// Use the URL field directly from the Burp data
	if raw.URL == "" {
		return WebpageRequest{}, fmt.Errorf("URL field is empty")
	}

	// Convert headers format from Burp's []map[string]string to map[string][]string
	headers := make(map[string][]string)
	for _, header := range raw.Headers {
		for key, value := range header {
			headers[key] = append(headers[key], value)
		}
	}

	return WebpageRequest{
		RawURL:  raw.URL,
		Method:  raw.Method,
		Headers: headers,
		Body:    raw.Body,
	}, nil
}

// convertRawResponseToWebpageResponse converts BurpRawResponseData to WebpageResponse (standalone function)
func convertRawResponseToWebpageResponse(raw *BurpRawResponseData) WebpageResponse {
	// Convert headers format from Burp's []map[string]string to map[string][]string
	headers := make(map[string][]string)
	for _, header := range raw.Headers {
		for key, value := range header {
			headers[key] = append(headers[key], value)
		}
	}

	// Extract status code from headers or default to 200
	statusCode := 200
	// Note: Burp Suite data doesn't typically include status code directly in the provided format

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
