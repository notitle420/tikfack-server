package user

import (
	"context"
	"sync"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// MemoryUserRepository stores users in memory for simplicity.
type MemoryUserRepository struct {
	mu           sync.RWMutex
	byKeycloakID map[string]*entity.User
	byUserID     map[string]*entity.User
}

// NewMemoryUserRepository constructs a new in-memory user repository.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		byKeycloakID: make(map[string]*entity.User),
		byUserID:     make(map[string]*entity.User),
	}
}

// UpsertByKeycloakID returns an existing user or creates a new one if absent.
func (r *MemoryUserRepository) UpsertByKeycloakID(ctx context.Context, keycloakID string) (*entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, ok := r.byKeycloakID[keycloakID]; ok {
		user.UpdatedAt = time.Now().UTC()
		return user, nil
	}

	user, err := entity.NewUserFromKeycloak(keycloakID)
	if err != nil {
		return nil, err
	}
	r.byKeycloakID[keycloakID] = user
	r.byUserID[user.UserID] = user
	return user, nil
}

// GetByUserID returns the user by user ID.
func (r *MemoryUserRepository) GetByUserID(ctx context.Context, userID string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.byUserID[userID]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}
