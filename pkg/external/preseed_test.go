package external

import (
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

func TestPreseed_ToTarget(t *testing.T) {
	tests := []struct {
		name    string
		preseed Preseed
		wantErr bool
	}{
		{
			name: "valid preseed",
			preseed: Preseed{
				Type:       "whois",
				Title:      "registrant_email",
				Value:      "test@example.com",
				Display:    "text",
				Status:     "A",
				Capability: "whois-lookup",
				Metadata:   map[string]string{"source": "manual"},
			},
			wantErr: false,
		},
		{
			name: "missing value",
			preseed: Preseed{
				Type:  "whois",
				Title: "registrant_email",
			},
			wantErr: true,
		},
		{
			name: "missing type",
			preseed: Preseed{
				Title: "registrant_email",
				Value: "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing title",
			preseed: Preseed{
				Type:  "whois",
				Value: "test@example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := tt.preseed.ToTarget()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if target == nil {
					t.Error("ToTarget() returned nil target")
					return
				}
				preseed, ok := target.(*model.Preseed)
				if !ok {
					t.Errorf("ToTarget() returned wrong type: %T", target)
					return
				}
				if preseed.Type != tt.preseed.Type {
					t.Errorf("Type = %v, want %v", preseed.Type, tt.preseed.Type)
				}
				if preseed.Title != tt.preseed.Title {
					t.Errorf("Title = %v, want %v", preseed.Title, tt.preseed.Title)
				}
				if preseed.Value != tt.preseed.Value {
					t.Errorf("Value = %v, want %v", preseed.Value, tt.preseed.Value)
				}
			}
		})
	}
}

func TestPreseed_ToModel(t *testing.T) {
	extPreseed := Preseed{
		Type:       "whois",
		Title:      "registrant_email",
		Value:      "test@example.com",
		Display:    "text",
		Status:     "A",
		Capability: "whois-lookup",
		Metadata:   map[string]string{"source": "manual"},
	}

	modelPreseed, err := extPreseed.ToModel()
	if err != nil {
		t.Fatalf("ToModel() error = %v", err)
	}

	if modelPreseed.Type != extPreseed.Type {
		t.Errorf("Type = %v, want %v", modelPreseed.Type, extPreseed.Type)
	}
	if modelPreseed.Title != extPreseed.Title {
		t.Errorf("Title = %v, want %v", modelPreseed.Title, extPreseed.Title)
	}
	if modelPreseed.Value != extPreseed.Value {
		t.Errorf("Value = %v, want %v", modelPreseed.Value, extPreseed.Value)
	}
	if modelPreseed.Display != extPreseed.Display {
		t.Errorf("Display = %v, want %v", modelPreseed.Display, extPreseed.Display)
	}
	if modelPreseed.Status != extPreseed.Status {
		t.Errorf("Status = %v, want %v", modelPreseed.Status, extPreseed.Status)
	}
	if modelPreseed.Capability != extPreseed.Capability {
		t.Errorf("Capability = %v, want %v", modelPreseed.Capability, extPreseed.Capability)
	}
	if len(modelPreseed.Metadata) != len(extPreseed.Metadata) {
		t.Errorf("Metadata length = %v, want %v", len(modelPreseed.Metadata), len(extPreseed.Metadata))
	}
	if modelPreseed.Key == "" {
		t.Error("Key should be set by NewPreseed")
	}
}

func TestPreseed_TargetInterface(t *testing.T) {
	extPreseed := Preseed{
		Type:  "whois",
		Title: "registrant_email",
		Value: "test@example.com",
	}

	if extPreseed.Group() != "whois" {
		t.Errorf("Group() = %v, want whois", extPreseed.Group())
	}
	if extPreseed.Identifier() != "test@example.com" {
		t.Errorf("Identifier() = %v, want test@example.com", extPreseed.Identifier())
	}

	// Verify it implements Target interface
	var _ Target = extPreseed
}

func TestPreseedFromModel(t *testing.T) {
	modelPreseed := model.NewPreseed("whois", "registrant_email", "test@example.com")
	modelPreseed.Display = "text"
	modelPreseed.Status = "A"
	modelPreseed.Capability = "whois-lookup"
	modelPreseed.Metadata = map[string]string{"source": "manual"}

	extPreseed := PreseedFromModel(&modelPreseed)

	if extPreseed.Type != modelPreseed.Type {
		t.Errorf("Type = %v, want %v", extPreseed.Type, modelPreseed.Type)
	}
	if extPreseed.Title != modelPreseed.Title {
		t.Errorf("Title = %v, want %v", extPreseed.Title, modelPreseed.Title)
	}
	if extPreseed.Value != modelPreseed.Value {
		t.Errorf("Value = %v, want %v", extPreseed.Value, modelPreseed.Value)
	}
	if extPreseed.Display != modelPreseed.Display {
		t.Errorf("Display = %v, want %v", extPreseed.Display, modelPreseed.Display)
	}
	if extPreseed.Status != modelPreseed.Status {
		t.Errorf("Status = %v, want %v", extPreseed.Status, modelPreseed.Status)
	}
	if extPreseed.Capability != modelPreseed.Capability {
		t.Errorf("Capability = %v, want %v", extPreseed.Capability, modelPreseed.Capability)
	}
	if len(extPreseed.Metadata) != len(modelPreseed.Metadata) {
		t.Errorf("Metadata length = %v, want %v", len(extPreseed.Metadata), len(modelPreseed.Metadata))
	}
}

func TestPreseed_DefaultValues(t *testing.T) {
	// Test that NewPreseed applies defaults correctly
	extPreseed := Preseed{
		Type:  "whois",
		Title: "registrant_email",
		Value: "test@example.com",
		// Not setting Display, Status - should get defaults from NewPreseed
	}

	modelPreseed, err := extPreseed.ToModel()
	if err != nil {
		t.Fatalf("ToModel() error = %v", err)
	}

	// NewPreseed should set default Status to "P" (Pending)
	if modelPreseed.Status != "P" {
		t.Errorf("Status = %v, want P (default from NewPreseed)", modelPreseed.Status)
	}

	// NewPreseed should set default Display to "text" for whois
	if modelPreseed.Display != "text" {
		t.Errorf("Display = %v, want text (default from NewPreseed)", modelPreseed.Display)
	}

	// Test overriding defaults
	extPreseedWithOverrides := Preseed{
		Type:    "whois",
		Title:   "registrant_email",
		Value:   "test@example.com",
		Display: "custom",
		Status:  "A",
	}

	modelPreseedWithOverrides, err := extPreseedWithOverrides.ToModel()
	if err != nil {
		t.Fatalf("ToModel() error = %v", err)
	}

	if modelPreseedWithOverrides.Status != "A" {
		t.Errorf("Status = %v, want A (override)", modelPreseedWithOverrides.Status)
	}
	if modelPreseedWithOverrides.Display != "custom" {
		t.Errorf("Display = %v, want custom (override)", modelPreseedWithOverrides.Display)
	}
}
