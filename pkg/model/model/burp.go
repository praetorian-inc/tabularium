package model

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

func (r *APIDefinitionResult) ToGraphQLInput() []map[string]any {
	definitions := make([]map[string]any, 0, 1)

	apiDef := make(map[string]any)

	if r.FileBasedDefinition != nil {
		apiDef["file_based_api_definition"] = map[string]any{
			"filename":          r.FileBasedDefinition.Filename,
			"contents":          r.FileBasedDefinition.Contents,
			"enabled_endpoints": r.FileBasedDefinition.EnabledEndpoints,
		}
	} else if r.URLBasedDefinition != nil {
		apiDef["url_based_api_definition"] = map[string]any{
			"url": r.URLBasedDefinition.URL,
		}
	}

	if len(r.Authentications) > 0 {
		apiDef["authentication_schemes_needed"] = r.Authentications
	}

	if len(r.Credentials) > 0 {
		apiDef["credentials"] = r.Credentials
	}

	definitions = append(definitions, apiDef)
	return definitions
}
