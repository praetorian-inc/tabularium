package model

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/lib/normalize"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

var webPageKeyRegex = regexp.MustCompile(`^#webpage#https?://.+(##.*)?$`)

const (
	DefaultMaxRequestsPerWebpage = 100
	ERR_PORT                     = -1
	DEFAULT_URL_PATH             = "/"
	DISPLAY_RESPONSE_FILE_PATH   = "file-path"
)

const (
	SSO_PROVIDER_OKTA    = "okta"
	SSO_PROVIDER_PINGONE = "pingone"
	SSO_PROVIDER_ENTRAID = "entraid"
)

type WebpageOption func(*Webpage) error

type WebpageForGob Webpage

type SSOWebpage struct {
	LastSeen            string `json:"last_seen" desc:"Timestamp when the webpage was last seen (RFC3339)." example:"2023-10-27T11:00:00Z"`
	Id                  string `json:"id" desc:"The ID of the webpage." example:"1234567890"`
	Name                string `json:"name" desc:"The webpage name." example:"Chariot"`
	OriginalProviderURL string `json:"original_provider_url" desc:"The original SSO provider URL before any redirects." example:"https://app.sso-provider.com/example"`
}

type WebpageCodeArtifact struct {
	Key    string `json:"key" desc:"The key of the ." example:"#file#source.zip"`
	Secret string `json:"secret" desc:"The secret id of the code artifact" example:"#file#source.zip"`
}

type GeneratorConfig struct {
	Type         string            `json:"type" neo4j:"type" desc:"Generator type (ollama, openai, rest)"`
	Endpoint     string            `json:"endpoint" neo4j:"endpoint" desc:"API endpoint URL"`
	APIKey       string            `json:"api_key,omitempty" neo4j:"api_key" desc:"API key for authentication"`
	Model        string            `json:"model,omitempty" neo4j:"model" desc:"Model name to use"`
	Method       string            `json:"method,omitempty" neo4j:"method" desc:"HTTP method (for REST generator)"`
	Headers      map[string]string `json:"headers,omitempty" neo4j:"headers" desc:"Custom HTTP headers"`
	Body         string            `json:"body,omitempty" neo4j:"body" desc:"Request body template"`
	ResponsePath string            `json:"response_path,omitempty" neo4j:"response_path" desc:"JSONPath to extract response"`
	ResponseType string            `json:"content_type,omitempty" neo4j:"content_type" desc:"Expected response content type"`
}

type EndpointFingerprint struct {
	Type             string            `json:"type,omitempty" neo4j:"type" desc:"Fingerprint type (llm, authentication, etc.)" example:"llm"`
	Component        string            `json:"component,omitempty" neo4j:"component" desc:"Detected component name for this specific endpoint" example:"okta"`
	Service          string            `json:"service,omitempty" neo4j:"service" desc:"Detected overall web application's service" example:"ollama"`
	GeneratorConfigs []GeneratorConfig `json:"generator_configs,omitempty" neo4j:"generator_configs" desc:"Augustus generator configurations from Julius fingerprinting"`
}

