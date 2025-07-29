package model

import (
	"fmt"
	"slices"
	"strings"
)

var settingValidators = map[string]func(value any) error{
	"scan-level": scanLevelValidator,
	"rate-limit": rateLimitValidator,
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
