package favorite

//go:generate mockgen -destination=../mock/mock_favorite_usecase.go -package=mock github.com/tikfack/server/internal/application/usecase/favorite FavoriteUsecase

import (
	"context"
	"fmt"

	"github.com/tikfack/server/internal/application/model"
	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
	"github.com/tikfack/server/internal/middleware/logger"
)

// FavoriteUsecase defines the operations for managing user favorites.
type FavoriteUsecase interface {
	AddFavoriteVideo(ctx context.Context, keycloakID, videoID string) (*model.FavoriteVideo, error)
	RemoveFavoriteVideo(ctx context.Context, keycloakID, videoID string) (*model.FavoriteVideo, error)
	ListFavoriteVideos(ctx context.Context, keycloakID string) ([]model.FavoriteVideo, error)

	AddFavoriteActor(ctx context.Context, keycloakID, actorID string) (*model.FavoriteActor, error)
	RemoveFavoriteActor(ctx context.Context, keycloakID, actorID string) (*model.FavoriteActor, error)
	ListFavoriteActors(ctx context.Context, keycloakID string) ([]model.FavoriteActor, error)
}

// usecase implements FavoriteUsecase.
type usecase struct {
	accountRepo       repository.AccountRepository
	favoriteVideoRepo repository.FavoriteVideoRepository
	favoriteActorRepo repository.FavoriteActorRepository
}

// NewFavoriteUsecase constructs a FavoriteUsecase.
func NewFavoriteUsecase(
	accountRepo repository.AccountRepository,
	favoriteVideoRepo repository.FavoriteVideoRepository,
	favoriteActorRepo repository.FavoriteActorRepository,
) FavoriteUsecase {
	return &usecase{
		accountRepo:       accountRepo,
		favoriteVideoRepo: favoriteVideoRepo,
		favoriteActorRepo: favoriteActorRepo,
	}
}

func (u *usecase) AddFavoriteVideo(ctx context.Context, keycloakID, videoID string) (*model.FavoriteVideo, error) {
	log := logger.LoggerWithCtx(ctx)
	account, err := u.ensureAccount(ctx, keycloakID)
	if err != nil {
		return nil, err
	}
	existing, err := u.favoriteVideoRepo.FindByUserAndVideoID(ctx, account.UserID, videoID)
	if err == nil && existing != nil {
		log.Debug("favorite video already exists", "video_id", videoID)
		fv := model.NewFavoriteVideoFromEntity(*existing)
		return &fv, nil
	}

	favorite, err := entity.NewFavoriteVideo(account.UserID, videoID)
	if err != nil {
		return nil, err
	}
	if err := u.favoriteVideoRepo.Add(ctx, favorite); err != nil {
		return nil, err
	}
	fv := model.NewFavoriteVideoFromEntity(*favorite)
	return &fv, nil
}

func (u *usecase) RemoveFavoriteVideo(ctx context.Context, keycloakID, videoID string) (*model.FavoriteVideo, error) {
	account, err := u.ensureAccount(ctx, keycloakID)
	if err != nil {
		return nil, err
	}
	removed, err := u.favoriteVideoRepo.RemoveByVideoID(ctx, account.UserID, videoID)
	if err != nil {
		return nil, err
	}
	fv := model.NewFavoriteVideoFromEntity(*removed)
	return &fv, nil
}

func (u *usecase) ListFavoriteVideos(ctx context.Context, keycloakID string) ([]model.FavoriteVideo, error) {
	account, err := u.ensureAccount(ctx, keycloakID)
	if err != nil {
		return nil, err
	}
	favorites, err := u.favoriteVideoRepo.ListByUserID(ctx, account.UserID)
	if err != nil {
		return nil, err
	}
	results := make([]model.FavoriteVideo, 0, len(favorites))
	for _, f := range favorites {
		results = append(results, model.NewFavoriteVideoFromEntity(f))
	}
	return results, nil
}

func (u *usecase) AddFavoriteActor(ctx context.Context, keycloakID, actorID string) (*model.FavoriteActor, error) {
	log := logger.LoggerWithCtx(ctx)
	account, err := u.ensureAccount(ctx, keycloakID)
	if err != nil {
		return nil, err
	}
	existing, err := u.favoriteActorRepo.FindByUserAndActorID(ctx, account.UserID, actorID)
	if err == nil && existing != nil {
		log.Debug("favorite actor already exists", "actor_id", actorID)
		fa := model.NewFavoriteActorFromEntity(*existing)
		return &fa, nil
	}
	favorite, err := entity.NewFavoriteActor(account.UserID, actorID)
	if err != nil {
		return nil, err
	}
	if err := u.favoriteActorRepo.Add(ctx, favorite); err != nil {
		return nil, err
	}
	fa := model.NewFavoriteActorFromEntity(*favorite)
	return &fa, nil
}

func (u *usecase) RemoveFavoriteActor(ctx context.Context, keycloakID, actorID string) (*model.FavoriteActor, error) {
	account, err := u.ensureAccount(ctx, keycloakID)
	if err != nil {
		return nil, err
	}
	removed, err := u.favoriteActorRepo.RemoveByActorID(ctx, account.UserID, actorID)
	if err != nil {
		return nil, err
	}
	fa := model.NewFavoriteActorFromEntity(*removed)
	return &fa, nil
}

func (u *usecase) ListFavoriteActors(ctx context.Context, keycloakID string) ([]model.FavoriteActor, error) {
	account, err := u.ensureAccount(ctx, keycloakID)
	if err != nil {
		return nil, err
	}
	favorites, err := u.favoriteActorRepo.ListByUserID(ctx, account.UserID)
	if err != nil {
		return nil, err
	}
	results := make([]model.FavoriteActor, 0, len(favorites))
	for _, f := range favorites {
		results = append(results, model.NewFavoriteActorFromEntity(f))
	}
	return results, nil
}

func (u *usecase) ensureAccount(ctx context.Context, keycloakID string) (*entity.Account, error) {
	if keycloakID == "" {
		return nil, fmt.Errorf("keycloak id is required")
	}
	return u.accountRepo.UpsertByKeycloakID(ctx, keycloakID)
}
