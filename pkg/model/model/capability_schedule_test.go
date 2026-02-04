package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilitySchedule_NewCapabilitySchedule(t *testing.T) {
	weeklySchedule := WeeklySchedule{
		Monday:   DaySchedule{Enabled: true, Time: "09:00"},
		Tuesday:  DaySchedule{Enabled: true, Time: "14:30"},
		Saturday: DaySchedule{Enabled: true, Time: "08:00"},
	}

	now := time.Now().UTC().Format(time.RFC3339)
	future := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)

	config := map[string]string{
		"Username": "admin",
		"Password": "secret",
		"Domain":   "example.com",
	}

	schedule := NewCapabilitySchedule(
		"user@example.com",
		"windows-ad-sharphound",
		"#addomain#example.com#S-1-5-21-123",
		config,
		weeklySchedule,
		now,
		future,
		"C.test-client-123", // clientID
	)

	assert.NotEmpty(t, schedule.ScheduleID)
	assert.Equal(t, "#capability_schedule#"+schedule.ScheduleID, schedule.Key)
	assert.Equal(t, "user@example.com", schedule.Username)
	assert.Equal(t, "windows-ad-sharphound", schedule.CapabilityName)
	assert.Equal(t, "C.test-client-123", schedule.ClientID)
	assert.Equal(t, ScheduleStatusActive, schedule.Status)
	assert.NotEmpty(t, schedule.CreatedAt)
	assert.NotEmpty(t, schedule.UpdatedAt)
	assert.NotEmpty(t, schedule.NextExecution)
}

