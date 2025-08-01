package model

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Risk is an instance of a vulnerability. A risk may be associated with multiple targets.
// Risks are grouped by the target's DNS value.
// Note this is not really the industry standard definition of risk and vulnerability.
// They are referred to as such in this context for historical reasons.
type Risk struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username" desc:"Chariot username associated with the risk." example:"user@example.com"`
	Key      string `neo4j:"key" json:"key" desc:"Unique key identifying the risk." example:"#risk#example.com#CVE-2023-12345"`
	// Attributes
	DNS        string `neo4j:"dns" json:"dns" desc:"Primary DNS or group associated with the risk." example:"example.com"`
	Name       string `neo4j:"name" json:"name" desc:"Name of the risk or vulnerability." example:"CVE-2023-12345"`
	Source     string `neo4j:"source" json:"source" desc:"Source that identified the risk." example:"nessus"`
	Status     string `neo4j:"status" json:"status" desc:"Current status of the risk (e.g., TH, OC, RM)." example:"TH"`
	Priority   int    `neo4j:"priority" json:"priority" desc:"Calculated priority score based on severity." example:"10"`
	Created    string `neo4j:"created" json:"created" desc:"Timestamp when the risk was first created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Updated    string `neo4j:"updated" json:"updated" desc:"Timestamp when the risk was last updated (RFC3339)." example:"2023-10-27T11:00:00Z"`
	Visited    string `neo4j:"visited" json:"visited" desc:"Timestamp when the risk was last visited or confirmed (RFC3339)." example:"2023-10-27T11:00:00Z"`
	TTL        int64  `neo4j:"ttl" json:"ttl" desc:"Time-to-live for the risk record (Unix timestamp)." example:"1706353200"`
	Comment    string `neo4j:"-" json:"comment,omitempty" desc:"User-provided comment about the risk." example:"Confirmed by manual check"`
	PlextracID string `neo4j:"plextracid" json:"plextracid" desc:"ID of the risk in PlexTrac." example:"#clientID#reportId#findingId"`
	Target     Target `neo4j:"-" json:"-"` // Internal use, not in schema
	History
	MLProperties
}

func init() {
	registry.Registry.MustRegisterModel(&Risk{})
	registry.Registry.MustRegisterModel(&MLProperties{})
	registry.Registry.MustRegisterModel(&RiskDefinition{})
}

// GetDescription returns a description for the Risk model.
func (r *Risk) GetDescription() string {
	return "Represents a security risk, typically linking an asset to a condition or vulnerability, including its status, severity, and associated metadata."
}

type MLProperties struct {
	registry.BaseModel
	Logit           *float32 `neo4j:"logit,omitempty" json:"logit,omitempty" desc:"Logit value from an ML model prediction." example:"0.75"`
	ProofSufficient *bool    `neo4j:"proofSufficient,omitempty" json:"proofSufficient,omitempty" desc:"Indicates if ML model considers proof sufficient for auto-triage." example:"true"`
	Agent           string   `neo4j:"-" json:"agent,omitempty" desc:"Name of the agent that provided the ML properties." example:"autotriage"`
}

// GetDescription returns a description for the MLProperties model.
func (mlp *MLProperties) GetDescription() string {
	return "Contains properties relevant to machine learning models, such as prediction scores and features."
}

type RiskDefinition struct {
	registry.BaseModel
	Description    string `json:"Description" desc:"Description of the risk or vulnerability." example:"This vulnerability allows..."`
	Impact         string `json:"Impact" desc:"Potential impact if the risk is exploited." example:"Remote code execution."`
	Recommendation string `json:"Recommendation" desc:"Recommended actions to mitigate the risk." example:"Apply vendor patch XYZ."`
	References     string `json:"References" desc:"Supporting references or links." example:"https://nvd.nist.gov/vuln/detail/CVE-2023-12345\nhttps://vendor.com/security/advisory"`
}

// GetDescription returns a description for the RiskDefinition model.
func (rd *RiskDefinition) GetDescription() string {
	return "Defines the static properties of a risk type, such as its name, description, and severity mappings."
}

var riskKey = regexp.MustCompile(`^#risk#([^#]+)#([^#]+)$`)

const RiskLabel = "Risk"

func (r *Risk) GetLabels() []string {
	return []string{RiskLabel, TTLLabel}
}

func (r *Risk) GetKey() string {
	return r.Key
}

func (r *Risk) Raw() string {
	rawJSON, _ := json.Marshal(r)
	return string(rawJSON)
}

func (r *Risk) Valid() bool {
	return riskKey.MatchString(r.Key) && r.Status != ""
}

func (r *Risk) Is(status string) bool {
	return strings.HasPrefix(r.Status, status)
}

func (r *Risk) Merge(update Risk) {
	if r.History.Update(r.Status, update.Status, update.Source, update.Comment, update.History) {
		r.setStatus(update.Status)
		r.Updated = Now()
	}
	if update.Created != "" {
		r.Created = update.Created
	}
	if !r.Is(Triage) {
		r.TTL = 0
	}
	if update.ProofSufficient != nil {
		r.ProofSufficient = update.ProofSufficient
	}
}

