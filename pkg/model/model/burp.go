package model

import (
	"fmt"
	"strings"
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

	PrimaryURL      string           `json:"primary_url"`
	Authentications []map[string]any `json:"authentications,omitempty"`
	Credentials     []map[string]any `json:"credentials,omitempty"`
}

func (r *APIDefinitionResult) GetAPIAuthenticationLabel() string {
	for _, auth := range r.Authentications {
		return auth["label"].(string)
	}
	return "Unknown"
}

func (r *APIDefinitionResult) ToGraphQLInput() ([]map[string]any, error) {
	if r.FileBasedDefinition == nil && r.URLBasedDefinition == nil {
		return nil, fmt.Errorf("no API definition found")
	}
	definitions := make([]map[string]any, 0, 1)

	apiDef := make(map[string]any)

	authentications := make(map[string]any, len(r.Authentications))

	if len(r.Authentications) > 0 {
		for _, auth := range r.Authentications {
			typ, ok := auth["type"].(string)
			if !ok {
				return nil, fmt.Errorf("no type found in authentication")
			}
			delete(auth, "type")
			typ = strings.ToLower(typ)
			if strings.HasPrefix(typ, "apibasic") {
				authentications["basic_authentication"] = auth
			}
			if strings.HasPrefix(typ, "apikey") {
				authentications["api_key_authentication"] = auth
			}
			if strings.HasPrefix(typ, "bearertoken") {
				authentications["bearer_token_authentication"] = auth
			}
		}
	}
	if len(r.Credentials) > 0 {
		for _, cred := range r.Credentials {
			delete(cred, "__typename")
		}
	}

	if r.FileBasedDefinition != nil {
		fileBasedDef := map[string]any{
			"filename":          r.FileBasedDefinition.Filename,
			"contents":          r.FileBasedDefinition.Contents,
			"enabled_endpoints": r.FileBasedDefinition.EnabledEndpoints,
		}
		if len(r.Authentications) > 0 {
			fileBasedDef["authentications"] = authentications
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
			urlBasedDef["authentications"] = authentications
		}
		if len(r.Credentials) > 0 {
			urlBasedDef["credentials"] = r.Credentials
		}
		apiDef["url_based_api_definition"] = urlBasedDef
	}

	definitions = append(definitions, apiDef)
	return definitions, nil
}
