package model

type BurpAuthType string

const (
	FieldBasicAuth       BurpAuthType = "basic_authentication"
	FieldAPIKeyAuth      BurpAuthType = "api_key_authentication"
	FieldBearerTokenAuth BurpAuthType = "bearer_token_authentication"
)

type BurpAPIDestination string

const (
	DestinationHeader BurpAPIDestination = "header"
	DestinationQuery  BurpAPIDestination = "query"
	DestinationCookie BurpAPIDestination = "cookie"
)

type BurpAPIRequestMethod string

const (
	RequestMethodGet  BurpAPIRequestMethod = "get"
	RequestMethodPost BurpAPIRequestMethod = "post"
)

type BurpDastCredentials struct {
	BurpURL string `json:"url"`
	Token   string `json:"token"`
}

type EnabledEndpoint struct {
	ID string `json:"id"`
}

type FileBasedAPIDefinition struct {
	Filename         string            `json:"filename"`
	Contents         string            `json:"contents"`
	EnabledEndpoints []EnabledEndpoint `json:"enabled_endpoints"`
}

type URLBasedAPIDefinition struct {
	URL string `json:"url"`
}

type APIDefinitionResult struct {
	FileBasedDefinition *FileBasedAPIDefinition `json:"file_based_definition,omitempty"`
	URLBasedDefinition  *URLBasedAPIDefinition  `json:"url_based_definition,omitempty"`

	PrimaryURL string `json:"primary_url"`
}

func (r *APIDefinitionResult) HydratableFilepath() string {
	dummyApp := NewWebApplication(r.PrimaryURL, r.PrimaryURL)
	return dummyApp.HydratableFilepath()
}

// See here: https://portswigger.net/burp/extensibility/dast/graphql-api/apidefinitioninput.html
func (r *APIDefinitionResult) ToAPIDefinitionArray() []map[string]any {
	apiDef := make(map[string]any)

	if r.FileBasedDefinition != nil {
		fileBasedDef := map[string]any{
			"filename":          r.FileBasedDefinition.Filename,
			"contents":          r.FileBasedDefinition.Contents,
			"enabled_endpoints": r.FileBasedDefinition.EnabledEndpoints,
			"authentications":   []map[string]any{},
			"credentials":       []map[string]any{},
		}

		apiDef["file_based_api_definition"] = fileBasedDef
	} else if r.URLBasedDefinition != nil {
		urlBasedDef := map[string]any{
			"url":             r.URLBasedDefinition.URL,
			"authentications": []map[string]any{},
		}
		apiDef["url_based_api_definition"] = urlBasedDef
	}

	return []map[string]any{apiDef}
}
