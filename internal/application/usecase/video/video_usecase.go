package usecase

//go:generate mockgen -destination=../mock/mock_video_usecase.go -package=mock github.com/tikfack/server/internal/application/usecase/video VideoUsecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
	"github.com/tikfack/server/internal/middleware/logger"
)

// VideoUsecase は動画関連のユースケースを定義するインターフェイス
type VideoUsecase interface {
	// GetVideosByDate は指定日付の動画一覧を取得する
	GetVideosByDate(ctx context.Context, targetDate time.Time, hits, offset int32) ([]entity.Video, *entity.SearchMetadata, error)

	// GetVideoById は指定されたDMMビデオIDの動画を取得する
	GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error)

	// SearchVideos はキーワードやIDを使って動画を検索する
	SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, *entity.SearchMetadata, error)

	// GetVideosByID は指定されたIDを使って動画を検索する
	GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits, offset int32, sort, gteDate, lteDate, site, service, floor string) ([]entity.Video, *entity.SearchMetadata, error)

	// GetVideosByKeyword はキーワードを使って動画を検索する
	GetVideosByKeyword(ctx context.Context, keyword string, hits, offset int32, sort, gteDate, lteDate, site, service, floor string) ([]entity.Video, *entity.SearchMetadata, error)
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

func (u *videoUsecase) loggerWithCtx(ctx context.Context) *slog.Logger {
    return u.logger.With(
        slog.String("user_id",  logger.UserIDFromContext(ctx)),   // 例: いずれかの場所で ctx に "sub" をセット済み
        slog.String("trace_id", logger.TraceIDFromContext(ctx)),  // 例: Interceptor などで ctx にセット済み
		slog.String("token_id", logger.TokenIDFromContext(ctx)),
    )
}

// GetVideosByDate は指定日付の動画一覧を取得する
func (u *videoUsecase) GetVideosByDate(ctx context.Context, targetDate time.Time, hits, offset int32) ([]entity.Video, *entity.SearchMetadata, error) {
	logger := u.loggerWithCtx(ctx)
	logger.Debug("GetVideosByDate called", 
		"targetDate", targetDate.Format("2006-01-02"),
		"hits", hits,
		"offset", offset,
	)
	return u.videoRepo.GetVideosByDate(ctx, targetDate, hits, offset)
}

// GetVideoById は、指定された DMMビデオID の動画を取得する
func (u *videoUsecase) GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error) {
	logger := u.loggerWithCtx(ctx)
	logger.Debug("GetVideoById called", "dmmId", dmmId)
	return u.videoRepo.GetVideoById(ctx, dmmId)
}

// SearchVideos はキーワードやIDを使って動画を検索する
func (u *videoUsecase) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, *entity.SearchMetadata, error) {
	logger := u.loggerWithCtx(ctx)
	logger.Debug("SearchVideos called",
		"keyword", keyword,
		"actressID", actressID,
		"genreID", genreID,
		"makerID", makerID,
		"seriesID", seriesID,
		"directorID", directorID,
	)
	return u.videoRepo.SearchVideos(ctx, keyword, actressID, genreID, makerID, seriesID, directorID)
}

// GetVideosByID は指定されたIDを使って動画を検索する
func (u *videoUsecase) GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits, offset int32, sort, gteDate, lteDate, site, service, floor string) ([]entity.Video, *entity.SearchMetadata, error) {
	logger := u.loggerWithCtx(ctx)
	logger.Debug("GetVideosByID called",
		"actressIDs", actressIDs,
		"genreIDs", genreIDs,
		"makerIDs", makerIDs,
		"seriesIDs", seriesIDs,
		"directorIDs", directorIDs,
		"hits", hits,
		"offset", offset,
		"sort", sort,
		"gteDate", gteDate,
		"lteDate", lteDate,
		"site", site,
		"service", service,
		"floor", floor,
	)
	return u.videoRepo.GetVideosByID(ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, hits, offset, sort, gteDate, lteDate, site, service, floor)
}

// GetVideosByKeyword はキーワードを使って動画を検索する
func (u *videoUsecase) GetVideosByKeyword(ctx context.Context, keyword string, hits, offset int32, sort, gteDate, lteDate, site, service, floor string) ([]entity.Video, *entity.SearchMetadata, error) {
	logger := u.loggerWithCtx(ctx)
	logger.Debug("GetVideosByKeyword called",
		"keyword", keyword,
		"hits", hits,
		"offset", offset,
		"sort", sort,
		"gteDate", gteDate,
		"lteDate", lteDate,
		"site", site,
		"service", service,
		"floor", floor,
	)
	return u.videoRepo.GetVideosByKeyword(ctx, keyword, hits, offset, sort, gteDate, lteDate, site, service, floor)
}
