package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FavoriteVideo represents a user's favorite video mapped to a DMM video ID.
type FavoriteVideo struct {
	FavoriteVideoUUID string
	UserID            string
	VideoID           string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewFavoriteVideo creates a new FavoriteVideo with a generated UUID.
func NewFavoriteVideo(userID, videoID string) (*FavoriteVideo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if videoID == "" {
		return nil, fmt.Errorf("video id is required")
	}
	now := time.Now().UTC()
	return &FavoriteVideo{
		FavoriteVideoUUID: uuid.NewString(),
		UserID:            userID,
		VideoID:           videoID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}
