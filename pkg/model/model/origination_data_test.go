package model

import (
	"reflect"
	"testing"
)

func TestOriginationData_Merge(t *testing.T) {
	tests := []struct {
		name     string
		initial  OriginationData
		other    OriginationData
		expected OriginationData
	}{
		{
			name:    "merge empty with populated",
			initial: OriginationData{},
			other: OriginationData{
				Capability:    []string{"amazon", "portscan"},
				AttackSurface: []string{"internal", "external"},
				Origins:       []string{"amazon", "ipv4"},
			},
			expected: OriginationData{
				Capability:    []string{"amazon", "portscan"},
				AttackSurface: []string{"internal", "external"},
				Origins:       []string{"amazon", "ipv4"},
			},
		},
		{
			name: "merge populated with other populated - overwrite behavior",
			initial: OriginationData{
				Capability:    []string{"dns"},
				AttackSurface: []string{"internal"},
				Origins:       []string{"dns"},
			},
			other: OriginationData{
				Capability:    []string{"amazon", "portscan"},
				AttackSurface: []string{"external"},
				Origins:       []string{"amazon", "ipv4"},
			},
			expected: OriginationData{
				Capability:    []string{"amazon", "portscan"},
				AttackSurface: []string{"external"},
				Origins:       []string{"amazon", "ipv4"},
			},
		},
		{
			name: "merge with nil fields - no change",
			initial: OriginationData{
				Capability:    []string{"dns"},
				AttackSurface: []string{"internal"},
				Origins:       []string{"dns"},
			},
			other: OriginationData{
				Capability:    nil,
				AttackSurface: nil,
				Origins:       nil,
			},
			expected: OriginationData{
				Capability:    []string{"dns"},
				AttackSurface: []string{"internal"},
				Origins:       []string{"dns"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.Merge(tt.other)
			if !reflect.DeepEqual(tt.initial, tt.expected) {
				t.Errorf("OriginationData.Merge() = %v, expected %v", tt.initial, tt.expected)
			}
		})
	}
}

func TestOriginationData_Visit(t *testing.T) {
	tests := []struct {
		name     string
		initial  OriginationData
		other    OriginationData
		expected OriginationData
	}{
		{
			name:    "visit empty with populated",
			initial: OriginationData{},
			other: OriginationData{
				Capability:    []string{"amazon", "portscan"},
				AttackSurface: []string{"internal", "external"},
				Origins:       []string{"amazon", "ipv4"},
			},
			expected: OriginationData{
				Capability:    []string{"amazon", "portscan"},
				AttackSurface: []string{"external", "internal"},
				Origins:       []string{"amazon", "ipv4"},
			},
		},
		{
			name: "visit populated with other populated - merge behavior",
			initial: OriginationData{
				Capability:    []string{"dns"},
				AttackSurface: []string{"internal"},
				Origins:       []string{"dns"},
			},
			other: OriginationData{
				Capability:    []string{"amazon", "portscan"},
				AttackSurface: []string{"external"},
				Origins:       []string{"amazon", "ipv4"},
			},
			expected: OriginationData{
				Capability:    []string{"amazon", "dns", "portscan"},
				AttackSurface: []string{"external", "internal"},
				Origins:       []string{"amazon", "dns", "ipv4"},
			},
		},
		{
			name: "visit with duplicates - should deduplicate",
			initial: OriginationData{
				Capability:    []string{"amazon"},
				AttackSurface: []string{"internal"},
				Origins:       []string{"dns", "amazon"},
			},
			other: OriginationData{
				Capability:    []string{"amazon", "portscan"},
				AttackSurface: []string{"internal", "external"},
				Origins:       []string{"amazon", "ipv4"},
			},
			expected: OriginationData{
				Capability:    []string{"amazon", "portscan"},
				AttackSurface: []string{"external", "internal"},
				Origins:       []string{"amazon", "dns", "ipv4"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.Visit(tt.other)
			
			// Sort slices for comparison since order might vary due to map iteration
			for i := range tt.initial.Capability {
				for j := range tt.initial.Capability {
					if i < j && tt.initial.Capability[i] > tt.initial.Capability[j] {
						tt.initial.Capability[i], tt.initial.Capability[j] = tt.initial.Capability[j], tt.initial.Capability[i]
					}
				}
			}
			for i := range tt.initial.AttackSurface {
				for j := range tt.initial.AttackSurface {
					if i < j && tt.initial.AttackSurface[i] > tt.initial.AttackSurface[j] {
						tt.initial.AttackSurface[i], tt.initial.AttackSurface[j] = tt.initial.AttackSurface[j], tt.initial.AttackSurface[i]
					}
				}
			}
			for i := range tt.initial.Origins {
				for j := range tt.initial.Origins {
					if i < j && tt.initial.Origins[i] > tt.initial.Origins[j] {
						tt.initial.Origins[i], tt.initial.Origins[j] = tt.initial.Origins[j], tt.initial.Origins[i]
					}
				}
			}

			// Sort expected for comparison too
			for i := range tt.expected.Capability {
				for j := range tt.expected.Capability {
					if i < j && tt.expected.Capability[i] > tt.expected.Capability[j] {
						tt.expected.Capability[i], tt.expected.Capability[j] = tt.expected.Capability[j], tt.expected.Capability[i]
					}
				}
			}
			for i := range tt.expected.AttackSurface {
				for j := range tt.expected.AttackSurface {
					if i < j && tt.expected.AttackSurface[i] > tt.expected.AttackSurface[j] {
						tt.expected.AttackSurface[i], tt.expected.AttackSurface[j] = tt.expected.AttackSurface[j], tt.expected.AttackSurface[i]
					}
				}
			}
			for i := range tt.expected.Origins {
				for j := range tt.expected.Origins {
					if i < j && tt.expected.Origins[i] > tt.expected.Origins[j] {
						tt.expected.Origins[i], tt.expected.Origins[j] = tt.expected.Origins[j], tt.expected.Origins[i]
					}
				}
			}

			if !reflect.DeepEqual(tt.initial, tt.expected) {
				t.Errorf("OriginationData.Visit() = %v, expected %v", tt.initial, tt.expected)
			}
		})
	}
}