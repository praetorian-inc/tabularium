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

func NewUser(username string, accounts []Account) User {
	return User{
		Name:       username,
		HomeTenant: username,
		Accounts:   accounts,
	}
}
