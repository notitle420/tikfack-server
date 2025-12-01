package port

import (
	"context"
	"time"

	"github.com/tikfack/server/internal/application/model"
)

// VideoCatalog は外部動画カタログへアクセスするポート。
type VideoCatalog interface {
	GetVideosByDate(ctx context.Context, targetDate time.Time, hits, offset int32) ([]model.Video, *model.SearchMetadata, error)
	GetVideoById(ctx context.Context, dmmId string) (*model.Video, error)
	SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]model.Video, *model.SearchMetadata, error)
	GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]model.Video, *model.SearchMetadata, error)
	GetVideosByKeyword(ctx context.Context, keyword string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]model.Video, *model.SearchMetadata, error)
}
