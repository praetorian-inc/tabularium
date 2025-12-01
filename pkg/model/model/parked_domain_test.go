package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	DOMAIN    = "test-domain.com"
	KEY       = "#parkeddomain#" + DOMAIN
	PRINCIPAL = "test@praetorian.com"
)

func TesTParkedDomain_New(t *testing.T) {
	p := NewParkedDomain(DOMAIN)
	assert.Equal(t, DOMAIN, p.Domain)
	assert.Equal(t, p.GetKey(), KEY)
	assert.NotEqual(t, p.Updated, "")
}

func TestParkedDomain_Merge_UpdatesParkedCategory(t *testing.T) {
	original := NewParkedDomain(DOMAIN)
	original.ParkedCategory = "legal"

	update := ParkedDomain{ParkedCategory: "finance"}
	original.Merge(update, []string{"parked_category"})

	assert.Equal(t, "finance", original.ParkedCategory)
	assert.NotEqual(t, original.Updated, "")
}

func TestParkedDomain_Merge_UpdatesCloudflareStatus(t *testing.T) {
	original := NewParkedDomain(DOMAIN)
	original.CloudflareStatus = "active"

	update := ParkedDomain{CloudflareStatus: "pending"}
	original.Merge(update, []string{"cloudflare_status"})

	assert.Equal(t, "pending", original.CloudflareStatus)
}

func TestParkedDomain_Merge_UpdatesStatus(t *testing.T) {
	original := NewParkedDomain(DOMAIN)
	original.Status = "available"

	update := ParkedDomain{Status: "reserved"}
	original.Merge(update, []string{"status"})

	assert.Equal(t, "reserved", original.Status)
}

func TestParkedDomain_Merge_UpdatesCheckoutFieldsWhenStatusInUse(t *testing.T) {
	original := NewParkedDomain(DOMAIN)
	original.Status = "available"

	update := ParkedDomain{
		Status:        "in-use",
		CheckoutStart: "2025-01-01T00:00:00Z",
		CheckoutEnd:   "2025-01-31T23:59:59Z",
		CheckoutUser:  PRINCIPAL,
		CheckoutNote:  "Customer engagement",
	}
	original.Merge(update, []string{"status"})

	assert.Equal(t, "in-use", original.Status)
	assert.Equal(t, "2025-01-01T00:00:00Z", original.CheckoutStart)
	assert.Equal(t, "2025-01-31T23:59:59Z", original.CheckoutEnd)
	assert.Equal(t, PRINCIPAL, original.CheckoutUser)
	assert.Equal(t, "Customer engagement", original.CheckoutNote)
}

func TestParkedDomain_Merge_UpdatesAutoRenew(t *testing.T) {
	original := NewParkedDomain(DOMAIN)
	original.AutoRenew = false

	update := ParkedDomain{AutoRenew: true}
	original.Merge(update, []string{"auto_renew"})

	assert.True(t, original.AutoRenew)
}

func TestParkedDomain_Merge_DoesNotUpdateAutoRenewWhenNotInList(t *testing.T) {
	original := NewParkedDomain(DOMAIN)
	original.AutoRenew = true

	update := ParkedDomain{AutoRenew: false, Status: "available"}
	original.Merge(update, []string{"status"}) // auto_renew not in list

	assert.True(t, original.AutoRenew, "AutoRenew should not change when not in fieldsToUpdate")
	assert.Equal(t, "available", original.Status)
}

func TestParkedDomain_Merge_WithMultipleFields(t *testing.T) {
	original := NewParkedDomain(DOMAIN)
	original.ParkedCategory = "legal"
	original.Status = "available"
	original.AutoRenew = false

	update := ParkedDomain{
		ParkedCategory: "finance",
		Status:         "in-use",
		CheckoutStart:  "2025-01-01T00:00:00Z",
		CheckoutEnd:    "2025-01-31T23:59:59Z",
		CheckoutUser:   "user@example.com",
		CheckoutNote:   "SOW-2025-001",
		AutoRenew:      true,
	}
	original.Merge(update, []string{"parked_category", "status", "auto_renew"})

	assert.Equal(t, "finance", original.ParkedCategory)
	assert.Equal(t, "in-use", original.Status)
	assert.Equal(t, "2025-01-01T00:00:00Z", original.CheckoutStart, "CheckoutStart auto-updated with status=in-use")
	assert.Equal(t, "2025-01-31T23:59:59Z", original.CheckoutEnd, "CheckoutEnd auto-updated with status=in-use")
	assert.Equal(t, "user@example.com", original.CheckoutUser, "CheckoutUser auto-updated with status=in-use")
	assert.Equal(t, "SOW-2025-001", original.CheckoutNote, "CheckoutNote auto-updated with status=in-use")
	assert.True(t, original.AutoRenew)
}

func TestParkedDomain_Merge_OnlyUpdatesSpecifiedFields(t *testing.T) {
	original := NewParkedDomain(DOMAIN)
	original.ParkedCategory = "legal"
	original.Status = "available"
	original.CheckoutNote = "Original note"

	update := ParkedDomain{
		ParkedCategory: "finance",
		Status:         "reserved",
		CheckoutNote:   "New note",
	}
	original.Merge(update, []string{"status"}) // only update status

	assert.Equal(t, "legal", original.ParkedCategory, "Should not update - not in fieldsToUpdate")
	assert.Equal(t, "reserved", original.Status, "Should update - in fieldsToUpdate")
	assert.Equal(t, "Original note", original.CheckoutNote, "Should not update - not in fieldsToUpdate")
}
