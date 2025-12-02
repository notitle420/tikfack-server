package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Account represents a user record tied to a Keycloak subject.
type Account struct {
	UserID     string
	KeycloakID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewAccountFromKeycloak constructs a new Account using the Keycloak subject as the identity source.
func NewAccountFromKeycloak(keycloakID string) (*Account, error) {
	if keycloakID == "" {
		return nil, fmt.Errorf("keycloak id is required")
	}
	if _, err := uuid.Parse(keycloakID); err != nil {
		return nil, fmt.Errorf("keycloak id must be uuid: %w", err)
	}

	now := time.Now().UTC()
	return &Account{
		UserID:     keycloakID,
		KeycloakID: keycloakID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
