package dmmapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

type Repository struct {
    client ClientInterface
    mapper MapperInterface
}



type defaultMapper struct{}
func (m defaultMapper) ConvertItem(item Item) entity.Video {
    return ConvertItem(item)
}

// NewRepository 新しい DMM 用 Repository を返す
func NewRepository() (repository.VideoRepository, error) {
    c, err := NewClient()
    if err != nil {
        return nil, err
    }
    return &Repository{
        client: c,
        mapper: defaultMapper{},
    }, nil
}

// テストやモック注入用
func NewRepositoryWithDeps(client ClientInterface, mapper MapperInterface) repository.VideoRepository {
    return &Repository{
        client: client,
        mapper: mapper,
    }
}


// GetVideosByDate は指定日付の動画一覧を取得する
func (r *Repository) GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error) {
    path := fmt.Sprintf(
        "/v3/ItemList?site=FANZA&service=digital&floor=videoa&sort=date&gte_date=%s&lte_date=%s",
        targetDate.Format("2006-01-02T00:00:00"),
        targetDate.AddDate(0, 0, 1).Format("2006-01-02T00:00:00"),
    )
    //log.Println(path)
    var resp Response
    if err := r.client.Call(path, &resp); err != nil {
        return nil, err
    }
    videos := make([]entity.Video, 0, len(resp.Result.Items))
    for _, item := range resp.Result.Items {
        videos = append(videos,r.mapper.ConvertItem(item))
    }
    return videos, nil
}

// GetVideoById は指定 ID の動画情報を取得する
func (r *Repository) GetVideoById(ctx context.Context, dmmID string) (*entity.Video, error) {
    path := fmt.Sprintf(
        "/v3/ItemList?site=FANZA&service=digital&floor=videoa&cid=%s",
        dmmID,
    )
    //log.Println(path)
    var resp Response
    if err := r.client.Call(path, &resp); err != nil {
        return nil, err
    }
    if len(resp.Result.Items) == 0 {
        return nil, fmt.Errorf("動画ID %s が見つかりませんでした", dmmID)
    }
    v := r.mapper.ConvertItem(resp.Result.Items[0])
    return &v, nil
}

// SearchVideos はキーワードや各種 ID で動画を検索する
func (r *Repository) SearchVideos(
    ctx context.Context,
    keyword, actressID, genreID, makerID, seriesID, directorID string,
) ([]entity.Video, error) {
    params := []string{"site=FANZA", "service=digital", "floor=videoa"}
    if keyword != "" {
        params = append(params, fmt.Sprintf("keyword=%s", keyword))
    }
    
    // 複数条件のfilterカウンター
    articleIdx := 0
    
    // 女優ID
    if actressID != "" {
        params = append(params, fmt.Sprintf("article[%d]=actress", articleIdx))
        params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, actressID))
        articleIdx++
    }
    
    // ジャンルID
    if genreID != "" {
        params = append(params, fmt.Sprintf("article[%d]=genre", articleIdx))
        params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, genreID))
        articleIdx++
    }
    
    // メーカーID
    if makerID != "" {
        params = append(params, fmt.Sprintf("article[%d]=maker", articleIdx))
        params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, makerID))
        articleIdx++
    }
    
    // シリーズID
    if seriesID != "" {
        params = append(params, fmt.Sprintf("article[%d]=series", articleIdx))
        params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, seriesID))
        articleIdx++
    }
    
    // 監督ID
    if directorID != "" {
        params = append(params, fmt.Sprintf("article[%d]=director", articleIdx))
        params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, directorID))
        articleIdx++
    }
    
    path := "/v3/ItemList?" + strings.Join(params, "&")
    var resp Response
    if err := r.client.Call(path, &resp); err != nil {
        return nil, err
    }
    videos := make([]entity.Video, 0, len(resp.Result.Items))
    for _, item := range resp.Result.Items {
        videos = append(videos,r.mapper.ConvertItem(item))
    }
    return videos, nil
}

// GetVideosByID は ID 一覧から動画を検索
func (r *Repository) GetVideosByID(
    ctx context.Context,
    actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string,
    hits int32,
    offset int32,
    sort, gteDate, lteDate, site, service, floor string,
) ([]entity.Video, error) {
    // パラメータ組み立て
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

    // 複数条件のfilterカウンター
    articleIdx := 0
    
    // actressIDs - 女優
    for _, id := range actressIDs {
        if id != "" {
            params = append(params, fmt.Sprintf("article[%d]=actress", articleIdx))
            params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
            articleIdx++
        }
    }
    
    // genreIDs - ジャンル
    for _, id := range genreIDs {
        if id != "" {
            params = append(params, fmt.Sprintf("article[%d]=genre", articleIdx))
            params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
            articleIdx++
        }
    }
    
    // makerIDs - メーカー
    for _, id := range makerIDs {
        if id != "" {
            params = append(params, fmt.Sprintf("article[%d]=maker", articleIdx))
            params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
            articleIdx++
        }
    }
    
    // seriesIDs - シリーズ
    for _, id := range seriesIDs {
        if id != "" {
            params = append(params, fmt.Sprintf("article[%d]=series", articleIdx))
            params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
            articleIdx++
        }
    }
    
    // directorIDs - 監督
    for _, id := range directorIDs {
        if id != "" {
            params = append(params, fmt.Sprintf("article[%d]=director", articleIdx))
            params = append(params, fmt.Sprintf("article_id[%d]=%s", articleIdx, id))
            articleIdx++
        }
    }
    
    path := "/v3/ItemList?" + strings.Join(params, "&")
    //log.Println(path)
    var resp Response
    if err := r.client.Call(path, &resp); err != nil {
        return nil, err
    }
    videos := make([]entity.Video, 0, len(resp.Result.Items))
    for _, item := range resp.Result.Items {
        videos = append(videos,r.mapper.ConvertItem(item))
    }
    return videos, nil
}

// GetVideosByKeyword はキーワード検索を行う
func (r *Repository) GetVideosByKeyword(
    ctx context.Context,
    keyword string,
    hits int32,
    offset int32,
    sort, gteDate, lteDate, site, service, floor string,
) ([]entity.Video, error) {
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
    //log.Println(path)
    var resp Response
    if err := r.client.Call(path, &resp); err != nil {
        return nil, err
    }
    videos := make([]entity.Video, 0, len(resp.Result.Items))
    for _, item := range resp.Result.Items {
        videos = append(videos,r.mapper.ConvertItem(item))
    }
    return videos, nil
}

// defaultIfEmpty は空文字列のときデフォルト値を返す
func defaultIfEmpty(s, def string) string {
    if s == "" {
        return def
    }
    return s
}