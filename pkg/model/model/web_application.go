package model

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"slices"

	"github.com/praetorian-inc/tabularium/pkg/lib/normalize"
	"github.com/praetorian-inc/tabularium/pkg/model/attacksurface"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type BurpMetadata struct {
	BurpType                 string   `neo4j:"burp_type" json:"burp_type" desc:"Burp type" example:"enterprise"`
	BurpSiteID               string   `neo4j:"burp_site_id" json:"burp_site_id" desc:"Burp Enterprise site identifier" example:"18865"`
	BurpFolderID             string   `neo4j:"burp_folder_id" json:"burp_folder_id" desc:"Burp Enterprise folder identifier" example:"17519"`
	BurpScheduleID           string   `neo4j:"burp_schedule_id" json:"burp_schedule_id" desc:"Burp Enterprise schedule identifier" example:"45934"`
	ApiDefinitionURL         string   `neo4j:"api_definition_url" json:"api_definition_url" desc:"URL to OpenAPI/Swagger specification" example:"https://api.example.com/openapi.json"`
	ApiDefinitionContentPath string   `neo4j:"api_definition_content_path" json:"api_definition_content_path" desc:"S3 path to API definition content for large files" example:"webapplication/user@example.com/api-definition-1234567890.json"`
	ExcludedExtensions       []string `neo4j:"excluded_extensions" json:"excluded_extensions" desc:"Excluded extensions" example:"[\"pdf\", \"doc\"]"`
	ScheduledInterval        int      `neo4j:"scheduledInterval" json:"scheduledInterval" desc:"Scheduled interval" example:"10"`
	MapType                  string   `neo4j:"mapType" json:"mapType" desc:"Map type" example:"proxy"`
	SizeThreshold            int      `neo4j:"sizeThreshold" json:"sizeThreshold" desc:"Size threshold" example:"1024"`
	AIEnabled                bool     `neo4j:"ai_enabled" json:"ai_enabled" desc:"AI enabled" example:"true"`
	ScopeEnabled             bool     `neo4j:"scope_enabled" json:"scope_enabled" desc:"Scope enabled" example:"true"`
	TimeUnit                 string   `neo4j:"timeUnit" json:"timeUnit" desc:"Time unit" example:"seconds"`
	TargetApplication        string   `neo4j:"target_application" json:"target_application" desc:"Target application" example:"https://example.com"`
}

// We wrap it so its still easy to marshal/unmarshal
type WebApplicationDetails struct {
	ApiDefinitionContent APIDefinitionResult `json:"api_definition_content" desc:"Full parsed content of the API definition file (OpenAPI/Swagger/Postman)"`
}

type WebApplicationForGob WebApplication

type WebApplication struct {
	BaseAsset
	LabelSettableEmbed
	PrimaryURL string   `neo4j:"primary_url" json:"primary_url" dynamodbav:"primary_url" desc:"The primary/canonical URL of the web application" example:"https://app.example.com"`
	URLs       []string `neo4j:"urls" json:"urls" dynamodbav:"urls" desc:"Additional URLs associated with this web application" example:"[\"https://api.example.com\", \"https://admin.example.com\"]"`
	Name       string   `neo4j:"name" json:"name" dynamodbav:"name" desc:"Name of the web application" example:"Example App"`
	BurpMetadata

	// S3-stored details (not saved to Neo4j/DynamoDB)
	WebApplicationDetails `neo4j:"-" json:"-" dynamodbav:"-"`
}

const WebApplicationLabel = "WebApplication"

var webAppKeyRegex = regexp.MustCompile(`^#webapplication#https?://[^?#]+$`)

func init() {
	MustRegisterLabel(WebApplicationLabel)
	registry.Registry.MustRegisterModel(&WebApplication{})
}

func (w *WebApplication) GetDescription() string {
	return "Represents a web application with a primary URL and associated URLs, designed for security testing and attack surface management."
}

func (w *WebApplication) GetLabels() []string {
	labels := []string{WebApplicationLabel, AssetLabel, TTLLabel}
	if w.Source == SeedSource {
		labels = append(labels, SeedLabel)
	}
	return labels
}

func (w *WebApplication) GetHooks() []registry.Hook {
	return []registry.Hook{
		useGroupAndIdentifier(w, &w.Name, &w.PrimaryURL),
		{
			Call: func() error {
				if w.PrimaryURL == "" {
					return fmt.Errorf("WebApplication requires non-empty PrimaryURL")
				}

				normalizedURL, err := normalize.Normalize(w.PrimaryURL)
				if err != nil {
					return fmt.Errorf("failed to normalize PrimaryURL: %w", err)
				}
				w.PrimaryURL = normalizedURL

				key := fmt.Sprintf("#webapplication#%s", w.PrimaryURL)
				if len(key) > 2048 {
					key = key[:2048]
				}
				w.Key = key

				normalizedURLs := make([]string, 0, len(w.URLs))
				for _, u := range w.URLs {
					if normalized, err := normalize.Normalize(u); err == nil {
						normalizedURLs = append(normalizedURLs, normalized)
					}
				}
				w.URLs = normalizedURLs

				return nil
			},
		},
		setGroupAndIdentifier(w, &w.Name, &w.PrimaryURL),
		{
			Call: func() error {
				if w.ExcludedExtensions == nil {
					w.ExcludedExtensions = []string{}
				}
				if !w.IsWebService() {
					w.BurpType = "webapplication"
				}
				return nil
			},
		},
	}
}

