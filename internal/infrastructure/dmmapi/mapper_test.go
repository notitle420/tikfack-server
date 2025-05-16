package dmmapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertItem(t *testing.T) {
	// テストケース1: レビュー情報あり
	t.Run("レビュー情報がある場合", func(t *testing.T) {
		item := Item{
			ContentID: "vid1",
			Title:     "テスト動画",
			Date:      "2024-01-01 00:00:00",
			URL:       "https://example.com/video1",
			ImageURL: struct {
				Large string `json:"large"`
			}{
				Large: "https://example.com/image1.jpg",
			},
			SampleMovieURL: &struct {
				Size720480 string `json:"size_720_480"`
			}{
				Size720480: "https://example.com/sample1.mp4",
			},
			Prices: struct {
				Price      string `json:"price,omitempty"`
				Deliveries *struct {
					Delivery []struct {
						Type      string `json:"type"`
						Price     string `json:"price"`
						ListPrice string `json:"list_price"`
					}
				} `json:"deliveries,omitempty"`
			}{
				Price: "1000円",
			},
			Review: &struct {
				Count   int    `json:"count"`
				Average string `json:"average"`
			}{
				Count:   42,
				Average: "4.5",
			},
			ItemInfo: struct {
				Actress  []Actress  `json:"actress,omitempty"`
				Genre    []Genre    `json:"genre,omitempty"`
				Maker    []Maker    `json:"maker,omitempty"`
				Series   []Series   `json:"series,omitempty"`
				Director []Director `json:"director,omitempty"`
			}{
				Actress: []Actress{
					{ID: 1, Name: "女優1"},
				},
				Genre: []Genre{
					{ID: 100, Name: "ジャンル1"},
				},
				Maker: []Maker{
					{ID: 200, Name: "メーカー1"},
				},
				Series: []Series{
					{ID: 300, Name: "シリーズ1"},
				},
				Director: []Director{
					{ID: 400, Name: "監督1"},
				},
			},
		}

		video := ConvertItem(item)

		// 基本情報の検証
		assert.Equal(t, "vid1", video.DmmID)
		assert.Equal(t, "テスト動画", video.Title)
		assert.Equal(t, "https://example.com/sample1.mp4", video.DirectURL)
		assert.Equal(t, "https://example.com/video1", video.URL)
		assert.Equal(t, "https://example.com/sample1.mp4", video.SampleURL)
		assert.Equal(t, "https://example.com/image1.jpg", video.ThumbnailURL)
		assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), video.CreatedAt)
		assert.Equal(t, 1000, video.Price)

		// レビュー情報の検証
		assert.Equal(t, 42, video.Review.Count)
		assert.Equal(t, float32(4.5), video.Review.Average)

		// 関連情報の検証
		assert.Len(t, video.Actresses, 1)
		assert.Equal(t, "1", video.Actresses[0].ID)
		assert.Equal(t, "女優1", video.Actresses[0].Name)

		assert.Len(t, video.Genres, 1)
		assert.Equal(t, "100", video.Genres[0].ID)
		assert.Equal(t, "ジャンル1", video.Genres[0].Name)

		assert.Len(t, video.Makers, 1)
		assert.Equal(t, "200", video.Makers[0].ID)
		assert.Equal(t, "メーカー1", video.Makers[0].Name)

		assert.Len(t, video.Series, 1)
		assert.Equal(t, "300", video.Series[0].ID)
		assert.Equal(t, "シリーズ1", video.Series[0].Name)

		assert.Len(t, video.Directors, 1)
		assert.Equal(t, "400", video.Directors[0].ID)
		assert.Equal(t, "監督1", video.Directors[0].Name)
	})

	// テストケース2: レビュー情報なし
	t.Run("レビュー情報がない場合", func(t *testing.T) {
		item := Item{
			ContentID: "vid2",
			Title:     "テスト動画2",
			Date:      "2024-01-01 00:00:00",
			Review:    nil,
		}

		video := ConvertItem(item)

		// レビュー情報の検証（デフォルト値）
		assert.Equal(t, 0, video.Review.Count)
		assert.Equal(t, float32(0), video.Review.Average)
	})

	// テストケース3: 数値変換エラー
	t.Run("レビュー評価の数値変換エラー", func(t *testing.T) {
		item := Item{
			ContentID: "vid3",
			Title:     "テスト動画3",
			Date:      "2024-01-01 00:00:00",
			Review: &struct {
				Count   int    `json:"count"`
				Average string `json:"average"`
			}{
				Count:   30,
				Average: "invalid", // 不正な数値
			},
		}

		video := ConvertItem(item)

		// レビュー情報の検証（Averageのデフォルト値）
		assert.Equal(t, 30, video.Review.Count)
		assert.Equal(t, float32(0), video.Review.Average) // 変換エラーなので0
	})
} 