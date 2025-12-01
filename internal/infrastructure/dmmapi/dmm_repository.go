package dmmapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/tikfack/server/internal/application/model"
	"github.com/tikfack/server/internal/application/port"
	"github.com/tikfack/server/internal/middleware/logger"
)

type Repository struct {
	client ClientInterface
	mapper MapperInterface
	logger *slog.Logger
}

type defaultMapper struct{}

func (m defaultMapper) ConvertEntityFromDMM(result Result) ([]model.Video, *model.SearchMetadata) {
	return ConvertEntityFromDMM(result)
}

// NewRepository 新しい DMM 用 Repository を返す
func NewRepository() (port.VideoCatalog, error) {
	c, err := NewClient()
	if err != nil {
		return nil, err
	}
	return &Repository{
		client: c,
		mapper: defaultMapper{},
		logger: slog.Default().With(slog.String("component", "dmmapi")),
	}, nil
}

// NewRepositoryWithDeps テストやモック注入用のコンストラクタ
func NewRepositoryWithDeps(client ClientInterface, mapper MapperInterface) port.VideoCatalog {
	return &Repository{
		client: client,
		mapper: mapper,
		logger: slog.Default().With(slog.String("component", "dmmapi")),
	}
}

// NewRepositoryWithLogger creates a repository with a custom logger
func NewRepositoryWithLogger(client ClientInterface, mapper MapperInterface, logger *slog.Logger) port.VideoCatalog {
	return &Repository{
		client: client,
		mapper: mapper,
		logger: logger.With(slog.String("component", "dmmapi")),
	}
}

// SetLogger sets a custom logger for the repository
func (r *Repository) SetLogger(logger *slog.Logger) {
	logger = logger.With(slog.String("component", "dmmapi"))
}

var ErrAPIError = errors.New("API error")

// GetVideosByDate は指定日付の動画一覧を取得する
func (r *Repository) GetVideosByDate(ctx context.Context, targetDate time.Time, hits, offset int32) ([]model.Video, *model.SearchMetadata, error) {
	path := fmt.Sprintf(
		"/v3/ItemList?site=FANZA&service=digital&floor=videoa&sort=date&hits=%d&offset=%d&gte_date=%s&lte_date=%s",
		hits,
		offset,
		targetDate.Format("2006-01-02T15:04:05"),
		time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 0, 0, targetDate.Location()).Format("2006-01-02T15:04:05"),
	)
	logger := logger.LoggerWithCtx(ctx)
	logger.Debug("calling API", "path", path)
	var resp Response
	if err := r.client.Call(path, &resp); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrAPIError, err)
	}

	videos, metadata := r.mapper.ConvertEntityFromDMM(resp.Result)

	// 先頭5件のみを抽出してログ出力
	if len(videos) > 0 {
		sampleSize := min(5, len(videos))
		sample := videos[:sampleSize]
		logger.Debug("video results sample", "count", sampleSize, "videos", sample)
	} else {
		logger.Debug("No videos found")
	}

	return videos, metadata, nil
}

// GetVideoById は指定 ID の動画情報を取得する
func (r *Repository) GetVideoById(ctx context.Context, dmmID string) (*model.Video, error) {
	path := fmt.Sprintf(
		"/v3/ItemList?site=FANZA&service=digital&floor=videoa&cid=%s",
		dmmID,
	)
	logger := logger.LoggerWithCtx(ctx)
	logger.Debug("calling API", "path", path)
	var resp Response
	if err := r.client.Call(path, &resp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAPIError, err)
	}
	if len(resp.Result.Items) == 0 {
		return nil, fmt.Errorf("動画ID %s が見つかりませんでした", dmmID)
	}
	videos, _ := r.mapper.ConvertEntityFromDMM(resp.Result)
	return &videos[0], nil
}

// SearchVideos はキーワードやIDを使って動画を検索する
func (r *Repository) SearchVideos(
	ctx context.Context,
	keyword, actressID, genreID, makerID, seriesID, directorID string,
) ([]model.Video, *model.SearchMetadata, error) {
	params := []string{"site=FANZA", "service=digital", "floor=videoa"}
	if keyword != "" {
		params = append(params, fmt.Sprintf("keyword=%s", keyword))
	}

	articleIdx := 0
	if actressID != "" {
		params = append(params, fmt.Sprintf("article[%d]=actress", articleIdx))
		params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, actressID))
		articleIdx++
	}
	if genreID != "" {
		params = append(params, fmt.Sprintf("article[%d]=genre", articleIdx))
		params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, genreID))
		articleIdx++
	}
	if makerID != "" {
		params = append(params, fmt.Sprintf("article[%d]=maker", articleIdx))
		params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, makerID))
		articleIdx++
	}
	if seriesID != "" {
		params = append(params, fmt.Sprintf("article[%d]=series", articleIdx))
		params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, seriesID))
		articleIdx++
	}
	if directorID != "" {
		params = append(params, fmt.Sprintf("article[%d]=director", articleIdx))
		params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, directorID))
		articleIdx++
	}

	path := "/v3/ItemList?" + strings.Join(params, "&")
	logger := logger.LoggerWithCtx(ctx)
	logger.Debug("calling API", "path", path)
	var resp Response
	if err := r.client.Call(path, &resp); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrAPIError, err)
	}

	if len(resp.Result.Items) == 0 {
		md := &model.SearchMetadata{
			ResultCount:   resp.Result.ResultCount,
			TotalCount:    resp.Result.TotalCount,
			FirstPosition: resp.Result.FirstPosition,
		}
		return []model.Video{}, md, nil
	}

	videos, metadata := r.mapper.ConvertEntityFromDMM(resp.Result)
	return videos, metadata, nil
}

