package model

import (
	"testing"
)

func TestNewPreseed(t *testing.T) {
	tests := []struct {
		name        string
		identifier  string
		value       string
		expectedKey string
	}{
		{"whois+email", "email", "gladiator@praetorian.com", "#preseed#whois+email#email#gladiator@praetorian.com"},
		{"whois+company", "company", "Chariot Systems", "#preseed#whois+company#company#Chariot Systems"},
		{"whois+name", "name", "Gladiator", "#preseed#whois+name#name#Gladiator"},
		{"csp", "1d38406f03ff445d25b48cbfbdd85b1a", "ZGVmYXVsdC1zcmMgJ3NlbGYnIHdpbjAuY29sb3NzZXVtLnN5c3RlbXMgd2luMS5jb2xvc3NldW0uc3lzdGVtcyB3aW4yLmNvbG9zc2V1bS5zeXN0ZW1z", "#preseed#csp#1d38406f03ff445d25b48cbfbdd85b1a#ZGVmYXVsdC1zcmMgJ3NlbGYnIHdpbjAuY29sb3NzZXVtLnN5c3RlbXMgd2luMS5jb2xvc3NldW0uc3lzdGVtcyB3aW4yLmNvbG9zc2V1bS5zeXN0ZW1z"},
		{"favicon", "", "https://www.praetorian.com/wp-content/uploads/2024/06/cropped-cropped-Praetorian-Favicon-192x192.png", "#preseed#favicon##https://www.praetorian.com/wp-content/uploads/2024/06/cropped-cropped-Praetorian-Favicon-192x192.png"},
	}

	for _, test := range tests {
		preseed := NewPreseed(test.name, test.identifier, test.value)
		if preseed.Key != test.expectedKey {
			t.Errorf("unexpected key %s, expected %s", preseed.Key, test.expectedKey)
		}
	}
}

func TestPreseedClass(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"whois+email", "whois"},
		{"csp", "csp"},
		{"favicon", "favicon"},
	}

	for _, test := range tests {
		preseed := NewPreseed(test.name, "", "")
		if preseed.Class() != test.want {
			t.Errorf("unexpected label %s, expected %s", preseed.Class(), test.want)
		}
	}
}

func TestPreseedAttribute(t *testing.T) {
	preseed := NewPreseed("whois", "email", "gladiator@praetorian.com")
	source := NewAsset("gladiator.systems", "54.89.228.191")
	attribute := preseed.ToAttribute(&source)

	expectedKey := "#attribute#preseed##preseed#whois#email#gladiator@praetorian.com" + source.GetKey()
	if attribute.Key != expectedKey {
		t.Errorf("unexpected key %s, expected %s", attribute.Key, expectedKey)
	}
}

func TestGeneratePreseedDisplay(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"whois", "text"},
		{"csp", "base64"},
		{"favicon", "image"},
		{"tlscert", "tlscert"},
		{"default", "text"},
	}

	for _, test := range tests {
		if got := generatePreseedDisplay(test.name); got != test.want {
			t.Errorf("generatePreseedDisplay(%s) = %s, want %s", test.name, got, test.want)
		}
	}
}
