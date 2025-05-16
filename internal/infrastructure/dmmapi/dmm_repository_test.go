package dmmapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tikfack/server/internal/domain/entity"
)

func TestGetVideosByDate(t *testing.T) {
	ctx := context.Background()
	fakeDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	item := Item{
		ContentID: "vid1",
		Title:     "テスト動画",
		Date:      "2024-01-01 00:00:00",
	}
	videoEntity := entity.Video{DmmID: "vid1", Title: "テスト動画"}
	expectedAPIError := errors.New("API error")
	
	tests := []struct {
		name        string
		date        time.Time
		setupMock   func(mockClient *MockClientInterface, mockMapper *MockMapperInterface)
		expected    []entity.Video
		expectedErr error
	}{
		{
			name: "APIから結果が返された場合は動画配列を返す",
			date: fakeDate,
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name: "クライアント呼び出しが失敗した場合はエラーを返す",
			date: fakeDate,
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					Return(expectedAPIError)
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    nil,
			expectedErr: expectedAPIError,
		},
		{
			name: "APIがnil項目を返した場合は空配列を返す",
			date: fakeDate,
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = nil
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    []entity.Video{},
			expectedErr: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockClient := NewMockClientInterface(ctrl)
			mockMapper := NewMockMapperInterface(ctrl)
			
			tt.setupMock(mockClient, mockMapper)
			
			repo := NewRepositoryWithDeps(mockClient, mockMapper)
			videos, err := repo.GetVideosByDate(ctx, tt.date)
			
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, videos)
			}
		})
	}
}

func TestGetVideoById(t *testing.T) {
	ctx := context.Background()
	
	item := Item{
		ContentID: "vid1",
		Title:     "テスト動画",
		Date:      "2024-01-01 00:00:00",
	}
	// 実際のコードで変換される時刻を使用
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	videoEntity := entity.Video{
		DmmID:     "vid1", 
		Title:     "テスト動画",
		CreatedAt: createdAt,
	}
	expectedAPIError := errors.New("API error")
	videoNotFoundErr := errors.New("video not found")
	
	tests := []struct {
		name        string
		videoID     string
		setupMock   func(mockClient *MockClientInterface, mockMapper *MockMapperInterface)
		expected    *entity.Video
		expectedErr error
	}{
		{
			name:    "IDで動画が見つかった場合は動画を返す",
			videoID: "vid1",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    &videoEntity,
			expectedErr: nil,
		},
		{
			name:    "クライアント呼び出しが失敗した場合はエラーを返す",
			videoID: "vid1",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					Return(expectedAPIError)
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    nil,
			expectedErr: expectedAPIError,
		},
		{
			name:    "動画が存在しない場合は見つからないエラーを返す",
			videoID: "nonexistent",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    nil,
			expectedErr: videoNotFoundErr,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockClient := NewMockClientInterface(ctrl)
			mockMapper := NewMockMapperInterface(ctrl)
			
			tt.setupMock(mockClient, mockMapper)
			
			repo := NewRepositoryWithDeps(mockClient, mockMapper)
			video, err := repo.GetVideoById(ctx, tt.videoID)
			
			if tt.expectedErr != nil {
				require.Error(t, err)
				if tt.expectedErr == expectedAPIError {
					require.ErrorIs(t, err, expectedAPIError)
				} else {
					require.ErrorContains(t, err, "見つかりませんでした")
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, video)
			}
		})
	}
}

