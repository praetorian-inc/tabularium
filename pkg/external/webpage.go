package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Webpage is a simplified webpage for external tool writers.
type Webpage struct {
	URL string `json:"url"` // The webpage URL
}

// Group implements Target interface.
func (w Webpage) Group() string {
	return w.URL
}

// Identifier implements Target interface.
func (w Webpage) Identifier() string {
	return w.URL
}

// ToTarget converts to a full Tabularium Webpage.
func (w Webpage) ToTarget() (model.Target, error) {
	if w.URL == "" {
		return nil, fmt.Errorf("webpage requires url")
	}

	webpage := model.NewWebpageFromString(w.URL, nil)
	if !webpage.Valid() {
		return nil, fmt.Errorf("invalid webpage url: %s", w.URL)
	}

	return &webpage, nil
}

// ToModel converts to a full Tabularium Webpage (convenience method).
func (w Webpage) ToModel() (*model.Webpage, error) {
	target, err := w.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Webpage), nil
}

// WebpageFromModel converts a Tabularium Webpage to an external Webpage.
func WebpageFromModel(m *model.Webpage) Webpage {
	return Webpage{
		URL: m.URL,
	}
}
