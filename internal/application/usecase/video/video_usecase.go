package usecase

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/domain/repository"
)

// VideoUsecase は動画関連のユースケースを定義するインターフェイス
type VideoUsecase interface {
	GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error)
	GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error)
	SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error)
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

// GetVideosByDate は指定日付の動画一覧を取得し、動画URLを検証する
func (u *videoUsecase) GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error) {
	videos, err := u.videoRepo.GetVideosByDate(ctx, targetDate)
	if err != nil {
		return nil, err
	}
	for i, v := range videos {
		validatedUrl, err := getValidVideoUrl(v.DmmID)
		if err != nil {
			log.Printf("動画ID [%s] URL検証失敗: %v", v.DmmID, err)
			videos[i].DirectURL = ""
		} else {
			videos[i].DirectURL = validatedUrl
		}
	}
	return videos, nil
}

// GetVideoById は、指定された DMMビデオID の動画を取得する
func (u *videoUsecase) GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error) {
	return u.videoRepo.GetVideoById(ctx, dmmId)
}

// SearchVideos は動画をキーワードや各種IDで検索する
func (u *videoUsecase) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error) {
	videos, err := u.videoRepo.SearchVideos(ctx, keyword, actressID, genreID, makerID, seriesID, directorID)
	if err != nil {
		return nil, err
	}
	for i, v := range videos {
		validatedUrl, err := getValidVideoUrl(v.DmmID)
		if err != nil {
			log.Printf("動画ID [%s] URL検証失敗: %v", v.DmmID, err)
			videos[i].DirectURL = ""
		} else {
			videos[i].DirectURL = validatedUrl
		}
	}
	return videos, nil
}

// getValidVideoUrl はDMMビデオIDを検証し、有効な動画URLを返す
func getValidVideoUrl(dmmId string) (string, error) {
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

	urls := []string{
		generateUrl(dmmId),
		generateUrl(reAlt0.ReplaceAllString(dmmId, "$1$2$3")),
		generateUrl(reAlt00.ReplaceAllString(dmmId, "$1$2$3")),
	}
	for _, url := range urls {
		resp, err := http.Head(url)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return url, nil
		}
	}
	return "", fmt.Errorf("有効な動画URLが見つかりませんでした")
}
