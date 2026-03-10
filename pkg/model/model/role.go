package model

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleAnalyst  Role = "analyst"
	RoleReadOnly Role = "readonly"
)

func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleAnalyst, RoleReadOnly:
		return true
	}
	return false
}

// AtLeast returns true if r has at least the permissions of minimum.
// Admin > Analyst > ReadOnly
func (r Role) AtLeast(minimum Role) bool {
	return roleRank(r) >= roleRank(minimum)
}

func roleRank(r Role) int {
	switch r {
	case RoleAdmin:
		return 3
	case RoleAnalyst:
		return 2
	case RoleReadOnly:
		return 1
	default:
		return 0
	}
}
