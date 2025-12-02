package favorite

import (
	"context"
	"sync"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// MemoryFavoriteVideoRepository provides in-memory storage for favorite videos.
type MemoryFavoriteVideoRepository struct {
	mu           sync.RWMutex
	videosByUser map[string]map[string]*entity.FavoriteVideo
}

// NewMemoryFavoriteVideoRepository constructs a new video repository instance.
func NewMemoryFavoriteVideoRepository() *MemoryFavoriteVideoRepository {
	return &MemoryFavoriteVideoRepository{
		videosByUser: make(map[string]map[string]*entity.FavoriteVideo),
	}
}

// Add adds a favorite video.
func (r *MemoryFavoriteVideoRepository) Add(ctx context.Context, favorite *entity.FavoriteVideo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userVideos, ok := r.videosByUser[favorite.UserID]
	if !ok {
		userVideos = make(map[string]*entity.FavoriteVideo)
		r.videosByUser[favorite.UserID] = userVideos
	}
	userVideos[favorite.VideoID] = favorite
	return nil
}

// RemoveByVideoID removes a favorite video entry and returns it.
func (r *MemoryFavoriteVideoRepository) RemoveByVideoID(ctx context.Context, userID, videoID string) (*entity.FavoriteVideo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	userVideos, ok := r.videosByUser[userID]
	if !ok {
		return nil, repository.ErrFavoriteVideoNotFound
	}
	favorite, ok := userVideos[videoID]
	if !ok {
		return nil, repository.ErrFavoriteVideoNotFound
	}
	delete(userVideos, videoID)
	return favorite, nil
}

// FindByUserAndVideoID finds a favorite video entry.
func (r *MemoryFavoriteVideoRepository) FindByUserAndVideoID(ctx context.Context, userID, videoID string) (*entity.FavoriteVideo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if userVideos, ok := r.videosByUser[userID]; ok {
		if favorite, ok := userVideos[videoID]; ok {
			return favorite, nil
		}
	}
	return nil, repository.ErrFavoriteVideoNotFound
}

// ListByUserID lists favorite videos for a user.
func (r *MemoryFavoriteVideoRepository) ListByUserID(ctx context.Context, userID string) ([]entity.FavoriteVideo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userVideos, ok := r.videosByUser[userID]
	if !ok {
		return []entity.FavoriteVideo{}, nil
	}
	result := make([]entity.FavoriteVideo, 0, len(userVideos))
	for _, favorite := range userVideos {
		result = append(result, *favorite)
	}
	return result, nil
}
