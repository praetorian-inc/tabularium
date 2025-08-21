# Organization Name Asset Type - Implementation Guide

## Overview

The Organization Name asset type enables comprehensive asset discovery by supporting multiple name variations for organizations. This implementation addresses the challenge of inconsistent naming conventions across platforms like GitHub, DockerHub, and LinkedIn.

## Core Components

### 1. Organization Model (`pkg/model/model/organization.go`)

The main `Organization` struct extends `BaseAsset` and includes:

```go
type Organization struct {
    BaseAsset
    PrimaryName     string             // Canonical organization name
    Names           []OrganizationName // All name variations
    Industry        string             // Industry classification
    Country         string             // Primary country of operation
    Region          string             // Primary region
    StockTicker     string             // Stock symbol if public
    Website         string             // Primary website
    Description     string             // Brief description
}
```

### 2. OrganizationName Model

Each name variation is tracked with metadata:

```go
type OrganizationName struct {
    Name         string    // The organization name
    Type         string    // primary, legal, dba, abbreviation, etc.
    Status       string    // active, inactive, historic
    DateAdded    string    // When discovered (RFC3339)
    EffectiveDate string   // When name became effective
    EndDate      string    // When discontinued
    Source       string    // Where discovered
    Metadata     map[string]interface{} // Additional data
}
```

**Name Types:**
- `primary`: Canonical/preferred name
- `legal`: Legal entity name
- `dba`: "Doing business as" name
- `abbreviation`: Common abbreviation
- `common`: Commonly used variation
- `former`: Previous name
- `regional`: Geographic variation

## Relationship Models

### Parent-Subsidiary Relationships

```go
type OrganizationParentSubsidiary struct {
    *BaseRelationship
    OwnershipPercentage float64  // 0-100%
    EffectiveDate       string   // When relationship started
    EndDate             string   // When ended (if applicable)
    RelationshipType    string   // wholly_owned, majority_owned, etc.
    Jurisdiction        string   // Legal jurisdiction
    Source              string   // Information source
}
```

### Name History Tracking

```go
type OrganizationNameHistory struct {
    *BaseRelationship
    OldName        string // Previous name
    NewName        string // New name
    ChangeDate     string // When changed
    ChangeReason   string // Why changed
    FilingReference string // Legal filing reference
}
```

### Merger & Acquisition Tracking

```go
type OrganizationMerger struct {
    *BaseRelationship
    MergerDate        string  // Transaction date
    TransactionValue  float64 // Value in USD
    Currency          string  // Transaction currency
    TransactionType   string  // merger, acquisition, spin_off
    Status            string  // pending, completed, cancelled
    RegulatoryApproval string // Approval status
}
```

## Search Expansion Service

### Core Functionality

The `OrganizationSearchExpansion` service provides:

1. **Name Normalization**: Consistent key generation
2. **Search Expansion**: Returns all name variations for a query
3. **Organization Discovery**: Find organizations by any name

```go
// Example usage
service := NewOrganizationSearchExpansion()
service.AddOrganization(&organization)

// Input: "Praetorian"
// Output: ["Praetorian", "Praetorian Inc", "Praetorian Security Inc", "Praetorian Labs"]
expansions := service.ExpandSearch("Praetorian")
```

### Name Normalization Algorithm

```go
func NormalizeOrganizationName(name string) string {
    // 1. Convert to lowercase
    // 2. Remove common suffixes (Inc, Corp, LLC, etc.)
    // 3. Remove special characters
    // 4. Collapse whitespace
    return normalized
}
```

## Integration Demo

### Discovery Service

The `OrganizationDiscoveryService` demonstrates real-world usage:

```go
service := NewOrganizationDiscoveryService()
service.SetupSampleData()

// Discover assets across platforms
results, err := service.DiscoverAssets("Praetorian")

// Generate comprehensive report
report, err := service.GenerateDiscoveryReport("Praetorian")
```

### Platform Integration

Mock platform searchers simulate:
- **GitHub**: Repository and organization discovery
- **DockerHub**: Container image discovery
- **LinkedIn**: Company page discovery

## API Usage Examples

### Creating Organizations

```go
// Create basic organization
org := NewOrganization("Walmart")

// Add name variations
org.AddName("Walmart Inc", NameTypeLegal, "sec_filings")
org.AddName("WMT", NameTypeAbbreviation, "stock_ticker")
org.AddName("Walmart Stores Inc", NameTypeFormer, "historical")

// Set additional properties
org.Industry = "Retail"
org.StockTicker = "WMT"
org.Website = "https://www.walmart.com"
```

### Managing Relationships

