package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Video は動画データのモデル
type Video struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	DmmVideoId   string    `json:"dmmVideoId"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	CreatedAt    time.Time `json:"createdAt"`
	LikesCount   int       `json:"likesCount"`
	SampleURL    string    `json:"sampleUrl"`
	URL          string    `json:"url"`
	Author       User      `json:"author"`
}

// User はユーザーデータのモデル
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl"`
}

// DMM APIのレスポンス構造体
type DMMResponse struct {
	Request struct {
		Parameters map[string]string `json:"parameters"`
	} `json:"request"`
	Result struct {
		Status       int `json:"status"`
		ResultCount  int `json:"result_count"`
		TotalCount   int `json:"total_count"`
		FirstPosition int `json:"first_position"`
		Items        []DMMItem `json:"items"`
	} `json:"result"`
}

type DMMItem struct {
	ServiceCode   string `json:"service_code"`
	ServiceName   string `json:"service_name"`
	FloorCode     string `json:"floor_code"`
	FloorName     string `json:"floor_name"`
	CategoryName  string `json:"category_name"`
	ContentID     string `json:"content_id"`
	ProductID     string `json:"product_id"`
	Title         string `json:"title"`
	Volume        string `json:"volume"`
	URL           string `json:"URL"`
	AffiliateURL  string `json:"affiliateURL"`
	ImageURL      struct {
		List  string `json:"list"`
		Small string `json:"small"`
		Large string `json:"large"`
	} `json:"imageURL"`
	SampleImageURL struct {
		SampleS struct {
			Image []string `json:"image"`
		} `json:"sample_s"`
		SampleL struct {
			Image []string `json:"image"`
		} `json:"sample_l"`
	} `json:"sampleImageURL"`
	SampleMovieURL *struct {
		Size476306  string `json:"size_476_306"`
		Size560360  string `json:"size_560_360"`
		Size644414  string `json:"size_644_414"`
		Size720480  string `json:"size_720_480"`
		PCFlag      int    `json:"pc_flag"`
		SPFlag      int    `json:"sp_flag"`
	} `json:"sampleMovieURL,omitempty"`
	Prices struct {
		Price     string `json:"price"`
		ListPrice string `json:"list_price"`
	} `json:"prices"`
	Date     string `json:"date"`
	ItemInfo struct {
		Genre []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"genre"`
		Maker []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"maker"`
		Actress []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Ruby string `json:"ruby"`
		} `json:"actress,omitempty"`
	} `json:"iteminfo"`
}

// グローバル変数としてのビデオリスト
var videos []Video

// アプリケーション初期化時にDMM APIからデータを取得
func initData() error {
	// DMM APIへのリクエストURL
	apiURL := "https://api.dmm.com/affiliate/v3/ItemList?api_id=PJTf3xEACNeaM7AbraTm&affiliate_id=notitle420-990&site=FANZA&service=digital&floor=videoa&hits=100&sort=date&output=json&gte_date=2025-03-05T00:00:00&lte_date=2025-03-06T00:00:00"

	// APIリクエスト
	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("APIリクエストエラー: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスの読み込み
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("レスポンス読み込みエラー: %v", err)
	}

	// JSONデコード
	var dmmResp DMMResponse
	if err := json.Unmarshal(body, &dmmResp); err != nil {
		return fmt.Errorf("JSONデコードエラー: %v", err)
	}

	// サンプル動画URLがある動画のみをフィルタリング
	var filteredItems []DMMItem
	for _, item := range dmmResp.Result.Items {
		if item.SampleMovieURL != nil {
			filteredItems = append(filteredItems, item)
		}
	}

	log.Printf("DMM APIから取得した動画数: %d", len(dmmResp.Result.Items))
	log.Printf("サンプル動画URLがある動画数: %d", len(filteredItems))

	// ビデオリストの作成
	videos = make([]Video, 0, len(filteredItems))
	
	for i, item := range filteredItems {
		// 女優名の取得（存在する場合）
		var authorName string
		if len(item.ItemInfo.Actress) > 0 {
			authorName = item.ItemInfo.Actress[0].Name
		} else if len(item.ItemInfo.Maker) > 0 {
			authorName = item.ItemInfo.Maker[0].Name + "所属女優"
		} else {
			authorName = "不明な女優"
		}
		
		// ジャンルの取得（説明用）
		var genres []string
		for _, genre := range item.ItemInfo.Genre {
			genres = append(genres, genre.Name)
		}
		
		// 動画オブジェクトの作成
		video := Video{
			ID:           fmt.Sprintf("%d", i+1),
			Title:        item.Title,
			Description:  fmt.Sprintf("%sの動画作品。人気ジャンル: %v", authorName, genres),
			DmmVideoId:   item.ContentID,
			ThumbnailURL: item.ImageURL.Large,
			CreatedAt:    parseDate(item.Date),
			LikesCount:   800 + i*10, // ダミーのいいね数
			SampleURL:    item.SampleMovieURL.Size720480,
			URL:          item.URL, // DMMのページURLを追加
			Author: User{
				ID:        fmt.Sprintf("actor%d", i+1),
				Username:  authorName,
				AvatarURL: fmt.Sprintf("/avatars/actor%d.jpg", i+1),
			},
		}
		
		videos = append(videos, video)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(videos), func(i, j int) {
		videos[i], videos[j] = videos[j], videos[i]
	})

	return nil
}

