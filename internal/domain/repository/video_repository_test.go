package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tikfack/server/internal/domain/entity"
)

type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) GetVideosByDate(ctx context.Context, targetDate time.Time, hits int32, offset int32) ([]entity.Video, *entity.SearchMetadata, error) {
	args := m.Called(ctx, targetDate, hits, offset)
	return args.Get(0).([]entity.Video), args.Get(1).(*entity.SearchMetadata), args.Error(2)
}

func (m *MockVideoRepository) GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error) {
	args := m.Called(ctx, dmmId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Video), args.Error(1)
}

func (m *MockVideoRepository) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, *entity.SearchMetadata, error) {
	args := m.Called(ctx, keyword, actressID, genreID, makerID, seriesID, directorID)
	return args.Get(0).([]entity.Video), args.Get(1).(*entity.SearchMetadata), args.Error(2)
}

func (m *MockVideoRepository) GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, *entity.SearchMetadata, error) {
	args := m.Called(ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, hits, offset, sort, gteDate, lteDate, site, service, floor)
	return args.Get(0).([]entity.Video), args.Get(1).(*entity.SearchMetadata), args.Error(2)
}

func (m *MockVideoRepository) GetVideosByKeyword(ctx context.Context, keyword string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, *entity.SearchMetadata, error) {
	args := m.Called(ctx, keyword, hits, offset, sort, gteDate, lteDate, site, service, floor)
	return args.Get(0).([]entity.Video), args.Get(1).(*entity.SearchMetadata), args.Error(2)
}

func TestMockVideoRepository(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	ctx := context.Background()
	
	expectedVideos := []entity.Video{
		{
			DmmID:  "test123",
			Title:  "Test Video",
			Price:  1000,
		},
	}
	
	expectedVideo := entity.Video{
		DmmID:  "test123",
		Title:  "Test Video",
		Price:  1000,
	}

	expectedMD := entity.SearchMetadata{
		ResultCount:   10,
		TotalCount:    100,
		FirstPosition: 1,
	}

	targetDate := time.Now()
	mockRepo.On("GetVideosByDate", ctx, targetDate, int32(10), int32(0)).Return(expectedVideos, &expectedMD, nil)
	videos, metadata, err := mockRepo.GetVideosByDate(ctx, targetDate, int32(10), int32(0))
	assert.NoError(t, err)
	assert.Equal(t, expectedVideos, videos)
	assert.Equal(t, &expectedMD, metadata)
	mockRepo.AssertExpectations(t)

	mockRepo.On("GetVideoById", ctx, "test123").Return(&expectedVideo, nil)
	video, err := mockRepo.GetVideoById(ctx, "test123")
	assert.NoError(t, err)
	assert.Equal(t, &expectedVideo, video)
	mockRepo.AssertExpectations(t)

	mockRepo.On("SearchVideos", ctx, "keyword", "1", "2", "3", "4", "5").Return(expectedVideos, &expectedMD, nil)
	videos, metadata, err = mockRepo.SearchVideos(ctx, "keyword", "1", "2", "3", "4", "5")
	assert.NoError(t, err)
	assert.Equal(t, expectedVideos, videos)
	assert.Equal(t, &expectedMD, metadata)
	mockRepo.AssertExpectations(t)

	mockRepo.On("GetVideosByID", ctx, 
		[]string{"1"}, []string{"2"}, []string{"3"}, []string{"4"}, []string{"5"}, 
		int32(10), int32(0), "rank", "2023-01-01", "2023-12-31", "FANZA", "digital", "videoa").
		Return(expectedVideos, &expectedMD, nil)
	videos, metadata, err = mockRepo.GetVideosByID(ctx, 
		[]string{"1"}, []string{"2"}, []string{"3"}, []string{"4"}, []string{"5"}, 
		int32(10), int32(0), "rank", "2023-01-01", "2023-12-31", "FANZA", "digital", "videoa")
	assert.NoError(t, err)
	assert.Equal(t, expectedVideos, videos)
	assert.Equal(t, &expectedMD, metadata)
	mockRepo.AssertExpectations(t)

	mockRepo.On("GetVideosByKeyword", ctx, "keyword", int32(10), int32(0), "rank", 
		"2023-01-01", "2023-12-31", "FANZA", "digital", "videoa").
		Return(expectedVideos, &expectedMD, nil)
	videos, metadata, err = mockRepo.GetVideosByKeyword(ctx, "keyword", int32(10), int32(0), "rank", 
		"2023-01-01", "2023-12-31", "FANZA", "digital", "videoa")
	assert.NoError(t, err)
	assert.Equal(t, expectedVideos, videos)
	assert.Equal(t, &expectedMD, metadata)
	mockRepo.AssertExpectations(t)
}
