package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// IAMPolicy represents an inline policy attached to an IAM principal.
// PolicyDocument is stored as raw JSON and persisted to S3, not Neo4j.
type IAMPolicy struct {
	PolicyName     string          `neo4j:"policyName" json:"policyName"`
	PolicyDocument json.RawMessage `json:"policyDocument"`
}

// IAMPolicyVersion represents a version of a managed IAM policy.
// Document is stored as raw JSON and persisted to S3, not Neo4j.
type IAMPolicyVersion struct {
	VersionId        string          `json:"versionId"`
	IsDefaultVersion bool            `json:"isDefaultVersion"`
	CreateDate       string          `json:"createDate"`
	Document         json.RawMessage `json:"document"`
}

const (
	AWSProvider      = "aws"
	AWSResourceLabel = "AWSResource"
)

type AWSResource struct {
	CloudResource

	OrgPolicy           []byte      `neo4j:"-" json:"orgPolicy"`
	HasOrgPolicy        bool        `neo4j:"hasOrgPolicy" json:"hasOrgPolicy"`
	IsManagementAccount bool        `neo4j:"isManagementAccount" json:"isManagementAccount"`
	InlinePolicies       []IAMPolicy      `neo4j:"-" json:"inlinePolicies"`
	HasInlinePolicies    bool             `neo4j:"hasInlinePolicies" json:"hasInlinePolicies"`
	TrustRelationship    json.RawMessage  `neo4j:"-" json:"trustRelationship"`
	HasTrustRelationship bool             `neo4j:"hasTrustRelationship" json:"hasTrustRelationship"`
	PolicyVersions       []IAMPolicyVersion `neo4j:"-" json:"policyVersions"`
	HasPolicyVersions    bool             `neo4j:"hasPolicyVersions" json:"hasPolicyVersions"`
	Tags                    []string           `neo4j:"tags" json:"tags,omitempty"`
	PermissionsBoundaryArn  string             `neo4j:"permissionsBoundaryArn" json:"permissionsBoundaryArn,omitempty"`
}

func init() {
	MustRegisterLabel(AWSResourceLabel)
	registry.Registry.MustRegisterModel(&AWSResource{})
}

// NewAWSResource creates an AWSResource from the given ARN, resource type, account reference, and properties.
func NewAWSResource(name, accountRef string, rtype CloudResourceType, properties map[string]any) (AWSResource, error) {
	parsedARN, err := arn.Parse(name)
	if err != nil {
		return AWSResource{}, fmt.Errorf("invalid ARN: %s", name)
	}

	r := AWSResource{
		CloudResource: CloudResource{
			Name:         name,
			DisplayName:  parsedARN.Resource,
			Provider:     AWSProvider,
			Properties:   properties,
			ResourceType: rtype,
			Region:       parsedARN.Region,
			AccountRef:   accountRef,
		},
	}

	r.Defaulted()
	registry.CallHooks(&r)
	return r, nil
}

func (a *AWSResource) Defaulted() {
	a.Origins = []string{"amazon"}
	a.AttackSurface = []string{"cloud"}
	a.CloudResource.Defaulted()
	a.BaseAsset.Defaulted()
}

func (a *AWSResource) GetHooks() []registry.Hook {
	hooks := []registry.Hook{
		useGroupAndIdentifier(a, &a.AccountRef, &a.Name),
		{
			Call: func() error {
				a.CloudResource.Key = fmt.Sprintf("#awsresource#%s#%s", a.AccountRef, a.Name)
				a.CloudResource.Labels = []string{AWSResourceLabel}
				a.CloudResource.IPs = a.GetIPs()
				a.CloudResource.URLs = a.GetURLs()

				a.IsManagementAccount = a.isManagementAccount()
				return nil
			},
		},
		setGroupAndIdentifier(a, &a.AccountRef, &a.Name),
	}

	hooks = append(hooks, a.CloudResource.GetHooks()...)
	return hooks
}

// WithStatus method for AWSResource to prevent type erasure
// Overrides the embedded CloudResource.WithStatus() to return proper type
func (a *AWSResource) WithStatus(status string) Target {
	ret := *a // Copy the full AWSResource, not just CloudResource
	ret.Status = status
	return &ret
}

func (c *AWSResource) SetOrgPolicy(policy []byte) {
	c.OrgPolicy = policy
	c.HasOrgPolicy = true
}

