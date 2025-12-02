package repository

import (
	"context"
	"errors"

	"github.com/tikfack/server/internal/domain/entity"
)

// FavoriteActorRepository defines persistence behavior for favorite actors.
type FavoriteActorRepository interface {
	Add(ctx context.Context, favorite *entity.FavoriteActor) error
	RemoveByActorID(ctx context.Context, userID, actorID string) (*entity.FavoriteActor, error)
	FindByUserAndActorID(ctx context.Context, userID, actorID string) (*entity.FavoriteActor, error)
	ListByUserID(ctx context.Context, userID string) ([]entity.FavoriteActor, error)
}

var (
	// ErrFavoriteActorNotFound indicates the requested favorite actor could not be located.
	ErrFavoriteActorNotFound = errors.New("favorite actor not found")
)
