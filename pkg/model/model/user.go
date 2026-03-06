package model

import (
	"strings"
)

type User struct {
	Name       string
	HomeTenant string // the home tenant of the user
	Accounts   []Account
}

func (u *User) Linked(username string) bool {
	for _, account := range u.Accounts {
		if account.Name == username && account.Member == u.Name {
			return true
		}
	}
	return false
}

func (u *User) Domain() string {
	parts := strings.Split(u.Name, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

func (u *User) Praetorian() bool {
	return strings.HasSuffix(u.Name, "@praetorian.com")
}

// RoleForAccount returns the user's effective role for the given account.
// It uses HomeTenant (critical for SSO users whose Name is the access key, not email).
func (u *User) RoleForAccount(accountName string) Role {
	tenant := u.HomeTenant
	if tenant == "" {
		tenant = u.Name
	}
	for _, account := range u.Accounts {
		if account.Name == accountName && account.Member == tenant {
			return account.EffectiveRole()
		}
	}
	return RoleReadOnly // safest default for unknown
}

func NewUser(username string, accounts []Account) User {
	return User{
		Name:       username,
		HomeTenant: username,
		Accounts:   accounts,
	}
}