func TestSearchVideos(t *testing.T) {
	ctx := context.Background()
	
	item := Item{
		ContentID: "vid1",
		Title:     "テスト動画",
		Date:      "2024-01-01 00:00:00",
	}
	videoEntity := entity.Video{DmmID: "vid1", Title: "テスト動画"}
	expectedAPIError := errors.New("API error")
	
	tests := []struct {
		name        string
		keyword     string
		actressID   string
		genreID     string
		makerID     string
		seriesID    string
		directorID  string
		setupMock   func(mockClient *MockClientInterface, mockMapper *MockMapperInterface)
		expected    []entity.Video
		expectedErr error
	}{
		{
			name:       "キーワードで検索した場合は動画を返す",
			keyword:    "テスト",
			actressID:  "",
			genreID:    "",
			makerID:    "",
			seriesID:   "",
			directorID: "",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:       "女優IDで検索した場合は動画を返す",
			keyword:    "",
			actressID:  "12345",
			genreID:    "",
			makerID:    "",
			seriesID:   "",
			directorID: "",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:       "ジャンルIDで検索した場合は動画を返す",
			keyword:    "",
			actressID:  "",
			genreID:    "12345",
			makerID:    "",
			seriesID:   "",
			directorID: "",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:       "メーカーIDで検索した場合は動画を返す",
			keyword:    "",
			actressID:  "",
			genreID:    "",
			makerID:    "12345",
			seriesID:   "",
			directorID: "",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:       "シリーズIDで検索した場合は動画を返す",
			keyword:    "",
			actressID:  "",
			genreID:    "",
			makerID:    "",
			seriesID:   "12345",
			directorID: "",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:       "監督IDで検索した場合は動画を返す",
			keyword:    "",
			actressID:  "",
			genreID:    "",
			makerID:    "",
			seriesID:   "",
			directorID: "12345",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:       "クライアント呼び出しが失敗した場合はエラーを返す",
			keyword:    "テスト",
			actressID:  "",
			genreID:    "",
			makerID:    "",
			seriesID:   "",
			directorID: "",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					Return(expectedAPIError)
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    nil,
			expectedErr: expectedAPIError,
		},
		{
			name:       "結果が見つからない場合は空配列を返す",
			keyword:    "存在しない",
			actressID:  "",
			genreID:    "",
			makerID:    "",
			seriesID:   "",
			directorID: "",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    []entity.Video{},
			expectedErr: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockClient := NewMockClientInterface(ctrl)
			mockMapper := NewMockMapperInterface(ctrl)
			
			tt.setupMock(mockClient, mockMapper)
			
			repo := NewRepositoryWithDeps(mockClient, mockMapper)
			videos, err := repo.SearchVideos(ctx, tt.keyword, tt.actressID, tt.genreID, tt.makerID, tt.seriesID, tt.directorID)
			
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, videos)
			}
		})
	}
}

func TestGetVideosByID(t *testing.T) {
	ctx := context.Background()
	
	item := Item{
		ContentID: "vid1",
		Title:     "テスト動画",
		Date:      "2024-01-01 00:00:00",
	}
	videoEntity := entity.Video{DmmID: "vid1", Title: "テスト動画"}
	expectedAPIError := errors.New("API error")
	
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
		setupMock    func(mockClient *MockClientInterface, mockMapper *MockMapperInterface)
		expected     []entity.Video
		expectedErr  error
	}{
		{
			name:        "女優IDで検索した場合は動画を返す",
			actressIDs:  []string{"12345"},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:        "ジャンルIDで検索した場合は動画を返す",
			actressIDs:  []string{},
			genreIDs:    []string{"12345"},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:        "メーカーIDで検索した場合は動画を返す",
			actressIDs:  []string{},
			genreIDs:    []string{},
			makerIDs:    []string{"12345"},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:        "シリーズIDで検索した場合は動画を返す",
			actressIDs:  []string{},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{"12345"},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:        "監督IDで検索した場合は動画を返す",
			actressIDs:  []string{},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{"12345"},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:        "オフセットがある場合は正しく動画を返す",
			actressIDs:  []string{"12345"},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      5,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:        "開始日が指定されている場合は正しく動画を返す",
			actressIDs:  []string{"12345"},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "2024-01-01T00:00:00",
			lteDate:     "",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:        "終了日が指定されている場合は正しく動画を返す",
			actressIDs:  []string{"12345"},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "2024-01-31T23:59:59",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:        "クライアント呼び出しが失敗した場合はエラーを返す",
			actressIDs:  []string{"12345"},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					Return(expectedAPIError)
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    nil,
			expectedErr: expectedAPIError,
		},
		{
			name:        "結果が見つからない場合は空配列を返す",
			actressIDs:  []string{"999999"},
			genreIDs:    []string{},
			makerIDs:    []string{},
			seriesIDs:   []string{},
			directorIDs: []string{},
			hits:        10,
			offset:      0,
			sort:        "rank",
			gteDate:     "",
			lteDate:     "",
			site:        "FANZA",
			service:     "digital",
			floor:       "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    []entity.Video{},
			expectedErr: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockClient := NewMockClientInterface(ctrl)
			mockMapper := NewMockMapperInterface(ctrl)
			
			tt.setupMock(mockClient, mockMapper)
			
			repo := NewRepositoryWithDeps(mockClient, mockMapper)
			videos, err := repo.GetVideosByID(ctx, 
				tt.actressIDs, tt.genreIDs, tt.makerIDs, tt.seriesIDs, tt.directorIDs,
				tt.hits, tt.offset, tt.sort, tt.gteDate, tt.lteDate, tt.site, tt.service, tt.floor)
			
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, videos)
			}
		})
	}
}

