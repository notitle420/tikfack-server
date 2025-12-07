package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FavoriteActor represents a favorite performer reference.
type FavoriteActor struct {
	FavoriteActorUUID string
	UserID            string
	ActorID           string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewFavoriteActor creates a new FavoriteActor with generated UUIDs.
func NewFavoriteActor(userID, actorID string) (*FavoriteActor, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if actorID == "" {
		return nil, fmt.Errorf("actor id is required")
	}
	now := time.Now().UTC()
	return &FavoriteActor{
		FavoriteActorUUID: uuid.NewString(),
		UserID:            userID,
		ActorID:           actorID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}