func (w *WebApplication) Defaulted() {
	w.BaseAsset.Defaulted()
	w.Class = "webapplication"
	w.AttackSurface = []string{string(attacksurface.Application)}
	if w.URLs == nil {
		w.URLs = []string{}
	}
}

func (w *WebApplication) Valid() bool {
	return webAppKeyRegex.MatchString(w.Key)
}

func (w *WebApplication) WithStatus(status string) Target {
	ret := *w
	ret.Status = status
	return &ret
}

func (w *WebApplication) GetPrimaryURL() url.URL {
	parsed, err := url.Parse(w.PrimaryURL)
	if err != nil {
		return url.URL{}
	}
	return *parsed
}

func (w *WebApplication) Group() string {
	return w.Name
}

func (w *WebApplication) Identifier() string {
	return w.PrimaryURL
}

func (w *WebApplication) IsSeed() bool {
	return slices.Contains(w.GetLabels(), SeedLabel)
}

func (w *WebApplication) IsPrivate() bool {
	return false
}

func (w *WebApplication) Merge(other Assetlike) {
	w.BaseAsset.Merge(other)
	otherApp, ok := other.(*WebApplication)
	if !ok {
		return
	}
	w.mergeDetails(otherApp)
	for _, u := range otherApp.URLs {
		if !slices.Contains(w.URLs, u) {
			w.URLs = append(w.URLs, u)
		}
	}
}

func (w *WebApplication) Visit(other Assetlike) {
	w.BaseAsset.Visit(other)
	otherApp, ok := other.(*WebApplication)
	if !ok {
		return
	}
	w.mergeDetails(otherApp)
}

func (w *WebApplication) mergeDetails(otherApp *WebApplication) {
	if w.Source != SeedSource && otherApp.Source == SeedSource {
		w.promoteToSeed()
	}

	if otherApp.Name != "" && (w.Name == w.PrimaryURL || otherApp.Name != otherApp.PrimaryURL) {
		w.Name = otherApp.Name
	}
	if otherApp.BurpSiteID != "" {
		w.BurpSiteID = otherApp.BurpSiteID
	}
	if otherApp.BurpFolderID != "" {
		w.BurpFolderID = otherApp.BurpFolderID
	}
	if otherApp.BurpScheduleID != "" {
		w.BurpScheduleID = otherApp.BurpScheduleID
	}
	if otherApp.ApiDefinitionContentPath != "" {
		w.ApiDefinitionContentPath = otherApp.ApiDefinitionContentPath
	}
}

func (w *WebApplication) promoteToSeed() {
	w.PendingLabelAddition = SeedLabel
	w.Source = SeedSource
}

func (w *WebApplication) Attribute(name, value string) Attribute {
	return NewAttribute(name, value, w)
}

func (w *WebApplication) GetHydratableFilepath() string {
	return fmt.Sprintf("webapplication/%s/api-definition.json", RemoveReservedCharacters(w.PrimaryURL))
}

func (w *WebApplication) HydratableFilepath() string {
	if !w.IsWebService() {
		return SKIP_HYDRATION
	}
	return w.GetHydratableFilepath()
}

func (w *WebApplication) Hydrate(data []byte) error {
	if len(data) == 0 {
		w.WebApplicationDetails = WebApplicationDetails{}
		return nil
	}

	w.ApiDefinitionContentPath = w.GetHydratableFilepath()
	if err := json.Unmarshal(data, &w.WebApplicationDetails.ApiDefinitionContent); err != nil {
		return fmt.Errorf("failed to hydrate WebApplication details: %w", err)
	}

	return nil
}

func (w *WebApplication) HydratedFile() File {
	bytes, err := json.Marshal(w.WebApplicationDetails.ApiDefinitionContent)
	if err != nil {
		slog.Error("failed to marshal WebApplicationDetails.ApiDefinitionContent", "error", err)
		bytes = []byte("{}")
	}

	filename := w.HydratableFilepath()
	detailsFile := NewFile(filename)
	detailsFile.Bytes = bytes

	w.ApiDefinitionContentPath = filename

	return detailsFile
}

func (w *WebApplication) Dehydrate() Hydratable {
	dehydratedApp := *w
	dehydratedApp.WebApplicationDetails = WebApplicationDetails{}
	return &dehydratedApp
}

func (w *WebApplication) IsWebService() bool {
	return w.BurpType == "webservice"
}

func NewWebApplication(primaryURL, name string) WebApplication {
	w := WebApplication{
		PrimaryURL: primaryURL,
		Name:       name,
		URLs:       []string{},
	}

	w.Defaulted()
	registry.CallHooks(&w)

	return w
}

func NewWebApplicationSeed(primaryURL string) WebApplication {
	w := NewWebApplication(primaryURL, primaryURL)
	w.Source = SeedSource
	w.Status = Pending
	w.TTL = 0
	return w
}

func (w *WebApplication) SeedModels() []Seedable {
	copy := *w
	return []Seedable{&copy}
}

func (w WebApplication) GobEncode() ([]byte, error) {
	temp := WebApplicationForGob(w)
	temp.WebApplicationDetails = WebApplicationDetails{}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(temp)
	return buf.Bytes(), err
}

func (w *WebApplication) GobDecode(data []byte) error {
	var temp WebApplicationForGob

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&temp); err != nil {
		return err
	}

	*w = WebApplication(temp)
	return nil
}
