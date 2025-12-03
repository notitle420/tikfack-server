package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// User represents a user record tied to a Keycloak subject.
// It is intentionally minimal and driven by Keycloak as the source of truth.
type User struct {
	UserID     string
	KeycloakID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewUserFromKeycloak constructs a new User using the Keycloak subject as the identity source.
func NewUserFromKeycloak(keycloakID string) (*User, error) {
	if keycloakID == "" {
		return nil, fmt.Errorf("keycloak id is required")
	}
	if _, err := uuid.Parse(keycloakID); err != nil {
		return nil, fmt.Errorf("keycloak id must be uuid: %w", err)
	}

	now := time.Now().UTC()
	return &User{
		UserID:     keycloakID,
		KeycloakID: keycloakID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
