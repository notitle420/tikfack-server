package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// 生成された proto メッセージ（proto ファイルの go_package に合わせる）
	pb "server/generated"
	// 生成された Connect 用ハンドラー。パッケージ名は生成設定に合わせて修正してください。
	protoconnect "server/generated/protoconnect"
)

// Video はサーバ内部で扱う動画データのモデル
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

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl"`
}

// DMMResponse と DMMItem は DMM API の JSON レスポンスをパースするための構造体
type DMMResponse struct {
	Request struct {
		Parameters map[string]string `json:"parameters"`
	} `json:"request"`
	Result struct {
		Status        int       `json:"status"`
		ResultCount   int       `json:"result_count"`
		TotalCount    int       `json:"total_count"`
		FirstPosition int       `json:"first_position"`
		Items         []DMMItem `json:"items"`
	} `json:"result"`
}

type DMMItem struct {
	ServiceCode  string `json:"service_code"`
	ServiceName  string `json:"service_name"`
	FloorCode    string `json:"floor_code"`
	FloorName    string `json:"floor_name"`
	CategoryName string `json:"category_name"`
	ContentID    string `json:"content_id"`
	ProductID    string `json:"product_id"`
	Title        string `json:"title"`
	Volume       string `json:"volume"`
	URL          string `json:"URL"`
	AffiliateURL string `json:"affiliateURL"`
	ImageURL     struct {
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
		Size476306 string `json:"size_476_306"`
		Size560360 string `json:"size_560_360"`
		Size644414 string `json:"size_644_414"`
		Size720480 string `json:"size_720_480"`
		PCFlag     int    `json:"pc_flag"`
		SPFlag     int    `json:"sp_flag"`
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

// parseDate は DMM API の日付文字列を time.Time にパースする補助関数
func parseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return time.Now()
	}
	return t
}

