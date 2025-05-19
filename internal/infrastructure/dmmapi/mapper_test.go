package dmmapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tikfack/server/internal/domain/entity"
)
func TestConvertEntityFromDMM(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name             string
		input            Result
		expectedVideos   []entity.Video
		expectedMetadata *entity.SearchMetadata
	}{
		{
			name: "正常系: 全てのフィールドが存在する場合",
			input: Result{
				Status:        200,
				ResultCount:   1,
				TotalCount:    1,
				FirstPosition: 1,
				Items: []Item{
					{
						ContentID: "test001",
						Title:     "テスト動画",
						Date:      "2024-01-01 00:00:00",
						URL:       "https://example.com/video",
						ImageURL: struct {
							Large string `json:"large"`
						}{
							Large: "https://example.com/image.jpg",
						},
						SampleMovieURL: &struct {
							Size720480 string `json:"size_720_480"`
						}{
							Size720480: "https://example.com/sample.mp4",
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
							Price: "1980円",
						},
						Review: &struct {
							Count   int    `json:"count"`
							Average string `json:"average"`
						}{
							Count:   10,
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
								{ID: 1, Name: "テスト女優1"},
								{ID: 2, Name: "テスト女優2"},
							},
							Genre: []Genre{
								{ID: 1, Name: "テストジャンル1"},
								{ID: 2, Name: "テストジャンル2"},
							},
							Maker: []Maker{
								{ID: 1, Name: "テストメーカー1"},
							},
							Series: []Series{
								{ID: 1, Name: "テストシリーズ1"},
							},
							Director: []Director{
								{ID: 1, Name: "テスト監督1"},
							},
						},
					},
				},
			},
			expectedVideos: []entity.Video{
				{
					DmmID:        "test001",
					Title:        "テスト動画",
					DirectURL:    "https://example.com/sample.mp4",
					URL:          "https://example.com/video",
					SampleURL:    "https://example.com/sample.mp4",
					ThumbnailURL: "https://example.com/image.jpg",
					CreatedAt:    createdAt,
					Price:        1980,
					LikesCount:   0,
					Actresses: []entity.Actress{
						{ID: "1", Name: "テスト女優1"},
						{ID: "2", Name: "テスト女優2"},
					},
					Genres: []entity.Genre{
						{ID: "1", Name: "テストジャンル1"},
						{ID: "2", Name: "テストジャンル2"},
					},
					Makers: []entity.Maker{
						{ID: "1", Name: "テストメーカー1"},
					},
					Series: []entity.Series{
						{ID: "1", Name: "テストシリーズ1"},
					},
					Directors: []entity.Director{
						{ID: "1", Name: "テスト監督1"},
					},
					Review: entity.Review{
						Count:   10,
						Average: 4.5,
					},
				},
			},
			expectedMetadata: &entity.SearchMetadata{
				ResultCount:   1,
				TotalCount:    1,
				FirstPosition: 1,
			},
		},
		{
			name: "正常系: オプショナルフィールドが存在しない場合",
			input: Result{
				Status:        200,
				ResultCount:   1,
				TotalCount:    1,
				FirstPosition: 1,
				Items: []Item{
					{
						ContentID: "test002",
						Title:     "テスト動画2",
						Date:      "2024-01-01 00:00:00",
						URL:       "https://example.com/video2",
						ImageURL: struct {
							Large string `json:"large"`
						}{
							Large: "https://example.com/image2.jpg",
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
							Price: "2980円",
						},
						ItemInfo: struct {
							Actress  []Actress  `json:"actress,omitempty"`
							Genre    []Genre    `json:"genre,omitempty"`
							Maker    []Maker    `json:"maker,omitempty"`
							Series   []Series   `json:"series,omitempty"`
							Director []Director `json:"director,omitempty"`
						}{},
					},
				},
			},
			expectedVideos: []entity.Video{
				{
					DmmID:        "test002",
					Title:        "テスト動画2",
					URL:          "https://example.com/video2",
					ThumbnailURL: "https://example.com/image2.jpg",
					CreatedAt:    createdAt,
					Price:        2980,
					LikesCount:   0,
					Actresses:    []entity.Actress{},
					Genres:       []entity.Genre{},
					Makers:       []entity.Maker{},
					Series:       []entity.Series{},
					Directors:    []entity.Director{},
					Review: entity.Review{
						Count:   0,
						Average: 0,
					},
				},
			},
			expectedMetadata: &entity.SearchMetadata{
				ResultCount:   1,
				TotalCount:    1,
				FirstPosition: 1,
			},
		},
		{
			name: "正常系: 空の結果を返す場合",
			input: Result{
				Status:        200,
				ResultCount:   0,
				TotalCount:    0,
				FirstPosition: 0,
				Items:         []Item{},
			},
			expectedVideos: []entity.Video{},
			expectedMetadata: &entity.SearchMetadata{
				ResultCount:   0,
				TotalCount:    0,
				FirstPosition: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			videos, metadata := ConvertEntityFromDMM(tt.input)
			
			require.Equal(t, tt.expectedVideos, videos)
			require.Equal(t, tt.expectedMetadata, metadata)
			
			// 日付の比較は別途行う（タイムゾーンの問題を避けるため）
			if len(videos) > 0 {
				for i, video := range videos {
					assert.True(t, video.CreatedAt.Equal(tt.expectedVideos[i].CreatedAt))
				}
			}
		})
	}
}

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name     string
		input    Item
		expected int
	}{
		{
			name: "通常の価格",
			input: Item{
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
					Price: "1980円",
				},
			},
			expected: 1980,
		},
		{
			name: "価格範囲の場合",
			input: Item{
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
					Price: "1980円~2980円",
				},
			},
			expected: 1980,
		},
		{
			name: "価格が空の場合",
			input: Item{
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
					Price: "",
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePrice(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "正常な日付",
			input:    "2024-01-01 00:00:00",
			expected: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "空の日付",
			input:    "",
			expected: time.Time{},
		},
		{
			name:     "不正な日付",
			input:    "invalid-date",
			expected: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDate(tt.input)
			if tt.input == "" || tt.input == "invalid-date" {
				assert.True(t, result.IsZero())
			} else {
				assert.True(t, result.Equal(tt.expected))
			}
		})
	}
} 