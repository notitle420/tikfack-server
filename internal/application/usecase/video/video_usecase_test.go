package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	require.NotNil(t, usecase)
}

func TestGetVideosByDate(t *testing.T) {
	ctx := context.Background()
	fixed := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	
	cases := []struct {
		name     string
		date     time.Time
		mockRet  []entity.Video
		mockErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name:    "正常系",
			date:    fixed,
			mockRet: []entity.Video{{DmmID: "test123", Title: "Test Video"}},
			mockErr: nil,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "異常系",
			date:    time.Time{},
			mockRet: []entity.Video{},
			mockErr: errors.New("repository error"),
			wantLen: 0,
			wantErr: true,
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			mockRepo := new(MockVideoRepository)
			usecase := NewVideoUsecase(mockRepo)
			
			mockRepo.
				On("GetVideosByDate", mock.Anything, c.date).
				Return(c.mockRet, c.mockErr)
			
			videos, err := usecase.GetVideosByDate(ctx, c.date)
			
			if c.wantErr {
				require.Error(tt, err)
				assert.Empty(tt, videos)
			} else {
				require.NoError(tt, err)
				assert.Equal(tt, c.wantLen, len(videos))
				assert.Equal(tt, c.mockRet, videos)
			}
			
			mockRepo.AssertExpectations(tt)
		})
	}
}

func TestGetVideoById(t *testing.T) {
	ctx := context.Background()
	
	expectedVideo := &entity.Video{
		DmmID: "test123",
		Title: "Test Video",
	}
	
	cases := []struct {
		name     string
		id       string
		mockRet  *entity.Video
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "正常系",
			id:      "test123",
			mockRet: expectedVideo,
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "異常系",
			id:      "notfound",
			mockRet: nil,
			mockErr: errors.New("repository error"),
			wantErr: true,
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			mockRepo := new(MockVideoRepository)
			usecase := NewVideoUsecase(mockRepo)
			
			mockRepo.
				On("GetVideoById", mock.Anything, c.id).
				Return(c.mockRet, c.mockErr)
			
			video, err := usecase.GetVideoById(ctx, c.id)
			
			if c.wantErr {
				require.Error(tt, err)
				assert.Nil(tt, video)
			} else {
				require.NoError(tt, err)
				require.NotNil(tt, video)
				assert.Equal(tt, c.mockRet, video)
			}
			
			mockRepo.AssertExpectations(tt)
		})
	}
}

func TestSearchVideos(t *testing.T) {
	ctx := context.Background()
	
	cases := []struct {
		name       string
		keyword    string
		actressID  string
		genreID    string
		makerID    string
		seriesID   string
		directorID string
		mockRet    []entity.Video
		mockErr    error
		wantLen    int
		wantErr    bool
	}{
		{
			name:       "正常系",
			keyword:    "keyword",
			actressID:  "1",
			genreID:    "2",
			makerID:    "3",
			seriesID:   "4",
			directorID: "5",
			mockRet:    []entity.Video{{DmmID: "test123", Title: "Test Video"}},
			mockErr:    nil,
			wantLen:    1,
			wantErr:    false,
		},
		{
			name:       "異常系",
			keyword:    "",
			actressID:  "",
			genreID:    "",
			makerID:    "",
			seriesID:   "",
			directorID: "",
			mockRet:    []entity.Video{},
			mockErr:    errors.New("repository error"),
			wantLen:    0,
			wantErr:    true,
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			mockRepo := new(MockVideoRepository)
			usecase := NewVideoUsecase(mockRepo)
			
			mockRepo.
				On("SearchVideos", mock.Anything, c.keyword, c.actressID, c.genreID, c.makerID, c.seriesID, c.directorID).
				Return(c.mockRet, c.mockErr)
			
			videos, err := usecase.SearchVideos(ctx, c.keyword, c.actressID, c.genreID, c.makerID, c.seriesID, c.directorID)
			
			if c.wantErr {
				require.Error(tt, err)
				assert.Empty(tt, videos)
			} else {
				require.NoError(tt, err)
				assert.Equal(tt, c.wantLen, len(videos))
				assert.Equal(tt, c.mockRet, videos)
			}
			
			mockRepo.AssertExpectations(tt)
		})
	}
}

