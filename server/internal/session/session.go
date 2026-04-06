package session

import "time"

// ClientType identifies the client platform that created a session.
type ClientType string

const (
	// ClientTypeWeb is used for browser-based sessions.
	ClientTypeWeb ClientType = "web"
)

// Session stores refresh-token state for a signed-in client.
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
