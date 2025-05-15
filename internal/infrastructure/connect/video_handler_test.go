package connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	
	pb "github.com/tikfack/server/gen/video"
	"github.com/tikfack/server/internal/domain/entity"
)

type MockVideoUsecase struct {
	mock.Mock
}

func (m *MockVideoUsecase) GetVideosByDate(ctx context.Context, targetDate time.Time) ([]entity.Video, error) {
	args := m.Called(ctx, targetDate)
	return args.Get(0).([]entity.Video), args.Error(1)
}

func (m *MockVideoUsecase) GetVideoById(ctx context.Context, dmmId string) (*entity.Video, error) {
	args := m.Called(ctx, dmmId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Video), args.Error(1)
}

func (m *MockVideoUsecase) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error) {
	args := m.Called(ctx, keyword, actressID, genreID, makerID, seriesID, directorID)
	return args.Get(0).([]entity.Video), args.Error(1)
}

func (m *MockVideoUsecase) GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	args := m.Called(ctx, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs, hits, offset, sort, gteDate, lteDate, site, service, floor)
	return args.Get(0).([]entity.Video), args.Error(1)
}

func (m *MockVideoUsecase) GetVideosByKeyword(ctx context.Context, keyword string, hits int32, offset int32, sort string, gteDate string, lteDate string, site string, service string, floor string) ([]entity.Video, error) {
	args := m.Called(ctx, keyword, hits, offset, sort, gteDate, lteDate, site, service, floor)
	return args.Get(0).([]entity.Video), args.Error(1)
}

func TestNewVideoServiceHandler(t *testing.T) {
	handler := NewVideoServiceHandler()
	assert.NotNil(t, handler)
}

func TestGetHandler(t *testing.T) {
	handler := NewVideoServiceHandler()
	pattern, httpHandler := handler.GetHandler()
	assert.NotEmpty(t, pattern)
	assert.NotNil(t, httpHandler)
}

func TestGetVideosByDate(t *testing.T) {
	mockUsecase := new(MockVideoUsecase)
	
	handler := &videoServiceServer{
		videoUsecase: mockUsecase,
	}
	
	ctx := context.Background()
	targetDate := time.Now()
	formattedDate := targetDate.Format("2006-01-02")
	
	expectedVideos := []entity.Video{
		{
			DmmID:        "test123",
			Title:        "Test Video",
			URL:          "https://example.com",
			SampleURL:    "https://example.com/sample",
			ThumbnailURL: "https://example.com/thumb.jpg",
			CreatedAt:    targetDate,
			Price:        1000,
			LikesCount:   500,
		},
	}
	
	mockUsecase.On("GetVideosByDate", mock.Anything, mock.Anything).Return(expectedVideos, nil)
	
	req := connect.NewRequest(&pb.GetVideosByDateRequest{
		Date: formattedDate,
	})
	
	resp, err := handler.GetVideosByDate(ctx, req)
	
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Msg)
	require.NotEmpty(t, resp.Msg.Videos)
	assert.Equal(t, "test123", resp.Msg.Videos[0].DmmId)
	assert.Equal(t, "Test Video", resp.Msg.Videos[0].Title)
	
	expectedError := errors.New("usecase error")
	mockUsecase.On("GetVideosByDate", mock.Anything, mock.Anything).Return([]entity.Video{}, expectedError)
	
	req = connect.NewRequest(&pb.GetVideosByDateRequest{
		Date: "invalid-date",
	})
	
	_, err = handler.GetVideosByDate(ctx, req)
	
	require.Error(t, err)
	
	mockUsecase.AssertExpectations(t)
}

func TestGetVideoById(t *testing.T) {
	mockUsecase := new(MockVideoUsecase)
	
	handler := &videoServiceServer{
		videoUsecase: mockUsecase,
	}
	
	ctx := context.Background()
	expectedVideo := &entity.Video{
		DmmID:        "test123",
		Title:        "Test Video",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    time.Now(),
		Price:        1000,
		LikesCount:   500,
	}
	
	mockUsecase.On("GetVideoById", mock.Anything, "test123").Return(expectedVideo, nil)
	
	req := connect.NewRequest(&pb.GetVideoByIdRequest{
		DmmId: "test123",
	})
	
	resp, err := handler.GetVideoById(ctx, req)
	
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Msg)
	require.NotNil(t, resp.Msg.Video)
	assert.Equal(t, "test123", resp.Msg.Video.DmmId)
	assert.Equal(t, "Test Video", resp.Msg.Video.Title)
	
	mockUsecase.On("GetVideoById", mock.Anything, "notfound").Return(nil, errors.New("not found"))
	
	req = connect.NewRequest(&pb.GetVideoByIdRequest{
		DmmId: "notfound",
	})
	
	_, err = handler.GetVideoById(ctx, req)
	
	require.Error(t, err)
	
	mockUsecase.AssertExpectations(t)
}

