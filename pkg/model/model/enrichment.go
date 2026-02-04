package model

import (
	"encoding/json"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const (
	PRIMARY   = "Primary"
	SECONDARY = "Secondary"
	BASE      = "Base"
	THREAT    = "Threat"
	TEMPORAL  = "Temporal"
)

func init() {
	registry.Registry.MustRegisterModel(&Enrichment{})
	registry.Registry.MustRegisterModel(&CvssMetrics{})
	registry.Registry.MustRegisterModel(&Epss{})
	registry.Registry.MustRegisterModel(&Ssvc{})
	registry.Registry.MustRegisterModel(&Weakness{})
	registry.Registry.MustRegisterModel(&MitreTechnique{})
	registry.Registry.MustRegisterModel(&Exploits{})
	registry.Registry.MustRegisterModel(&ThreatActor{})
	registry.Registry.MustRegisterModel(&ExploitCounts{})
	registry.Registry.MustRegisterModel(&ExploitTimeline{})
}

// GetDescription returns a description for the CvssMetrics model.
func (c *CvssMetrics) GetDescription() string {
	return "Represents Common Vulnerability Scoring System (CVSS) metrics for a vulnerability."
}

type CvssMetrics struct {
	registry.BaseModel
	Version             string   `json:"version" desc:"CVSS version (e.g., v2, v3.0, v3.1, v4.0)." example:"v3.1"`
	Type                string   `json:"type" desc:"Type of CVSS metric (Primary or Secondary)." example:"Primary"`
	MetricGroup         string   `json:"metric_group" desc:"Metric group (Base, Temporal, Threat)." example:"Base"`
	BaseVector          *string  `json:"base_vector,omitempty" desc:"CVSS base vector string." example:"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"`
	BaseScore           *float32 `json:"base_score,omitempty" desc:"CVSS base score." example:"9.8"`
	BaseSeverity        *string  `json:"base_severity,omitempty" desc:"CVSS base severity rating." example:"Critical"`
	ExploitabilityScore *float32 `json:"exploitability_score,omitempty" desc:"CVSS exploitability subscore." example:"3.9"`
	ImpactScore         *float32 `json:"impact_score,omitempty" desc:"CVSS impact subscore." example:"5.9"`
	TemporalVector      *string  `json:"temporal_vector,omitempty" desc:"CVSS temporal vector string." example:"E:P/RL:O/RC:C"`
	TemporalScore       *float32 `json:"temporal_score,omitempty" desc:"CVSS temporal score." example:"8.1"`
	ExploitMaturity     *string  `json:"exploit_maturity,omitempty" desc:"Exploit maturity level." example:"Proof-of-Concept"`
	ThreatScore         *float32 `json:"threat_score,omitempty" desc:"CVSS threat score (v4.0+)." example:"9.1"`
	ThreatSeverity      *string  `json:"threat_severity,omitempty" desc:"CVSS threat severity (v4.0+)." example:"High"`
}

// GetDescription returns a description for the Epss model.
func (e *Epss) GetDescription() string {
	return "Represents EPSS score and percentile for a vulnerability."
}

type Epss struct {
	registry.BaseModel
	Score      *float32 `json:"score,omitempty" desc:"EPSS score (probability of exploitation)." example:"0.95"`
	Percentile *float32 `json:"percentile,omitempty" desc:"EPSS percentile rank." example:"0.99"`
}

// GetDescription returns a description for the Ssvc model.
func (s *Ssvc) GetDescription() string {
	return "Represents SSVC assessment information for a vulnerability."
}

type Ssvc struct {
	registry.BaseModel
	Source          *string `json:"source,omitempty" desc:"Source of the SSVC assessment." example:"CISA"`
	Exploitation    *string `json:"exploitation,omitempty" desc:"SSVC exploitation status." example:"active"`
	TechnicalImpact *string `json:"technical_impact,omitempty" desc:"SSVC technical impact level." example:"total"`
	Automatable     *string `json:"automatable,omitempty" desc:"SSVC automatable status." example:"yes"`
}

// GetDescription returns a description for the Weakness model.
func (w *Weakness) GetDescription() string {
	return "Represents weakness information for a vulnerability."
}

type Weakness struct {
	registry.BaseModel
	Source *string `json:"source,omitempty" desc:"Source of the weakness information (e.g., NVD)." example:"NVD"`
	Type   *string `json:"type,omitempty" desc:"Type of weakness classification (Primary or Secondary)." example:"Primary"`
	Value  *string `json:"value,omitempty" desc:"Weakness identifier (e.g., CWE ID)." example:"CWE-79"`
	Name   *string `json:"name,omitempty" desc:"Name of the weakness." example:"Improper Neutralization of Input During Web Page Generation ('Cross-site Scripting')"`
	Url    *string `json:"url,omitempty" desc:"URL for more information about the weakness." example:"https://cwe.mitre.org/data/definitions/79.html"`
}

// GetDescription returns a description for the MitreTechnique model.
func (m *MitreTechnique) GetDescription() string {
	return "Represents a MITRE ATT&CK technique for a vulnerability."
}

type MitreTechnique struct {
	registry.BaseModel
	Id           *string   `json:"id,omitempty" desc:"MITRE ATT&CK technique ID." example:"T1566"`
	Name         *string   `json:"name,omitempty" desc:"MITRE ATT&CK technique name." example:"Phishing"`
	Domain       *string   `json:"domain,omitempty" desc:"MITRE ATT&CK domain." example:"enterprise-attack"`
	Subtechnique *bool     `json:"subtechnique,omitempty" desc:"Indicates if this is a sub-technique." example:"false"`
	Tactics      *[]string `json:"tactics,omitempty" desc:"List of MITRE ATT&CK tactics associated with the technique." example:"[\"initial-access\"]"`
	Url          *string   `json:"url,omitempty" desc:"URL for more information about the technique." example:"https://attack.mitre.org/techniques/T1566/"`
}

// GetDescription returns a description for the Exploits model.
func (e *Exploits) GetDescription() string {
	return "Represents exploit counts and timeline information for a vulnerability."
}

type Exploits struct {
	registry.BaseModel
	Counts   ExploitCounts   `json:"counts,omitempty" desc:"Counts of various exploit-related indicators."`
	Timeline ExploitTimeline `json:"timeline,omitempty" desc:"Timeline of exploit-related events."`
}

// GetDescription returns a description for the ThreatActor model.
func (t *ThreatActor) GetDescription() string {
	return "Represents a threat actor group for a vulnerability."
}

type ThreatActor struct {
	registry.BaseModel
	Name       string   `json:"name" desc:"Name of the threat actor group." example:"APT28"`
	Alias      []string `json:"aliases" desc:"Known aliases for the threat actor." example:"[\"Fancy Bear\", \"Sofacy Group\"]"`
	Country    *string  `json:"country,omitempty" desc:"Country associated with the threat actor." example:"RU"`
	Categories []string `json:"categories" desc:"Categories the threat actor falls into." example:"[\"State-Sponsored\", \"Espionage\"]"`
}

// GetDescription returns a description for the ExploitCounts model.
func (e *ExploitCounts) GetDescription() string {
	return "Represents counts of various exploit-related indicators for a vulnerability."
}

type ExploitCounts struct {
	registry.BaseModel
	Exploits           int `json:"exploits" desc:"Total number of known exploits." example:"5"`
	Botnets            int `json:"botnets" desc:"Number of botnets associated with exploits." example:"1"`
	RansomwareFamilies int `json:"ransomware_families" desc:"Number of ransomware families associated with exploits." example:"2"`
	ThreatActors       int `json:"threat_actors" desc:"Number of threat actors associated with exploits." example:"3"`
}

// GetDescription returns a description for the ExploitTimeline model.
func (e *ExploitTimeline) GetDescription() string {
	return "Represents timeline information for exploit-related events for a vulnerability."
}

type ExploitTimeline struct {
	registry.BaseModel
	CisaKevDateAdded                        *string `json:"cisa_kev_date_added,omitempty" desc:"Date added to CISA KEV catalog (RFC3339)." example:"2023-01-10T00:00:00Z"`
	CisaKevDateDue                          *string `json:"cisa_kev_date_due,omitempty" desc:"Due date for patching according to CISA KEV (RFC3339)." example:"2023-01-31T00:00:00Z"`
	FirstExploitPublished                   *string `json:"first_exploit_published,omitempty" desc:"Date the first exploit was published (RFC3339)." example:"2022-12-01T00:00:00Z"`
	FirstExploitPublishedWeaponizedOrHigher *string `json:"first_exploit_published_weaponized_or_higher,omitempty" desc:"Date the first weaponized (or higher) exploit was published (RFC3339)." example:"2022-12-15T00:00:00Z"`
	FirstReportedBotnet                     *string `json:"first_reported_botnet,omitempty" desc:"Date the first associated botnet was reported (RFC3339)." example:"2023-02-01T00:00:00Z"`
	FirstReportedRansomware                 *string `json:"first_reported_ransomware,omitempty" desc:"Date the first associated ransomware was reported (RFC3339)." example:"2023-01-20T00:00:00Z"`
	FirstReportedThreatActor                *string `json:"first_reported_threat_actor,omitempty" desc:"Date the first associated threat actor was reported (RFC3339)." example:"2023-01-05T00:00:00Z"`
	MostRecentExploitPublished              *string `json:"most_recent_exploit_published,omitempty" desc:"Date the most recent exploit was published (RFC3339)." example:"2023-03-01T00:00:00Z"`
	MostRecentReportedBotnet                *string `json:"most_recent_reported_botnet,omitempty" desc:"Date the most recent associated botnet was reported (RFC3339)." example:"2023-03-10T00:00:00Z"`
	MostRecentReportedRansomware            *string `json:"most_recent_reported_ransomware,omitempty" desc:"Date the most recent associated ransomware was reported (RFC3339)." example:"2023-02-20T00:00:00Z"`
	MostRecentReportedThreatActor           *string `json:"most_recent_reported_threat_actor,omitempty" desc:"Date the most recent associated threat actor was reported (RFC3339)." example:"2023-02-15T00:00:00Z"`
	NvdLastModified                         *string `json:"nvd_last_modified,omitempty" desc:"NVD last modified date (RFC3339)." example:"2023-04-01T10:00:00Z"`
	NvdPublished                            *string `json:"nvd_published,omitempty" desc:"NVD published date (RFC3339)." example:"2022-11-01T00:00:00Z"`
	VulncheckKevDateAdded                   *string `json:"vulncheck_kev_date_added,omitempty" desc:"Date added to VulnCheck KEV (RFC3339)." example:"2023-01-11T00:00:00Z"`
	VulncheckKevDateDue                     *string `json:"vulncheck_kev_date_due,omitempty" desc:"Due date for patching according to VulnCheck KEV (RFC3339)." example:"2023-02-01T00:00:00Z"`
}

// GetDescription returns a description for the Enrichment model.
func (e *Enrichment) GetDescription() string {
	return "Represents enrichment data for a vulnerability."
}

type Enrichment struct {
	registry.BaseModel
	Id              string           `json:"id" desc:"Unique identifier for the enrichment data (often CVE ID)." example:"CVE-2023-12345"`
	Name            string           `json:"name" desc:"Common name or title (e.g., from KEV)." example:"Microsoft Exchange Server Remote Code Execution Vulnerability"`
	IsKev           bool             `json:"is_kev" desc:"Indicates if the vulnerability is listed in the CISA KEV catalog." example:"true"`
	Description     string           `json:"description" desc:"Detailed description of the vulnerability or enrichment context." example:"A remote code execution vulnerability exists..."`
	Published       string           `json:"published" desc:"Date the vulnerability or enrichment data was published (RFC3339)." example:"2023-11-01T00:00:00Z"`
	Modified        string           `json:"modified" desc:"Date the vulnerability or enrichment data was last modified (RFC3339)." example:"2023-11-10T12:00:00Z"`
	Cvss            []CvssMetrics    `json:"cvss,omitempty" desc:"List of associated CVSS metrics."`
	Epss            *Epss            `json:"epss,omitempty" desc:"Associated EPSS score and percentile."`
	Ssvces          []Ssvc           `json:"ssvc,omitempty" desc:"List of associated SSVC assessments."`
	Weaknesses      []Weakness       `json:"weaknesses,omitempty" desc:"List of associated weaknesses (e.g., CWEs)."`
	MitreTechniques []MitreTechnique `json:"mitre_techniques,omitempty" desc:"List of associated MITRE ATT&CK techniques."`
	Exploits        *Exploits        `json:"exploits,omitempty" desc:"Exploit counts and timeline information."`
	ThreatActors    []ThreatActor    `json:"threat_actors,omitempty" desc:"List of associated threat actors."`
}

func (e *Enrichment) Vulnerability() Vulnerability {
	v := NewVulnerability(e.Id)
	v.Kev = e.IsKev
	v.Exploit = v.Kev || (e.Exploits != nil && e.Exploits.Counts.Exploits > 0)
	v.Title = &e.Name
	if v.Title == nil || *v.Title == "" {
		v.Title = &e.Id
	}

	if e.Published != "" {
		v.Created = &e.Published
	}
	if e.Modified != "" {
		v.Updated = &e.Modified
	}

	if data, err := json.Marshal(e); err == nil {
		dataStr := string(data)
		v.Data = &dataStr
	}

	if e.IsKev {
		feed := "cisa-kev"
		v.Feed = &feed

		if e.Exploits != nil && e.Exploits.Timeline.CisaKevDateAdded != nil {
			v.KevDateAdded = e.Exploits.Timeline.CisaKevDateAdded
		}
		if e.Exploits != nil && e.Exploits.Timeline.CisaKevDateDue != nil {
			v.KevDueDate = e.Exploits.Timeline.CisaKevDateDue
		}
	}

	version := ""
	for _, cvss := range e.Cvss {
		if cvss.Version < version || cvss.BaseScore == nil {
			continue
		}
		version = cvss.Version
		v.CVSS = cvss.BaseScore
	}

	if e.Epss != nil {
		v.EPSS = e.Epss.Score
	}

	return v
}