func (r *Risk) Visit(n Risk) {
	r.Visited = n.Visited

	if r.Is(Triage) {
		r.TTL = n.TTL
	}

	if r.Is(Remediated) {
		r.Set(Open)
	}

	r.Comment = n.Comment
}

func (r *Risk) SetSeverity(state string) {
	if len(r.Status) < 2 || len(state) < 2 {
		return
	}
	update := *r
	update.Status = r.Status[:1] + state[1:]
	r.Merge(update)
}

func (r *Risk) Set(state string) {
	update := *r
	update.setStatus(state + r.Severity()) // reset substate here
	r.Merge(update)
}

func (r *Risk) setStatus(status string) {
	r.Status = status
	r.Priority = riskPriority[r.Severity()]
}

func (r *Risk) Proof(bits []byte) File {
	file := NewFile(fmt.Sprintf("proofs/%s/%s", r.DNS, r.Name))
	file.Bytes = bits
	return file
}

func (r *Risk) Definition(definition RiskDefinition) File {
	file := NewFile(fmt.Sprintf("definitions/%s", r.Name))
	file.Overwrite = false

	body := ""
	if definition.Description != "" {
		body += fmt.Sprintf("#### Vulnerability Description\n%s\n", definition.Description)
	}
	if definition.Impact != "" {
		body += fmt.Sprintf("#### Impact\n%s\n", definition.Impact)
	}
	if definition.Recommendation != "" {
		body += fmt.Sprintf("#### Recommendation\n%s\n", definition.Recommendation)
	}
	if definition.References != "" {
		body += fmt.Sprintf("#### References\n%s\n", definition.References)
	}

	file.Bytes = []byte(body)
	return file
}

func (r *Risk) SubState() string {
	if len(r.Status) < 3 {
		return ""
	}
	return string(r.Status[2])
}

func (r *Risk) Severity() string {
	if len(r.Status) < 2 {
		return ""
	}
	return string(r.Status[1])
}

func (r *Risk) State() string {
	if len(r.Status) < 1 {
		return ""
	}
	return string(r.Status[0])
}

func (r *Risk) Link(username string) string {
	username = fmt.Sprintf("\"%s\"", username)
	return fmt.Sprintf("https://chariot.praetorian.com/%s/vulnerabilities?vulnerabilityDrawerKey=%s&drawerOrder=%%5B%%22vulnerability%%22%%5D",
		url.PathEscape(base64.StdEncoding.EncodeToString([]byte(username))),
		url.QueryEscape(r.Key))
}

func (r *Risk) Attribute(name, value string) Attribute {
	return NewAttribute(name, value, r)
}

func (r *Risk) PendingAsset() (Asset, bool) {
	// For use solely in VM Integrations in case we are not importing assets
	switch t := r.Target.(type) {
	case *Asset:
		asset := *t
		asset.Status = Pending
		return asset, true
	}
	return Asset{}, false
}

func (r *Risk) GetAgent() string {
	return r.MLProperties.Agent
}

func (r *Risk) SetUsername(username string) {
	r.Username = username
}

func GeneratePlexTracID(clientID, reportID, findingID string) string {
	return fmt.Sprintf("#%s#%s#%s", clientID, reportID, findingID)
}

func (r *Risk) SetPlexTracID(clientID, reportID, findingID string) {
	r.PlextracID = GeneratePlexTracID(clientID, reportID, findingID)
}

func NewRisk(target Target, name, status string) Risk {
	return NewRiskWithDNS(target, name, target.Group(), status)
}

func (r *Risk) Defaulted() {
	r.Source = ProvidedSource
	r.Created = Now()
	r.Updated = Now()
	r.Visited = Now()
	r.TTL = Future(14 * 24)
}

func (r *Risk) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				r.formatName()
				r.Key = fmt.Sprintf("#risk#%s#%s", r.DNS, r.Name)
				r.Priority = riskPriority[r.Severity()]
				return nil
			},
		},
	}
}

var cveRegex = regexp.MustCompile(`(?i)^cve-\d+-\d+$`)

func (r *Risk) formatName() {
	if cveRegex.MatchString(r.Name) {
		r.Name = strings.ToUpper(r.Name)
		return
	}

	r.Name = strings.ToLower(r.Name)
	r.Name = strings.ReplaceAll(r.Name, " ", "-")
}

func NewRiskWithDNS(target Target, name, dns, status string) Risk {
	r := Risk{
		DNS:    dns,
		Name:   name,
		Status: status,
		Target: target,
	}
	r.Defaulted()
	registry.CallHooks(&r)
	return r
}

var riskPriority = map[string]int{
	"C": 0,
	"H": 10,
	"M": 20,
	"L": 30,
	"I": 40,
	"E": 50,
	"":  60,
}