func TestSearchVideos(t *testing.T) {
	mockUsecase := new(MockVideoUsecase)
	
	handler := &videoServiceServer{
		videoUsecase: mockUsecase,
	}
	
	ctx := context.Background()
	expectedVideos := []entity.Video{
		{
			DmmID:        "test123",
			Title:        "Test Video",
			URL:          "https://example.com",
			SampleURL:    "https://example.com/sample",
			ThumbnailURL: "https://example.com/thumb.jpg",
			CreatedAt:    time.Now(),
			Price:        1000,
			LikesCount:   500,
		},
	}
	
	mockUsecase.On("SearchVideos", mock.Anything, "keyword", "1", "2", "3", "4", "5").
		Return(expectedVideos, nil)
	
	req := connect.NewRequest(&pb.SearchVideosRequest{
		Keyword:    "keyword",
		ActressId:  "1",
		GenreId:    "2",
		MakerId:    "3",
		SeriesId:   "4",
		DirectorId: "5",
	})
	
	resp, err := handler.SearchVideos(ctx, req)
	
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Msg)
	require.NotEmpty(t, resp.Msg.Videos)
	assert.Equal(t, "test123", resp.Msg.Videos[0].DmmId)
	assert.Equal(t, "Test Video", resp.Msg.Videos[0].Title)
	
	mockUsecase.On("SearchVideos", mock.Anything, "", "", "", "", "", "").
		Return([]entity.Video{}, errors.New("search error"))
	
	req = connect.NewRequest(&pb.SearchVideosRequest{})
	
	_, err = handler.SearchVideos(ctx, req)
	
	require.Error(t, err)
	
	mockUsecase.AssertExpectations(t)
}

func TestGetVideosByID(t *testing.T) {
	mockUsecase := new(MockVideoUsecase)
	
	handler := &videoServiceServer{
		videoUsecase: mockUsecase,
	}
	
	ctx := context.Background()
	expectedVideos := []entity.Video{
		{
			DmmID:        "test123",
			Title:        "Test Video",
			URL:          "https://example.com",
			SampleURL:    "https://example.com/sample",
			ThumbnailURL: "https://example.com/thumb.jpg",
			CreatedAt:    time.Now(),
			Price:        1000,
			LikesCount:   500,
		},
	}
	
	mockUsecase.On("GetVideosByID", mock.Anything, 
		[]string{"1"}, []string{"2"}, []string{"3"}, []string{"4"}, []string{"5"}, 
		int32(10), int32(0), "rank", "2023-01-01", "2023-12-31", "FANZA", "digital", "videoa").
		Return(expectedVideos, nil)
	
	req := connect.NewRequest(&pb.GetVideosByIDRequest{
		ActressId:  []string{"1"},
		GenreId:    []string{"2"},
		MakerId:    []string{"3"},
		SeriesId:   []string{"4"},
		DirectorId: []string{"5"},
		Hits:       10,
		Offset:     0,
		Sort:       "rank",
		GteDate:    "2023-01-01",
		LteDate:    "2023-12-31",
		Site:       "FANZA",
		Service:    "digital",
		Floor:      "videoa",
	})
	
	resp, err := handler.GetVideosByID(ctx, req)
	
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Msg)
	require.NotEmpty(t, resp.Msg.Videos)
	assert.Equal(t, "test123", resp.Msg.Videos[0].DmmId)
	assert.Equal(t, "Test Video", resp.Msg.Videos[0].Title)
	
	mockUsecase.On("GetVideosByID", mock.Anything, 
		[]string{}, []string{}, []string{}, []string{}, []string{}, 
		int32(0), int32(0), "", "", "", "", "", "").
		Return([]entity.Video{}, errors.New("search error"))
	
	req = connect.NewRequest(&pb.GetVideosByIDRequest{})
	
	_, err = handler.GetVideosByID(ctx, req)
	
	require.Error(t, err)
	
	mockUsecase.AssertExpectations(t)
}

func TestGetVideosByKeyword(t *testing.T) {
	mockUsecase := new(MockVideoUsecase)
	
	handler := &videoServiceServer{
		videoUsecase: mockUsecase,
	}
	
	ctx := context.Background()
	expectedVideos := []entity.Video{
		{
			DmmID:        "test123",
			Title:        "Test Video",
			URL:          "https://example.com",
			SampleURL:    "https://example.com/sample",
			ThumbnailURL: "https://example.com/thumb.jpg",
			CreatedAt:    time.Now(),
			Price:        1000,
			LikesCount:   500,
		},
	}
	
	mockUsecase.On("GetVideosByKeyword", mock.Anything, 
		"test", int32(10), int32(0), "rank", "2023-01-01", "2023-12-31", "FANZA", "digital", "videoa").
		Return(expectedVideos, nil)
	
	req := connect.NewRequest(&pb.GetVideosByKeywordRequest{
		Keyword:  "test",
		Hits:     10,
		Offset:   0,
		Sort:     "rank",
		GteDate:  "2023-01-01",
		LteDate:  "2023-12-31",
		Site:     "FANZA",
		Service:  "digital",
		Floor:    "videoa",
	})
	
	resp, err := handler.GetVideosByKeyword(ctx, req)
	
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Msg)
	require.NotEmpty(t, resp.Msg.Videos)
	assert.Equal(t, "test123", resp.Msg.Videos[0].DmmId)
	assert.Equal(t, "Test Video", resp.Msg.Videos[0].Title)
	
	mockUsecase.On("GetVideosByKeyword", mock.Anything, 
		"", int32(0), int32(0), "", "", "", "", "", "").
		Return([]entity.Video{}, errors.New("search error"))
	
	req = connect.NewRequest(&pb.GetVideosByKeywordRequest{})
	
	_, err = handler.GetVideosByKeyword(ctx, req)
	
	require.Error(t, err)
	
	mockUsecase.AssertExpectations(t)
}
