package dmmapi

//go:generate mockgen -destination=mock_mapper.go -package=dmmapi github.com/tikfack/server/internal/infrastructure/dmmapi MapperInterface

import (
	"strconv"
	"strings"
	"time"

	"github.com/tikfack/server/internal/application/model"
	"github.com/tikfack/server/internal/infrastructure/util"
)

var resolveDirectURL = func(dmmID string) (string, error) {
	return util.GetValidVideoUrl(dmmID)
}

type MapperInterface interface {
	ConvertEntityFromDMM(Result) ([]model.Video, *model.SearchMetadata)
}

// ConvertEntityFromDMM は DMM API のレスポンス結果を model.Video と model.SearchMetadata に変換する
func ConvertEntityFromDMM(result Result) ([]model.Video, *model.SearchMetadata) {
	metadata := &model.SearchMetadata{
		ResultCount:   result.ResultCount,
		TotalCount:    result.TotalCount,
		FirstPosition: result.FirstPosition,
	}

	// 動画の変換
	videos := make([]model.Video, 0, len(result.Items))
	for _, item := range result.Items {
		// 価格や日付のパースなどロジックをここに集約
		price := parsePrice(item)
		created := parseDate(item.Date)

		// actor/genre/maker などの変換
		actresses := make([]model.Actress, 0, len(item.ItemInfo.Actress))
		for _, a := range item.ItemInfo.Actress {
			actresses = append(actresses, model.Actress{ID: strconv.Itoa(a.ID), Name: a.Name})
		}

		// genres変換
		genres := make([]model.Genre, 0, len(item.ItemInfo.Genre))
		for _, g := range item.ItemInfo.Genre {
			genres = append(genres, model.Genre{ID: strconv.Itoa(g.ID), Name: g.Name})
		}

		// makers変換
		makers := make([]model.Maker, 0, len(item.ItemInfo.Maker))
		for _, m := range item.ItemInfo.Maker {
			makers = append(makers, model.Maker{ID: strconv.Itoa(m.ID), Name: m.Name})
		}

		// series変換
		series := make([]model.Series, 0, len(item.ItemInfo.Series))
		for _, s := range item.ItemInfo.Series {
			series = append(series, model.Series{ID: strconv.Itoa(s.ID), Name: s.Name})
		}

		// directors変換
		directors := make([]model.Director, 0, len(item.ItemInfo.Director))
		for _, d := range item.ItemInfo.Director {
			directors = append(directors, model.Director{ID: strconv.Itoa(d.ID), Name: d.Name})
		}

		// サンプル動画URL
		var sampleURL string
		if item.SampleMovieURL != nil {
			sampleURL = item.SampleMovieURL.Size720480
		}

		directURL := ""
		if resolved, err := resolveDirectURL(item.ContentID); err == nil && resolved != "" {
			directURL = resolved
		}

		// レビュー情報
		review := model.Review{
			Count:   0,
			Average: 0,
		}
		if item.Review != nil {
			review.Count = item.Review.Count
			if average, err := strconv.ParseFloat(item.Review.Average, 32); err == nil {
				review.Average = float32(average)
			}
		}

		video := model.Video{
			DmmID:        item.ContentID,
			Title:        item.Title,
			DirectURL:    directURL,
			URL:          item.URL,
			SampleURL:    sampleURL,
			ThumbnailURL: item.ImageURL.Large,
			CreatedAt:    created,
			Price:        price,
			LikesCount:   0,
			Actresses:    actresses,
			Genres:       genres,
			Makers:       makers,
			Series:       series,
			Directors:    directors,
			Review:       review,
		}
		videos = append(videos, video)
	}

	return videos, metadata
}

func parsePrice(item Item) int {
	s := strings.TrimSpace(item.Prices.Price)
	if s == "" {
		return 0
	}

	// 価格範囲の場合は先頭の値を使用する
	if strings.Contains(s, "~") {
		s = strings.SplitN(s, "~", 2)[0]
	}

	s = strings.TrimSuffix(s, "円")
	p, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return p
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", dateStr)
	return t
}
