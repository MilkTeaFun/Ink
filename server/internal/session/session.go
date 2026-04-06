package session

import "time"

type ClientType string

const (
	ClientTypeWeb ClientType = "web"
)

type Session struct {
	ID               string
	FamilyID         string
	UserID           string
	RefreshTokenHash string
	ClientType       ClientType
	UserAgent        string
	IPAddress        string
	ExpiresAt        time.Time
	RotatedAt        *time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
	LastUsedAt       time.Time
}
