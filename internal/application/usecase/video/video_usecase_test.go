package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tikfack/server/internal/domain/entity"
	mockrepo "github.com/tikfack/server/internal/domain/repository/mock"
)

func TestNewVideoUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockRepo := mockrepo.NewMockVideoRepository(ctrl)
	usecase := NewVideoUsecase(mockRepo)
	require.NotNil(t, usecase)
}

func TestGetVideosByDate(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name        string
		date        time.Time
		setupMock   func(mockRepo *mockrepo.MockVideoRepository)
		expected    []entity.Video
		expectError bool
	}{
		{
			name: "正常系",
			date: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)).
					Return([]entity.Video{{DmmID: "test123", Title: "Test Video"}}, nil)
			},
			expected:    []entity.Video{{DmmID: "test123", Title: "Test Video"}},
			expectError: false,
		},
		{
			name: "異常系",
			date: time.Time{},
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), time.Time{}).
					Return([]entity.Video{}, errors.New("repository error"))
			},
			expected:    []entity.Video{},
			expectError: true,
		},
		{
			name: "空配列を返す場合",
			date: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)).
					Return([]entity.Video{}, nil)
			},
			expected:    []entity.Video{},
			expectError: false,
		},
		{
			name: "タイムアウトエラーの場合",
			date: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)).
					Return(nil, context.DeadlineExceeded)
			},
			expected:    nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockRepo := mockrepo.NewMockVideoRepository(ctrl)
			uc := NewVideoUsecase(mockRepo)
			
			tt.setupMock(mockRepo)
			
			videos, err := uc.GetVideosByDate(ctx, tt.date)
			
			if tt.expectError {
				require.Error(t, err)
				if tt.name == "タイムアウトエラーの場合" {
					require.Equal(t, context.DeadlineExceeded, err)
				}
			} else {
				require.NoError(t, err)
			}
			
			require.Equal(t, tt.expected, videos)
		})
	}
}

func TestGetVideoById(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name        string
		id          string
		setupMock   func(mockRepo *mockrepo.MockVideoRepository)
		expected    *entity.Video
		expectError bool
	}{
		{
			name: "正常系",
			id:   "test123",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideoById(gomock.Any(), "test123").
					Return(&entity.Video{DmmID: "test123", Title: "Test Video"}, nil)
			},
			expected:    &entity.Video{DmmID: "test123", Title: "Test Video"},
			expectError: false,
		},
		{
			name: "異常系",
			id:   "notfound",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideoById(gomock.Any(), "notfound").
					Return(nil, errors.New("repository error"))
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "リポジトリがnilとnilを返す場合",
			id:   "empty",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideoById(gomock.Any(), "empty").
					Return(nil, nil)
			},
			expected:    nil,
			expectError: false,
		},
		{
			name: "タイムアウトエラーの場合",
			id:   "timeout",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideoById(gomock.Any(), "timeout").
					Return(nil, context.DeadlineExceeded)
			},
			expected:    nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockRepo := mockrepo.NewMockVideoRepository(ctrl)
			uc := NewVideoUsecase(mockRepo)
			
			tt.setupMock(mockRepo)
			
			video, err := uc.GetVideoById(ctx, tt.id)
			
			if tt.expectError {
				require.Error(t, err)
				if tt.name == "タイムアウトエラーの場合" {
					require.Equal(t, context.DeadlineExceeded, err)
				}
			} else {
				require.NoError(t, err)
			}
			
			require.Equal(t, tt.expected, video)
		})
	}
}

