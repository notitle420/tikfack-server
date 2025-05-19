package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tikfack/server/internal/domain/entity"
	mockrepo "github.com/tikfack/server/internal/domain/repository/mock"
	"go.uber.org/mock/gomock"
)

var (
	testTime = time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	testVideo = entity.Video{
		DmmID:        "test123",
		Title:        "Test Video",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    testTime,
		Price:        1000,
		LikesCount:   500,
		Actresses:    []entity.Actress{{ID: "a1", Name: "女優A"}},
		Genres:       []entity.Genre{{ID: "g1", Name: "ジャンルA"}},
		Makers:       []entity.Maker{{ID: "m1", Name: "メーカーA"}},
		Series:       []entity.Series{{ID: "s1", Name: "シリーズA"}},
		Directors:    []entity.Director{{ID: "d1", Name: "監督A"}},
		Review:       entity.Review{Count: 100, Average: 4.5},
	}
	testMetadata = &entity.SearchMetadata{
		ResultCount:   10,
		TotalCount:    100,
		FirstPosition: 1,
	}
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
		hits        int32
		offset      int32
		setupMock   func(mockRepo *mockrepo.MockVideoRepository)
		expected    []entity.Video
		expectedMD  *entity.SearchMetadata
		expectError bool
	}{
		{
			name:   "正常系",
			date:   time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			hits:   10,
			offset: 0,
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC), int32(10), int32(0)).
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expected:    []entity.Video{testVideo},
			expectedMD:  testMetadata,
			expectError: false,
		},
		{
			name:   "異常系",
			date:   time.Time{},
			hits:   0,
			offset: 0,
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), time.Time{}, int32(0), int32(0)).
					Return(nil, nil, errors.New("repository error"))
			},
			expected:    nil,
			expectedMD:  nil,
			expectError: true,
		},
		{
			name:   "空配列を返す場合",
			date:   time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			hits:   20,
			offset: 0,
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), int32(20), int32(0)).
					Return([]entity.Video{}, testMetadata, nil)
			},
			expected:    []entity.Video{},
			expectedMD:  testMetadata,
			expectError: false,
		},
		{
			name:   "タイムアウトエラーの場合",
			date:   time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			hits:   10,
			offset: 0,
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), int32(10), int32(0)).
					Return(nil, nil, context.DeadlineExceeded)
			},
			expected:    nil,
			expectedMD:  nil,
			expectError: true,
		},
		{
			name:   "ページネーションのテスト",
			date:   time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
			hits:   5,
			offset: 10,
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC), int32(5), int32(10)).
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expected:    []entity.Video{testVideo},
			expectedMD:  testMetadata,
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
			
			videos, metadata, err := uc.GetVideosByDate(ctx, tt.date, tt.hits, tt.offset)
			
			if tt.expectError {
				require.Error(t, err)
				if tt.name == "タイムアウトエラーの場合" {
					require.Equal(t, context.DeadlineExceeded, err)
				}
			} else {
				require.NoError(t, err)
			}
			
			require.Equal(t, tt.expected, videos)
			require.Equal(t, tt.expectedMD, metadata)
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
					Return(&entity.Video{
						DmmID: "test123", 
						Title: "Test Video",
						Review: entity.Review{
							Count:   15,
							Average: 3.8,
						},
					}, nil)
			},
			expected: &entity.Video{
				DmmID: "test123", 
				Title: "Test Video",
				Review: entity.Review{
					Count:   15,
					Average: 3.8,
				},
			},
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
		expectedMD  *entity.SearchMetadata
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
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expected:    []entity.Video{testVideo},
			expectedMD:  testMetadata,
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
					Return(nil, nil, errors.New("repository error"))
			},
			expected:    nil,
			expectedMD:  nil,
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
					Return([]entity.Video{}, testMetadata, nil)
			},
			expected:    []entity.Video{},
			expectedMD:  testMetadata,
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
					Return(nil, nil, context.DeadlineExceeded)
			},
			expected:    nil,
			expectedMD:  nil,
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
			
			videos, metadata, err := uc.SearchVideos(ctx, tt.keyword, tt.actressID, tt.genreID, tt.makerID, tt.seriesID, tt.directorID)
			
			if tt.expectError {
				require.Error(t, err)
				if tt.name == "タイムアウトエラーの場合" {
					require.Equal(t, context.DeadlineExceeded, err)
				}
			} else {
				require.NoError(t, err)
			}
			
			require.Equal(t, tt.expected, videos)
			require.Equal(t, tt.expectedMD, metadata)
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
		expectedMD   *entity.SearchMetadata
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
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expected:    []entity.Video{testVideo},
			expectedMD:  testMetadata,
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
						gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
						int32(0), int32(0), "", "", "", "", "", "").
					Return(nil, nil, errors.New("repository error"))
			},
			expected:    nil,
			expectedMD:  nil,
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
					Return([]entity.Video{}, testMetadata, nil)
			},
			expected:    []entity.Video{},
			expectedMD:  testMetadata,
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
					Return(nil, nil, context.DeadlineExceeded)
			},
			expected:    nil,
			expectedMD:  nil,
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
					Return(nil, nil, nil)
			},
			expected:    nil,
			expectedMD:  nil,
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
					Return([]entity.Video{{DmmID: "mass1", Title: "Mass Test"}}, testMetadata, nil)
			},
			expected:    []entity.Video{{DmmID: "mass1", Title: "Mass Test"}},
			expectedMD:  testMetadata,
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
			
			videos, metadata, err := uc.GetVideosByID(ctx, 
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
			require.Equal(t, tt.expectedMD, metadata)
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
		expectedMD  *entity.SearchMetadata
		expectError bool
	}{
		{
			name:    "正常系",
			keyword: "keyword",
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
					GetVideosByKeyword(gomock.Any(), 
						"keyword", int32(10), int32(0), "rank", 
						"2023-01-01", "2023-12-31", "FANZA", "digital", "videoa").
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expected:    []entity.Video{testVideo},
			expectedMD:  testMetadata,
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
					Return(nil, nil, errors.New("repository error"))
			},
			expected:    nil,
			expectedMD:  nil,
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
					Return([]entity.Video{}, testMetadata, nil)
			},
			expected:    []entity.Video{},
			expectedMD:  testMetadata,
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
					Return(nil, nil, context.DeadlineExceeded)
			},
			expected:    nil,
			expectedMD:  nil,
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
					Return([]entity.Video{}, testMetadata, nil)
			},
			expected:    []entity.Video{},
			expectedMD:  testMetadata,
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
					}, testMetadata, nil)
			},
			expected: []entity.Video{
				{DmmID: "invalid1", Title: "Invalid Video", Price: -100},
				{DmmID: "invalid2", Title: "Invalid Date", CreatedAt: time.Time{}},
			},
			expectedMD: testMetadata,
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
			
			videos, metadata, err := uc.GetVideosByKeyword(ctx, tt.keyword, tt.hits, tt.offset, tt.sort,
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
			require.Equal(t, tt.expectedMD, metadata)
		})
	}
}
