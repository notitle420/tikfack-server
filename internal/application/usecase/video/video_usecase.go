package usecase

import (
	"context"
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
}

// NewVideoUsecase は依存する VideoRepository を受け取り VideoUsecase を返す
func NewVideoUsecase(repo repository.VideoRepository) VideoUsecase {
	return &videoUsecase{
		videoRepo: repo,
	}
}

// GetVideosByDate は指定日付の動画一覧を取得する
func (u *videoUsecase) GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error) {
	return u.videoRepo.GetVideosByDate(ctx, targetDate)
}

// GetVideoById は、指定された DMMビデオID の動画を取得する
func (u *videoUsecase) GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error) {
	return u.videoRepo.GetVideoById(ctx, dmmId)
}

// SearchVideos は動画をキーワードや各種IDで検索する
func (u *videoUsecase) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error) {
	return u.videoRepo.SearchVideos(ctx, keyword, actressID, genreID, makerID, seriesID, directorID)
}

// GetVideosByID は複数のIDを使用して動画を検索する
func (u *videoUsecase) GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	return u.videoRepo.GetVideosByID(ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, hits, offset, sort, gteDate, lteDate, site, service, floor)
}

// GetVideosByKeyword はキーワードを使用して動画を検索する
func (u *videoUsecase) GetVideosByKeyword(ctx context.Context, keyword string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	return u.videoRepo.GetVideosByKeyword(ctx, keyword, hits, offset, sort, gteDate, lteDate, site, service, floor)
}
