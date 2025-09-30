package model

import (
	"fmt"
	"time"
)

type Subscription map[string]any

func (sub Subscription) Within(t time.Time) bool {
	var startDate, endDate *time.Time

	if start, ok := sub["startDate"].(string); ok {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = &parsed
		}
	}

	afterStart := startDate == nil || t.After(*startDate) || t.Equal(*startDate)

	if end, ok := sub["endDate"].(string); ok {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = &parsed
		}
	}

	beforeEnd := endDate == nil || t.Before(*endDate) || t.Equal(*endDate)

	return afterStart && beforeEnd
}

func subscriptionValidator(v any) error {
	m, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("input %T was not map[string]any", v)
	}

	validate := func(k string) error {
		if a, ok := m[k]; ok {
			if start, ok := a.(string); ok {
				_, err := time.Parse("2006-01-02", start)
				if err != nil {
					return fmt.Errorf("invalid start date: %v", err)
				}
			} else {
				return fmt.Errorf("startDate was specified but was not a string")
			}
		}
		return nil
	}

	if err := validate("startDate"); err != nil {
		return err
	}
	if err := validate("endDate"); err != nil {
		return err
	}
	return nil
}
