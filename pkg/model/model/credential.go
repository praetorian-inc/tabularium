package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type CredentialCategory string
type CredentialType string
type CredentialLifecycle string
type CredentialOperation string
type AdditionalCredParams map[string]any

const (
	// Credential Categories
	CategoryInternal    CredentialCategory = "internal"        // Internal credentials (db password, etc.)
	CategoryIntegration CredentialCategory = "env-integration" // Used to integrate and access client environment
	CategoryAdHoc       CredentialCategory = "ad-hoc"          // Found during in-flight execution

	// Credential Types
	StaticCredential      CredentialType = "static-credential" // un/pw pairs
	TokenCredential       CredentialType = "static-token"      // API key, PAT, etc.
	AWSCredential         CredentialType = "aws"
	GCloudCredential      CredentialType = "gcloud"
	AzureCredential       CredentialType = "azure"
	SSHKeyCredential      CredentialType = "ssh-key"         // SSH private key
	JSONCredential        CredentialType = "json-credential" // Generic JSON credential format
	AegisConfigCredential CredentialType = "aegis-config"    // Aegis config credential

	// API Keys
	ApolloCredential                CredentialType = "apollo_api" // Apollo.io API key
	AgentCredential                 CredentialType = "agent"
	MailgunCredentialGitHubPhishing CredentialType = "mailgun_phishing_github" // Mailgun API key for GitHub phishing
	AxoniousCredential              CredentialType = "axonious"                // Axonious API key
	BitbucketCredential             CredentialType = "bitbucket"               // Bitbucket API key
	BuiltWithCredential             CredentialType = "builtwith"               // Built with credentials
	CloudflareCredential            CredentialType = "cloudflare"              // Cloudflare API key
	DigitalOceanCredential          CredentialType = "digitalocean"            // DigitalOcean API key
	GithubCredential                CredentialType = "github"                  // Github API key
	GitlabCredential                CredentialType = "gitlab"                  // Gitlab API key
	InsightVMCredential             CredentialType = "insight-vm"              // InsightVM credentials
	MicrosoftDefenderCredential     CredentialType = "microsoft-defender"      // Microsoft Defender credentials
	CrowdStrikeSpotlightCredential  CredentialType = "crowdstrike-spotlight"   // CrowdStrike Spotlight credentials
	NessusCredential                CredentialType = "nessus"                  // Nessus API key
	NS1Credential                   CredentialType = "ns1"                     // NS1 API key
	BurpSuiteCredential             CredentialType = "portswigger"             // Customer BurpSuite API key
	BurpSuiteInternalCredential     CredentialType = "burp-internal"           // Internal BurpSuite instance API creds
	BurpSiteLoginCredential         CredentialType = "burp-site-login"         // Burp site login credentials
	BurpRecordingCredential         CredentialType = "burp-recording"          // Burp login recording scripts
	QualysCredential                CredentialType = "qualys"                  // Qualys credentials
	ShodanCredential                CredentialType = "shodan"                  // Shodan API key
	TenableVMCredential             CredentialType = "tenable-vm"              // Tenable API key
	WizCredential                   CredentialType = "wiz"                     // Wiz credentials
	WhoxyCredential                 CredentialType = "whoxy"                   // Whoxy API key
	XpanseCredential                CredentialType = "xpanse"                  // Xpanse credentials

	LegacyCloudCredential CredentialType = "legacy-cloud" // Legacy cloud credentials (AWS, GCP, Azure)

	// Credential Formats for capabilities to use (and advertise)
	CredentialFormatEnv   CredentialFormat = "env"   // things like tokens can be release into env vars for caps to use
	CredentialFormatFile  CredentialFormat = "file"  // credentials requested as files to be stored at a specific location
	CredentialFormatToken CredentialFormat = "token" // returned in a struct for direct use by caps

	// Credential Lifecycles
	CredentialLifecycleStatic    CredentialLifecycle = "static"
	CredentialLifecycleTemporary CredentialLifecycle = "temporary"

	// Credential Operations
	CredentialOperationGet    CredentialOperation = "get"
	CredentialOperationAdd    CredentialOperation = "add"
	CredentialOperationDelete CredentialOperation = "delete"
)

// LoginCredentialParams defines the structure for login credential parameters
type LoginCredentialParams struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Label      string `json:"label"`
	WebappKey  string `json:"webappKey"`
}

