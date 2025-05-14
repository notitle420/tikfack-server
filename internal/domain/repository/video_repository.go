package repository

import (
	"context"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
)

// VideoRepository は動画データの取得方法を定義するインターフェイス
type VideoRepository interface {
	// 指定日付の動画一覧を取得する
	GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error)

	// ID で動画を取得する
	GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error)

	// キーワードやIDを使ってVideoを検索する
	SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error)
}