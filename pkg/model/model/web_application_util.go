package model

// Authentication type constants - GraphQL __typename format
// These are what Burp returns in parse_api_definition and what CreateSite expects
const (
	AuthTypeBasic       = "ApiBasicAuthenticationWithoutCredentials"
	AuthTypeAPIKey      = "ApiKeyAuthenticationWithoutCredentials"
	AuthTypeBearerToken = "ApiBearerTokenAuthenticationWithoutCredentials"
	AuthTypeUnsupported = "ApiUnsupportedAuthenticationWithoutCredentials"
)

// Credential type constants - GraphQL __typename format
const (
	CredentialTypeBasic       = "ApiBasicAuthenticationCredentials"
	CredentialTypeAPIKey      = "ApiKeyAuthenticationCredentials"
	CredentialTypeBearerToken = "ApiBearerTokenAuthenticationCredentials"
)

// GraphQL field names for authentication union types
const (
	FieldBasicAuth       = "basic_authentication"
	FieldAPIKeyAuth      = "api_key_authentication"
	FieldBearerTokenAuth = "bearer_token_authentication"
)

// API destination constants - MUST BE LOWERCASE (Burp API requirement)
const (
	DestinationHeader = "header"
	DestinationQuery  = "query"
	DestinationCookie = "cookie"
)

// Request method constants - MUST BE LOWERCASE (Burp API requirement)
const (
	RequestMethodGet  = "get"
	RequestMethodPost = "post"
)

// GetFieldNameForType maps __typename to GraphQL field name
// This allows conversion from Burp parse response to GraphQL CreateSite format
func GetFieldNameForType(typename string) string {
	switch typename {
	case AuthTypeBasic, CredentialTypeBasic:
		return FieldBasicAuth
	case AuthTypeAPIKey, CredentialTypeAPIKey:
		return FieldAPIKeyAuth
	case AuthTypeBearerToken, CredentialTypeBearerToken:
		return FieldBearerTokenAuth
	default:
		return ""
	}
}

// IsCredentialType checks if typename represents credentials (with values)
// as opposed to authentication schemes (structure only)
func IsCredentialType(typename string) bool {
	return typename == CredentialTypeBasic ||
		typename == CredentialTypeAPIKey ||
		typename == CredentialTypeBearerToken
}

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

	PrimaryURL      string           `json:"primary_url"`
	Authentications []map[string]any `json:"authentications,omitempty"`
	Credentials     []map[string]any `json:"credentials,omitempty"`
}

func (r *APIDefinitionResult) ToAPIDefinitionArray() []map[string]any {
	apiDef := make(map[string]any)

	if r.FileBasedDefinition != nil {
		fileBasedDef := map[string]any{
			"filename":          r.FileBasedDefinition.Filename,
			"contents":          r.FileBasedDefinition.Contents,
			"enabled_endpoints": r.FileBasedDefinition.EnabledEndpoints,
		}
		if len(r.Authentications) > 0 {
			fileBasedDef["authentications"] = r.Authentications
		}
		if len(r.Credentials) > 0 {
			fileBasedDef["credentials"] = r.Credentials
		}
		apiDef["file_based_api_definition"] = fileBasedDef
	} else if r.URLBasedDefinition != nil {
		urlBasedDef := map[string]any{
			"url": r.URLBasedDefinition.URL,
		}
		if len(r.Authentications) > 0 {
			urlBasedDef["authentications"] = r.Authentications
		}
		if len(r.Credentials) > 0 {
			urlBasedDef["credentials"] = r.Credentials
		}
		apiDef["url_based_api_definition"] = urlBasedDef
	}

	return []map[string]any{apiDef}
}
