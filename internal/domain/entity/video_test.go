package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVideoEntity(t *testing.T) {
	createdAt := time.Now()
	video := Video{
		DmmID:        "test123",
		Title:        "Test Video",
		DirectURL:    "https://example.com/direct",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    createdAt,
		Price:        1000,
		LikesCount:   500,
		Actresses: []Actress{
			{ID: "1", Name: "Test Actress"},
		},
		Genres: []Genre{
			{ID: "1", Name: "Test Genre"},
		},
		Makers: []Maker{
			{ID: "1", Name: "Test Maker"},
		},
		Series: []Series{
			{ID: "1", Name: "Test Series"},
		},
		Directors: []Director{
			{ID: "1", Name: "Test Director"},
		},
	}

	assert.Equal(t, "test123", video.DmmID)
	assert.Equal(t, "Test Video", video.Title)
	assert.Equal(t, "https://example.com/direct", video.DirectURL)
	assert.Equal(t, "https://example.com", video.URL)
	assert.Equal(t, "https://example.com/sample", video.SampleURL)
	assert.Equal(t, "https://example.com/thumb.jpg", video.ThumbnailURL)
	assert.Equal(t, createdAt, video.CreatedAt)
	assert.Equal(t, 1000, video.Price)
	assert.Equal(t, 500, video.LikesCount)

	assert.Len(t, video.Actresses, 1)
	assert.Equal(t, "1", video.Actresses[0].ID)
	assert.Equal(t, "Test Actress", video.Actresses[0].Name)

	assert.Len(t, video.Genres, 1)
	assert.Equal(t, "1", video.Genres[0].ID)
	assert.Equal(t, "Test Genre", video.Genres[0].Name)

	assert.Len(t, video.Makers, 1)
	assert.Equal(t, "1", video.Makers[0].ID)
	assert.Equal(t, "Test Maker", video.Makers[0].Name)

	assert.Len(t, video.Series, 1)
	assert.Equal(t, "1", video.Series[0].ID)
	assert.Equal(t, "Test Series", video.Series[0].Name)

	assert.Len(t, video.Directors, 1)
	assert.Equal(t, "1", video.Directors[0].ID)
	assert.Equal(t, "Test Director", video.Directors[0].Name)
}

func TestActressEntity(t *testing.T) {
	actress := Actress{
		ID:   "123",
		Name: "Test Name",
	}

	assert.Equal(t, "123", actress.ID)
	assert.Equal(t, "Test Name", actress.Name)
}

func TestGenreEntity(t *testing.T) {
	genre := Genre{
		ID:   "456",
		Name: "Test Genre",
	}

	assert.Equal(t, "456", genre.ID)
	assert.Equal(t, "Test Genre", genre.Name)
}

func TestMakerEntity(t *testing.T) {
	maker := Maker{
		ID:   "789",
		Name: "Test Maker",
	}

	assert.Equal(t, "789", maker.ID)
	assert.Equal(t, "Test Maker", maker.Name)
}

func TestSeriesEntity(t *testing.T) {
	series := Series{
		ID:   "012",
		Name: "Test Series",
	}

	assert.Equal(t, "012", series.ID)
	assert.Equal(t, "Test Series", series.Name)
}

func TestDirectorEntity(t *testing.T) {
	director := Director{
		ID:   "345",
		Name: "Test Director",
	}

	assert.Equal(t, "345", director.ID)
	assert.Equal(t, "Test Director", director.Name)
}
