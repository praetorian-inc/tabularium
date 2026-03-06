package model

import "testing"

func TestRoleValid(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleAdmin, true},
		{RoleAnalyst, true},
		{RoleReadOnly, true},
		{"", false},
		{"superadmin", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := tt.role.Valid(); got != tt.want {
				t.Errorf("Role(%q).Valid() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestRoleAtLeast(t *testing.T) {
	tests := []struct {
		role    Role
		minimum Role
		want    bool
	}{
		// Admin can do everything
		{RoleAdmin, RoleAdmin, true},
		{RoleAdmin, RoleAnalyst, true},
		{RoleAdmin, RoleReadOnly, true},

		// Analyst can do analyst and readonly
		{RoleAnalyst, RoleAdmin, false},
		{RoleAnalyst, RoleAnalyst, true},
		{RoleAnalyst, RoleReadOnly, true},

		// ReadOnly can only do readonly
		{RoleReadOnly, RoleAdmin, false},
		{RoleReadOnly, RoleAnalyst, false},
		{RoleReadOnly, RoleReadOnly, true},

		// Invalid role can't do anything
		{"", RoleReadOnly, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.role)+"_atleast_"+string(tt.minimum), func(t *testing.T) {
			if got := tt.role.AtLeast(tt.minimum); got != tt.want {
				t.Errorf("Role(%q).AtLeast(%q) = %v, want %v", tt.role, tt.minimum, got, tt.want)
			}
		})
	}
}

func TestEffectiveRole(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want Role
	}{
		{"empty defaults to admin", "", RoleAdmin},
		{"admin stays admin", RoleAdmin, RoleAdmin},
		{"analyst stays analyst", RoleAnalyst, RoleAnalyst},
		{"readonly stays readonly", RoleReadOnly, RoleReadOnly},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{Role: tt.role}
			if got := a.EffectiveRole(); got != tt.want {
				t.Errorf("EffectiveRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleForAccount(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		accountName string
		want        Role
	}{
		{
			name: "finds matching account role",
			user: User{
				Name:       "user@example.com",
				HomeTenant: "user@example.com",
				Accounts: []Account{
					{Name: "customer@example.com", Member: "user@example.com", Role: RoleAnalyst},
				},
			},
			accountName: "customer@example.com",
			want:        RoleAnalyst,
		},
		{
			name: "empty role defaults to admin",
			user: User{
				Name:       "user@example.com",
				HomeTenant: "user@example.com",
				Accounts: []Account{
					{Name: "customer@example.com", Member: "user@example.com", Role: ""},
				},
			},
			accountName: "customer@example.com",
			want:        RoleAdmin,
		},
		{
			name: "unknown account returns readonly",
			user: User{
				Name:       "user@example.com",
				HomeTenant: "user@example.com",
				Accounts:   []Account{},
			},
			accountName: "unknown@example.com",
			want:        RoleReadOnly,
		},
		{
			name: "uses HomeTenant for member matching",
			user: User{
				Name:       "api-key-id",
				HomeTenant: "user@example.com",
				Accounts: []Account{
					{Name: "customer@example.com", Member: "user@example.com", Role: RoleReadOnly},
				},
			},
			accountName: "customer@example.com",
			want:        RoleReadOnly,
		},
		{
			name: "falls back to Name when HomeTenant empty",
			user: User{
				Name:       "user@example.com",
				HomeTenant: "",
				Accounts: []Account{
					{Name: "customer@example.com", Member: "user@example.com", Role: RoleAnalyst},
				},
			},
			accountName: "customer@example.com",
			want:        RoleAnalyst,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.RoleForAccount(tt.accountName); got != tt.want {
				t.Errorf("RoleForAccount(%q) = %v, want %v", tt.accountName, got, tt.want)
			}
		})
	}
}