// RecordingCredentialParams defines the structure for recording credential parameters
type RecordingCredentialParams struct {
	Script    string `json:"script"`
	Label     string `json:"label"`
	WebappKey string `json:"webappKey"`
}

// LoginCredentialDeleteParams defines the structure for login credential deletion parameters
type LoginCredentialDeleteParams struct {
	LoginCredentialID string `json:"login_credential_id"`
}

// RecordingCredentialDeleteParams defines the structure for recording credential deletion parameters
type RecordingCredentialDeleteParams struct {
	RecordingID string `json:"recording_id"`
}

type CredentialRequest struct {
	Username     string               `json:"username"`
	CredentialID string               `json:"credentialId,omitempty"` // unique identifier for the credential
	ResourceKey  string               `json:"resourceKey,omitempty"`  // key of the resource to get the credential for
	Category     CredentialCategory   `json:"category"`               // internal, client integration, etc.
	Type         CredentialType       `json:"type"`                   // static, AWS, GCP, etc.
	Operation    CredentialOperation  `json:"operation"`              // get, add, delete operation type
	Format       []CredentialFormat   `json:"format"`                 // formats to get credentials in - file, env, token, etc.
	Parameters   AdditionalCredParams `json:"parameters,omitempty"`   // additional parameters to help with retrieval (if any)
}

type CredentialFile struct {
	CredentialFileContent  any    `json:"credentialFileContent,omitempty"`  // any allows it to be things like SSH key and can be a dictionary if needed too
	CredentialFileLocation string `json:"credentialFileLocation,omitempty"` // location to put the content in (eg. .azure/token.json, .ssh/key, etc.)
}

type CredentialResponse struct {
	CredentialCategory   CredentialCategory  `json:"credentialCategory"`
	CredentialType       CredentialType      `json:"credentialType"`
	CredentialLifecycle  CredentialLifecycle `json:"credentialLifecycle"`
	CredentialValue      map[string]string   `json:"credentialValue"`     // For tokens
	CredentialValueEnv   map[string]string   `json:"credentialValueEnv"`  // For environment variables (receiver can consider this a `.env` file and load it in context for Janus template runners)
	CredentialValueFiles []CredentialFile    `json:"credentialValueFile"` // For credential file (list of files that need to be placed at specific locations; Janus will place files in runners)
}

type Credential struct {
	registry.BaseModel
	Username     string             `neo4j:"username" json:"username" desc:"Username associated with the credential"`
	CredentialID string             `neo4j:"credentialId" json:"credentialId" desc:"UUID for the credential"`
	Key          string             `neo4j:"key" json:"key" desc:"Key of the credential (#credential#category#type#credentialId)"`
	AccountKey   string             `neo4j:"accountKey" json:"accountKey" desc:"Key of the associated account"`
	Category     CredentialCategory `neo4j:"category" json:"category" desc:"Category of the credential"`
	Type         CredentialType     `neo4j:"type" json:"type" desc:"Type of credential"`
	Created      string             `neo4j:"created" json:"created" desc:"Timestamp when the credential was created"`
	Updated      string             `neo4j:"updated" json:"updated" desc:"Timestamp when the credential was last updated"`
}

func (c *Credential) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				c.Key = fmt.Sprintf("#credential#%s#%s#%s", c.Category, c.Type, c.CredentialID)
				return nil
			},
		},
	}
}

func (c *Credential) Defaulted() {
	c.Created = Now()
	c.Updated = Now()
}

func NewCredential(accountKey string, category CredentialCategory, credType CredentialType) *Credential {
	credId := uuid.New().String()
	c := Credential{
		CredentialID: credId,
		AccountKey:   accountKey,
		Category:     category,
		Type:         credType,
	}
	c.Defaulted()
	registry.CallHooks(&c)
	return &c
}

func (c *Credential) GetDescription() string {
	return "Credential model for storing client credentials in the graph database"
}

func (c *Credential) GetKey() string {
	return c.Key
}

func (c *Credential) GetCredentialID() string {
	return c.CredentialID
}

func (c *Credential) GetLabels() []string {
	return []string{"Credential", string(c.Category), string(c.Type)}
}

func (c *Credential) Valid() bool {
	return c.CredentialID != "" && c.Key != "" && c.Category != "" && c.Type != ""
}

func init() {
	registry.Registry.MustRegisterModel(&Credential{})
}