// getVideosFromDMM は指定した日付の動画データを DMM API から取得し内部モデルに変換する
func getVideosFromDMM(targetDate time.Time) ([]Video, error) {
	apiID := os.Getenv("DMM_API_ID")
	if apiID == "" {
		return nil, fmt.Errorf("DMM_API_ID not set")
	}
	affiliateID := os.Getenv("DMM_API_AFFILIATE_ID")
	if affiliateID == "" {
		return nil, fmt.Errorf("DMM_API_AFFILIATE_ID not set")
	}

	startDate := targetDate.Format("2006-01-02T00:00:00")
	endDate := targetDate.AddDate(0, 0, 1).Format("2006-01-02T00:00:00")
	apiURL := fmt.Sprintf("https://api.dmm.com/affiliate/v3/ItemList?api_id=%s&affiliate_id=%s&site=FANZA&service=digital&floor=videoa&hits=10&sort=date&output=json&gte_date=%s&lte_date=%s",
		apiID, affiliateID, startDate, endDate)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dmmResp DMMResponse
	if err := json.Unmarshal(body, &dmmResp); err != nil {
		return nil, err
	}

	var videos []Video
	for i, item := range dmmResp.Result.Items {
		sampleURL := ""
		if item.SampleMovieURL != nil {
			sampleURL = item.SampleMovieURL.Size720480
		}
		var authorName string
		if len(item.ItemInfo.Actress) > 0 {
			authorName = item.ItemInfo.Actress[0].Name
		} else if len(item.ItemInfo.Maker) > 0 {
			authorName = item.ItemInfo.Maker[0].Name + "所属女優"
		} else {
			authorName = "不明な女優"
		}
		var genres []string
		for _, genre := range item.ItemInfo.Genre {
			genres = append(genres, genre.Name)
		}
		video := Video{
			ID:           item.ContentID,
			Title:        item.Title,
			Description:  fmt.Sprintf("%sの動画作品。人気ジャンル: %v", authorName, genres),
			DmmVideoId:   item.ContentID,
			ThumbnailURL: item.ImageURL.Large,
			CreatedAt:    parseDate(item.Date),
			LikesCount:   800 + i*10,
			SampleURL:    sampleURL,
			URL:          item.URL,
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
	return videos, nil
}

// サーバー側で直接の動画URLを検証し取得するための関数



func getValidVideoUrl(dmmVideoId string) (string, error) {
	log.Printf("validate")
	var reAlt0 = regexp.MustCompile(`^([A-Za-z0-9_]+?)0(\d+)([A-Za-z])?$`)
	var reAlt00 = regexp.MustCompile(`^([A-Za-z0-9_]+?)00(\d+)([A-Za-z])?$`)
	generateUrl := func(id string) string {
		if len(id) < 3 {
			return ""
		}
		firstChar := id[0:1]
		firstThreeChars := id[0:3]
		return fmt.Sprintf("https://cc3001.dmm.co.jp/litevideo/freepv/%s/%s/%s/%smhb.mp4", firstChar, firstThreeChars, id, id)
	}

	originalUrl := generateUrl(dmmVideoId)
	alternativeUrl0 := generateUrl(reAlt0.ReplaceAllString(dmmVideoId, "$1$2$3"))
	alternativeUrl00 := generateUrl(reAlt00.ReplaceAllString(dmmVideoId, "$1$2$3"))

	urls := []string{originalUrl, alternativeUrl0, alternativeUrl00}
	for _, url := range urls {
		resp, err := http.Head(url)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return url, nil
		}
	}
	return "", fmt.Errorf("有効な動画URLが見つかりませんでした")
}

//
// Connect サーバー実装
//

type videoServiceServer struct{}

func (s *videoServiceServer) GetVideos(ctx context.Context, req *connect.Request[pb.GetVideosRequest]) (*connect.Response[pb.GetVideosResponse], error) {
	var targetDate time.Time
	if req.Msg.Date == "" {
		targetDate = time.Now()
	} else {
		t, err := time.Parse("2006-01-02", req.Msg.Date)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "不正な日付形式です")
		}
		targetDate = t
	}
	videos, err := getVideosFromDMM(targetDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "動画の取得に失敗しました: %v", err)
	}
	var pbVideos []*pb.Video
	for _, v := range videos {
		// サーバー側で有効な動画URLを検証
		directUrl, err := getValidVideoUrl(v.DmmVideoId)
		if err != nil {
			// もし検証に失敗したら、fallback として getDirectVideoUrl を使用
			directUrl = "none"
		}
		pbVideo := &pb.Video{
			Id:           v.ID,
			Title:        v.Title,
			Description:  v.Description,
			DmmId:        v.DmmVideoId,
			ThumbnailUrl: v.ThumbnailURL,
			CreatedAt:    v.CreatedAt.Format("2006-01-02 15:04:05"),
			LikesCount:   int32(v.LikesCount),
			SampleUrl:    v.SampleURL,
			Url:          v.URL,
			DirectUrl:    directUrl, // 新たに追加したフィールド（proto に定義済み）
			Author: &pb.User{
				Id:        v.Author.ID,
				Username:  v.Author.Username,
				AvatarUrl: v.Author.AvatarURL,
			},
		}
		pbVideos = append(pbVideos, pbVideo)
	}
	res := &pb.GetVideosResponse{Videos: pbVideos}
	return connect.NewResponse(res), nil
}

func (s *videoServiceServer) GetVideoById(ctx context.Context, req *connect.Request[pb.GetVideoByIdRequest]) (*connect.Response[pb.GetVideoByIdResponse], error) {
	videos, err := getVideosFromDMM(time.Now())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "動画の取得に失敗しました: %v", err)
	}
	for _, v := range videos {
		if v.ID == req.Msg.Id {
			directUrl, err := getValidVideoUrl(v.DmmVideoId)
			if err != nil {
				directUrl = "None"
			}
			pbVideo := &pb.Video{
				Id:           v.ID,
				Title:        v.Title,
				Description:  v.Description,
				DmmId:        v.DmmVideoId,
				ThumbnailUrl: v.ThumbnailURL,
				CreatedAt:    v.CreatedAt.Format("2006-01-02 15:04:05"),
				LikesCount:   int32(v.LikesCount),
				SampleUrl:    v.SampleURL,
				Url:          v.URL,
				DirectUrl:    directUrl,
				Author: &pb.User{
					Id:        v.Author.ID,
					Username:  v.Author.Username,
					AvatarUrl: v.Author.AvatarURL,
				},
			}
			res := &pb.GetVideoByIdResponse{Video: pbVideo}
			return connect.NewResponse(res), nil
		}
	}
	return nil, status.Error(codes.NotFound, "video not found")
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env ファイルの読み込みに失敗しました: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	mux := http.NewServeMux()
	pattern, handler := protoconnect.NewVideoServiceHandler(&videoServiceServer{}, connect.WithCompressMinBytes(0))
	mux.Handle(pattern, handler)
	handlerWithCORS := cors.AllowAll().Handler(mux)
	log.Printf("Connect gRPC server is running on :%s", port)
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}