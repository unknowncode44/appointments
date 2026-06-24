package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	ID        uuid.UUID  `json:"id"`
	UserID    int32      `json:"user_id"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	IssuedAt  time.Time  `json:"issued_at"`
	ExpiredAt time.Time  `json:"expired_at"`
	jwt.RegisteredClaims
}

// NewPayload creates a new token payload with a specific username and duration
func NewPayload(username, role string, userID int32, tenantID *uuid.UUID, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	payload := &Payload{
		ID:        tokenID,
		UserID:    userID,
		Username:  username,
		Role:      role,
		TenantID:  tenantID,
		IssuedAt:  now,
		ExpiredAt: now.Add(duration),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	return payload, nil
}

// Valid checks if the token payload is valid or not
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return errors.New("token has expired")
	}
	if payload.IssuedAt.After(time.Now()) {
		return errors.New("token issued in the future")
	}
	return nil
}
