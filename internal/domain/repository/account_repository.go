package repository

import (
	"context"
	"fmt"

	"github.com/tikfack/server/internal/domain/entity"
)

// AccountRepository manages user records linked to Keycloak subjects.
type AccountRepository interface {
	UpsertByKeycloakID(ctx context.Context, keycloakID string) (*entity.Account, error)
	GetByUserID(ctx context.Context, userID string) (*entity.Account, error)
}

var (
	// ErrAccountNotFound is returned when no account exists for the given identifier.
	ErrAccountNotFound = fmt.Errorf("account not found")
)
