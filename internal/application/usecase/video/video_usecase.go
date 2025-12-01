package usecase

//go:generate mockgen -destination=../mock/mock_video_usecase.go -package=mock github.com/tikfack/server/internal/application/usecase/video VideoUsecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/tikfack/server/internal/application/model"
	"github.com/tikfack/server/internal/application/port"
	"github.com/tikfack/server/internal/middleware/logger"
)

const (
	maxHits   int32 = 100
	maxOffset int32 = 50000
)

// VideoUsecase は動画関連のユースケースを定義するインターフェイス
// DMM のような外部カタログに依存するアクセスをアプリケーション層で
// 一元的に扱うためのポートを切っている。
type VideoUsecase interface {
	// GetVideosByDate は指定日付の動画一覧を取得する
	GetVideosByDate(ctx context.Context, targetDate time.Time, hits, offset int32) ([]model.Video, *model.SearchMetadata, error)

	// GetVideoById は指定されたDMMビデオIDの動画を取得する
	GetVideoById(ctx context.Context, dmmId string) (*model.Video, error)

	// SearchVideos はキーワードやIDを使って動画を検索する
	SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]model.Video, *model.SearchMetadata, error)

	// GetVideosByID は複数ID条件で動画を検索する
	GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits, offset int32, sort, gteDate, lteDate, site, service, floor string) ([]model.Video, *model.SearchMetadata, error)

	// GetVideosByKeyword はキーワード検索を行う
	GetVideosByKeyword(ctx context.Context, keyword string, hits, offset int32, sort, gteDate, lteDate, site, service, floor string) ([]model.Video, *model.SearchMetadata, error)
}

// videoUsecase は VideoUsecase の実装
type videoUsecase struct {
	catalog port.VideoCatalog
	logger  *slog.Logger
}

// NewVideoUsecase は VideoCatalog ポートを受け取り VideoUsecase を返す
func NewVideoUsecase(catalog port.VideoCatalog) VideoUsecase {
	if catalog == nil {
		panic("video catalog must be provided")
	}
	return &videoUsecase{
		catalog: catalog,
		logger:  slog.Default().With(slog.String("component", "video_usecase")),
	}
}

func (u *videoUsecase) loggerWithCtx(ctx context.Context) *slog.Logger {
	return u.logger.With(
		slog.String("user_id", logger.UserIDFromContext(ctx)),
		slog.String("trace_id", logger.TraceIDFromContext(ctx)),
		slog.String("token_id", logger.TokenIDFromContext(ctx)),
	)
}

// GetVideosByDate は指定日付の動画一覧を取得する
func (u *videoUsecase) GetVideosByDate(ctx context.Context, targetDate time.Time, hits, offset int32) ([]model.Video, *model.SearchMetadata, error) {
	logger := u.loggerWithCtx(ctx)
	normHits := clampHits(hits)
	normOffset := clampOffset(offset)
	logger.Debug("GetVideosByDate called",
		"targetDate", targetDate.Format("2006-01-02"),
		"hits", normHits,
		"offset", normOffset,
	)
	return u.catalog.GetVideosByDate(ctx, targetDate, normHits, normOffset)
}

// GetVideoById は、指定された DMMビデオID の動画を取得する
func (u *videoUsecase) GetVideoById(ctx context.Context, dmmId string) (*model.Video, error) {
	logger := u.loggerWithCtx(ctx)
	logger.Debug("GetVideoById called", "dmmId", dmmId)
	return u.catalog.GetVideoById(ctx, dmmId)
}

// SearchVideos はキーワードやIDを使って動画を検索する
func (u *videoUsecase) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]model.Video, *model.SearchMetadata, error) {
	logger := u.loggerWithCtx(ctx)
	logger.Debug("SearchVideos called",
		"keyword", keyword,
		"actressID", actressID,
		"genreID", genreID,
		"makerID", makerID,
		"seriesID", seriesID,
		"directorID", directorID,
	)
	return u.catalog.SearchVideos(ctx, keyword, actressID, genreID, makerID, seriesID, directorID)
}

// GetVideosByID は指定されたIDを使って動画を検索する
func (u *videoUsecase) GetVideosByID(
	ctx context.Context,
	actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string,
	hits, offset int32,
	sort, gteDate, lteDate, site, service, floor string,
) ([]model.Video, *model.SearchMetadata, error) {
	logger := u.loggerWithCtx(ctx)
	normHits := clampHits(hits)
	normOffset := clampOffset(offset)
	logger.Debug("GetVideosByID called",
		"actressIDs", actressIDs,
		"genreIDs", genreIDs,
		"makerIDs", makerIDs,
		"seriesIDs", seriesIDs,
		"directorIDs", directorIDs,
		"hits", normHits,
		"offset", normOffset,
		"sort", sort,
		"gteDate", gteDate,
		"lteDate", lteDate,
		"site", site,
		"service", service,
		"floor", floor,
	)
	return u.catalog.GetVideosByID(ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, normHits, normOffset, sort, gteDate, lteDate, site, service, floor)
}

// GetVideosByKeyword はキーワードを使って動画を検索する
func (u *videoUsecase) GetVideosByKeyword(
	ctx context.Context,
	keyword string,
	hits, offset int32,
	sort, gteDate, lteDate, site, service, floor string,
) ([]model.Video, *model.SearchMetadata, error) {
	logger := u.loggerWithCtx(ctx)
	normHits := clampHits(hits)
	normOffset := clampOffset(offset)
	logger.Debug("GetVideosByKeyword called",
		"keyword", keyword,
		"hits", normHits,
		"offset", normOffset,
		"sort", sort,
		"gteDate", gteDate,
		"lteDate", lteDate,
		"site", site,
		"service", service,
		"floor", floor,
	)
	return u.catalog.GetVideosByKeyword(ctx, keyword, normHits, normOffset, sort, gteDate, lteDate, site, service, floor)
}

func clampHits(hits int32) int32 {
	if hits <= 0 {
		return hits
	}
	if hits > maxHits {
		return maxHits
	}
	return hits
}

func clampOffset(offset int32) int32 {
	if offset < 0 {
		return 0
	}
	if offset > maxOffset {
		return maxOffset
	}
	return offset
}
