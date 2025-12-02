package model

import (
	"time"

	"github.com/tikfack/server/internal/domain/entity"
)

// FavoriteVideo represents a favorite video DTO used by the application layer.
type FavoriteVideo struct {
	FavoriteVideoUUID string
	UserID            string
	VideoID           string
	CreatedAt         string
}

// FavoriteActor represents a favorite actor DTO used by the application layer.
type FavoriteActor struct {
	FavoriteActorUUID string
	UserID            string
	ActorID           string
	CreatedAt         string
}

// NewFavoriteVideoFromEntity converts a domain entity to an application model.
func NewFavoriteVideoFromEntity(e entity.FavoriteVideo) FavoriteVideo {
	return FavoriteVideo{
		FavoriteVideoUUID: e.FavoriteVideoUUID,
		UserID:            e.UserID,
		VideoID:           e.VideoID,
		CreatedAt:         e.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// NewFavoriteActorFromEntity converts a domain entity to an application model.
func NewFavoriteActorFromEntity(e entity.FavoriteActor) FavoriteActor {
	return FavoriteActor{
		FavoriteActorUUID: e.FavoriteActorUUID,
		UserID:            e.UserID,
		ActorID:           e.ActorID,
		CreatedAt:         e.CreatedAt.UTC().Format(time.RFC3339),
	}
}
