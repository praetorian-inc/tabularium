package model

type OrganizationRelationshipService struct {
	organizations map[string]*Organization
}

func NewOrganizationRelationshipService() *OrganizationRelationshipService {
	return &OrganizationRelationshipService{
		organizations: make(map[string]*Organization),
	}
}

func (ors *OrganizationRelationshipService) AddOrganization(org *Organization) {
	ors.organizations[org.GetKey()] = org
}

func (ors *OrganizationRelationshipService) GetSubsidiaries(orgKey string) []*Organization {
	var subsidiaries []*Organization

	for _, org := range ors.organizations {
		if org.ParentOrganization == orgKey {
			subsidiaries = append(subsidiaries, org)
		}
	}

	return subsidiaries
}

func (ors *OrganizationRelationshipService) GetParentOrganizations(orgKey string) []*Organization {
	var parents []*Organization

	if org, exists := ors.organizations[orgKey]; exists && org.ParentOrganization != "" {
		if parent, exists := ors.organizations[org.ParentOrganization]; exists {
			parents = append(parents, parent)
		}
	}

	return parents
}

func (ors *OrganizationRelationshipService) GetOrganizationFamily(orgKey string) []*Organization {
	visited := make(map[string]bool)
	family := make([]*Organization, 0)

	queue := []string{orgKey}
	visited[orgKey] = true

	for len(queue) > 0 {
		currentKey := queue[0]
		queue = queue[1:]

		if org, exists := ors.organizations[currentKey]; exists {
			family = append(family, org)

			if org.ParentOrganization != "" && !visited[org.ParentOrganization] {
				queue = append(queue, org.ParentOrganization)
				visited[org.ParentOrganization] = true
			}
		}

		for _, otherOrg := range ors.organizations {
			if otherOrg.ParentOrganization == currentKey && !visited[otherOrg.GetKey()] {
				queue = append(queue, otherOrg.GetKey())
				visited[otherOrg.GetKey()] = true
			}
		}
	}

	return family
}
