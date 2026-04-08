package user

import "time"

// Status describes whether a user can authenticate.
type Status string

// Role describes the permissions granted to an account.
type Role string

const (
	// StatusActive marks an account that can sign in normally.
	StatusActive Status = "active"
	// StatusDisabled marks an account that is blocked from authentication.
	StatusDisabled Status = "disabled"

	// RoleMember marks a regular account.
	RoleMember Role = "member"
	// RoleAdmin marks an administrator account.
	RoleAdmin Role = "admin"
)

// User represents an authenticated account persisted in storage.
type User struct {
	ID           string
	Email        string
	DisplayName  string
	Role         Role
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
	PasswordHash string
}
