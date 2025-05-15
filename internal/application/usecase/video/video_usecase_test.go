package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tikfack/server/internal/domain/entity"
)

type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error) {
	args := m.Called(ctx, targetDate)
	return args.Get(0).([]entity.Video), args.Error(1)
}

func (m *MockVideoRepository) GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error) {
	args := m.Called(ctx, dmmId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Video), args.Error(1)
}

func (m *MockVideoRepository) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error) {
	args := m.Called(ctx, keyword, actressID, genreID, makerID, seriesID, directorID)
	return args.Get(0).([]entity.Video), args.Error(1)
}

func (m *MockVideoRepository) GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	args := m.Called(ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, hits, offset, sort, gteDate, lteDate, site, service, floor)
	return args.Get(0).([]entity.Video), args.Error(1)
}

func (m *MockVideoRepository) GetVideosByKeyword(ctx context.Context, keyword string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	args := m.Called(ctx, keyword, hits, offset, sort, gteDate, lteDate, site, service, floor)
	return args.Get(0).([]entity.Video), args.Error(1)
}

func TestNewVideoUsecase(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	usecase := NewVideoUsecase(mockRepo)
	assert.NotNil(t, usecase)
}

func TestGetVideosByDate(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	usecase := NewVideoUsecase(mockRepo)
	ctx := context.Background()
	
	targetDate := time.Now()
	expectedVideos := []entity.Video{
		{
			DmmID: "test123",
			Title: "Test Video",
		},
	}
	
	mockRepo.On("GetVideosByDate", ctx, targetDate).Return(expectedVideos, nil)
	videos, err := usecase.GetVideosByDate(ctx, targetDate)
	assert.NoError(t, err)
	assert.Equal(t, expectedVideos, videos)
	
	expectedError := errors.New("repository error")
	mockRepo.On("GetVideosByDate", ctx, time.Time{}).Return([]entity.Video{}, expectedError)
	videos, err = usecase.GetVideosByDate(ctx, time.Time{})
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, videos)
	
	mockRepo.AssertExpectations(t)
}

func TestGetVideoById(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	usecase := NewVideoUsecase(mockRepo)
	ctx := context.Background()
	
	expectedVideo := &entity.Video{
		DmmID: "test123",
		Title: "Test Video",
	}
	
	mockRepo.On("GetVideoById", ctx, "test123").Return(expectedVideo, nil)
	video, err := usecase.GetVideoById(ctx, "test123")
	assert.NoError(t, err)
	assert.Equal(t, expectedVideo, video)
	
	expectedError := errors.New("repository error")
	mockRepo.On("GetVideoById", ctx, "notfound").Return(nil, expectedError)
	video, err = usecase.GetVideoById(ctx, "notfound")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, video)
	
	mockRepo.AssertExpectations(t)
}

func TestSearchVideos(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	usecase := NewVideoUsecase(mockRepo)
	ctx := context.Background()
	
	expectedVideos := []entity.Video{
		{
			DmmID: "test123",
			Title: "Test Video",
		},
	}
	
	mockRepo.On("SearchVideos", ctx, "keyword", "1", "2", "3", "4", "5").
		Return(expectedVideos, nil)
	videos, err := usecase.SearchVideos(ctx, "keyword", "1", "2", "3", "4", "5")
	assert.NoError(t, err)
	assert.Equal(t, expectedVideos, videos)
	
	expectedError := errors.New("repository error")
	mockRepo.On("SearchVideos", ctx, "", "", "", "", "", "").
		Return([]entity.Video{}, expectedError)
	videos, err = usecase.SearchVideos(ctx, "", "", "", "", "", "")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, videos)
	
	mockRepo.AssertExpectations(t)
}

func TestGetVideosByID(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	usecase := NewVideoUsecase(mockRepo)
	ctx := context.Background()
	
	expectedVideos := []entity.Video{
		{
			DmmID: "test123",
			Title: "Test Video",
		},
	}
	
	actressIDs := []string{"1"}
	genreIDs := []string{"2"}
	makerIDs := []string{"3"}
	seriesIDs := []string{"4"}
	directorIDs := []string{"5"}
	hits := int32(10)
	offset := int32(0)
	sort := "rank"
	gteDate := "2023-01-01"
	lteDate := "2023-12-31"
	site := "FANZA"
	service := "digital"
	floor := "videoa"
	
	mockRepo.On("GetVideosByID", ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, 
		hits, offset, sort, gteDate, lteDate, site, service, floor).
		Return(expectedVideos, nil)
	
	videos, err := usecase.GetVideosByID(ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, 
		hits, offset, sort, gteDate, lteDate, site, service, floor)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedVideos, videos)
	
	expectedError := errors.New("repository error")
	mockRepo.On("GetVideosByID", ctx, []string{}, []string{}, []string{}, []string{}, []string{}, 
		int32(0), int32(0), "", "", "", "", "", "").
		Return([]entity.Video{}, expectedError)
	
	videos, err = usecase.GetVideosByID(ctx, []string{}, []string{}, []string{}, []string{}, []string{}, 
		int32(0), int32(0), "", "", "", "", "", "")
	
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, videos)
	
	mockRepo.AssertExpectations(t)
}

func TestGetVideosByKeyword(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	usecase := NewVideoUsecase(mockRepo)
	ctx := context.Background()
	
	expectedVideos := []entity.Video{
		{
			DmmID: "test123",
			Title: "Test Video",
		},
	}
	
	keyword := "test"
	hits := int32(10)
	offset := int32(0)
	sort := "rank"
	gteDate := "2023-01-01"
	lteDate := "2023-12-31"
	site := "FANZA"
	service := "digital"
	floor := "videoa"
	
	mockRepo.On("GetVideosByKeyword", ctx, keyword, hits, offset, sort, gteDate, lteDate, site, service, floor).
		Return(expectedVideos, nil)
	
	videos, err := usecase.GetVideosByKeyword(ctx, keyword, hits, offset, sort, gteDate, lteDate, site, service, floor)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedVideos, videos)
	
	expectedError := errors.New("repository error")
	mockRepo.On("GetVideosByKeyword", ctx, "", int32(0), int32(0), "", "", "", "", "", "").
		Return([]entity.Video{}, expectedError)
	
	videos, err = usecase.GetVideosByKeyword(ctx, "", int32(0), int32(0), "", "", "", "", "", "")
	
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, videos)
	
	mockRepo.AssertExpectations(t)
}
