package user

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

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
