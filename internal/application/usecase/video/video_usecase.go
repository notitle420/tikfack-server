package usecase

//go:generate mockgen -destination=../mock/mock_video_usecase.go -package=mock github.com/tikfack/server/internal/application/usecase/video VideoUsecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

const (
	maxHits   int32 = 100
	maxOffset int32 = 50000
)

var (
	// ErrInvalidDateFormat は入力の日付文字列が不正な場合に返される
	ErrInvalidDateFormat = errors.New("invalid date format")
)

// GetVideosByDateInput は日付検索の入力をまとめた値オブジェクト
type GetVideosByDateInput struct {
	Date   string
	Hits   int32
	Offset int32
}

// GetVideosByIDInput は ID 条件検索の入力値を表す
type GetVideosByIDInput struct {
	ActressIDs  []string
	GenreIDs    []string
	MakerIDs    []string
	SeriesIDs   []string
	DirectorIDs []string
	Hits        int32
	Offset      int32
	Sort        string
	GteDate     string
	LteDate     string
	Site        string
	Service     string
	Floor       string
}

// GetVideosByKeywordInput はキーワード検索の入力値を表す
type GetVideosByKeywordInput struct {
	Keyword string
	Hits    int32
	Offset  int32
	Sort    string
	GteDate string
	LteDate string
	Site    string
	Service string
	Floor   string
}

// GetVideosByDateOutput は日付検索の結果
type GetVideosByDateOutput struct {
	Videos     []entity.Video
	Metadata   *entity.SearchMetadata
	TargetDate time.Time
	Hits       int32
	Offset     int32
}

// GetVideosOutput は ID / Keyword 検索の結果
type GetVideosOutput struct {
	Videos   []entity.Video
	Metadata *entity.SearchMetadata
	Hits     int32
	Offset   int32
}

// VideoUsecase は動画関連のユースケースを定義するインターフェイス
type VideoUsecase interface {
	GetVideosByDate(ctx context.Context, input GetVideosByDateInput) (*GetVideosByDateOutput, error)
	GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error)
	SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, *entity.SearchMetadata, error)
	GetVideosByID(ctx context.Context, input GetVideosByIDInput) (*GetVideosOutput, error)
	GetVideosByKeyword(ctx context.Context, input GetVideosByKeywordInput) (*GetVideosOutput, error)
}

// videoUsecase は VideoUsecase の実装
type videoUsecase struct {
	videoRepo repository.VideoRepository
	logger    *slog.Logger
	now       func() time.Time
}

// NewVideoUsecase は依存する VideoRepository を受け取り VideoUsecase を返す
func NewVideoUsecase(repo repository.VideoRepository) VideoUsecase {
	return NewVideoUsecaseWithDeps(repo, nil)
}

// NewVideoUsecaseWithDeps はテスト用に現在時刻関数を注入可能なコンストラクタ
func NewVideoUsecaseWithDeps(repo repository.VideoRepository, now func() time.Time) VideoUsecase {
	if now == nil {
		now = time.Now
	}
	return &videoUsecase{
		videoRepo: repo,
		logger:    slog.Default().With(slog.String("component", "video_usecase")),
		now:       now,
	}
}

func (u *videoUsecase) loggerWithCtx(ctx context.Context) *slog.Logger {
	return u.logger.With(
		slog.String("user_id", ctxkeys.UserIDFromContext(ctx)),
		slog.String("trace_id", ctxkeys.TraceIDFromContext(ctx)),
		slog.String("token_id", ctxkeys.TokenIDFromContext(ctx)),
	)
}

// GetVideosByDate は指定日付の動画一覧を取得する
func (u *videoUsecase) GetVideosByDate(ctx context.Context, input GetVideosByDateInput) (*GetVideosByDateOutput, error) {
	logger := u.loggerWithCtx(ctx)
	targetDate, err := u.resolveDate(input.Date)
	if err != nil {
		logger.Debug("invalid date supplied", "date", input.Date, "error", err)
		return nil, fmt.Errorf("%w: %v", ErrInvalidDateFormat, err)
	}
	hits := clampHits(input.Hits)
	offset := clampOffset(input.Offset)
	logger.Debug("GetVideosByDate called",
		"targetDate", targetDate.Format("2006-01-02"),
		"hits", hits,
		"offset", offset,
	)
	videos, metadata, err := u.videoRepo.GetVideosByDate(ctx, targetDate, hits, offset)
	if err != nil {
		return nil, err
	}
	return &GetVideosByDateOutput{
		Videos:     videos,
		Metadata:   metadata,
		TargetDate: targetDate,
		Hits:       hits,
		Offset:     offset,
	}, nil
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
func (u *videoUsecase) GetVideosByID(ctx context.Context, input GetVideosByIDInput) (*GetVideosOutput, error) {
	logger := u.loggerWithCtx(ctx)
	hits := clampHits(input.Hits)
	offset := clampOffset(input.Offset)
	logger.Debug("GetVideosByID called",
		"actressIDs", input.ActressIDs,
		"genreIDs", input.GenreIDs,
		"makerIDs", input.MakerIDs,
		"seriesIDs", input.SeriesIDs,
		"directorIDs", input.DirectorIDs,
		"hits", hits,
		"offset", offset,
		"sort", input.Sort,
		"gteDate", input.GteDate,
		"lteDate", input.LteDate,
		"site", input.Site,
		"service", input.Service,
		"floor", input.Floor,
	)
	videos, metadata, err := u.videoRepo.GetVideosByID(
		ctx,
		input.ActressIDs,
		input.GenreIDs,
		input.MakerIDs,
		input.SeriesIDs,
		input.DirectorIDs,
		hits,
		offset,
		input.Sort,
		input.GteDate,
		input.LteDate,
		input.Site,
		input.Service,
		input.Floor,
	)
	if err != nil {
		return nil, err
	}
	return &GetVideosOutput{
		Videos:   videos,
		Metadata: metadata,
		Hits:     hits,
		Offset:   offset,
	}, nil
}

// GetVideosByKeyword はキーワードを使って動画を検索する
func (u *videoUsecase) GetVideosByKeyword(ctx context.Context, input GetVideosByKeywordInput) (*GetVideosOutput, error) {
	logger := u.loggerWithCtx(ctx)
	hits := clampHits(input.Hits)
	offset := clampOffset(input.Offset)
	logger.Debug("GetVideosByKeyword called",
		"keyword", input.Keyword,
		"hits", hits,
		"offset", offset,
		"sort", input.Sort,
		"gteDate", input.GteDate,
		"lteDate", input.LteDate,
		"site", input.Site,
		"service", input.Service,
		"floor", input.Floor,
	)
	videos, metadata, err := u.videoRepo.GetVideosByKeyword(
		ctx,
		input.Keyword,
		hits,
		offset,
		input.Sort,
		input.GteDate,
		input.LteDate,
		input.Site,
		input.Service,
		input.Floor,
	)
	if err != nil {
		return nil, err
	}
	return &GetVideosOutput{
		Videos:   videos,
		Metadata: metadata,
		Hits:     hits,
		Offset:   offset,
	}, nil
}

func (u *videoUsecase) resolveDate(date string) (time.Time, error) {
	if date == "" {
		return u.now(), nil
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
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