```go
// Create parent-subsidiary relationship
parent := NewOrganization("Walmart Inc")
subsidiary := NewOrganization("Sam's Club")

relationship := NewOrganizationParentSubsidiary(
    &parent, &subsidiary, 
    100.0, RelationshipTypeWhollyOwned
)

// Track name history
nameHistory := NewOrganizationNameHistory(
    &org, 
    "Wal-Mart Stores Inc", 
    "Walmart Inc", 
    "2018-02-01T00:00:00Z"
)
nameHistory.ChangeReason = "Corporate rebranding"
```

### Search Operations

```go
// Query OrganizationName nodes via relationships in Neo4j
// MATCH (org:Organization)-[:HAS_ORGANIZATION_NAME]->(name:OrganizationName {state: 'active'})
// WHERE org.key = $orgKey RETURN name

// Search expansion
expansion := searchService.ExpandSearch("Praetorian")

// Find organization
foundOrg := searchService.FindOrganization("Praetorian Labs")
```

## Validation Rules

### Organization Validation
- Primary name required
- Must have at least one name of type "primary"
- Key must match organization key format
- All names must be valid

### Name Validation
- Name cannot be empty
- Type must be valid (primary, legal, dba, etc.)
- Status must be valid (active, inactive, historic)

### Relationship Validation
- Ownership percentage must be 0-100%
- Transaction types must be valid
- Required fields must be present

## Performance Considerations

### Scalability
- Tested with 1000+ organizations
- Efficient lookup using normalized names
- BFS algorithm for family relationships

### Benchmarks
```go
BenchmarkOrganizationSearchExpansion_ExpandSearch-8     	   10000	    150 ns/op

```

## Integration Patterns

### External Platform APIs

```go
// Example: GitHub API integration
type GitHubSearcher struct {
    client *github.Client
}

func (g *GitHubSearcher) SearchOrganizations(names []string) []Asset {
    var assets []Asset
    for _, name := range names {
        // Search GitHub for repositories, organizations
        repos := g.client.SearchRepositories(name)
        orgs := g.client.SearchOrganizations(name)
        // Convert to Asset objects
    }
    return assets
}
```

## Error Handling

### Common Scenarios
- Invalid organization names
- Duplicate name registrations
- Missing required fields
- Relationship constraint violations

### Best Practices
```go
// Validate before adding names
if err := org.AddName(name, nameType, source); err != nil {
    log.Printf("Failed to add name %s: %v", name, err)
    return err
}

// Check organization validity
if !org.Valid() {
    return fmt.Errorf("invalid organization: %+v", org)
}
```

## Testing Strategy

### Unit Tests
- Model validation
- Name normalization
- Search expansion
- Relationship management

### Integration Tests
- Cross-platform discovery
- End-to-end workflows
- Performance benchmarks

### Acceptance Tests
- Jira story requirements
- Real-world scenarios
- Example use cases

## Schema Generation

The models automatically generate:

1. **OpenAPI 3.1 Schema** (`client/api.yaml`)
2. **Python Pydantic Models** (`client/python/tabularium/models.py`)
3. **Type-safe Client Libraries**

### Example Generated Schema
```yaml
Organization:
  type: object
  properties:
    primaryName:
      type: string
      description: "Primary canonical name of the organization"
      example: "Walmart"
    names:
      type: array
      items:
        $ref: "#/components/schemas/OrganizationName"
```

## Deployment

### Registration
Models are automatically registered in `init()` functions:

```go
func init() {
    registry.Registry.MustRegisterModel(&Organization{})
    registry.Registry.MustRegisterModel(&OrganizationName{})
}
```

### Code Generation
Run the generation pipeline:

```bash
# Generate OpenAPI schema
go run ./cmd/schemagen -output client/api.yaml

# Generate Python client
go run ./cmd/codegen -input client/api.yaml -gen py:client/python/tabularium
```

## Troubleshooting

### Common Issues

**Empty Search Results**
```go
// Check if organization is registered
org := service.FindOrganization("MyOrg")
if org == nil {
    log.Printf("Organization not found: %s", "MyOrg")
}
```

**Name Conflicts**
```go
// Handle duplicate names
if err := org.AddName(name, nameType, source); err != nil {
    if strings.Contains(err.Error(), "already exists") {
        log.Printf("Name already registered: %s", name)
    }
}
```

**Validation Failures**
```go
// Debug validation issues
if !org.Valid() {
    log.Printf("Organization invalid:")
    log.Printf("  PrimaryName: %s", org.PrimaryName)
    log.Printf("  Names count: %d", len(org.Names))
    log.Printf("  Key: %s", org.Key)
}
```
