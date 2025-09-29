package model

import (
	"fmt"
	"time"
)

type Subscription map[string]any

func (sub Subscription) Within(t time.Time) bool {
	startDate, endDate := &time.Time{}, &time.Time{}

	if start, ok := sub["startDate"].(string); ok {
		t, _ = time.Parse("2006-01-02", start)
		startDate = &t
	}

	afterStart := startDate == nil || t.After(*startDate)

	if end, ok := sub["endDate"].(string); ok {
		t, _ = time.Parse("2006-01-02", end)
		endDate = &t
	}

	beforeEnd := endDate == nil || t.Before(*endDate)

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
