package user

import (
	"context"
	"database/sql"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// PostgresUserRepository persists users to the users table.
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository constructs a Postgres-backed user repository.
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// UpsertByKeycloakID inserts or updates a user record identified by the Keycloak subject.
func (r *PostgresUserRepository) UpsertByKeycloakID(ctx context.Context, keycloakID string) (*entity.User, error) {
	if _, err := entity.NewUserFromKeycloak(keycloakID); err != nil {
		return nil, err
	}

	query := `
INSERT INTO users (user_id, created_at, updated_at)
VALUES ($1, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
RETURNING user_id, created_at, updated_at
`

	row := r.db.QueryRowContext(ctx, query, keycloakID)
	user := &entity.User{KeycloakID: keycloakID}
	if err := row.Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}

	return user, nil
}

// GetByUserID retrieves a persisted user by its user ID.
func (r *PostgresUserRepository) GetByUserID(ctx context.Context, userID string) (*entity.User, error) {
	query := `
SELECT user_id, created_at, updated_at
FROM users
WHERE user_id = $1
`

	row := r.db.QueryRowContext(ctx, query, userID)
	user := &entity.User{KeycloakID: userID}
	if err := row.Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// ensure interface compliance
var _ repository.UserRepository = (*PostgresUserRepository)(nil)