func TestGetVideosByID(t *testing.T) {
	ctx := context.Background()
	
	cases := []struct {
		name        string
		actressIDs  []string
		genreIDs    []string
		makerIDs    []string
		seriesIDs   []string
		directorIDs []string
		hits        int32
		offset      int32
		sort        string
		gteDate     string
		lteDate     string
		site        string
		service     string
		floor       string
		mockRet     []entity.Video
		mockErr     error
		wantLen     int
		wantErr     bool
	}{
		{
			name:        "正常系",
			actressIDs:  []string{"1"},
			genreIDs:    []string{"2"},
			makerIDs:    []string{"3"},
			seriesIDs:   []string{"4"},
			directorIDs: []string{"5"},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "2023-01-01",
			lteDate:     "2023-12-31",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			mockRet:     []entity.Video{{DmmID: "test123", Title: "Test Video"}},
			mockErr:     nil,
			wantLen:     1,
			wantErr:     false,
		},
		{
			name:        "異常系",
			actressIDs:  []string{},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        0,
			offset:      0,
			sort:        "",
			gteDate:     "",
			lteDate:     "",
			site:        "",
			service:     "",
			floor:       "",
			mockRet:     []entity.Video{},
			mockErr:     errors.New("repository error"),
			wantLen:     0,
			wantErr:     true,
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			mockRepo := new(MockVideoRepository)
			usecase := NewVideoUsecase(mockRepo)
			
			mockRepo.
				On("GetVideosByID", mock.Anything, c.actressIDs, c.genreIDs, c.makerIDs, c.seriesIDs, c.directorIDs,
					c.hits, c.offset, c.sort, c.gteDate, c.lteDate, c.site, c.service, c.floor).
				Return(c.mockRet, c.mockErr)
			
			videos, err := usecase.GetVideosByID(ctx, c.actressIDs, c.genreIDs, c.makerIDs, c.seriesIDs, c.directorIDs,
				c.hits, c.offset, c.sort, c.gteDate, c.lteDate, c.site, c.service, c.floor)
			
			if c.wantErr {
				require.Error(tt, err)
				assert.Empty(tt, videos)
			} else {
				require.NoError(tt, err)
				assert.Equal(tt, c.wantLen, len(videos))
				assert.Equal(tt, c.mockRet, videos)
			}
			
			mockRepo.AssertExpectations(tt)
		})
	}
}

func TestGetVideosByKeyword(t *testing.T) {
	ctx := context.Background()
	
	cases := []struct {
		name     string
		keyword  string
		hits     int32
		offset   int32
		sort     string
		gteDate  string
		lteDate  string
		site     string
		service  string
		floor    string
		mockRet  []entity.Video
		mockErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "正常系",
			keyword:  "test",
			hits:     10,
			offset:   0,
			sort:     "rank",
			gteDate:  "2023-01-01",
			lteDate:  "2023-12-31",
			site:     "FANZA",
			service:  "digital",
			floor:    "videoa",
			mockRet:  []entity.Video{{DmmID: "test123", Title: "Test Video"}},
			mockErr:  nil,
			wantLen:  1,
			wantErr:  false,
		},
		{
			name:     "異常系",
			keyword:  "",
			hits:     0,
			offset:   0,
			sort:     "",
			gteDate:  "",
			lteDate:  "",
			site:     "",
			service:  "",
			floor:    "",
			mockRet:  []entity.Video{},
			mockErr:  errors.New("repository error"),
			wantLen:  0,
			wantErr:  true,
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			mockRepo := new(MockVideoRepository)
			usecase := NewVideoUsecase(mockRepo)
			
			mockRepo.
				On("GetVideosByKeyword", mock.Anything, c.keyword, c.hits, c.offset, c.sort,
					c.gteDate, c.lteDate, c.site, c.service, c.floor).
				Return(c.mockRet, c.mockErr)
			
			videos, err := usecase.GetVideosByKeyword(ctx, c.keyword, c.hits, c.offset, c.sort,
				c.gteDate, c.lteDate, c.site, c.service, c.floor)
			
			if c.wantErr {
				require.Error(tt, err)
				assert.Empty(tt, videos)
			} else {
				require.NoError(tt, err)
				assert.Equal(tt, c.wantLen, len(videos))
				assert.Equal(tt, c.mockRet, videos)
			}
			
			mockRepo.AssertExpectations(tt)
		})
	}
}
