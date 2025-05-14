package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// dmmVideoRepository は DMM API から動画情報を取得するリポジトリ実装例
type dmmVideoRepository struct{}

// NewDMMVideoRepository は dmmVideoRepository の新しいインスタンスを返す
func NewDMMVideoRepository() repository.VideoRepository {
	return &dmmVideoRepository{}
}

// GetVideosByDate は指定日付の動画一覧を DMM API から取得する
func (r *dmmVideoRepository) GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error) {
	log.Printf("connect")
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("BASE URL Not Set")
	}

	apiID := os.Getenv("DMM_API_ID")
	if apiID == "" {
		return nil, fmt.Errorf("DMM_API_ID not set")
	}

	affiliateID := os.Getenv("DMM_API_AFFILIATE_ID")
	if affiliateID == "" {
		return nil, fmt.Errorf("DMM_API_AFFILIATE_ID not set")
	}

	hits := os.Getenv("HITS")
	if hits == "" {
		return nil, fmt.Errorf("HITS not set")
	}

	startDate := targetDate.Format("2006-01-02T00:00:00")
	endDate := targetDate.AddDate(0, 0, 1).Format("2006-01-02T00:00:00")
	apiURL := fmt.Sprintf("%s/v3/ItemList?api_id=%s&affiliate_id=%s&site=FANZA&service=digital&floor=videoa&hits=%s&sort=date&output=json&gte_date=%s&lte_date=%s",
		baseURL, apiID, affiliateID, hits, startDate, endDate)
	log.Printf(apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// DMM API レスポンス用構造体（内部で使用）
	var dmmResp struct {
		Result struct {
			Items []struct {
				ContentID string `json:"content_id"`
				Title     string `json:"title"`
				Date      string `json:"date"`
				URL       string `json:"URL"`
				ImageURL  struct {
					Large string `json:"large"`
				} `json:"imageURL"`
				SampleMovieURL *struct {
					Size720480 string `json:"size_720_480"`
				} `json:"sampleMovieURL,omitempty"`
				Prices struct {
					Price string `json:"price"`
				} `json:"prices"`
				ItemInfo struct {
					Actress  []entity.Actress  `json:"actress"`
					Genre    []entity.Genre    `json:"genre"`
					Maker    []entity.Maker    `json:"maker"`
					Series   []entity.Series   `json:"series"`
					Director []entity.Director `json:"director"`
				} `json:"iteminfo"`
			} `json:"items"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &dmmResp); err != nil {
		return nil, err
	}

	var videos []entity.Video
	for _, item := range dmmResp.Result.Items {
		price, _ := strconv.Atoi(item.Prices.Price)
		sampleURL := ""
		if item.SampleMovieURL != nil {
			sampleURL = item.SampleMovieURL.Size720480
		}

		video := entity.Video{
			DmmID:        item.ContentID,
			Title:        item.Title,
			ThumbnailURL: item.ImageURL.Large,
			CreatedAt:    parseDate(item.Date),
			Price:        price,
			LikesCount:   rand.Intn(1000),
			SampleURL:    sampleURL,
			URL:          item.URL,
			Actresses:    item.ItemInfo.Actress,
			Genres:       item.ItemInfo.Genre,
			Makers:       item.ItemInfo.Maker,
			Series:       item.ItemInfo.Series,
			Directors:    item.ItemInfo.Director,
		}
		videos = append(videos, video)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(videos), func(i, j int) { videos[i], videos[j] = videos[j], videos[i] })
	return videos, nil
}

func (r *dmmVideoRepository) GetVideoById(ctx context.Context, dmmID string) (*entity.Video, error) {
	// TODO: 実装
	return nil, nil
}

// SearchVideos はキーワードや各種IDで動画を検索する
func (r *dmmVideoRepository) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error) {
	// TODO: 実装
	return nil, nil
}

func parseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return time.Now()
	}
	return t
}