func TestGetVideosByKeyword(t *testing.T) {
	ctx := context.Background()
	
	item := Item{
		ContentID: "vid1",
		Title:     "テスト動画",
		Date:      "2024-01-01 00:00:00",
	}
	videoEntity := entity.Video{DmmID: "vid1", Title: "テスト動画"}
	expectedAPIError := errors.New("API error")
	
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
		setupMock   func(mockClient *MockClientInterface, mockMapper *MockMapperInterface)
		expected    []entity.Video
		expectedErr error
	}{
		{
			name:    "キーワードで検索した場合は動画を返す",
			keyword: "テスト",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "",
			lteDate: "",
			site:    "FANZA",
			service: "digital",
			floor:   "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:    "オフセットがある場合は正しく動画を返す",
			keyword: "テスト",
			hits:    10,
			offset:  5,
			sort:    "rank",
			gteDate: "",
			lteDate: "",
			site:    "FANZA",
			service: "digital",
			floor:   "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:    "開始日が指定されている場合は正しく動画を返す",
			keyword: "テスト",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "2024-01-01T00:00:00",
			lteDate: "",
			site:    "FANZA",
			service: "digital",
			floor:   "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:    "終了日が指定されている場合は正しく動画を返す",
			keyword: "テスト",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "",
			lteDate: "2024-01-31T23:59:59",
			site:    "FANZA",
			service: "digital",
			floor:   "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:    "サイト指定が空文字の場合はデフォルト値を使用する",
			keyword: "テスト",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "",
			lteDate: "",
			site:    "",
			service: "digital",
			floor:   "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{item}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().
					ConvertItem(gomock.Eq(item)).
					Return(videoEntity)
			},
			expected:    []entity.Video{videoEntity},
			expectedErr: nil,
		},
		{
			name:    "クライアント呼び出しが失敗した場合はエラーを返す",
			keyword: "テスト",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "",
			lteDate: "",
			site:    "FANZA",
			service: "digital",
			floor:   "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					Return(expectedAPIError)
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    nil,
			expectedErr: expectedAPIError,
		},
		{
			name:    "結果が見つからない場合は空配列を返す",
			keyword: "存在しない",
			hits:    10,
			offset:  0,
			sort:    "rank",
			gteDate: "",
			lteDate: "",
			site:    "FANZA",
			service: "digital",
			floor:   "videoa",
			setupMock: func(mockClient *MockClientInterface, mockMapper *MockMapperInterface) {
				resp := &Response{}
				resp.Result.Items = []Item{}
				mockClient.EXPECT().
					Call(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, v interface{}) error {
						*v.(*Response) = *resp
						return nil
					})
				mockMapper.EXPECT().ConvertItem(gomock.Any()).Times(0)
			},
			expected:    []entity.Video{},
			expectedErr: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockClient := NewMockClientInterface(ctrl)
			mockMapper := NewMockMapperInterface(ctrl)
			
			tt.setupMock(mockClient, mockMapper)
			
			repo := NewRepositoryWithDeps(mockClient, mockMapper)
			videos, err := repo.GetVideosByKeyword(ctx, tt.keyword, tt.hits, tt.offset, tt.sort,
				tt.gteDate, tt.lteDate, tt.site, tt.service, tt.floor)
			
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, videos)
			}
		})
	}
}

func TestDefaultIfEmpty(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		def      string
		expected string
	}{
		{
			name:     "空文字列の場合はデフォルト値を返す",
			value:    "",
			def:      "DEFAULT",
			expected: "DEFAULT",
		},
		{
			name:     "値が存在する場合はその値を返す",
			value:    "VALUE",
			def:      "DEFAULT",
			expected: "VALUE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultIfEmpty(tt.value, tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}
