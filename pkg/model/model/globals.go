package model

import (
	"encoding/gob"
	"time"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	// This is required for sending items to be processed in batch
	// Not added as a pointer, because aegiscli, which uses ingest client
	// uses another library which registers a pointer creating a clash
	// If that ever happens again, this will have to be redone as a separate
	// composite model
	gob.Register([]any{})
}

type Notification interface {
	Push(risk Risk)
	CreateTicket(risk Risk, templateID string) (Attribute, error)
	AssociateTicket(risk Risk, ticketID string) (Attribute, error)
	ValidateCredentials() (map[string]any, error)
}

type Export interface {
	ScheduledExport() error
}

// Filter returns a boolean to indicate whether processing of this entity should be stopped
// Methods of Filter may modify the input entity
type Filter interface {
	Asset(asset Assetlike) bool
	Risk(risk *Risk) bool
}

type HasKey interface {
	GetKey() string
}

// HasIP is an interface that can be used to get the IP address of the resource
// If the resource does not have an IP address, it will return an empty string
type HasIP interface {
	GetIPs() []string
}

// HasURL is an interface that can be used to get the URL of the resource
// If the resource does not have a URL, it will return an empty string
type HasURL interface {
	GetURLs() []string
}

// HasSecret is an interface that can be used to get the secret reference of the resource
// If the resource does not have a secret, it will return an empty string
type HasSecret interface {
	GetSecret() string
}

// TableModel is a noop interface that is used to make DynamoDB inserts type-safe
type TableModel interface {
	TableModel()
	registry.Model
}

type baseTableModel struct{}

func (baseTableModel) TableModel() {}

type GraphModel interface {
	registry.Model
	HasKey
	GetLabels() []string
	Valid() bool
}

type GraphRelationship interface {
	registry.Model
	Label() string
	Nodes() (source GraphModel, target GraphModel)
	Base() *BaseRelationship
	Visit(GraphRelationship)
	Valid() bool
}

type AgentSendable interface {
	SetUsername(string)
	GetAgent() string
	GraphModel
}

type Hydratable interface {
	Hydrate([]byte) error
	HydratableFilepath() string
	HydratedFile() File
	Dehydrate() Hydratable
}

const SKIP_HYDRATION = ""

type HydratableTarget interface {
	Hydratable
	Target
}

type Insertable interface {
	Merge(any)
	Visit(any) error
	SetStatus(string)
	GraphModel
}

const (
	TTLLabel = "TTL" // TTLLabel attach to models that utilize TTL cleanup in neo4j
)

const (
	// Global Flags
	FrozenFlag     string = "frozen"
	BurpModuleFlag string = "enable_burp_module"

	// seed:status
	Rejected string = "FR"

	// asset:status (in database)
	Deleted        string = "D"
	Pending        string = "P"
	Active         string = "A"
	Frozen         string = "F"
	FrozenRejected string = "FR"

	// asset:status (at runtime)
	ActiveLow     string = "AL"
	ActivePassive string = "AP"
	ActiveHigh    string = "AH"

	// job:status
	Queued         string = "JQ"
	Running        string = "JR"
	AsyncRunning   string = "JRA"
	AsyncReturning string = "JQA"
	Fail           string = "JF"
	Pass           string = "JP"

	// risk:state (not valid statuses alone)
	Triage     string = "T"
	Ignored    string = "I"
	Open       string = "O"
	Remediated string = "R"

	// risk:status
	TriageInfo     string = "TI"
	TriageLow      string = "TL"
	TriageMedium   string = "TM"
	TriageHigh     string = "TH"
	TriageCritical string = "TC"

	OpenExposure string = "OE"
	OpenInfo     string = "OI"
	OpenLow      string = "OL"
	OpenMedium   string = "OM"
	OpenHigh     string = "OH"
	OpenCritical string = "OC"

	AcceptedExposure string = "IE"
	AcceptedInfo     string = "II"
	AcceptedLow      string = "IL"
	AcceptedMedium   string = "IM"
	AcceptedHigh     string = "IH"
	AcceptedCritical string = "IC"

	RemediatedExposure string = "RE"
	RemediatedInfo     string = "RI"
	RemediatedLow      string = "RL"
	RemediatedMedium   string = "RM"
	RemediatedHigh     string = "RH"
	RemediatedCritical string = "RC"

	DeletedExposureFalsePositive string = "DEF"
	DeletedInfoFalsePositive     string = "DIF"
	DeletedLowFalsePositive      string = "DLF"
	DeletedMediumFalsePositive   string = "DMF"
	DeletedHighFalsePositive     string = "DHF"
	DeletedCriticalFalsePositive string = "DCF"

	DeletedExposureOutOfScope string = "DES"
	DeletedInfoOutOfScope     string = "DIS"
	DeletedLowOutOfScope      string = "DLS"
	DeletedMediumOutOfScope   string = "DMS"
	DeletedHighOutOfScope     string = "DHS"
	DeletedCriticalOutOfScope string = "DCS"

	DeletedExposureOther string = "DEO"
	DeletedInfoOther     string = "DIO"
	DeletedLowOther      string = "DLO"
	DeletedMediumOther   string = "DMO"
	DeletedHighOther     string = "DHO"
	DeletedCriticalOther string = "DCO"

	DeletedExposureDuplicate string = "DED"
	DeletedInfoDuplicate     string = "DID"
	DeletedLowDuplicate      string = "DLD"
	DeletedMediumDuplicate   string = "DMD"
	DeletedHighDuplicate     string = "DHD"
	DeletedCriticalDuplicate string = "DCD"

	// risk:source and asset:source
	ProvidedSource string = "provided"

	// asset:source
	SeedSource    string = "seed"
	AccountSource string = "account"
	SelfSource    string = "self"

	// job:queue
	Standard    string = "standard"
	Priority    string = "priority"
	Synchronous string = "synchronous"

	// praetorian-ai agent constants
	AffiliationAgentName string = "affiliation"
	AutoTriageAgentName  string = "autotriage"
	ScreenshotAgentName  string = "screenshotter"

	// system user for globally shared records
	SystemUser string = "global"

	// webpage:state
	Unanalyzed    string = "unanalyzed"
	Interesting   string = "interesting"
	Uninteresting string = "uninteresting"

	AllHooks string = "all"
)

var Intensity map[string]int = map[string]int{
	ActivePassive: 1,
	ActiveLow:     2,
	Active:        3,
	ActiveHigh:    4,
}

// RiskSeverity maps severity codes to their full text representations
var RiskSeverity = map[string]string{
	"I": "Info",
	"L": "Low",
	"M": "Medium",
	"H": "High",
	"C": "Critical",
	"E": "Exposure",
}

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func Future(hours int) int64 {
	return time.Now().UTC().Add(time.Duration(hours) * time.Hour).Unix()
}

var AgentDataTypes = map[string]map[string]bool{
	AutoTriageAgentName: {
		"risks": true,
	},
	AffiliationAgentName: {
		"assets": true,
		"risks":  true,
	},
}

var AgentClasses = map[string]map[string]bool{
	AffiliationAgentName: {
		"ipv4":   true,
		"ipv6":   true,
		"domain": true,
		"tld":    true,
	},
}

const LargeArtifactsUploadExpiration = 6 * 24 * time.Hour

const GenericPraetorianAegisInstallerMsi = "PraetorianAegisInstaller_generic.msi"
const GenericPraetorianAegisInstallerDeb = "PraetorianAegisInstaller_generic.deb"
const GenericPraetorianAegisInstallerRpm = "PraetorianAegisInstaller_generic.rpm"