type Webpage struct {
	registry.BaseModel
	Username  string                `neo4j:"username" json:"username" desc:"The username associated with this webpage, if authenticated." example:"user@example.com"`
	Key       string                `neo4j:"key" json:"key" desc:"Unique key identifying the webpage." example:"#webpage#https://example.com#parentKey"`
	Created   string                `neo4j:"created" json:"created" desc:"Timestamp when the webpage was first discovered (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited   string                `neo4j:"visited" json:"visited" desc:"Timestamp when the webpage was last visited (RFC3339)." example:"2023-10-27T11:00:00Z"`
	TTL       int64                 `neo4j:"ttl" json:"ttl" desc:"Timestamp when the webpage will be deleted from the database in Unix seconds." example:"1747636791"`
	Status    string                `neo4j:"status" json:"status" desc:"Current status of the webpage (e.g., Active, Inactive)." example:"Active"`
	Source    []string              `neo4j:"source" json:"source" desc:"Sources that identified this webpage (e.g., seed, crawl)" example:"[\"crawl\", \"login\"]"`
	Artifacts []WebpageCodeArtifact `neo4j:"artifacts" json:"artifacts" desc:"Source code repositories or files for analysis (e.g., repositories, file keys)"`
	Private   bool                  `neo4j:"private" json:"private" desc:"Whether the webpage is on a public web server." example:"false"`
	History
	// Neo4j fields
	URL             string                `neo4j:"url" json:"url" desc:"The basic URL of the webpage." example:"https://example.com/path"`
	Metadata        map[string]any        `neo4j:"metadata" json:"metadata" dynamodbav:"metadata" desc:"Deprecated: Additional metadata associated with the webpage." example:"{\"title\": \"Example Domain\"}"`
	SSOIdentified   map[string]SSOWebpage `neo4j:"sso_identified" json:"sso_identified" desc:"SSO providers that have identified this webpage with their last seen timestamps." example:"{\"okta\": {\"last_seen\": \"2023-10-27T11:00:00Z\", \"id\": \"1234567890\", \"name\": \"Chariot\"}}"`
	DetailsFilepath string                `neo4j:"details_filepath" json:"details_filepath" dynamodbav:"details_filepath" desc:"The path to the details file for the webpage." example:"webpage/1234567890/details-1234567890.json"`
	Screenshot      string                `neo4j:"screenshot" json:"screenshot" desc:"Path to screenshot file" example:"webpage/example.com/443/screenshot.jpeg"`
	Resources       string                `neo4j:"resources" json:"resources" desc:"Path to network resources zip" example:"webpage/example.com/443/network_resources.zip"`
	EndpointFingerprint
	// S3 / Hydratable fields
	WebpageDetails
	// Not Saved but useful for internal processing
	Parent *WebApplication `neo4j:"-" json:"parent" desc:"The parent entity from which this webpage was discovered. Only used for creating a relationship. Pointer for easy reference"`
}

type WebpageDetails struct {
	Requests []WebpageRequest `json:"requests" desc:"A list of the HTTP requests under the webpage's URL. Limited to 100 requests."`
}

type WebpageRequest struct {
	OriginalURL    string              `json:"original_url" desc:"The original URL of the request before taking into account redirects like Location headers." example:"https://example.com/path?query=value"`
	RawURL         string              `json:"raw_url" desc:"The raw URL of the request after any redirects." example:"https://example.com/path?query=value"`
	Method         string              `json:"method" desc:"HTTP method used for the request (e.g., GET, POST)." example:"GET"`
	Headers        map[string][]string `json:"headers" desc:"Headers sent in the request." example:"{\"User-Agent\": [\"TabulariumCrawler/1.0\"]}"`
	Body           string              `json:"body" desc:"Body content of the request, if applicable." example:"{\"key\": \"value\"}"`
	WasIntercepted bool                `json:"was_intercepted" desc:"Whether the request was intercepted by a proxy." example:"false"`
	WasModified    bool                `json:"was_modified" desc:"Whether the request was modified by a proxy." example:"false"`
	Response       *WebpageResponse    `json:"response" desc:"Details of the HTTP response received from the webpage." example:"{\"status_code\": 200, \"headers\": {\"Content-Type\": [\"text/html\"]}, \"body\": \"<html><body>Example Domain</body></html>\"}"`
	Notes          string              `json:"notes" desc:"Notes about the request." example:"This is a note about the request."`
}

type WebpageResponse struct {
	StatusCode     int                 `json:"status_code" desc:"HTTP status code of the response." example:"200"`
	Headers        map[string][]string `json:"headers" desc:"Headers received in the response." example:"{\"Content-Type\": [\"text/html\"]}"`
	Body           string              `json:"body" desc:"Body content of the response." example:"<html><body>Example Domain</body></html>"`
	WasIntercepted bool                `json:"was_intercepted" desc:"Whether the response was intercepted by a proxy." example:"false"`
	WasModified    bool                `json:"was_modified" desc:"Whether the response was modified by a proxy." example:"false"`
	Notes          string              `json:"notes" desc:"Notes about the response." example:"This is a note about the response."`
}

const WebpageLabel = "Webpage"

func init() {
	registry.Registry.MustRegisterModel(&Webpage{})
}

func (w *Webpage) GetDescription() string {
	return "Represents a webpage, including its URL, status, and metadata."
}

func (w *Webpage) IsPrivate() bool {
	return w.Private
}

func (w *Webpage) GetKey() string {
	return w.Key
}

func (w *Webpage) GetLabels() []string {
	return []string{WebpageLabel, TTLLabel}
}

func (w *Webpage) Valid() bool {
	return webPageKeyRegex.MatchString(w.Key)
}

func (w *Webpage) SetUsername(username string) {
	w.Username = username
}

// Custom gob encoding to ensure WebpageDetails is always empty
func (w Webpage) GobEncode() ([]byte, error) {
	temp := WebpageForGob(w)
	temp.WebpageDetails = WebpageDetails{}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(temp)
	return buf.Bytes(), err
}

func (w *Webpage) GobDecode(data []byte) error {
	var temp WebpageForGob

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&temp); err != nil {
		return err
	}

	*w = Webpage(temp)
	return nil
}

func (w *Webpage) Visit(other Webpage) error {
	if w.TTL != 0 {
		w.TTL = other.TTL
	}
	w.Visited = Now()
	w.Merge(other)
	return nil
}

func (w *Webpage) PopulateResponses(refresh bool) {
	for i := range w.Requests {
		if w.Requests[i].Response != nil && !refresh {
			continue
		}
		w.PopulateResponse(&w.Requests[i])
	}
}

func (w *Webpage) AddRequest(req WebpageRequest) {
	w.Requests = append(w.Requests, req)
}

func (w *Webpage) Group() string {
	return fmt.Sprintf("%s://%s", w.Protocol(), w.Host())
}

func (w *Webpage) Identifier() string {
	return w.UrlPath()
}

func (w *Webpage) GetStatus() string {
	return w.Status
}

func (w *Webpage) IsStatus(value string) bool {
	return strings.HasPrefix(w.Status, value)
}

func (w *Webpage) IsClass(value string) bool {
	return strings.HasPrefix(w.URL, value)
}

func (w *Webpage) WithStatus(status string) Target {
	ret := *w
	ret.Status = status
	return &ret
}

func (w *Webpage) Merge(other Webpage) {
	if other.DetailsFilepath != "" {
		w.DetailsFilepath = other.DetailsFilepath
	}
	if other.Parent != nil {
		w.Parent = other.Parent
	}
	if other.Screenshot != "" {
		w.Screenshot = other.Screenshot
	}
	if other.Resources != "" {
		w.Resources = other.Resources
	}
	if other.Type != "" {
		w.Type = other.Type
	}
	if other.Component != "" {
		w.Component = other.Component
	}
	if other.Service != "" {
		w.Service = other.Service
	}
	if len(other.GeneratorConfigs) > 0 {
		w.GeneratorConfigs = other.GeneratorConfigs
	}
	w.MergeSSOIdentified(other)
	w.MergeMetadata(other)
	w.MergeSource(other)
	w.MergeArtifacts(other)
	w.MergeRequests(other.Requests...)
}

func (w *Webpage) HydratableFilepath() string {
	return w.DetailsFilePath()
}

func (w *Webpage) Hydrate(data []byte) error {
	return json.Unmarshal(data, &w)
}

func (w *Webpage) HydratedFile() File {
	if len(w.WebpageDetails.Requests) > DefaultMaxRequestsPerWebpage {
		w.WebpageDetails.Requests = w.WebpageDetails.Requests[:DefaultMaxRequestsPerWebpage]
	}

	detailsFile := w.GetDetailsFile(w.WebpageDetails)
	w.DetailsFilepath = detailsFile.Name

	return detailsFile
}

func (w *Webpage) Dehydrate() Hydratable {
	dehydratedWebpage := *w
	dehydratedWebpage.WebpageDetails = WebpageDetails{}
	return &dehydratedWebpage
}

func (w *Webpage) Defaulted() {
	w.Source = []string{}
	w.Artifacts = []WebpageCodeArtifact{}
	w.Status = Active
	w.Created = Now()
	w.Visited = Now()
	w.TTL = Future(30 * 24)
	w.Metadata = map[string]any{}
}

func (w *Webpage) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if w.URL != "" {
					key := fmt.Sprintf("#webpage#%s", w.URL)
					if len(key) > 2048 {
						key = key[:2048]
					}
					w.Key = key
				}
				return nil
			},
			Description: "Construction and basic analysis",
		},
		{
			Call: func() error {
				if w.Parent == nil {
					return nil
				}
				w.Private = w.Parent.IsPrivate()
				return nil
			},
			Description: "Determine if the webpage parent is a public server",
		},
	}
}

func NewWebpageFromString(urlString string, parent *WebApplication, options ...WebpageOption) Webpage {
	url, err := url.Parse(urlString)
	if err != nil {
		return Webpage{}
	}
	return NewWebpage(*url, parent, options...)
}

func NewWebpage(url url.URL, parent *WebApplication, options ...WebpageOption) Webpage {
	if url.Path == "" {
		url.Path = DEFAULT_URL_PATH
	}
	url = normalize.RemoveDefaultPorts(url)
	urlString := fmt.Sprintf("%s://%s%s", url.Scheme, url.Host, url.Path)
	w := Webpage{URL: urlString}
	w.Parent = parent
	w.Defaulted()
	// We run hooks twice to ensure construction and analysis are run
	registry.CallHooks(&w)

	for _, option := range options {
		option(&w)
	}

	registry.CallHooks(&w)
	return w
}

func (w *Webpage) CreateParent() *WebApplication {
	parsedURL, err := url.Parse(w.URL)
	if err != nil {
		return nil
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	webappName := baseURL
	if w.Service != "" {
		webappName = w.Service
	}
	webapp := NewWebApplication(baseURL, webappName)
	webapp.Status = Pending
	return &webapp
}

func (w *Webpage) AddSSOProvider(provider string, ssoData SSOWebpage) {
	if w.SSOIdentified == nil {
		w.SSOIdentified = make(map[string]SSOWebpage)
	}
	w.SSOIdentified[provider] = ssoData
}
