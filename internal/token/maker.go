package token

import (
	"time"

	"github.com/google/uuid"
)

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new token for a specific username, role, userID, tenant and duration
	CreateToken(username, role string, userID int32, tenantID *uuid.UUID, duration time.Duration) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}
