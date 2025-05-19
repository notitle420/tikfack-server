package dmmapi

//go:generate mockgen -destination=mock_mapper.go -package=dmmapi github.com/tikfack/server/internal/infrastructure/dmmapi MapperInterface

import (
	"strconv"
	"strings"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
)

type MapperInterface interface {
	ConvertEntityFromDMM(Result) ([]entity.Video, *entity.SearchMetadata)
}

// ConvertEntityFromDMM は DMM API のレスポンス結果を entity.Video と entity.SearchMetadata に変換する
func ConvertEntityFromDMM(result Result) ([]entity.Video, *entity.SearchMetadata) { // メタデータの変換
	metadata := &entity.SearchMetadata{
		ResultCount:   result.ResultCount,
		TotalCount:    result.TotalCount,
		FirstPosition: result.FirstPosition,
	}

	// 動画の変換
	videos := make([]entity.Video, 0, len(result.Items))
	for _, item := range result.Items {
		// 価格や日付のパースなどロジックをここに集約
		price := parsePrice(item)
		created := parseDate(item.Date)

		// actor/genre/maker などの変換
		actresses := make([]entity.Actress, 0, len(item.ItemInfo.Actress))
		for _, a := range item.ItemInfo.Actress {
			actresses = append(actresses, entity.Actress{ID: strconv.Itoa(a.ID), Name: a.Name})
		}

		// genres変換
		genres := make([]entity.Genre, 0, len(item.ItemInfo.Genre))
		for _, g := range item.ItemInfo.Genre {
			genres = append(genres, entity.Genre{ID: strconv.Itoa(g.ID), Name: g.Name})
		}

		// makers変換
		makers := make([]entity.Maker, 0, len(item.ItemInfo.Maker))
		for _, m := range item.ItemInfo.Maker {
			makers = append(makers, entity.Maker{ID: strconv.Itoa(m.ID), Name: m.Name})
		}

		// series変換
		series := make([]entity.Series, 0, len(item.ItemInfo.Series))
		for _, s := range item.ItemInfo.Series {
			series = append(series, entity.Series{ID: strconv.Itoa(s.ID), Name: s.Name})
		}

		// directors変換
		directors := make([]entity.Director, 0, len(item.ItemInfo.Director))
		for _, d := range item.ItemInfo.Director {
			directors = append(directors, entity.Director{ID: strconv.Itoa(d.ID), Name: d.Name})
		}

		// サンプル動画URL
		var sampleURL string
		if item.SampleMovieURL != nil {
			sampleURL = item.SampleMovieURL.Size720480
		}

		// レビュー情報
		review := entity.Review{
			Count:   0,
			Average: 0,
		}
		if item.Review != nil {
			review.Count = item.Review.Count
			if average, err := strconv.ParseFloat(item.Review.Average, 32); err == nil {
				review.Average = float32(average)
			}
		}

		video := entity.Video{
			DmmID:        item.ContentID,
			Title:        item.Title,
			DirectURL:    sampleURL,
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
