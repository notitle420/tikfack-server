package repository

import (
	"context"
	"fmt"

	"github.com/tikfack/server/internal/domain/entity"
)

// UserRepository manages user records linked to Keycloak subjects.
type UserRepository interface {
	UpsertByKeycloakID(ctx context.Context, keycloakID string) (*entity.User, error)
	GetByUserID(ctx context.Context, userID string) (*entity.User, error)
}

var (
	// ErrUserNotFound is returned when no user exists for the given identifier.
	ErrUserNotFound = fmt.Errorf("user not found")
)
