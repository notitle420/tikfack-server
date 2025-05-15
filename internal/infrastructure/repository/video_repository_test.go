package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDMMVideoRepository(t *testing.T) {
	repo := NewDMMVideoRepository()
	assert.NotNil(t, repo)
}

func TestGetVideosByDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "ItemList")
		assert.Contains(t, r.URL.String(), "api_id=")
		assert.Contains(t, r.URL.String(), "affiliate_id=")
		assert.Contains(t, r.URL.String(), "site=FANZA")
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		response := map[string]interface{}{
			"result": map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"content_id": "test123",
						"title":     "Test Video",
						"date":      "2023-12-31T12:34:56",
						"URL":       "https://example.com",
						"imageURL": map[string]interface{}{
							"large": "https://example.com/thumb.jpg",
						},
						"sampleMovieURL": map[string]interface{}{
							"size_720_480": "https://example.com/sample.mp4",
						},
						"prices": map[string]interface{}{
							"price": "1000円",
						},
						"iteminfo": map[string]interface{}{
							"actress": []map[string]interface{}{
								{"id": 1, "name": "Test Actress"},
							},
							"genre": []map[string]interface{}{
								{"id": 1, "name": "Test Genre"},
							},
							"maker": []map[string]interface{}{
								{"id": 1, "name": "Test Maker"},
							},
							"series": []map[string]interface{}{
								{"id": 1, "name": "Test Series"},
							},
							"director": []map[string]interface{}{
								{"id": 1, "name": "Test Director"},
							},
						},
					},
				},
			},
		}
		
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	oldBaseURL := os.Getenv("BASE_URL")
	oldApiID := os.Getenv("DMM_API_ID")
	oldAffiliateID := os.Getenv("DMM_API_AFFILIATE_ID")
	oldHits := os.Getenv("HITS")
	
	os.Setenv("BASE_URL", server.URL)
	os.Setenv("DMM_API_ID", "test_api_id")
	os.Setenv("DMM_API_AFFILIATE_ID", "test_affiliate_id")
	os.Setenv("HITS", "10")
	
	defer func() {
		os.Setenv("BASE_URL", oldBaseURL)
		os.Setenv("DMM_API_ID", oldApiID)
		os.Setenv("DMM_API_AFFILIATE_ID", oldAffiliateID)
		os.Setenv("HITS", oldHits)
	}()
	
	repo := NewDMMVideoRepository()
	ctx := context.Background()
	targetDate := time.Now()
	
	videos, err := repo.GetVideosByDate(ctx, targetDate)
	
	require.NoError(t, err)
	require.NotEmpty(t, videos)
	assert.Equal(t, "test123", videos[0].DmmID)
	assert.Equal(t, "Test Video", videos[0].Title)
	assert.Equal(t, "https://example.com", videos[0].URL)
	assert.Equal(t, "https://example.com/thumb.jpg", videos[0].ThumbnailURL)
	assert.Equal(t, "https://example.com/sample.mp4", videos[0].SampleURL)
	assert.Equal(t, 1000, videos[0].Price)
	
	require.Len(t, videos[0].Actresses, 1)
	assert.Equal(t, "1", videos[0].Actresses[0].ID)
	assert.Equal(t, "Test Actress", videos[0].Actresses[0].Name)
	
	require.Len(t, videos[0].Genres, 1)
	assert.Equal(t, "1", videos[0].Genres[0].ID)
	assert.Equal(t, "Test Genre", videos[0].Genres[0].Name)
}

func TestGetVideoById(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		response := map[string]interface{}{
			"result": map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"content_id": "test123",
						"title":     "Test Video",
						"date":      "2023-12-31T12:34:56",
						"URL":       "https://example.com",
						"imageURL": map[string]interface{}{
							"large": "https://example.com/thumb.jpg",
						},
						"prices": map[string]interface{}{
							"price": "1000円",
						},
						"iteminfo": map[string]interface{}{},
					},
				},
			},
		}
		
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	oldBaseURL := os.Getenv("BASE_URL")
	oldApiID := os.Getenv("DMM_API_ID")
	oldAffiliateID := os.Getenv("DMM_API_AFFILIATE_ID")
	oldHits := os.Getenv("HITS")
	
	os.Setenv("BASE_URL", server.URL)
	os.Setenv("DMM_API_ID", "test_api_id")
	os.Setenv("DMM_API_AFFILIATE_ID", "test_affiliate_id")
	os.Setenv("HITS", "10")
	
	defer func() {
		os.Setenv("BASE_URL", oldBaseURL)
		os.Setenv("DMM_API_ID", oldApiID)
		os.Setenv("DMM_API_AFFILIATE_ID", oldAffiliateID)
		os.Setenv("HITS", oldHits)
	}()
	
	repo := NewDMMVideoRepository()
	ctx := context.Background()
	
	video, err := repo.GetVideoById(ctx, "test123")
	
	require.NoError(t, err)
	require.NotNil(t, video)
	assert.Equal(t, "test123", video.DmmID)
	assert.Equal(t, "Test Video", video.Title)
}

