package user

import "time"

// Status describes whether a user can authenticate.
type Status string

const (
	// StatusActive marks an account that can sign in normally.
	StatusActive Status = "active"
	// StatusDisabled marks an account that is blocked from authentication.
	StatusDisabled Status = "disabled"
)

// User represents an authenticated account persisted in storage.
type User struct {
	ID           string
	Email        string
	DisplayName  string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
	PasswordHash string
}
