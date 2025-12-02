package favorite

import (
	"context"
	"database/sql"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// PostgresFavoriteActorRepository stores favorite actors in Postgres.
type PostgresFavoriteActorRepository struct {
	db *sql.DB
}

// NewPostgresFavoriteActorRepository creates a new PostgresFavoriteActorRepository.
func NewPostgresFavoriteActorRepository(db *sql.DB) *PostgresFavoriteActorRepository {
	return &PostgresFavoriteActorRepository{db: db}
}

// Add inserts a favorite actor, returning the existing record if already present.
func (r *PostgresFavoriteActorRepository) Add(ctx context.Context, favorite *entity.FavoriteActor) error {
	query := `
INSERT INTO favorite_actors (favorite_actor_uuid, user_id, actor_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (user_id, actor_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
RETURNING favorite_actor_uuid, user_id, actor_id, created_at, updated_at
`

	return r.scanFavorite(r.db.QueryRowContext(ctx, query, favorite.FavoriteActorUUID, favorite.UserID, favorite.ActorID), favorite)
}

// RemoveByActorID deletes a favorite actor for a user.
func (r *PostgresFavoriteActorRepository) RemoveByActorID(ctx context.Context, userID, actorID string) (*entity.FavoriteActor, error) {
	query := `
DELETE FROM favorite_actors
WHERE user_id = $1 AND actor_id = $2
RETURNING favorite_actor_uuid, user_id, actor_id, created_at, updated_at
`
	row := r.db.QueryRowContext(ctx, query, userID, actorID)
	favorite := &entity.FavoriteActor{}
	if err := r.scanFavorite(row, favorite); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrFavoriteActorNotFound
		}
		return nil, err
	}
	return favorite, nil
}

// FindByUserAndActorID returns a favorite actor when it exists.
func (r *PostgresFavoriteActorRepository) FindByUserAndActorID(ctx context.Context, userID, actorID string) (*entity.FavoriteActor, error) {
	query := `
SELECT favorite_actor_uuid, user_id, actor_id, created_at, updated_at
FROM favorite_actors
WHERE user_id = $1 AND actor_id = $2
`
	row := r.db.QueryRowContext(ctx, query, userID, actorID)
	favorite := &entity.FavoriteActor{}
	if err := r.scanFavorite(row, favorite); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrFavoriteActorNotFound
		}
		return nil, err
	}
	return favorite, nil
}

// ListByUserID returns all favorite actors for a user.
func (r *PostgresFavoriteActorRepository) ListByUserID(ctx context.Context, userID string) ([]entity.FavoriteActor, error) {
	query := `
SELECT favorite_actor_uuid, user_id, actor_id, created_at, updated_at
FROM favorite_actors
WHERE user_id = $1
ORDER BY created_at ASC
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []entity.FavoriteActor
	for rows.Next() {
		var favorite entity.FavoriteActor
		if err := rows.Scan(&favorite.FavoriteActorUUID, &favorite.UserID, &favorite.ActorID, &favorite.CreatedAt, &favorite.UpdatedAt); err != nil {
			return nil, err
		}
		favorites = append(favorites, favorite)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return favorites, nil
}

func (r *PostgresFavoriteActorRepository) scanFavorite(row *sql.Row, favorite *entity.FavoriteActor) error {
	return row.Scan(&favorite.FavoriteActorUUID, &favorite.UserID, &favorite.ActorID, &favorite.CreatedAt, &favorite.UpdatedAt)
}

// ensure interface compliance
var _ repository.FavoriteActorRepository = (*PostgresFavoriteActorRepository)(nil)
