package model

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var labelRegistry map[string]string

func getRegistry() map[string]string {
	if labelRegistry == nil {
		labelRegistry = map[string]string{}
	}

	return labelRegistry
}

func MustRegisterLabel(label string) {
	registry := getRegistry()
	_, exists := registry[label]
	if exists {
		panic(fmt.Sprintf("label '%s' already registered", label))
	}

	lowercase := strings.ToLower(label)
	registry[lowercase] = label
}

func GetLabel(label string) string {
	lowercase := strings.ToLower(label)

	registry := getRegistry()
	registered, ok := registry[lowercase]
	if ok {
		return registered
	}
	return cases.Title(language.English).String(lowercase)
}
