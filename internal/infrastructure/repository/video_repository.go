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
	"strings"
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
	log.Printf("API URL: %s", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	log.Printf("Response status: %s", resp.Status)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// レスポンスボディの先頭部分をデバッグ出力
	if len(body) > 100 {
		log.Printf("Response body (first 100 chars): %s...", string(body[:100]))
	} else {
		log.Printf("Response body: %s", string(body))
	}

	// DMM API レスポンス用構造体（内部で使用）
	// APIのレスポンスに合わせて型を定義
	type dmmGenre struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmActress struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmMaker struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmSeries struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmDirector struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
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
					Price string `json:"price,omitempty"`
					Deliveries *struct {
						Delivery []struct {
							Type      string `json:"type"`
							Price     string `json:"price"`
							ListPrice string `json:"list_price"`
						} `json:"delivery"`
					} `json:"deliveries,omitempty"`
				} `json:"prices"`
				ItemInfo struct {
					Actress  []dmmActress  `json:"actress,omitempty"`
					Genre    []dmmGenre    `json:"genre,omitempty"`
					Maker    []dmmMaker    `json:"maker,omitempty"`
					Series   []dmmSeries   `json:"series,omitempty"`
					Director []dmmDirector `json:"director,omitempty"`
				} `json:"iteminfo"`
			} `json:"items"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &dmmResp); err != nil {
		log.Printf("JSON unmarshaling error: %v", err)
		return nil, err
	}

	var videos []entity.Video
	for i, item := range dmmResp.Result.Items {
		// 価格の取得方法を調整
		price := 0
		if item.Prices.Price != "" {
			// 単純な価格文字列から数値への変換を試みる
			cleanedPrice := strings.Replace(strings.Replace(item.Prices.Price, "~", "", -1), "円", "", -1)
			parsedPrice, err := strconv.Atoi(cleanedPrice)
			if err == nil {
				price = parsedPrice
			}
		} else if item.Prices.Deliveries != nil && len(item.Prices.Deliveries.Delivery) > 0 {
			// 配信タイプ別の価格から一つ選択
			for _, delivery := range item.Prices.Deliveries.Delivery {
				if delivery.Type == "download" || delivery.Type == "stream" {
					parsedPrice, err := strconv.Atoi(delivery.Price)
					if err == nil {
						price = parsedPrice
						break
					}
				}
			}
		}
		
		sampleURL := ""
		if item.SampleMovieURL != nil {
			sampleURL = item.SampleMovieURL.Size720480
		}
		
		// DMM APIからのデータをエンティティに変換
		// ID型を文字列に変換
		actresses := make([]entity.Actress, 0, len(item.ItemInfo.Actress))
		for _, a := range item.ItemInfo.Actress {
			actresses = append(actresses, entity.Actress{
				ID:   strconv.Itoa(a.ID),
				Name: a.Name,
			})
		}
		
		genres := make([]entity.Genre, 0, len(item.ItemInfo.Genre))
		for _, g := range item.ItemInfo.Genre {
			genres = append(genres, entity.Genre{
				ID:   strconv.Itoa(g.ID),
				Name: g.Name,
			})
		}
		
		makers := make([]entity.Maker, 0, len(item.ItemInfo.Maker))
		for _, m := range item.ItemInfo.Maker {
			makers = append(makers, entity.Maker{
				ID:   strconv.Itoa(m.ID),
				Name: m.Name,
			})
		}
		
		series := make([]entity.Series, 0, len(item.ItemInfo.Series))
		for _, s := range item.ItemInfo.Series {
			series = append(series, entity.Series{
				ID:   strconv.Itoa(s.ID),
				Name: s.Name,
			})
		}
		
		directors := make([]entity.Director, 0, len(item.ItemInfo.Director))
		for _, d := range item.ItemInfo.Director {
			directors = append(directors, entity.Director{
				ID:   strconv.Itoa(d.ID),
				Name: d.Name,
			})
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
			Actresses:    actresses,
			Genres:       genres,
			Makers:       makers,
			Series:       series,
			Directors:    directors,
		}
		videos = append(videos, video)
		
		// 最初の動画の詳細のみログに出力
		if i == 0 {
			log.Printf("最初の動画データ: ID=%s, タイトル=%s", video.DmmID, video.Title)
			log.Printf("  URL: %s", video.URL)
			log.Printf("  サムネイル: %s", video.ThumbnailURL)
			log.Printf("  サンプル動画: %s", video.SampleURL)
			log.Printf("  価格: %d円", video.Price)
			
			if len(video.Actresses) > 0 {
				actressNames := make([]string, 0, len(video.Actresses))
				for _, a := range video.Actresses {
					actressNames = append(actressNames, a.Name)
				}
				log.Printf("  女優: %s", strings.Join(actressNames, ", "))
			}
			
			if len(video.Genres) > 0 {
				genreNames := make([]string, 0, len(video.Genres))
				for _, g := range video.Genres {
					genreNames = append(genreNames, g.Name)
				}
				log.Printf("  ジャンル: %s", strings.Join(genreNames, ", "))
			}
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(videos), func(i, j int) { videos[i], videos[j] = videos[j], videos[i] })
	return videos, nil
}

func (r *dmmVideoRepository) GetVideoById(ctx context.Context, dmmID string) (*entity.Video, error) {
	// まず日付を指定せずに動画を取得し、その中からIDで検索する
	videos, err := r.GetVideosByDate(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	// IDに一致する動画を探す
	for _, video := range videos {
		if video.DmmID == dmmID {
			return &video, nil
		}
	}

	// 見つからなかった場合
	return nil, fmt.Errorf("動画ID %s が見つかりませんでした", dmmID)
}

// SearchVideos はキーワードや各種IDで動画を検索する
func (r *dmmVideoRepository) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error) {
	// TODO: 実装
	return nil, nil
}

// GetVideosByID は複数のIDを使用して動画を検索する
func (r *dmmVideoRepository) GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	log.Printf("GetVideosByID called")
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

	defaultHits := os.Getenv("HITS")
	if defaultHits == "" && hits == 0 {
		return nil, fmt.Errorf("HITS not set")
	}
	
	requestHits := defaultHits
	if hits > 0 {
		requestHits = strconv.Itoa(int(hits))
	}
	
	apiURL := fmt.Sprintf("%s/v3/ItemList?api_id=%s&affiliate_id=%s&hits=%s", baseURL, apiID, affiliateID, requestHits)
	
	if site != "" {
		apiURL += fmt.Sprintf("&site=%s", site)
	} else {
		apiURL += "&site=FANZA"
	}
	
	if service != "" {
		apiURL += fmt.Sprintf("&service=%s", service)
	} else {
		apiURL += "&service=digital"
	}
	
	if floor != "" {
		apiURL += fmt.Sprintf("&floor=%s", floor)
	} else {
		apiURL += "&floor=videoa"
	}
	
	if sort != "" {
		apiURL += fmt.Sprintf("&sort=%s", sort)
	} else {
		apiURL += "&sort=rank"
	}
	
	if offset > 0 {
		apiURL += fmt.Sprintf("&offset=%d", offset)
	}
	
	if gteDate != "" {
		apiURL += fmt.Sprintf("&gte_date=%s", gteDate)
	}
	
	if lteDate != "" {
		apiURL += fmt.Sprintf("&lte_date=%s", lteDate)
	}
	
	for _, id := range actressIDs {
		if id != "" {
			apiURL += fmt.Sprintf("&article=actress&article_id=%s", id)
		}
	}
	
	for _, id := range genreIDs {
		if id != "" {
			apiURL += fmt.Sprintf("&article=genre&article_id=%s", id)
		}
	}
	
	for _, id := range makerIDs {
		if id != "" {
			apiURL += fmt.Sprintf("&article=maker&article_id=%s", id)
		}
	}
	
	for _, id := range seriesIDs {
		if id != "" {
			apiURL += fmt.Sprintf("&article=series&article_id=%s", id)
		}
	}
	
	for _, id := range directorIDs {
		if id != "" {
			apiURL += fmt.Sprintf("&article=director&article_id=%s", id)
		}
	}
	
	apiURL += "&output=json"
	
	log.Printf("API URL: %s", apiURL)
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	log.Printf("Response status: %s", resp.Status)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// レスポンスボディの先頭部分をデバッグ出力
	if len(body) > 100 {
		log.Printf("Response body (first 100 chars): %s...", string(body[:100]))
	} else {
		log.Printf("Response body: %s", string(body))
	}
	
	// DMM API レスポンス用構造体（内部で使用）
	type dmmGenre struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmActress struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmMaker struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmSeries struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmDirector struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
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
					Price string `json:"price,omitempty"`
					Deliveries *struct {
						Delivery []struct {
							Type      string `json:"type"`
							Price     string `json:"price"`
							ListPrice string `json:"list_price"`
						} `json:"delivery"`
					} `json:"deliveries,omitempty"`
				} `json:"prices"`
				ItemInfo struct {
					Actress  []dmmActress  `json:"actress,omitempty"`
					Genre    []dmmGenre    `json:"genre,omitempty"`
					Maker    []dmmMaker    `json:"maker,omitempty"`
					Series   []dmmSeries   `json:"series,omitempty"`
					Director []dmmDirector `json:"director,omitempty"`
				} `json:"iteminfo"`
			} `json:"items"`
		} `json:"result"`
	}
	
	if err := json.Unmarshal(body, &dmmResp); err != nil {
		log.Printf("JSON unmarshaling error: %v", err)
		return nil, err
	}
	
	var videos []entity.Video
	for i, item := range dmmResp.Result.Items {
		// 価格の取得方法を調整
		price := 0
		if item.Prices.Price != "" {
			// 単純な価格文字列から数値への変換を試みる
			cleanedPrice := strings.Replace(strings.Replace(item.Prices.Price, "~", "", -1), "円", "", -1)
			parsedPrice, err := strconv.Atoi(cleanedPrice)
			if err == nil {
				price = parsedPrice
			}
		} else if item.Prices.Deliveries != nil && len(item.Prices.Deliveries.Delivery) > 0 {
			// 配信タイプ別の価格から一つ選択
			for _, delivery := range item.Prices.Deliveries.Delivery {
				if delivery.Type == "download" || delivery.Type == "stream" {
					parsedPrice, err := strconv.Atoi(delivery.Price)
					if err == nil {
						price = parsedPrice
						break
					}
				}
			}
		}
		
		sampleURL := ""
		if item.SampleMovieURL != nil {
			sampleURL = item.SampleMovieURL.Size720480
		}
		
		// DMM APIからのデータをエンティティに変換
		// ID型を文字列に変換
		actresses := make([]entity.Actress, 0, len(item.ItemInfo.Actress))
		for _, a := range item.ItemInfo.Actress {
			actresses = append(actresses, entity.Actress{
				ID:   strconv.Itoa(a.ID),
				Name: a.Name,
			})
		}
		
		genres := make([]entity.Genre, 0, len(item.ItemInfo.Genre))
		for _, g := range item.ItemInfo.Genre {
			genres = append(genres, entity.Genre{
				ID:   strconv.Itoa(g.ID),
				Name: g.Name,
			})
		}
		
		makers := make([]entity.Maker, 0, len(item.ItemInfo.Maker))
		for _, m := range item.ItemInfo.Maker {
			makers = append(makers, entity.Maker{
				ID:   strconv.Itoa(m.ID),
				Name: m.Name,
			})
		}
		
		series := make([]entity.Series, 0, len(item.ItemInfo.Series))
		for _, s := range item.ItemInfo.Series {
			series = append(series, entity.Series{
				ID:   strconv.Itoa(s.ID),
				Name: s.Name,
			})
		}
		
		directors := make([]entity.Director, 0, len(item.ItemInfo.Director))
		for _, d := range item.ItemInfo.Director {
			directors = append(directors, entity.Director{
				ID:   strconv.Itoa(d.ID),
				Name: d.Name,
			})
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
			Actresses:    actresses,
			Genres:       genres,
			Makers:       makers,
			Series:       series,
			Directors:    directors,
		}
		videos = append(videos, video)
		
		// 最初の動画の詳細のみログに出力
		if i == 0 {
			log.Printf("最初の動画データ: ID=%s, タイトル=%s", video.DmmID, video.Title)
		}
	}
	
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(videos), func(i, j int) { videos[i], videos[j] = videos[j], videos[i] })
	return videos, nil
}

// GetVideosByKeyword はキーワードを使用して動画を検索する
func (r *dmmVideoRepository) GetVideosByKeyword(ctx context.Context, keyword string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	log.Printf("GetVideosByKeyword called")
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

	defaultHits := os.Getenv("HITS")
	if defaultHits == "" && hits == 0 {
		return nil, fmt.Errorf("HITS not set")
	}
	
	requestHits := defaultHits
	if hits > 0 {
		requestHits = strconv.Itoa(int(hits))
	}
	
	apiURL := fmt.Sprintf("%s/v3/ItemList?api_id=%s&affiliate_id=%s&hits=%s", baseURL, apiID, affiliateID, requestHits)
	
	if site != "" {
		apiURL += fmt.Sprintf("&site=%s", site)
	} else {
		apiURL += "&site=FANZA"
	}
	
	if service != "" {
		apiURL += fmt.Sprintf("&service=%s", service)
	} else {
		apiURL += "&service=digital"
	}
	
	if floor != "" {
		apiURL += fmt.Sprintf("&floor=%s", floor)
	} else {
		apiURL += "&floor=videoa"
	}
	
	if keyword != "" {
		apiURL += fmt.Sprintf("&keyword=%s", keyword)
	}
	
	if sort != "" {
		apiURL += fmt.Sprintf("&sort=%s", sort)
	} else {
		apiURL += "&sort=match" // キーワード検索の場合はマッチング順がデフォルト
	}
	
	if offset > 0 {
		apiURL += fmt.Sprintf("&offset=%d", offset)
	}
	
	if gteDate != "" {
		apiURL += fmt.Sprintf("&gte_date=%s", gteDate)
	}
	
	if lteDate != "" {
		apiURL += fmt.Sprintf("&lte_date=%s", lteDate)
	}
	
	apiURL += "&output=json"
	
	log.Printf("API URL: %s", apiURL)
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	log.Printf("Response status: %s", resp.Status)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// レスポンスボディの先頭部分をデバッグ出力
	if len(body) > 100 {
		log.Printf("Response body (first 100 chars): %s...", string(body[:100]))
	} else {
		log.Printf("Response body: %s", string(body))
	}
	
	// DMM API レスポンス用構造体（内部で使用）
	type dmmGenre struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmActress struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmMaker struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmSeries struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	type dmmDirector struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
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
					Price string `json:"price,omitempty"`
					Deliveries *struct {
						Delivery []struct {
							Type      string `json:"type"`
							Price     string `json:"price"`
							ListPrice string `json:"list_price"`
						} `json:"delivery"`
					} `json:"deliveries,omitempty"`
				} `json:"prices"`
				ItemInfo struct {
					Actress  []dmmActress  `json:"actress,omitempty"`
					Genre    []dmmGenre    `json:"genre,omitempty"`
					Maker    []dmmMaker    `json:"maker,omitempty"`
					Series   []dmmSeries   `json:"series,omitempty"`
					Director []dmmDirector `json:"director,omitempty"`
				} `json:"iteminfo"`
			} `json:"items"`
		} `json:"result"`
	}
	
	if err := json.Unmarshal(body, &dmmResp); err != nil {
		log.Printf("JSON unmarshaling error: %v", err)
		return nil, err
	}
	
	var videos []entity.Video
	for i, item := range dmmResp.Result.Items {
		// 価格の取得方法を調整
		price := 0
		if item.Prices.Price != "" {
			// 単純な価格文字列から数値への変換を試みる
			cleanedPrice := strings.Replace(strings.Replace(item.Prices.Price, "~", "", -1), "円", "", -1)
			parsedPrice, err := strconv.Atoi(cleanedPrice)
			if err == nil {
				price = parsedPrice
			}
		} else if item.Prices.Deliveries != nil && len(item.Prices.Deliveries.Delivery) > 0 {
			// 配信タイプ別の価格から一つ選択
			for _, delivery := range item.Prices.Deliveries.Delivery {
				if delivery.Type == "download" || delivery.Type == "stream" {
					parsedPrice, err := strconv.Atoi(delivery.Price)
					if err == nil {
						price = parsedPrice
						break
					}
				}
			}
		}
		
		sampleURL := ""
		if item.SampleMovieURL != nil {
			sampleURL = item.SampleMovieURL.Size720480
		}
		
		// DMM APIからのデータをエンティティに変換
		// ID型を文字列に変換
		actresses := make([]entity.Actress, 0, len(item.ItemInfo.Actress))
		for _, a := range item.ItemInfo.Actress {
			actresses = append(actresses, entity.Actress{
				ID:   strconv.Itoa(a.ID),
				Name: a.Name,
			})
		}
		
		genres := make([]entity.Genre, 0, len(item.ItemInfo.Genre))
		for _, g := range item.ItemInfo.Genre {
			genres = append(genres, entity.Genre{
				ID:   strconv.Itoa(g.ID),
				Name: g.Name,
			})
		}
		
		makers := make([]entity.Maker, 0, len(item.ItemInfo.Maker))
		for _, m := range item.ItemInfo.Maker {
			makers = append(makers, entity.Maker{
				ID:   strconv.Itoa(m.ID),
				Name: m.Name,
			})
		}
		
		series := make([]entity.Series, 0, len(item.ItemInfo.Series))
		for _, s := range item.ItemInfo.Series {
			series = append(series, entity.Series{
				ID:   strconv.Itoa(s.ID),
				Name: s.Name,
			})
		}
		
		directors := make([]entity.Director, 0, len(item.ItemInfo.Director))
		for _, d := range item.ItemInfo.Director {
			directors = append(directors, entity.Director{
				ID:   strconv.Itoa(d.ID),
				Name: d.Name,
			})
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
			Actresses:    actresses,
			Genres:       genres,
			Makers:       makers,
			Series:       series,
			Directors:    directors,
		}
		videos = append(videos, video)
		
		// 最初の動画の詳細のみログに出力
		if i == 0 {
			log.Printf("最初の動画データ: ID=%s, タイトル=%s", video.DmmID, video.Title)
		}
	}
	
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(videos), func(i, j int) { videos[i], videos[j] = videos[j], videos[i] })
	return videos, nil
}

func parseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return time.Now()
	}
	return t
}
