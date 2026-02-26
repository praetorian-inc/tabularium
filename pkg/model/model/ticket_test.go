package model

import "testing"

func TestAddTicket_New(t *testing.T) {
	td := TicketData{}
	td.AddTicket(Ticket{Provider: "jira", ID: "PROJ-1", Link: "https://jira.example.com/PROJ-1", Status: "Open"})

	if len(td.Tickets) != 1 {
		t.Fatalf("expected 1 ticket, got %d", len(td.Tickets))
	}
	if td.Tickets[0].Provider != "jira" || td.Tickets[0].ID != "PROJ-1" {
		t.Fatalf("unexpected ticket: %+v", td.Tickets[0])
	}
}

func TestAddTicket_Replace(t *testing.T) {
	td := TicketData{
		Tickets: []Ticket{
			{Provider: "jira", ID: "PROJ-1", Status: "Open"},
		},
	}
	td.AddTicket(Ticket{Provider: "jira", ID: "PROJ-1", Status: "Done"})

	if len(td.Tickets) != 1 {
		t.Fatalf("expected 1 ticket after replace, got %d", len(td.Tickets))
	}
	if td.Tickets[0].Status != "Done" {
		t.Fatalf("expected status Done, got %s", td.Tickets[0].Status)
	}
}

func TestAddTicket_MultipleProviders(t *testing.T) {
	td := TicketData{}
	td.AddTicket(Ticket{Provider: "jira", ID: "PROJ-1"})
	td.AddTicket(Ticket{Provider: "azuredevops", ID: "42"})

	if len(td.Tickets) != 2 {
		t.Fatalf("expected 2 tickets, got %d", len(td.Tickets))
	}
}

func TestRemoveTicket(t *testing.T) {
	td := TicketData{
		Tickets: []Ticket{
			{Provider: "jira", ID: "PROJ-1"},
			{Provider: "azuredevops", ID: "42"},
		},
	}
	td.RemoveTicket("jira", "PROJ-1")

	if len(td.Tickets) != 1 {
		t.Fatalf("expected 1 ticket after remove, got %d", len(td.Tickets))
	}
	if td.Tickets[0].Provider != "azuredevops" {
		t.Fatalf("wrong ticket remaining: %+v", td.Tickets[0])
	}
}

func TestRemoveTicket_NotFound(t *testing.T) {
	td := TicketData{
		Tickets: []Ticket{
			{Provider: "jira", ID: "PROJ-1"},
		},
	}
	td.RemoveTicket("jira", "PROJ-999")

	if len(td.Tickets) != 1 {
		t.Fatalf("expected 1 ticket (no-op), got %d", len(td.Tickets))
	}
}