func TestSearchVideos(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name        string
		keyword     string
		actressID   string
		genreID     string
		makerID     string
		seriesID    string
		directorID  string
		setupMock   func(mockRepo *mockrepo.MockVideoRepository)
		expected    []entity.Video
		expectError bool
	}{
		{
			name:       "正常系",
			keyword:    "keyword",
			actressID:  "1",
			genreID:    "2",
			makerID:    "3",
			seriesID:   "4",
			directorID: "5",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					SearchVideos(gomock.Any(), "keyword", "1", "2", "3", "4", "5").
					Return([]entity.Video{{DmmID: "test123", Title: "Test Video"}}, nil)
			},
			expected:    []entity.Video{{DmmID: "test123", Title: "Test Video"}},
			expectError: false,
		},
		{
			name:       "異常系",
			keyword:    "",
			actressID:  "",
			genreID:    "",
			makerID:    "",
			seriesID:   "",
			directorID: "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					SearchVideos(gomock.Any(), "", "", "", "", "", "").
					Return([]entity.Video{}, errors.New("repository error"))
			},
			expected:    []entity.Video{},
			expectError: true,
		},
		{
			name:       "空配列を返す場合",
			keyword:    "notexist",
			actressID:  "",
			genreID:    "",
			makerID:    "",
			seriesID:   "",
			directorID: "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					SearchVideos(gomock.Any(), "notexist", "", "", "", "", "").
					Return([]entity.Video{}, nil)
			},
			expected:    []entity.Video{},
			expectError: false,
		},
		{
			name:       "タイムアウトエラーの場合",
			keyword:    "timeout",
			actressID:  "",
			genreID:    "",
			makerID:    "",
			seriesID:   "",
			directorID: "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					SearchVideos(gomock.Any(), "timeout", "", "", "", "", "").
					Return(nil, context.DeadlineExceeded)
			},
			expected:    nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockRepo := mockrepo.NewMockVideoRepository(ctrl)
			uc := NewVideoUsecase(mockRepo)
			
			tt.setupMock(mockRepo)
			
			videos, err := uc.SearchVideos(ctx, tt.keyword, tt.actressID, tt.genreID, tt.makerID, tt.seriesID, tt.directorID)
			
			if tt.expectError {
				require.Error(t, err)
				if tt.name == "タイムアウトエラーの場合" {
					require.Equal(t, context.DeadlineExceeded, err)
				}
			} else {
				require.NoError(t, err)
			}
			
			require.Equal(t, tt.expected, videos)
		})
	}
}

func TestGetVideosByID(t *testing.T) {
	ctx := context.Background()
	
	// 大量のIDを含むテスト用のデータを作成
	largeActressIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeActressIDs[i] = "actress_" + string(rune('a'+i%26))
	}
	
	tests := []struct {
		name         string
		actressIDs   []string
		genreIDs     []string
		makerIDs     []string
		seriesIDs    []string
		directorIDs  []string
		hits         int32
		offset       int32
		sort         string
		gteDate      string
		lteDate      string
		site         string
		service      string
		floor        string
		setupMock    func(mockRepo *mockrepo.MockVideoRepository)
		expected     []entity.Video
		expectError  bool
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
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByID(gomock.Any(), 
						[]string{"1"}, []string{"2"}, []string{"3"}, []string{"4"}, []string{"5"},
						int32(10), int32(0), "rank", "2023-01-01", "2023-12-31", "FANZA", "digital", "videoa").
					Return([]entity.Video{{DmmID: "test123", Title: "Test Video"}}, nil)
			},
			expected:    []entity.Video{{DmmID: "test123", Title: "Test Video"}},
			expectError: false,
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
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByID(gomock.Any(), 
						[]string{}, []string{}, []string{}, []string{}, []string{},
						int32(0), int32(0), "", "", "", "", "", "").
					Return([]entity.Video{}, errors.New("repository error"))
			},
			expected:    []entity.Video{},
			expectError: true,
		},
		{
			name:        "空配列を返す場合",
			actressIDs:  []string{"999"},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "",
			service:     "",
			floor:       "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByID(gomock.Any(), 
						[]string{"999"}, []string{}, []string{}, []string{}, []string{},
						int32(10), int32(0), "rank", "", "", "", "", "").
					Return([]entity.Video{}, nil)
			},
			expected:    []entity.Video{},
			expectError: false,
		},
		{
			name:        "タイムアウトエラーの場合",
			actressIDs:  []string{"timeout"},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "",
			service:     "",
			floor:       "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByID(gomock.Any(), 
						[]string{"timeout"}, []string{}, []string{}, []string{}, []string{},
						int32(10), int32(0), "rank", "", "", "", "", "").
					Return(nil, context.DeadlineExceeded)
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:        "nilを返す場合",
			actressIDs:  []string{"nil_case"},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "",
			service:     "",
			floor:       "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByID(gomock.Any(), 
						[]string{"nil_case"}, []string{}, []string{}, []string{}, []string{},
						int32(10), int32(0), "rank", "", "", "", "", "").
					Return(nil, nil)
			},
			expected:    nil,
			expectError: false,
		},
		{
			name:        "大量のIDを指定した場合",
			actressIDs:  largeActressIDs,
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        100,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "",
			service:     "",
			floor:       "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByID(gomock.Any(), 
						gomock.Any(), []string{}, []string{}, []string{}, []string{},
						int32(100), int32(0), "rank", "", "", "", "", "").
					Return([]entity.Video{{DmmID: "mass1", Title: "Mass Test"}}, nil)
			},
			expected:    []entity.Video{{DmmID: "mass1", Title: "Mass Test"}},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockRepo := mockrepo.NewMockVideoRepository(ctrl)
			uc := NewVideoUsecase(mockRepo)
			
			tt.setupMock(mockRepo)
			
			videos, err := uc.GetVideosByID(ctx, 
				tt.actressIDs, tt.genreIDs, tt.makerIDs, tt.seriesIDs, tt.directorIDs,
				tt.hits, tt.offset, tt.sort, tt.gteDate, tt.lteDate, tt.site, tt.service, tt.floor)
			
			if tt.expectError {
				require.Error(t, err)
				if tt.name == "タイムアウトエラーの場合" {
					require.Equal(t, context.DeadlineExceeded, err)
				}
			} else {
				require.NoError(t, err)
			}
			
			require.Equal(t, tt.expected, videos)
		})
	}
}

