package model

import (
	"testing"
)

func TestTechnology_NewTechnology(t *testing.T) {
	tests := []struct {
		name    string
		cpe     string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid CPE",
			cpe:     "cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:sp2:*:*:*:*",
			want:    "#technology#cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:sp2:*:*:*:*",
			wantErr: false,
		},
		{
			name:    "Valid CPE with spaces",
			cpe:     "   cpe:2.3:a:microsoft:internet_explorer:8.0.6001:*:*:*:*:*:*:*    ",
			want:    "#technology#cpe:2.3:a:microsoft:internet_explorer:8.0.6001:*:*:*:*:*:*:*",
			wantErr: false,
		},
		{
			name:    "Invalid CPE",
			cpe:     "invalid_cpe",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTechnology(tt.cpe)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTechnology(%q) error = %v, wantErr %v", tt.cpe, err, tt.wantErr)
				return
			}
			if got.Key != tt.want {
				t.Errorf("NewTechnology(%q).Key = %q, want %q", tt.cpe, got.Key, tt.want)
			}
		})
	}
}

func TestTechnology_Proof(t *testing.T) {
	technology, _ := NewTechnology("cpe:2.3:a:microsoft:internet_explorer:8.0.6001:*:*:*:*:*:*:*")
	asset := NewAsset("example.com", "1.1.1.1")
	expectedName := "proofs/cpe:2.3:a:microsoft:internet_explorer:8.0.6001:*:*:*:*:*:*:*/example.com/1.1.1.1/tcp/80"

	proof := technology.Proof([]byte("proof"), &asset, "tcp", "80")

	if proof.Name != expectedName {
		t.Errorf("Proof() = %q, want %q", proof.Name, expectedName)
	}
	if len(proof.Bytes) == 0 {
		t.Errorf("Proof().Bytes = %q, want non-empty", proof.Bytes)
	}
}

func TestTechnology_Valid(t *testing.T) {
	tests := []struct {
		name string
		cpe  string
		want bool
	}{
		{
			name: "Valid CPE",
			cpe:  "#technology#cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:sp2:*:*:*:*",
			want: true,
		},
		{
			name: "Too short CPE",
			cpe:  "#technology#cpe:2.3:a:microsoft:internet_explorer",
			want: false,
		},
		{
			name: "Too long CPE",
			cpe:  "#technology#cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:sp2:*:*:*:*:*:*",
			want: false,
		},
		{
			name: "Invalid key",
			cpe:  "invalid_key",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technology := Technology{Key: tt.cpe}
			if got := technology.Valid(); got != tt.want {
				t.Errorf("Technology.Valid() = %v, want %v, cpe: %v", got, tt.want, tt.cpe)
			}
		})
	}
}
