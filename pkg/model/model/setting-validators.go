package model

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

var settingValidators = map[string]func(value any) error{
	"scan-level":           scanLevelValidator,
	"rate-limit":           rateLimitValidator,
	"blocked_capabilities": blockedCapabilitiesValidator,
}

func scanLevelValidator(value any) error {
	var permittedScanLevels = []string{Active, ActiveLow, ActivePassive}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("scan level must be a string")
	}

	if !slices.Contains(permittedScanLevels, str) {
		return fmt.Errorf("scan level must be one of [%s]", strings.Join(permittedScanLevels, ", "))
	}

	return nil
}

func blockedCapabilitiesValidator(value any) error {
	bites, err := json.Marshal(value)
	if err != nil || value == nil {
		return fmt.Errorf("failed to marshal JSON")
	}

	type blocked struct {
		Capabilities []string `json:"capabilities"`
	}

	var b blocked
	if err = json.Unmarshal(bites, &b); err != nil {
		return fmt.Errorf("failed to unmarshal JSON")
	}

	if len(b.Capabilities) == 0 {
		return fmt.Errorf("blocked capabilities must contain at least one capability")
	}

	return nil
}
