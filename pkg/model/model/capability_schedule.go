package model

import (
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
	"time"

	"github.com/google/uuid"
)

func init() {
	shared.Registry.MustRegisterModel(&CapabilitySchedule{}, "capability_schedule")
}

// WeeklySchedule defines execution times for each day of the week
type WeeklySchedule struct {
	Monday    DaySchedule `json:"monday" dynamodbav:"monday"`
	Tuesday   DaySchedule `json:"tuesday" dynamodbav:"tuesday"`
	Wednesday DaySchedule `json:"wednesday" dynamodbav:"wednesday"`
	Thursday  DaySchedule `json:"thursday" dynamodbav:"thursday"`
	Friday    DaySchedule `json:"friday" dynamodbav:"friday"`
	Saturday  DaySchedule `json:"saturday" dynamodbav:"saturday"`
	Sunday    DaySchedule `json:"sunday" dynamodbav:"sunday"`
}

// DaySchedule represents a single day's execution configuration
type DaySchedule struct {
	Enabled bool   `json:"enabled" dynamodbav:"enabled"`
	Time    string `json:"time" dynamodbav:"time"` // HH:MM format (24-hour, UTC)
}

// ScheduleStatus represents the state of a schedule
type ScheduleStatus string

const (
	ScheduleStatusActive  ScheduleStatus = "active"
	ScheduleStatusPaused  ScheduleStatus = "paused"
	ScheduleStatusExpired ScheduleStatus = "expired"
)

// CapabilitySchedule represents a scheduled execution configuration for any capability
type CapabilitySchedule struct {
	baseTableModel  // Embed for TableModel interface
	model.BaseModel // Embed BaseModel for default implementations

	ScheduleID string `json:"scheduleId" dynamodbav:"schedule_id" desc:"Unique identifier for this schedule"`

	Key string `json:"key" dynamodbav:"key" desc:"DynamoDB partition key"`

	Username string `json:"username" dynamodbav:"username" desc:"Account owner email"`

	CapabilityName string `json:"capabilityName" dynamodbav:"capability_name" desc:"Capability name to execute"`

	ClientID string `json:"clientId,omitempty" dynamodbav:"client_id,omitempty" desc:"Aegis agent client ID (only for Aegis capabilities)"`

	TargetKey string `json:"targetKey" dynamodbav:"target_key" desc:"Asset or target key for capability execution"`

	Config map[string]string `json:"config" dynamodbav:"config" desc:"Capability configuration parameters including credentials"`

	WeeklySchedule WeeklySchedule `json:"weeklySchedule" dynamodbav:"weekly_schedule" desc:"Weekly execution schedule configuration"`

	StartDate string `json:"startDate" dynamodbav:"start_date" desc:"Schedule start date in RFC3339 format"`

	EndDate string `json:"endDate,omitempty" dynamodbav:"end_date,omitempty" desc:"Schedule end date in RFC3339 format (optional)"`

	Status ScheduleStatus `json:"status" dynamodbav:"status" desc:"Schedule status (active, paused, or expired)"`

	NextExecution string `json:"nextExecution,omitempty" dynamodbav:"next_execution,omitempty" desc:"Calculated next execution time"`

	LastExecution string `json:"lastExecution,omitempty" dynamodbav:"last_execution,omitempty" desc:"Last execution timestamp"`

	CreatedAt string `json:"createdAt" dynamodbav:"created_at" desc:"Creation timestamp"`

	UpdatedAt string `json:"updatedAt" dynamodbav:"updated_at" desc:"Last modification timestamp"`
}

// GetDescription returns a description of the CapabilitySchedule model
func (s *CapabilitySchedule) GetDescription() string {
	return "Scheduled execution configuration for capabilities"
}

// GetKey returns the DynamoDB partition key for this schedule
func (s *CapabilitySchedule) GetKey() string {
	return s.Key
}