func TestSearchVideos(t *testing.T) {
	repo := NewDMMVideoRepository()
	ctx := context.Background()
	
	videos, err := repo.SearchVideos(ctx, "keyword", "1", "2", "3", "4", "5")
	
	assert.Nil(t, err)
	assert.Nil(t, videos)
}

func TestGetVideosByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "ItemList")
		assert.Contains(t, r.URL.String(), "api_id=")
		assert.Contains(t, r.URL.String(), "affiliate_id=")
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		response := map[string]interface{}{
			"result": map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"content_id": "test123",
						"title":     "Test Video",
						"date":      "2023-12-31T12:34:56",
						"URL":       "https://example.com",
						"imageURL": map[string]interface{}{
							"large": "https://example.com/thumb.jpg",
						},
						"prices": map[string]interface{}{
							"price": "1000円",
						},
						"iteminfo": map[string]interface{}{},
					},
				},
			},
		}
		
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	oldBaseURL := os.Getenv("BASE_URL")
	oldApiID := os.Getenv("DMM_API_ID")
	oldAffiliateID := os.Getenv("DMM_API_AFFILIATE_ID")
	oldHits := os.Getenv("HITS")
	
	os.Setenv("BASE_URL", server.URL)
	os.Setenv("DMM_API_ID", "test_api_id")
	os.Setenv("DMM_API_AFFILIATE_ID", "test_affiliate_id")
	os.Setenv("HITS", "10")
	
	defer func() {
		os.Setenv("BASE_URL", oldBaseURL)
		os.Setenv("DMM_API_ID", oldApiID)
		os.Setenv("DMM_API_AFFILIATE_ID", oldAffiliateID)
		os.Setenv("HITS", oldHits)
	}()
	
	repo := NewDMMVideoRepository()
	ctx := context.Background()
	
	actressIDs := []string{"1"}
	genreIDs := []string{"2"}
	makerIDs := []string{"3"}
	seriesIDs := []string{"4"}
	directorIDs := []string{"5"}
	
	videos, err := repo.GetVideosByID(ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, 
		10, 0, "rank", "2023-01-01", "2023-12-31", "FANZA", "digital", "videoa")
	
	require.NoError(t, err)
	require.NotEmpty(t, videos)
	assert.Equal(t, "test123", videos[0].DmmID)
	assert.Equal(t, "Test Video", videos[0].Title)
}

func TestGetVideosByKeyword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "ItemList")
		assert.Contains(t, r.URL.String(), "api_id=")
		assert.Contains(t, r.URL.String(), "affiliate_id=")
		assert.Contains(t, r.URL.String(), "keyword=test")
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		response := map[string]interface{}{
			"result": map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"content_id": "test123",
						"title":     "Test Video",
						"date":      "2023-12-31T12:34:56",
						"URL":       "https://example.com",
						"imageURL": map[string]interface{}{
							"large": "https://example.com/thumb.jpg",
						},
						"prices": map[string]interface{}{
							"price": "1000円",
						},
						"iteminfo": map[string]interface{}{},
					},
				},
			},
		}
		
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	oldBaseURL := os.Getenv("BASE_URL")
	oldApiID := os.Getenv("DMM_API_ID")
	oldAffiliateID := os.Getenv("DMM_API_AFFILIATE_ID")
	oldHits := os.Getenv("HITS")
	
	os.Setenv("BASE_URL", server.URL)
	os.Setenv("DMM_API_ID", "test_api_id")
	os.Setenv("DMM_API_AFFILIATE_ID", "test_affiliate_id")
	os.Setenv("HITS", "10")
	
	defer func() {
		os.Setenv("BASE_URL", oldBaseURL)
		os.Setenv("DMM_API_ID", oldApiID)
		os.Setenv("DMM_API_AFFILIATE_ID", oldAffiliateID)
		os.Setenv("HITS", oldHits)
	}()
	
	repo := NewDMMVideoRepository()
	ctx := context.Background()
	
	videos, err := repo.GetVideosByKeyword(ctx, "test", 10, 0, "rank", 
		"2023-01-01", "2023-12-31", "FANZA", "digital", "videoa")
	
	require.NoError(t, err)
	require.NotEmpty(t, videos)
	assert.Equal(t, "test123", videos[0].DmmID)
	assert.Equal(t, "Test Video", videos[0].Title)
}