func (c *AWSResource) GetOrgPolicy() []byte {
	return c.OrgPolicy
}

func (c *AWSResource) SetInlinePolicies(policies []IAMPolicy) {
	c.InlinePolicies = policies
	c.HasInlinePolicies = len(policies) > 0
}

func (a *AWSResource) InlinePoliciesFilename() string {
	h := sha256.Sum256([]byte(a.Name))
	return fmt.Sprintf("awsresource/%s/inline-policies/%x.json", a.AccountRef, h[:8])
}

func (c *AWSResource) SetTrustRelationship(doc json.RawMessage) {
	c.TrustRelationship = doc
	c.HasTrustRelationship = len(doc) > 0
}

func (a *AWSResource) TrustRelationshipFilename() string {
	h := sha256.Sum256([]byte(a.Name))
	return fmt.Sprintf("awsresource/%s/trust-relationship/%x.json", a.AccountRef, h[:8])
}

func (c *AWSResource) SetPolicyVersions(versions []IAMPolicyVersion) {
	c.PolicyVersions = versions
	c.HasPolicyVersions = len(versions) > 0
}

func (a *AWSResource) PolicyVersionsFilename() string {
	h := sha256.Sum256([]byte(a.Name))
	return fmt.Sprintf("awsresource/%s/policy-versions/%x.json", a.AccountRef, h[:8])
}

func (c *AWSResource) CanHydrate() bool {
	return c.HasOrgPolicy || c.HasInlinePolicies || c.HasTrustRelationship || c.HasPolicyVersions
}

func (c *AWSResource) Hydrate(getFile func(string) ([]byte, error)) error {
	if c.HasOrgPolicy {
		data, err := getFile(c.OrgPolicyFilename())
		if err != nil {
			return err
		}
		if data == nil {
			return fmt.Errorf("no data")
		}
		c.SetOrgPolicy(data)
	}

	if c.HasInlinePolicies {
		data, err := getFile(c.InlinePoliciesFilename())
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &c.InlinePolicies); err != nil {
			return err
		}
	}

	if c.HasTrustRelationship {
		data, err := getFile(c.TrustRelationshipFilename())
		if err != nil {
			return err
		}
		c.TrustRelationship = data
	}

	if c.HasPolicyVersions {
		data, err := getFile(c.PolicyVersionsFilename())
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &c.PolicyVersions); err != nil {
			return err
		}
	}

	return nil
}

func (c *AWSResource) Dehydrate() ([]File, Hydratable) {
	var files []File
	dehydrated := *c

	if c.OrgPolicy != nil {
		file := NewFile(c.OrgPolicyFilename())
		file.Bytes = c.OrgPolicy
		files = append(files, file)
		dehydrated.OrgPolicy = nil
	}

	if len(c.InlinePolicies) > 0 {
		data, _ := json.Marshal(c.InlinePolicies)
		file := NewFile(c.InlinePoliciesFilename())
		file.Bytes = data
		files = append(files, file)
		dehydrated.InlinePolicies = nil
	}

	if len(c.TrustRelationship) > 0 {
		file := NewFile(c.TrustRelationshipFilename())
		file.Bytes = SmartBytes(c.TrustRelationship)
		files = append(files, file)
		dehydrated.TrustRelationship = nil
	}

	if len(c.PolicyVersions) > 0 {
		data, _ := json.Marshal(c.PolicyVersions)
		file := NewFile(c.PolicyVersionsFilename())
		file.Bytes = data
		files = append(files, file)
		dehydrated.PolicyVersions = nil
	}

	return files, &dehydrated
}

func (a *AWSResource) OrgPolicyFilename() string {
	return fmt.Sprintf("awsresource/%s/org-policies.json", a.AccountRef)
}

func (a *AWSResource) Visit(other Assetlike) {
	otherResource, ok := other.(*AWSResource)
	if !ok {
		return
	}

	if otherResource.HasOrgPolicy {
		a.HasOrgPolicy = otherResource.HasOrgPolicy
	}

	if otherResource.HasInlinePolicies {
		a.HasInlinePolicies = otherResource.HasInlinePolicies
	}

	if otherResource.HasTrustRelationship {
		a.HasTrustRelationship = otherResource.HasTrustRelationship
	}

	if otherResource.HasPolicyVersions {
		a.HasPolicyVersions = otherResource.HasPolicyVersions
	}

	if len(otherResource.Tags) > 0 {
		a.Tags = otherResource.Tags
	}

	if otherResource.PermissionsBoundaryArn != "" {
		a.PermissionsBoundaryArn = otherResource.PermissionsBoundaryArn
	}

	a.CloudResource.Visit(&otherResource.CloudResource)
	a.BaseAsset.Visit(otherResource)
}

