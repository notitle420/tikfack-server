package repository

import (
	"context"
	"errors"

	"github.com/tikfack/server/internal/domain/entity"
)

// FavoriteVideoRepository defines persistence behavior for favorite videos.
type FavoriteVideoRepository interface {
	Add(ctx context.Context, favorite *entity.FavoriteVideo) error
	RemoveByVideoID(ctx context.Context, userID, videoID string) (*entity.FavoriteVideo, error)
	FindByUserAndVideoID(ctx context.Context, userID, videoID string) (*entity.FavoriteVideo, error)
	ListByUserID(ctx context.Context, userID string) ([]entity.FavoriteVideo, error)
}

var (
	// ErrFavoriteVideoNotFound indicates the requested favorite video could not be located.
	ErrFavoriteVideoNotFound = errors.New("favorite video not found")
)
