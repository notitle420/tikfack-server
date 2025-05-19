package usecase

//go:generate mockgen -destination=../mock/mock_video_usecase.go -package=mock github.com/tikfack/server/internal/application/usecase/video VideoUsecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// VideoUsecase は動画関連のユースケースを定義するインターフェイス
type VideoUsecase interface {
	GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error)
	GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error)
	SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error)
	GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error)
	GetVideosByKeyword(ctx context.Context, keyword string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error)
}

// videoUsecase は VideoUsecase の実装
type videoUsecase struct {
	videoRepo repository.VideoRepository
	logger    *slog.Logger
}

// NewVideoUsecase は依存する VideoRepository を受け取り VideoUsecase を返す
func NewVideoUsecase(repo repository.VideoRepository) VideoUsecase {
	return &videoUsecase{
		videoRepo: repo,
		logger:    slog.Default().With(slog.String("component", "video_usecase")),
	}
}

// GetVideosByDate は指定日付の動画一覧を取得する
func (u *videoUsecase) GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error) {
	u.logger.Debug("GetVideosByDate called", "targetDate", targetDate.Format("2006-01-02"))
	return u.videoRepo.GetVideosByDate(ctx, targetDate)
}

// GetVideoById は、指定された DMMビデオID の動画を取得する
func (u *videoUsecase) GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error) {
	u.logger.Debug("GetVideoById called", "dmmId", dmmId)
	return u.videoRepo.GetVideoById(ctx, dmmId)
}

// SearchVideos は動画をキーワードや各種IDで検索する
func (u *videoUsecase) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error) {
	u.logger.Debug("SearchVideos called", 
		"keyword", keyword,
		"actressID", actressID,
		"genreID", genreID,
		"makerID", makerID,
		"seriesID", seriesID,
		"directorID", directorID)
	return u.videoRepo.SearchVideos(ctx, keyword, actressID, genreID, makerID, seriesID, directorID)
}

// GetVideosByID は複数のIDを使用して動画を検索する
func (u *videoUsecase) GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	u.logger.Debug("GetVideosByID called", 
		"actressIDs_count", len(actressIDs),
		"genreIDs_count", len(genreIDs),
		"makerIDs_count", len(makerIDs),
		"seriesIDs_count", len(seriesIDs),
		"directorIDs_count", len(directorIDs),
		"hits", hits,
		"offset", offset,
		"sort", sort)
	return u.videoRepo.GetVideosByID(ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, hits, offset, sort, gteDate, lteDate, site, service, floor)
}

// GetVideosByKeyword はキーワードを使用して動画を検索する
func (u *videoUsecase) GetVideosByKeyword(ctx context.Context, keyword string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	u.logger.Debug("GetVideosByKeyword called", 
		"keyword", keyword,
		"hits", hits,
		"offset", offset,
		"sort", sort,
		"gteDate", gteDate,
		"lteDate", lteDate)
	return u.videoRepo.GetVideosByKeyword(ctx, keyword, hits, offset, sort, gteDate, lteDate, site, service, floor)
}