// 日付文字列をパースする補助関数
func parseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		// エラーの場合は現在時刻を返す
		return time.Now()
	}
	return t
}

// APIエラーレスポンスの構造体
type ErrorResponse struct {
	Error string `json:"error"`
}

// 動画リスト取得API
func getVideos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}

// 動画メタデータ取得API
func getVideoByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	
	for _, video := range videos {
		if video.ID == params["id"] {
			json.NewEncoder(w).Encode(video)
			return
		}
	}
	
	// 見つからない場合は404を返す
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(ErrorResponse{Error: "Video not found"})
}

func main() {
	// 起動時にDMM APIからデータを取得
	log.Println("DMM APIからデータを取得中...")
	if err := initData(); err != nil {
		log.Printf("データ取得エラー: %v", err)
		log.Println("フォールバックとしてモックデータを使用します")
		// エラー時はフォールバックデータを使用
		createFallbackData()
	}
	
	r := mux.NewRouter()
	
	// ルーティング設定
	r.HandleFunc("/api/videos", getVideos).Methods("GET")
	r.HandleFunc("/api/videos/{id}", getVideoByID).Methods("GET")
	
	// CORS設定
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})
	
	// サーバー起動
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", c.Handler(r)))
}

// APIリクエストに失敗した場合のフォールバックデータ
func createFallbackData() {
	videos = []Video{
		{
			ID:           "1",
			Title:        "配信限定 マドンナ専属女優の『リアル』解禁。 MADOOOON！！！！ 通野未帆 ハメ撮り",
			Description:  "通野未帆のリアル解禁作品。マドンナレーベルからの配信限定コンテンツ。",
			DmmVideoId:   "mdon00072",
			ThumbnailURL: "https://pics.dmm.co.jp/digital/video/mdon00072/mdon00072pl.jpg",
			CreatedAt:    parseDate("2025-04-01 00:00:00"),
			LikesCount:   950,
			URL:          "https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=mdon00072/",
			Author: User{
				ID:        "actor1",
				Username:  "通野未帆",
				AvatarURL: "/avatars/tonomidori.jpg",
			},
		},
		{
			ID:           "2",
			Title:        "配信限定 マドンナ専属女優の『リアル』解禁。SEASON2 MADOOOON！！！！ 一色桃子 ハメ撮り",
			Description:  "一色桃子のリアル解禁作品。マドンナレーベルからのSEASON2配信コンテンツ。",
			DmmVideoId:   "mdon00071",
			ThumbnailURL: "https://pics.dmm.co.jp/digital/video/mdon00071/mdon00071pl.jpg",
			CreatedAt:    parseDate("2025-04-01 00:00:00"),
			LikesCount:   900,
			URL:          "https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=mdon00071/",
			Author: User{
				ID:        "actor2",
				Username:  "一色桃子",
				AvatarURL: "/avatars/isshikimomoko.jpg",
			},
		},
	}
}