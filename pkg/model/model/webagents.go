package model

import "github.com/praetorian-inc/tabularium/pkg/registry"

func init() {
	registry.Registry.MustRegisterModel(&LoginPortalTemplate{})
}

type LoginPortalTemplate struct {
	registry.BaseModel
	FormReport   FormReport        `json:"form_report"`
	StatusReport LoginStatusReport `json:"status_report,omitempty"`
}

func (l *LoginPortalTemplate) GetDescription() string {
	return "Represents a login portal template, containing a form report and a status report. Can be used with lib/rod to login to a website."
}

type LoginStatusReport struct {
	Selector string   `json:"selector" description:"The exact goQuery selector string for the element that indicates the login failed."`
	Texts    []string `json:"texts" description:"Snippets of text that ONLY appear when the login attempt is invalid."`
}

type FormReport struct {
	LoginSelector string      `json:"login_selector" description:"The exact go-rod selector string to call page.MustElement(<X>).MustClick() to login to the website"`
	InputObjects  []FormInput `json:"input_objects" description:"Ordered list of required form inputs with their expected types. Each one will be used in page.MustElement(<selector>).MustInput(<value>)"`
}

type FormInput struct {
	Type     string `json:"type" description:"An ENUM of 'username', 'password', or 'other'. Used to replace the value later with test case values."`
	Selector string `json:"selector" description:"The exact go-rod selector string for the form field."`
	Value    string `json:"value" description:"An example value to be filled in the form field or request parameter"`
}