func TestCapabilitySchedule_Validate_Success(t *testing.T) {
	schedule := &CapabilitySchedule{
		CapabilityName: "test-capability",
		ClientID:       "C.test-123",
		TargetKey:      "#asset#test.com#test.com",
		StartDate:      "2025-11-22T00:00:00Z",
		EndDate:        "2025-12-10T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	err := schedule.Validate()
	assert.NoError(t, err)
}

func TestCapabilitySchedule_Validate_Success_NonAegis(t *testing.T) {
	schedule := &CapabilitySchedule{
		CapabilityName: "test-capability",
		TargetKey:      "#asset#test.com#test.com",
		StartDate:      "2025-11-22T00:00:00Z",
		EndDate:        "2025-12-10T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	err := schedule.Validate()
	assert.NoError(t, err)
}

func TestCapabilitySchedule_Validate_MissingCapability(t *testing.T) {
	schedule := &CapabilitySchedule{
		ClientID:  "C.test-123",
		TargetKey: "#asset#test.com#test.com",
		StartDate: "2025-11-22T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	err := schedule.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "capability name is required")
}

func TestCapabilitySchedule_Validate_InvalidStartDate(t *testing.T) {
	schedule := &CapabilitySchedule{
		CapabilityName: "test-capability",
		ClientID:       "C.test-123",
		TargetKey:      "#asset#test.com#test.com",
		StartDate:      "invalid-date",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	err := schedule.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid start date format")
}

func TestCapabilitySchedule_Validate_EndDateBeforeStart(t *testing.T) {
	schedule := &CapabilitySchedule{
		CapabilityName: "test-capability",
		ClientID:       "C.test-123",
		TargetKey:      "#asset#test.com#test.com",
		StartDate:      "2025-12-10T00:00:00Z",
		EndDate:        "2025-11-22T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	err := schedule.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end date must be after start date")
}

func TestCapabilitySchedule_Validate_NoDaysEnabled(t *testing.T) {
	schedule := &CapabilitySchedule{
		CapabilityName: "test-capability",
		ClientID:       "C.test-123",
		TargetKey:      "#asset#test.com#test.com",
		StartDate:      "2025-11-22T00:00:00Z",
		WeeklySchedule: WeeklySchedule{},
	}

	err := schedule.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one day must be enabled")
}

func TestCapabilitySchedule_Validate_InvalidTimeFormat(t *testing.T) {
	schedule := &CapabilitySchedule{
		CapabilityName: "test-capability",
		ClientID:       "C.test-123",
		TargetKey:      "#asset#test.com#test.com",
		StartDate:      "2025-11-22T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "invalid"},
		},
	}

	err := schedule.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid time format")
}

func TestCapabilitySchedule_Validate_InvalidTimeRange(t *testing.T) {
	schedule := &CapabilitySchedule{
		CapabilityName: "test-capability",
		ClientID:       "C.test-123",
		TargetKey:      "#asset#test.com#test.com",
		StartDate:      "2025-11-22T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "25:00"}, // Invalid hour
		},
	}

	err := schedule.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid time range")
}

func TestCapabilitySchedule_CalculateNextExecution_FutureStart(t *testing.T) {
	// Schedule starts in the future
	futureStart := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	schedule := &CapabilitySchedule{
		StartDate: futureStart,
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	schedule.CalculateNextExecution()

	assert.NotEmpty(t, schedule.NextExecution)
	// Next execution should be at or after start date
	nextExec, _ := time.Parse(time.RFC3339, schedule.NextExecution)
	startTime, _ := time.Parse(time.RFC3339, futureStart)
	assert.True(t, nextExec.After(startTime) || nextExec.Equal(startTime))
}

func TestCapabilitySchedule_CalculateNextExecution_Expired(t *testing.T) {
	// Schedule ended in the past
	pastEnd := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)

	schedule := &CapabilitySchedule{
		StartDate: "2025-01-01T00:00:00Z",
		EndDate:   pastEnd,
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
		Status: ScheduleStatusActive,
	}

	schedule.CalculateNextExecution()

	assert.Equal(t, ScheduleStatusExpired, schedule.Status)
	assert.Empty(t, schedule.NextExecution)
}

func TestCapabilitySchedule_ShouldExecuteNow_NotStarted(t *testing.T) {
	futureStart := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	schedule := &CapabilitySchedule{
		StartDate:     futureStart,
		Status:        ScheduleStatusActive,
		NextExecution: futureStart,
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	assert.False(t, schedule.ShouldExecuteNow())
}

func TestCapabilitySchedule_ShouldExecuteNow_Paused(t *testing.T) {
	schedule := &CapabilitySchedule{
		StartDate:     "2025-01-01T00:00:00Z",
		Status:        ScheduleStatusPaused,
		NextExecution: time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339),
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	assert.False(t, schedule.ShouldExecuteNow())
}

func TestCapabilitySchedule_ShouldExecuteNow_ReadyToExecute(t *testing.T) {
	// Schedule should execute if next execution time has passed
	pastExecution := time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339)

	schedule := &CapabilitySchedule{
		StartDate:     "2025-01-01T00:00:00Z",
		Status:        ScheduleStatusActive,
		NextExecution: pastExecution,
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	assert.True(t, schedule.ShouldExecuteNow())
}

func TestCapabilitySchedule_Pause(t *testing.T) {
	schedule := &CapabilitySchedule{
		Status: ScheduleStatusActive,
	}

	schedule.Pause()

	assert.Equal(t, ScheduleStatusPaused, schedule.Status)
	assert.NotEmpty(t, schedule.UpdatedAt)
}

func TestCapabilitySchedule_Resume(t *testing.T) {
	schedule := &CapabilitySchedule{
		Status:    ScheduleStatusPaused,
		StartDate: "2025-11-22T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	schedule.Resume()

	assert.Equal(t, ScheduleStatusActive, schedule.Status)
	assert.NotEmpty(t, schedule.UpdatedAt)
	assert.NotEmpty(t, schedule.NextExecution)
}

func TestCapabilitySchedule_Resume_OnlyWhenPaused(t *testing.T) {
	schedule := &CapabilitySchedule{
		Status: ScheduleStatusExpired,
	}

	oldStatus := schedule.Status
	schedule.Resume()

	// Should not change status if not paused
	assert.Equal(t, oldStatus, schedule.Status)
}

func TestCapabilitySchedule_MarkExecuted(t *testing.T) {
	schedule := &CapabilitySchedule{
		StartDate: "2025-11-22T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	oldLastExecution := schedule.LastExecution
	schedule.MarkExecuted()

	assert.NotEqual(t, oldLastExecution, schedule.LastExecution)
	assert.NotEmpty(t, schedule.LastExecution)
	assert.NotEmpty(t, schedule.UpdatedAt)
}

func TestCapabilitySchedule_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   ScheduleStatus
		expected bool
	}{
		{"active", ScheduleStatusActive, true},
		{"paused", ScheduleStatusPaused, false},
		{"expired", ScheduleStatusExpired, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := &CapabilitySchedule{Status: tt.status}
			assert.Equal(t, tt.expected, schedule.IsActive())
		})
	}
}

func TestCapabilitySchedule_GetDescription(t *testing.T) {
	schedule := &CapabilitySchedule{}
	desc := schedule.GetDescription()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "capabilities")
}

func TestCapabilitySchedule_GetKey(t *testing.T) {
	schedule := &CapabilitySchedule{
		Key: "#capability_schedule#test-123",
	}

	assert.Equal(t, "#capability_schedule#test-123", schedule.GetKey())
}

func TestCapabilitySchedule_MarshalJSON(t *testing.T) {
	schedule := &CapabilitySchedule{
		ScheduleID:     "test-123",
		Key:            "#capability_schedule#test-123",
		CapabilityName: "test-capability",
		ClientID:       "C.test-123",
		TargetKey:      "#asset#test.com#test.com",
		Status:         ScheduleStatusActive,
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	data, err := json.Marshal(schedule)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test-123")
	assert.Contains(t, string(data), "test-capability")
}

func TestCapabilitySchedule_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"scheduleId": "test-123",
		"key": "#capability_schedule#test-123",
		"capabilityName": "test-capability",
		"clientId": "C.test-123",
		"targetKey": "#asset#test.com#test.com",
		"config": {"param1": "value1"},
		"weeklySchedule": {
			"monday": {"enabled": true, "time": "09:00"},
			"tuesday": {"enabled": false, "time": "09:00"},
			"wednesday": {"enabled": false, "time": "09:00"},
			"thursday": {"enabled": false, "time": "09:00"},
			"friday": {"enabled": false, "time": "09:00"},
			"saturday": {"enabled": false, "time": "09:00"},
			"sunday": {"enabled": false, "time": "09:00"}
		},
		"startDate": "2025-11-22T00:00:00Z",
		"status": "active"
	}`

	var schedule CapabilitySchedule
	err := json.Unmarshal([]byte(jsonData), &schedule)

	require.NoError(t, err)
	assert.Equal(t, "test-123", schedule.ScheduleID)
	assert.Equal(t, "test-capability", schedule.CapabilityName)
	assert.Equal(t, "C.test-123", schedule.ClientID)
	assert.True(t, schedule.WeeklySchedule.Monday.Enabled)
	assert.Equal(t, "09:00", schedule.WeeklySchedule.Monday.Time)
	assert.False(t, schedule.WeeklySchedule.Tuesday.Enabled)
}

func TestCapabilitySchedule_findNextExecutionFrom_SameDay(t *testing.T) {
	// Test finding next execution on the same day
	now := time.Now().UTC()

	schedule := &CapabilitySchedule{
		StartDate: "2025-01-01T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday:    DaySchedule{Enabled: true, Time: "09:00"},
			Tuesday:   DaySchedule{Enabled: true, Time: "14:00"},
			Wednesday: DaySchedule{Enabled: true, Time: "10:00"},
			Thursday:  DaySchedule{Enabled: true, Time: "15:00"},
			Friday:    DaySchedule{Enabled: true, Time: "16:00"},
			Saturday:  DaySchedule{Enabled: true, Time: "08:00"},
			Sunday:    DaySchedule{Enabled: true, Time: "11:00"},
		},
	}

	nextExec := schedule.findNextExecutionFrom(now)
	assert.NotEmpty(t, nextExec)

	// Parse and verify it's in the future
	nextTime, err := time.Parse(time.RFC3339, nextExec)
	require.NoError(t, err)
	assert.True(t, nextTime.After(now), "Next execution should be in the future")
}

func TestCapabilitySchedule_findNextExecutionFrom_NextWeek(t *testing.T) {
	// Schedule only enabled for a specific day in the future
	now := time.Now().UTC()

	// Enable only one day per week
	schedule := &CapabilitySchedule{
		StartDate: "2025-01-01T00:00:00Z",
		WeeklySchedule: WeeklySchedule{
			Monday: DaySchedule{Enabled: int(now.Weekday()) != 1, Time: "09:00"},
		},
	}

	nextExec := schedule.findNextExecutionFrom(now)

	if int(now.Weekday()) != 1 {
		assert.NotEmpty(t, nextExec)
	}
}

func TestCapabilitySchedule_findNextExecutionFrom_RespectEndDate(t *testing.T) {
	now := time.Now().UTC()
	// End date is tomorrow
	tomorrow := now.Add(24 * time.Hour).Format(time.RFC3339)

	schedule := &CapabilitySchedule{
		StartDate: "2025-01-01T00:00:00Z",
		EndDate:   tomorrow,
		WeeklySchedule: WeeklySchedule{
			Monday:    DaySchedule{Enabled: true, Time: "09:00"},
			Tuesday:   DaySchedule{Enabled: true, Time: "09:00"},
			Wednesday: DaySchedule{Enabled: true, Time: "09:00"},
			Thursday:  DaySchedule{Enabled: true, Time: "09:00"},
			Friday:    DaySchedule{Enabled: true, Time: "09:00"},
			Saturday:  DaySchedule{Enabled: true, Time: "09:00"},
			Sunday:    DaySchedule{Enabled: true, Time: "09:00"},
		},
	}

	nextExec := schedule.findNextExecutionFrom(now)

	if nextExec != "" {
		nextTime, err := time.Parse(time.RFC3339, nextExec)
		require.NoError(t, err)
		endTime, _ := time.Parse(time.RFC3339, tomorrow)
		assert.True(t, nextTime.Before(endTime), "Next execution should be before end date")
	}
}

func TestWeeklySchedule_AllDaysEnabled(t *testing.T) {
	schedule := WeeklySchedule{
		Monday:    DaySchedule{Enabled: true, Time: "09:00"},
		Tuesday:   DaySchedule{Enabled: true, Time: "10:00"},
		Wednesday: DaySchedule{Enabled: true, Time: "11:00"},
		Thursday:  DaySchedule{Enabled: true, Time: "12:00"},
		Friday:    DaySchedule{Enabled: true, Time: "13:00"},
		Saturday:  DaySchedule{Enabled: true, Time: "14:00"},
		Sunday:    DaySchedule{Enabled: true, Time: "15:00"},
	}

	data, err := json.Marshal(schedule)
	require.NoError(t, err)

	var unmarshaled WeeklySchedule
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.True(t, unmarshaled.Monday.Enabled)
	assert.Equal(t, "09:00", unmarshaled.Monday.Time)
	assert.True(t, unmarshaled.Sunday.Enabled)
	assert.Equal(t, "15:00", unmarshaled.Sunday.Time)
}

func TestDaySchedule_Disabled(t *testing.T) {
	day := DaySchedule{Enabled: false, Time: "09:00"}

	data, err := json.Marshal(day)
	require.NoError(t, err)

	var unmarshaled DaySchedule
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.False(t, unmarshaled.Enabled)
	assert.Equal(t, "09:00", unmarshaled.Time)
}

func TestScheduleStatus_Constants(t *testing.T) {
	assert.Equal(t, ScheduleStatus("active"), ScheduleStatusActive)
	assert.Equal(t, ScheduleStatus("paused"), ScheduleStatusPaused)
	assert.Equal(t, ScheduleStatus("expired"), ScheduleStatusExpired)
}
