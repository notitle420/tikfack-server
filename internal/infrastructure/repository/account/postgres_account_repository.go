package account

import (
	"context"
	"database/sql"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// PostgresAccountRepository persists accounts to the users table.
type PostgresAccountRepository struct {
	db *sql.DB
}

// NewPostgresAccountRepository constructs a Postgres-backed account repository.
func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

// UpsertByKeycloakID inserts or updates a user record identified by the Keycloak subject.
func (r *PostgresAccountRepository) UpsertByKeycloakID(ctx context.Context, keycloakID string) (*entity.Account, error) {
	if _, err := entity.NewAccountFromKeycloak(keycloakID); err != nil {
		return nil, err
	}

	query := `
INSERT INTO users (user_id, created_at, updated_at)
VALUES ($1, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
RETURNING user_id, created_at, updated_at
`

	row := r.db.QueryRowContext(ctx, query, keycloakID)
	account := &entity.Account{KeycloakID: keycloakID}
	if err := row.Scan(&account.UserID, &account.CreatedAt, &account.UpdatedAt); err != nil {
		return nil, err
	}

	return account, nil
}

// GetByUserID retrieves a persisted account by its user ID.
func (r *PostgresAccountRepository) GetByUserID(ctx context.Context, userID string) (*entity.Account, error) {
	query := `
SELECT user_id, created_at, updated_at
FROM users
WHERE user_id = $1
`

	row := r.db.QueryRowContext(ctx, query, userID)
	account := &entity.Account{KeycloakID: userID}
	if err := row.Scan(&account.UserID, &account.CreatedAt, &account.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrAccountNotFound
		}
		return nil, err
	}

	return account, nil
}

// ensure interface compliance
var _ repository.AccountRepository = (*PostgresAccountRepository)(nil)