// NewCapabilitySchedule creates a new CapabilitySchedule with default values
// Config should include all capability parameters including credentials
func NewCapabilitySchedule(
	username, capabilityName, targetKey string,
	config map[string]string,
	weeklySchedule WeeklySchedule,
	startDate, endDate string,
	clientID string,
) *CapabilitySchedule {
	now := Now()
	scheduleID := uuid.New().String()

	schedule := &CapabilitySchedule{
		ScheduleID:     scheduleID,
		Key:            fmt.Sprintf("#capability_schedule#%s", scheduleID),
		Username:       username,
		CapabilityName: capabilityName,
		ClientID:       clientID,
		TargetKey:      targetKey,
		Config:         config,
		WeeklySchedule: weeklySchedule,
		StartDate:      startDate,
		EndDate:        endDate,
		Status:         ScheduleStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	schedule.CalculateNextExecution()
	return schedule
}

// CalculateNextExecution calculates and sets the next execution time based on the schedule
func (s *CapabilitySchedule) CalculateNextExecution() {
	now := time.Now().UTC()

	// Check if schedule has expired
	if s.EndDate != "" {
		endTime, err := time.Parse(time.RFC3339, s.EndDate)
		if err == nil && now.After(endTime) {
			s.Status = ScheduleStatusExpired
			s.NextExecution = ""
			return
		}
	}

	// Check if schedule hasn't started yet
	startTime, err := time.Parse(time.RFC3339, s.StartDate)
	if err != nil || now.Before(startTime) {
		// Find first execution after start date
		s.NextExecution = s.findNextExecutionFrom(startTime)
		return
	}

	// Find next execution from now
	s.NextExecution = s.findNextExecutionFrom(now)
}

// findNextExecutionFrom finds the next execution time starting from the given time
func (s *CapabilitySchedule) findNextExecutionFrom(from time.Time) string {
	days := []struct {
		weekday  time.Weekday
		schedule DaySchedule
	}{
		{time.Sunday, s.WeeklySchedule.Sunday},
		{time.Monday, s.WeeklySchedule.Monday},
		{time.Tuesday, s.WeeklySchedule.Tuesday},
		{time.Wednesday, s.WeeklySchedule.Wednesday},
		{time.Thursday, s.WeeklySchedule.Thursday},
		{time.Friday, s.WeeklySchedule.Friday},
		{time.Saturday, s.WeeklySchedule.Saturday},
	}

	// Check up to 7 days ahead to find next execution (0-7 to include same day next week)
	for daysAhead := 0; daysAhead <= 7; daysAhead++ {
		checkDate := from.AddDate(0, 0, daysAhead)
		checkWeekday := checkDate.Weekday()

		for _, day := range days {
			if day.weekday != checkWeekday || !day.schedule.Enabled {
				continue
			}

			// Parse time (HH:MM format)
			var hour, minute int
			_, err := fmt.Sscanf(day.schedule.Time, "%d:%d", &hour, &minute)
			if err != nil {
				continue
			}

			execTime := time.Date(
				checkDate.Year(), checkDate.Month(), checkDate.Day(),
				hour, minute, 0, 0, time.UTC,
			)

			// If this time hasn't passed yet (or we're checking future days), use it
			if execTime.After(from) {
				// Check if this is before end date (if set)
				if s.EndDate != "" {
					endTime, err := time.Parse(time.RFC3339, s.EndDate)
					if err == nil && execTime.After(endTime) {
						continue
					}
				}
				return execTime.Format(time.RFC3339)
			}
		}
	}

	return ""
}

// IsActive returns true if the schedule is currently active
func (s *CapabilitySchedule) IsActive() bool {
	return s.Status == ScheduleStatusActive
}

// ShouldExecuteNow checks if the schedule should execute at the current time
func (s *CapabilitySchedule) ShouldExecuteNow() bool {
	if !s.IsActive() {
		return false
	}

	now := time.Now().UTC()

	startTime, err := time.Parse(time.RFC3339, s.StartDate)
	if err != nil || now.Before(startTime) {
		return false
	}

	if s.EndDate != "" {
		endTime, err := time.Parse(time.RFC3339, s.EndDate)
		if err == nil && now.After(endTime) {
			return false
		}
	}

	if s.NextExecution == "" {
		return false
	}

	nextExecTime, err := time.Parse(time.RFC3339, s.NextExecution)
	if err != nil {
		return false
	}

	return now.After(nextExecTime)
}

// Pause pauses the schedule
func (s *CapabilitySchedule) Pause() {
	s.Status = ScheduleStatusPaused
	s.UpdatedAt = Now()
}

// Resume resumes a paused schedule
func (s *CapabilitySchedule) Resume() {
	if s.Status == ScheduleStatusPaused {
		s.Status = ScheduleStatusActive
		s.CalculateNextExecution()
		s.UpdatedAt = Now()
	}
}

// MarkExecuted marks the schedule as executed and calculates next execution
func (s *CapabilitySchedule) MarkExecuted() {
	s.LastExecution = Now()
	s.CalculateNextExecution()
	s.UpdatedAt = Now()
}

// Validate validates the schedule configuration
func (s *CapabilitySchedule) Validate() error {
	if s.CapabilityName == "" {
		return fmt.Errorf("capability name is required")
	}

	if s.TargetKey == "" {
		return fmt.Errorf("target key is required")
	}

	if s.StartDate == "" {
		return fmt.Errorf("start date is required")
	}

	_, err := time.Parse(time.RFC3339, s.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start date format: %w", err)
	}

	if s.EndDate != "" {
		endTime, err := time.Parse(time.RFC3339, s.EndDate)
		if err != nil {
			return fmt.Errorf("invalid end date format: %w", err)
		}

		startTime, _ := time.Parse(time.RFC3339, s.StartDate)
		if endTime.Before(startTime) {
			return fmt.Errorf("end date must be after start date")
		}
	}

	hasEnabledDay := s.WeeklySchedule.Monday.Enabled ||
		s.WeeklySchedule.Tuesday.Enabled ||
		s.WeeklySchedule.Wednesday.Enabled ||
		s.WeeklySchedule.Thursday.Enabled ||
		s.WeeklySchedule.Friday.Enabled ||
		s.WeeklySchedule.Saturday.Enabled ||
		s.WeeklySchedule.Sunday.Enabled

	if !hasEnabledDay {
		return fmt.Errorf("at least one day must be enabled in the schedule")
	}

	days := []DaySchedule{
		s.WeeklySchedule.Monday,
		s.WeeklySchedule.Tuesday,
		s.WeeklySchedule.Wednesday,
		s.WeeklySchedule.Thursday,
		s.WeeklySchedule.Friday,
		s.WeeklySchedule.Saturday,
		s.WeeklySchedule.Sunday,
	}

	for _, day := range days {
		if day.Enabled {
			var hour, minute int
			_, err := fmt.Sscanf(day.Time, "%d:%d", &hour, &minute)
			if err != nil {
				return fmt.Errorf("invalid time format for enabled day: %s", day.Time)
			}
			if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
				return fmt.Errorf("invalid time range for enabled day: %s", day.Time)
			}
		}
	}

	return nil
}
