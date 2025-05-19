package repository

//go:generate mockgen -destination=mock/mock_video_repository.go -package=mock github.com/tikfack/server/internal/domain/repository VideoRepository

import (
	"context"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
)

// VideoRepository は動画データの取得方法を定義するインターフェイス
type VideoRepository interface {
	// 指定日付の動画一覧を取得する
	GetVideosByDate(ctx context.Context, targetDate time.Time, hits, offset int32) ([]entity.Video, *entity.SearchMetadata, error)

	// ID で動画を取得する
	GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error)

	// キーワードやIDを使ってVideoを検索する
	SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, *entity.SearchMetadata, error)
	
	GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, *entity.SearchMetadata, error)
	
	GetVideosByKeyword(ctx context.Context, keyword string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, *entity.SearchMetadata, error)
}
