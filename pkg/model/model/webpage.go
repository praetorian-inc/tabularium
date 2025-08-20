package model

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

var webPageKeyRegex = regexp.MustCompile(`^#webpage#https?://.+(##.*)?$`)

const (
	DefaultMaxRequestsPerWebpage = 100
	ERR_PORT                     = -1
	DEFAULT_URL_PATH             = "/"
	PARAMETERS_IDENTIFIED        = "parameters-identified"
	WEB_LOGIN_IDENTIFIED         = "login-identified"
	WEB_SECRET_IDENTIFIED        = "web-secret-identified"
	DISPLAY_RESPONSE_FILE_PATH   = "file-path"
	SCREENSHOT                   = "screenshot"
	SC_RESOURCES                 = "resources"
)

type WebpageOption func(*Webpage) error

type WebpageForGob Webpage

// SSOIdentification represents SSO provider information
type SSOIdentification struct {
	LastSeen string `json:"lastSeen" desc:"Timestamp when the SSO provider was last seen (RFC3339)." example:"2023-10-27T11:00:00Z"`
}

type Webpage struct {
	registry.BaseModel
	Username string   `neo4j:"username" json:"username" desc:"The username associated with this webpage, if authenticated." example:"user@example.com"`
	Key      string   `neo4j:"key" json:"key" desc:"Unique key identifying the webpage." example:"#webpage#https://example.com#parentKey"`
	Created  string   `neo4j:"created" json:"created" desc:"Timestamp when the webpage was first discovered (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited  string   `neo4j:"visited" json:"visited" desc:"Timestamp when the webpage was last visited (RFC3339)." example:"2023-10-27T11:00:00Z"`
	TTL      int64    `neo4j:"ttl" json:"ttl" desc:"Timestamp when the webpage will be deleted from the database in Unix seconds." example:"1747636791"`
	Status   string   `neo4j:"status" json:"status" desc:"Current status of the webpage (e.g., Active, Inactive)." example:"Active"`
	Source   []string `neo4j:"source" json:"source" desc:"Sources that identified this webpage (e.g., seed, crawl)" example:"[\"crawl\", \"login\"]"`
	Private  bool     `neo4j:"private" json:"private" desc:"Whether the webpage is on a public web server." example:"false"`
	History
	// Neo4j fields
	URL             string                       `neo4j:"url" json:"url" desc:"The basic URL of the webpage." example:"https://example.com/path"`
	State           string                       `neo4j:"state" json:"state" desc:"Current analysis state of the webpage (e.g., Unanalyzed, Interesting, Uninteresting)." example:"Unanalyzed"`
	Metadata        map[string]any               `neo4j:"metadata" json:"metadata" dynamodbav:"metadata" desc:"Additional metadata associated with the webpage." example:"{\"title\": \"Example Domain\"}"`
	SSOIdentified   map[string]SSOIdentification `neo4j:"sso_identified" json:"sso_identified" desc:"SSO providers that have identified this webpage with their last seen timestamps." example:"{\"okta\": {\"lastSeen\": \"2023-10-27T11:00:00Z\"}}"`
	DetailsFilepath string                       `neo4j:"details_filepath" json:"details_filepath" dynamodbav:"details_filepath" desc:"The path to the details file for the webpage." example:"webpage/1234567890/details-1234567890.json"`
	// S3 fields
	WebpageDetails
	// Not Saved but useful for internal processing
	Parent GraphModelWrapper `neo4j:"-" json:"parent" desc:"The parent entity from which this webpage was discovered. Only used for creating a relationship"`
}

type WebpageDetails struct {
	Requests []WebpageRequest `json:"requests" desc:"A list of the HTTP requests under the webpage's URL. Limited to 100 requests."`
}

type WebpageRequest struct {
	RawURL   string              `json:"raw_url" desc:"The raw URL of the request." example:"https://example.com/path?query=value"`
	Method   string              `json:"method" desc:"HTTP method used for the request (e.g., GET, POST)." example:"GET"`
	Headers  map[string][]string `json:"headers" desc:"Headers sent in the request." example:"{\"User-Agent\": [\"TabulariumCrawler/1.0\"]}"`
	Body     string              `json:"body" desc:"Body content of the request, if applicable." example:"{\"key\": \"value\"}"`
	Response *WebpageResponse    `json:"response" desc:"Details of the HTTP response received from the webpage." example:"{\"status_code\": 200, \"headers\": {\"Content-Type\": [\"text/html\"]}, \"body\": \"<html><body>Example Domain</body></html>\"}"`
}

type WebpageResponse struct {
	StatusCode int                 `json:"status_code" desc:"HTTP status code of the response." example:"200"`
	Headers    map[string][]string `json:"headers" desc:"Headers received in the response." example:"{\"Content-Type\": [\"text/html\"]}"`
	Body       string              `json:"body" desc:"Body content of the response." example:"<html><body>Example Domain</body></html>"`
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

func (w *Webpage) GetAgent() string {
	return ScreenshotAgentName
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
	switch w.State {
	case Unanalyzed:
		w.State = other.State
	case Uninteresting:
		if other.State == Interesting {
			w.State = other.State
		}
	}
	w.MergeSSOIdentified(other)
	w.MergeMetadata(other)
	w.MergeSource(other)
	w.MergeRequests(other.Requests...)
}

func (w *Webpage) Hydrate() (path string, hydrate func([]byte) error) {
	hydrate = func(fileContents []byte) error {
		if err := json.Unmarshal(fileContents, &w); err != nil {
			return err
		}
		return nil
	}
	return w.DetailsFilepath, hydrate
}

func (w *Webpage) Dehydrate() (File, Hydratable) {
	dehydratedWebpage := *w
	if len(dehydratedWebpage.Requests) > DefaultMaxRequestsPerWebpage {
		dehydratedWebpage.Requests = dehydratedWebpage.Requests[:DefaultMaxRequestsPerWebpage]
	}

	detailsFile := w.GetDetailsFile(dehydratedWebpage.WebpageDetails)
	dehydratedWebpage.DetailsFilepath = detailsFile.Name

	dehydratedWebpage.WebpageDetails = WebpageDetails{}

	return detailsFile, &dehydratedWebpage
}

func (w *Webpage) Defaulted() {
	w.Source = []string{}
	w.Status = Active
	w.State = Unanalyzed
	w.Created = Now()
	w.Visited = Now()
	w.TTL = Future(7 * 24)
	w.SSOIdentified = map[string]SSOIdentification{}
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
				w.basicAnalysis()
				return nil
			},
			Description: "Construction and basic analysis",
		},
		{
			Call: func() error {
				if w.Parent.Model == nil {
					return nil
				}
				if target, ok := w.Parent.Model.(Target); ok {
					w.Private = target.IsPrivate()
				}
				return nil
			},
			Description: "Determine if the webpage parent is a public server",
		},
	}
}

func NewWebpageFromString(urlString string, parent GraphModel, options ...WebpageOption) Webpage {
	url, err := url.Parse(urlString)
	if err != nil {
		return Webpage{}
	}
	return NewWebpage(*url, parent, options...)
}

func NewWebpage(url url.URL, parent GraphModel, options ...WebpageOption) Webpage {
	if url.Path == "" {
		url.Path = DEFAULT_URL_PATH
	}
	urlString := fmt.Sprintf("%s://%s%s", url.Scheme, url.Host, url.Path)
	w := Webpage{URL: urlString, Parent: NewGraphModelWrapper(parent)}
	w.Defaulted()
	// We run hooks twice to ensure construction and analysis are run
	registry.CallHooks(&w)

	for _, option := range options {
		option(&w)
	}

	registry.CallHooks(&w)
	return w
}
