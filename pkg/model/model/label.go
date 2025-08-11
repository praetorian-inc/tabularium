package model

import (
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type labeled interface {
	GetLabels() []string
}

func FormatLabel(term string) (string, bool) {
	for tipe := range registry.Registry.GetAllTypes() {
		m, ok := registry.Registry.MakeType(tipe)
		if !ok {
			continue
		}

		labeled, ok := m.(labeled)
		if !ok {
			continue
		}

		for _, label := range labeled.GetLabels() {
			if strings.EqualFold(label, term) {
				return label, true
			}
		}
	}

	return "", false
}