func (a *AWSResource) Merge(other Assetlike) {
	otherResource, ok := other.(*AWSResource)
	if !ok {
		return
	}

	if otherResource.HasOrgPolicy {
		a.HasOrgPolicy = otherResource.HasOrgPolicy
	}

	if otherResource.HasInlinePolicies {
		a.HasInlinePolicies = otherResource.HasInlinePolicies
	}

	if otherResource.HasTrustRelationship {
		a.HasTrustRelationship = otherResource.HasTrustRelationship
	}

	if otherResource.HasPolicyVersions {
		a.HasPolicyVersions = otherResource.HasPolicyVersions
	}

	if len(otherResource.Tags) > 0 {
		a.Tags = otherResource.Tags
	}

	if otherResource.PermissionsBoundaryArn != "" {
		a.PermissionsBoundaryArn = otherResource.PermissionsBoundaryArn
	}

	a.CloudResource.Merge(&otherResource.CloudResource)
	a.BaseAsset.Merge(otherResource)
}

func (a *AWSResource) isManagementAccount() bool {
	if a.ResourceType != AWSOrganization {
		return false
	}

	// management account IDs are listed in the middle of Organization ARNs, and the actual account's ID is listed at the end
	// If they match, that means the actual account IS the management account
	// e.g., management account ARN: arn:aws:organizations::123456789012:account/o-b5qlad4a9o/123456789012
	//   non-management account ARN: arn:aws:organizations::123456789012:account/o-b5qlad4a9o/098765432109
	parts := strings.Split(a.Identifier(), ":")
	if len(parts) < 5 {
		return false
	}

	managementAccountID := parts[4]
	isManagementAccount := strings.HasSuffix(a.Identifier(), "/"+managementAccountID)

	return isManagementAccount
}

func (a *AWSResource) GetIPs() []string {
	ips := make([]string, 0) // Initialize with empty slice instead of nil
	switch a.ResourceType {
	case AWSEC2Instance:
		if ip, ok := a.Properties["PublicIp"].(string); ok && ip != "" {
			ips = append(ips, ip)
		}
		// Also check for PrivateIp in case it's needed for validation
		if ip, ok := a.Properties["PrivateIp"].(string); ok && ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

func (a *AWSResource) GetURLs() []string {
	return []string{}
}

func (a *AWSResource) GetDNS() string {
	if dns, ok := a.Properties["PublicDnsName"].(string); ok {
		return dns
	}
	return ""
}

func (a *AWSResource) Group() string {
	return a.AccountRef
}

func (a *AWSResource) Identifier() string {
	return a.Name
}

// Return an Asset that matches the legacy integration
func (a *AWSResource) NewAssets() []Asset {
	assets := make([]Asset, 0)
	dns := a.GetDNS()
	ips := a.GetIPs()
	urls := a.GetURLs()

	// Extract service name from ARN (same logic as Amazon capability)
	service := a.extractService()

	record := func(asset Asset) {
		asset.CloudId = a.Name
		asset.CloudService = service
		asset.CloudAccount = a.AccountRef
		assets = append(assets, asset)
	}

	// Create assets from URLs - NewAsset(url, arn)
	for _, url := range urls {
		record(NewAsset(url, a.Name))
	}

	for _, ip := range ips {
		if dns != "" {
			record(NewAsset(dns, ip))
		}
		record(NewAsset(ip, ip))
	}

	if len(assets) == 0 {
		identifier := a.Name // Use ARN as fallback
		if dns != "" {
			identifier = dns
		}
		record(NewAsset(identifier, identifier))
	}

	return assets
}

// extractService extracts the service name from the ARN (same logic as Amazon capability)
func (a *AWSResource) extractService() string {
	parts := strings.Split(a.Name, ":")
	if len(parts) > 2 {
		return parts[2]
	}
	return "Unknown Service"
}

func (a *AWSResource) IsPrivate() bool {
	return false
}