func TestGetVideosByKeyword(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name        string
		keyword     string
		hits        int32
		offset      int32
		sort        string
		gteDate     string
		lteDate     string
		site        string
		service     string
		floor       string
		setupMock   func(mockRepo *mockrepo.MockVideoRepository)
		expected    []entity.Video
		expectError bool
	}{
		{
			name:    "正常系",
			keyword: "test",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "2023-01-01",
			lteDate: "2023-12-31",
			site:    "FANZA",
			service: "digital",
			floor:   "videoa",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByKeyword(gomock.Any(), "test", int32(10), int32(0), "rank",
						"2023-01-01", "2023-12-31", "FANZA", "digital", "videoa").
					Return([]entity.Video{{DmmID: "test123", Title: "Test Video"}}, nil)
			},
			expected:    []entity.Video{{DmmID: "test123", Title: "Test Video"}},
			expectError: false,
		},
		{
			name:    "異常系",
			keyword: "",
			hits:    0,
			offset:  0,
			sort:    "",
			gteDate: "",
			lteDate: "",
			site:    "",
			service: "",
			floor:   "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByKeyword(gomock.Any(), "", int32(0), int32(0), "",
						"", "", "", "", "").
					Return([]entity.Video{}, errors.New("repository error"))
			},
			expected:    []entity.Video{},
			expectError: true,
		},
		{
			name:    "空配列を返す場合",
			keyword: "notexist",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "",
			lteDate: "",
			site:    "",
			service: "",
			floor:   "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByKeyword(gomock.Any(), "notexist", int32(10), int32(0), "rank",
						"", "", "", "", "").
					Return([]entity.Video{}, nil)
			},
			expected:    []entity.Video{},
			expectError: false,
		},
		{
			name:    "タイムアウトエラーの場合",
			keyword: "timeout",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "",
			lteDate: "",
			site:    "",
			service: "",
			floor:   "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByKeyword(gomock.Any(), "timeout", int32(10), int32(0), "rank",
						"", "", "", "", "").
					Return(nil, context.DeadlineExceeded)
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:    "非常に長いキーワードの場合",
			keyword: string(make([]rune, 500)),
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "",
			lteDate: "",
			site:    "",
			service: "",
			floor:   "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByKeyword(gomock.Any(), gomock.Any(), int32(10), int32(0), "rank",
						"", "", "", "", "").
					Return([]entity.Video{}, nil)
			},
			expected:    []entity.Video{},
			expectError: false,
		},
		{
			name:    "異常なフィールド値を含む動画を返す場合",
			keyword: "invalid",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "",
			lteDate: "",
			site:    "",
			service: "",
			floor:   "",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				// 価格がマイナス値、日付が不正等の異常値を含むケース
				// CreatedAtはtime.Time型なので空のtime.Timeを使用
				m.EXPECT().
					GetVideosByKeyword(gomock.Any(), "invalid", int32(10), int32(0), "rank",
						"", "", "", "", "").
					Return([]entity.Video{
						{DmmID: "invalid1", Title: "Invalid Video", Price: -100},
						{DmmID: "invalid2", Title: "Invalid Date", CreatedAt: time.Time{}},
					}, nil)
			},
			expected: []entity.Video{
				{DmmID: "invalid1", Title: "Invalid Video", Price: -100},
				{DmmID: "invalid2", Title: "Invalid Date", CreatedAt: time.Time{}},
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockRepo := mockrepo.NewMockVideoRepository(ctrl)
			uc := NewVideoUsecase(mockRepo)
			
			tt.setupMock(mockRepo)
			
			videos, err := uc.GetVideosByKeyword(ctx, tt.keyword, tt.hits, tt.offset, tt.sort,
				tt.gteDate, tt.lteDate, tt.site, tt.service, tt.floor)
			
			if tt.expectError {
				require.Error(t, err)
				if tt.name == "タイムアウトエラーの場合" {
					require.Equal(t, context.DeadlineExceeded, err)
				}
			} else {
				require.NoError(t, err)
			}
			
			require.Equal(t, tt.expected, videos)
		})
	}
}
