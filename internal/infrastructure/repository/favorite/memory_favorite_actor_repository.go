package favorite

import (
	"context"
	"sync"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// MemoryFavoriteActorRepository provides in-memory storage for favorite actors.
type MemoryFavoriteActorRepository struct {
	mu           sync.RWMutex
	actorsByUser map[string]map[string]*entity.FavoriteActor
}

// NewMemoryFavoriteActorRepository constructs a new actor repository instance.
func NewMemoryFavoriteActorRepository() *MemoryFavoriteActorRepository {
	return &MemoryFavoriteActorRepository{
		actorsByUser: make(map[string]map[string]*entity.FavoriteActor),
	}
}

// Add adds a favorite actor.
func (r *MemoryFavoriteActorRepository) Add(ctx context.Context, favorite *entity.FavoriteActor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userActors, ok := r.actorsByUser[favorite.UserID]
	if !ok {
		userActors = make(map[string]*entity.FavoriteActor)
		r.actorsByUser[favorite.UserID] = userActors
	}
	userActors[favorite.ActorID] = favorite
	return nil
}

// RemoveByActorID removes a favorite actor entry and returns it.
func (r *MemoryFavoriteActorRepository) RemoveByActorID(ctx context.Context, userID, actorID string) (*entity.FavoriteActor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	userActors, ok := r.actorsByUser[userID]
	if !ok {
		return nil, repository.ErrFavoriteActorNotFound
	}
	favorite, ok := userActors[actorID]
	if !ok {
		return nil, repository.ErrFavoriteActorNotFound
	}
	delete(userActors, actorID)
	return favorite, nil
}

// FindByUserAndActorID finds a favorite actor entry.
func (r *MemoryFavoriteActorRepository) FindByUserAndActorID(ctx context.Context, userID, actorID string) (*entity.FavoriteActor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if userActors, ok := r.actorsByUser[userID]; ok {
		if favorite, ok := userActors[actorID]; ok {
			return favorite, nil
		}
	}
	return nil, repository.ErrFavoriteActorNotFound
}

// ListByUserID lists favorite actors for a user.
func (r *MemoryFavoriteActorRepository) ListByUserID(ctx context.Context, userID string) ([]entity.FavoriteActor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userActors, ok := r.actorsByUser[userID]
	if !ok {
		return []entity.FavoriteActor{}, nil
	}
	result := make([]entity.FavoriteActor, 0, len(userActors))
	for _, favorite := range userActors {
		result = append(result, *favorite)
	}
	return result, nil
}
