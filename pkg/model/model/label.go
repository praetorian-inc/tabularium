package model

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetLabel(label string) string {
	lowercase := strings.ToLower(label)
	if strings.HasPrefix(lowercase, "ad") {
		return GetADLabel(lowercase)
	}
	return cases.Title(language.English).String(lowercase)
}