// GetVideosByID は複数のIDで動画を検索する
func (r *Repository) GetVideosByID(
	ctx context.Context,
	actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string,
	hits, offset int32,
	sort, gteDate, lteDate, site, service, floor string,
) ([]model.Video, *model.SearchMetadata, error) {
	params := []string{
		fmt.Sprintf("site=%s", defaultIfEmpty(site, "FANZA")),
		fmt.Sprintf("service=%s", defaultIfEmpty(service, "digital")),
		fmt.Sprintf("floor=%s", defaultIfEmpty(floor, "videoa")),
	}

	if hits > 0 {
		params = append(params, fmt.Sprintf("hits=%d", hits))
	}
	if offset > 0 {
		params = append(params, fmt.Sprintf("offset=%d", offset))
	}
	if sort != "" {
		params = append(params, fmt.Sprintf("sort=%s", sort))
	}
	if gteDate != "" {
		params = append(params, fmt.Sprintf("gte_date=%s", gteDate))
	}
	if lteDate != "" {
		params = append(params, fmt.Sprintf("lte_date=%s", lteDate))
	}

	articleIdx := 0
	for _, id := range actressIDs {
		if id != "" {
			params = append(params, fmt.Sprintf("article[%d]=actress", articleIdx))
			params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
			articleIdx++
		}
	}
	for _, id := range genreIDs {
		if id != "" {
			params = append(params, fmt.Sprintf("article[%d]=genre", articleIdx))
			params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
			articleIdx++
		}
	}
	for _, id := range makerIDs {
		if id != "" {
			params = append(params, fmt.Sprintf("article[%d]=maker", articleIdx))
			params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
			articleIdx++
		}
	}
	for _, id := range seriesIDs {
		if id != "" {
			params = append(params, fmt.Sprintf("article[%d]=series", articleIdx))
			params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
			articleIdx++
		}
	}
	for _, id := range directorIDs {
		if id != "" {
			params = append(params, fmt.Sprintf("article[%d]=director", articleIdx))
			params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
			articleIdx++
		}
	}

	path := "/v3/ItemList?" + strings.Join(params, "&")
	logger := logger.LoggerWithCtx(ctx)
	logger.Debug("calling API", "path", path)
	var resp Response
	if err := r.client.Call(path, &resp); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrAPIError, err)
	}

	if len(resp.Result.Items) == 0 {
		md := &model.SearchMetadata{
			ResultCount:   resp.Result.ResultCount,
			TotalCount:    resp.Result.TotalCount,
			FirstPosition: resp.Result.FirstPosition,
		}
		return []model.Video{}, md, nil
	}

	videos, metadata := r.mapper.ConvertEntityFromDMM(resp.Result)
	return videos, metadata, nil
}

// GetVideosByKeyword はキーワードで動画を検索する
func (r *Repository) GetVideosByKeyword(
	ctx context.Context,
	keyword string,
	hits, offset int32,
	sort, gteDate, lteDate, site, service, floor string,
) ([]model.Video, *model.SearchMetadata, error) {
	params := []string{
		fmt.Sprintf("site=%s", defaultIfEmpty(site, "FANZA")),
		fmt.Sprintf("service=%s", defaultIfEmpty(service, "digital")),
		fmt.Sprintf("floor=%s", defaultIfEmpty(floor, "videoa")),
	}

	if keyword != "" {
		params = append(params, fmt.Sprintf("keyword=%s", keyword))
	}
	if hits > 0 {
		params = append(params, fmt.Sprintf("hits=%d", hits))
	}
	if offset > 0 {
		params = append(params, fmt.Sprintf("offset=%d", offset))
	}
	if sort != "" {
		params = append(params, fmt.Sprintf("sort=%s", sort))
	}
	if gteDate != "" {
		params = append(params, fmt.Sprintf("gte_date=%s", gteDate))
	}
	if lteDate != "" {
		params = append(params, fmt.Sprintf("lte_date=%s", lteDate))
	}

	path := "/v3/ItemList?" + strings.Join(params, "&")
	logger := logger.LoggerWithCtx(ctx)
	logger.Debug("calling API", "path", path)
	var resp Response
	if err := r.client.Call(path, &resp); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrAPIError, err)
	}

	videos, metadata := r.mapper.ConvertEntityFromDMM(resp.Result)
	return videos, metadata, nil
}

// defaultIfEmpty は空文字列のときデフォルト値を返す
func defaultIfEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// min は2つの整数の小さい方を返す
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
