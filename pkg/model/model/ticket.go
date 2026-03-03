package model

// Ticket represents an ITSM ticket associated with a Risk.
// Generic across all ticketing providers (Jira, AzureDevOps, Freshdesk, etc.).
type Ticket struct {
	Provider string `json:"provider" neo4j:"provider" desc:"ITSM provider name (e.g., jira, azuredevops, freshdesk)." example:"jira"`
	ID       string `json:"id" neo4j:"id" desc:"Ticket identifier in the external system." example:"PROJ-123"`
	Link     string `json:"link,omitempty" neo4j:"link,omitempty" desc:"URL to the ticket in the external system." example:"https://jira.example.com/browse/PROJ-123"`
	Status     string `json:"status,omitempty" neo4j:"status,omitempty" desc:"Current status of the ticket." example:"In Progress"`
	Assignee   string `json:"assignee,omitempty" neo4j:"assignee,omitempty" desc:"Display name of the person assigned to the ticket." example:"John Doe"`
	Resolution string `json:"resolution,omitempty" neo4j:"resolution,omitempty" desc:"Resolution of the ticket when closed." example:"Fixed"`
}

// TicketData is embedded in Risk to store associated tickets.
// Following the same pattern as History, OriginationData, Tags.
type TicketData struct {
	Tickets []Ticket `neo4j:"tickets,omitempty" json:"tickets,omitempty" desc:"ITSM tickets associated with this risk."`
}

// AddTicket adds a ticket to the TicketData, replacing any existing ticket
// from the same provider with the same ID.
func (td *TicketData) AddTicket(t Ticket) {
	for i, existing := range td.Tickets {
		if existing.Provider == t.Provider && existing.ID == t.ID {
			td.Tickets[i] = t
			return
		}
	}
	td.Tickets = append(td.Tickets, t)
}

// RemoveTicket removes a ticket by provider and ID.
func (td *TicketData) RemoveTicket(provider, id string) {
	for i, existing := range td.Tickets {
		if existing.Provider == provider && existing.ID == id {
			td.Tickets = append(td.Tickets[:i], td.Tickets[i+1:]...)
			return
		}
	}
}
