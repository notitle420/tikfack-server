package favorite

import (
	"context"
	"database/sql"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// PostgresFavoriteVideoRepository stores favorite videos in Postgres.
type PostgresFavoriteVideoRepository struct {
	db *sql.DB
}

// NewPostgresFavoriteVideoRepository creates a new PostgresFavoriteVideoRepository.
func NewPostgresFavoriteVideoRepository(db *sql.DB) *PostgresFavoriteVideoRepository {
	return &PostgresFavoriteVideoRepository{db: db}
}

// Add inserts a favorite video, returning the existing record if already present.
func (r *PostgresFavoriteVideoRepository) Add(ctx context.Context, favorite *entity.FavoriteVideo) error {
	query := `
INSERT INTO favorite_videos (favorite_video_uuid, user_id, video_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (user_id, video_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
RETURNING favorite_video_uuid, user_id, video_id, created_at, updated_at
`

	return r.scanFavorite(r.db.QueryRowContext(ctx, query, favorite.FavoriteVideoUUID, favorite.UserID, favorite.VideoID), favorite)
}

// RemoveByVideoID deletes a favorite video for a user.
func (r *PostgresFavoriteVideoRepository) RemoveByVideoID(ctx context.Context, userID, videoID string) (*entity.FavoriteVideo, error) {
	query := `
DELETE FROM favorite_videos
WHERE user_id = $1 AND video_id = $2
RETURNING favorite_video_uuid, user_id, video_id, created_at, updated_at
`
	row := r.db.QueryRowContext(ctx, query, userID, videoID)
	favorite := &entity.FavoriteVideo{}
	if err := r.scanFavorite(row, favorite); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrFavoriteVideoNotFound
		}
		return nil, err
	}
	return favorite, nil
}

// FindByUserAndVideoID returns a favorite video when it exists.
func (r *PostgresFavoriteVideoRepository) FindByUserAndVideoID(ctx context.Context, userID, videoID string) (*entity.FavoriteVideo, error) {
	query := `
SELECT favorite_video_uuid, user_id, video_id, created_at, updated_at
FROM favorite_videos
WHERE user_id = $1 AND video_id = $2
`
	row := r.db.QueryRowContext(ctx, query, userID, videoID)
	favorite := &entity.FavoriteVideo{}
	if err := r.scanFavorite(row, favorite); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrFavoriteVideoNotFound
		}
		return nil, err
	}
	return favorite, nil
}

// ListByUserID returns all favorite videos for a user.
func (r *PostgresFavoriteVideoRepository) ListByUserID(ctx context.Context, userID string) ([]entity.FavoriteVideo, error) {
	query := `
SELECT favorite_video_uuid, user_id, video_id, created_at, updated_at
FROM favorite_videos
WHERE user_id = $1
ORDER BY created_at ASC
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []entity.FavoriteVideo
	for rows.Next() {
		var favorite entity.FavoriteVideo
		if err := rows.Scan(&favorite.FavoriteVideoUUID, &favorite.UserID, &favorite.VideoID, &favorite.CreatedAt, &favorite.UpdatedAt); err != nil {
			return nil, err
		}
		favorites = append(favorites, favorite)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return favorites, nil
}

func (r *PostgresFavoriteVideoRepository) scanFavorite(row *sql.Row, favorite *entity.FavoriteVideo) error {
	return row.Scan(&favorite.FavoriteVideoUUID, &favorite.UserID, &favorite.VideoID, &favorite.CreatedAt, &favorite.UpdatedAt)
}

// ensure interface compliance
var _ repository.FavoriteVideoRepository = (*PostgresFavoriteVideoRepository)(nil)
