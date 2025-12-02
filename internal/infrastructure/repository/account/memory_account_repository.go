package account

import (
	"context"
	"sync"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// MemoryAccountRepository stores accounts in memory for simplicity.
type MemoryAccountRepository struct {
	mu           sync.RWMutex
	byKeycloakID map[string]*entity.Account
	byUserID     map[string]*entity.Account
}

// NewMemoryAccountRepository constructs a new in-memory account repository.
func NewMemoryAccountRepository() *MemoryAccountRepository {
	return &MemoryAccountRepository{
		byKeycloakID: make(map[string]*entity.Account),
		byUserID:     make(map[string]*entity.Account),
	}
}

// UpsertByKeycloakID returns an existing account or creates a new one if absent.
func (r *MemoryAccountRepository) UpsertByKeycloakID(ctx context.Context, keycloakID string) (*entity.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if account, ok := r.byKeycloakID[keycloakID]; ok {
		account.UpdatedAt = time.Now().UTC()
		return account, nil
	}

	account, err := entity.NewAccountFromKeycloak(keycloakID)
	if err != nil {
		return nil, err
	}
	r.byKeycloakID[keycloakID] = account
	r.byUserID[account.UserID] = account
	return account, nil
}

// GetByUserID returns the account by user ID.
func (r *MemoryAccountRepository) GetByUserID(ctx context.Context, userID string) (*entity.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, ok := r.byUserID[userID]
	if !ok {
		return nil, repository.ErrAccountNotFound
	}
	return account, nil
}
