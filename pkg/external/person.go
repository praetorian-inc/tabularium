package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Person is a simplified person for external tool writers.
// It contains essential fields for identifying and enriching person data.
type Person struct {
	Email            string `json:"email"`                       // Person's email address
	Name             string `json:"name"`                        // Person's full name
	Title            string `json:"title,omitempty"`             // Job title
	OrganizationName string `json:"organization_name,omitempty"` // Organization they work for
	LinkedinURL      string `json:"linkedin_url,omitempty"`      // LinkedIn profile URL
}

// Group implements Target interface.
func (p Person) Group() string { return p.Email }

// Identifier implements Target interface.
func (p Person) Identifier() string {
	if p.Email != "" {
		return fmt.Sprintf("#person#%s#%s", p.Email, p.Name)
	}
	return fmt.Sprintf("#person#%s#%s", p.Name, p.Name)
}

// ToTarget converts to a full Tabularium Person.
func (p Person) ToTarget() (model.Target, error) {
	if p.Email == "" && p.Name == "" {
		return nil, fmt.Errorf("person requires email or name")
	}

	var person *model.Person
	if p.Email != "" {
		person = model.NewPerson(p.Email, p.Name, "")
	} else {
		person = model.NewPersonFromName(p.Name, "")
	}

	if p.Title != "" {
		person.Title = &p.Title
	}
	if p.OrganizationName != "" {
		person.OrganizationName = &p.OrganizationName
	}
	if p.LinkedinURL != "" {
		person.LinkedinURL = &p.LinkedinURL
	}

	return person, nil
}

// ToModel converts to a full Tabularium Person (convenience method).
func (p Person) ToModel() (*model.Person, error) {
	target, err := p.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Person), nil
}

