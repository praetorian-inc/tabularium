package model

import (
	"fmt"
	"slices"
	"strings"
)

var configurationValidator = map[string]func(value any) error{
	"subscription": subscriptionValidator,
}
