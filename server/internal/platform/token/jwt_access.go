package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/user"
)

type JWTAccessManager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

type claims struct {
	Type      string `json:"typ"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

func NewJWTAccessManager(secret string, issuer string, ttl time.Duration) *JWTAccessManager {
	return &JWTAccessManager{
		secret: []byte(secret),
		ttl:    ttl,
		issuer: issuer,
	}
}

func (m *JWTAccessManager) Issue(account user.User, sessionID string, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(m.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Type:      "access",
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account.ID,
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signed, expiresAt, nil
}

func (m *JWTAccessManager) Parse(raw string) (*auth.AccessClaims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer))
	if err != nil {
		return nil, err
	}

	typed, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid || typed.Type != "access" {
		return nil, errors.New("invalid access token")
	}

	return &auth.AccessClaims{
		UserID:    typed.Subject,
		SessionID: typed.SessionID,
	}, nil
}
