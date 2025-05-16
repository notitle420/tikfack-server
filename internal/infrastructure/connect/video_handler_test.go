package connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/tikfack/server/gen/video"
	mockvideo "github.com/tikfack/server/internal/application/usecase/mock"
	"github.com/tikfack/server/internal/domain/entity"
)

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
	ctx := context.Background()
	targetDate, _ := time.Parse("2006-01-02", "2024-01-01")
	
	tests := []struct {
		name        string
		request     *pb.GetVideosByDateRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    []entity.Video
		expectError bool
	}{
		{
			name: "正常系",
			request: &pb.GetVideosByDateRequest{
				Date: "2024-01-01",
			},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate).
					Return([]entity.Video{
						{
							DmmID:        "test123",
							Title:        "動画1",
							URL:          "https://example.com",
							SampleURL:    "https://example.com/sample",
							ThumbnailURL: "https://example.com/thumb.jpg",
							CreatedAt:    targetDate,
							Price:        1000,
							LikesCount:   500,
						},
					}, nil)
			},
			expected: []entity.Video{
				{
					DmmID:        "test123",
					Title:        "動画1",
					URL:          "https://example.com",
					SampleURL:    "https://example.com/sample",
					ThumbnailURL: "https://example.com/thumb.jpg",
					CreatedAt:    targetDate,
					Price:        1000,
					LikesCount:   500,
				},
			},
			expectError: false,
		},
		{
			name: "異常系 - 不正な日付形式",
			request: &pb.GetVideosByDateRequest{
				Date: "invalid-date",
			},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				// モックは呼ばれないので設定不要
			},
			expected:    nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := &videoServiceServer{videoUsecase: mockUsecase}
			
			tt.setupMock(mockUsecase)
			
			req := connect.NewRequest(tt.request)
			resp, err := handler.GetVideosByDate(ctx, req)
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotEmpty(t, resp.Msg.Videos)
				assert.Equal(t, tt.expected[0].DmmID, resp.Msg.Videos[0].DmmId)
				assert.Equal(t, tt.expected[0].Title, resp.Msg.Videos[0].Title)
			}
		})
	}
}

func TestGetVideoById(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	
	tests := []struct {
		name        string
		request     *pb.GetVideoByIdRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    *entity.Video
		expectError bool
	}{
		{
			name: "正常系",
			request: &pb.GetVideoByIdRequest{
				DmmId: "test123",
			},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideoById(gomock.Any(), "test123").
					Return(&entity.Video{
						DmmID:        "test123",
						Title:        "動画1",
						URL:          "https://example.com",
						SampleURL:    "https://example.com/sample",
						ThumbnailURL: "https://example.com/thumb.jpg",
						CreatedAt:    now,
						Price:        1000,
						LikesCount:   500,
					}, nil)
			},
			expected: &entity.Video{
				DmmID:        "test123",
				Title:        "動画1",
				URL:          "https://example.com",
				SampleURL:    "https://example.com/sample",
				ThumbnailURL: "https://example.com/thumb.jpg",
				CreatedAt:    now,
				Price:        1000,
				LikesCount:   500,
			},
			expectError: false,
		},
		{
			name: "異常系 - 動画が見つからない",
			request: &pb.GetVideoByIdRequest{
				DmmId: "notfound",
			},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideoById(gomock.Any(), "notfound").
					Return(nil, errors.New("not found"))
			},
			expected:    nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := &videoServiceServer{videoUsecase: mockUsecase}
			
			tt.setupMock(mockUsecase)
			
			req := connect.NewRequest(tt.request)
			resp, err := handler.GetVideoById(ctx, req)
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.Video)
				assert.Equal(t, tt.expected.DmmID, resp.Msg.Video.DmmId)
				assert.Equal(t, tt.expected.Title, resp.Msg.Video.Title)
			}
		})
	}
}

func TestSearchVideos(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	
	tests := []struct {
		name        string
		request     *pb.SearchVideosRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    []entity.Video
		expectError bool
	}{
		{
			name: "正常系",
			request: &pb.SearchVideosRequest{
				Keyword:    "keyword",
				ActressId:  "1",
				GenreId:    "2",
				MakerId:    "3",
				SeriesId:   "4",
				DirectorId: "5",
			},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					SearchVideos(gomock.Any(), "keyword", "1", "2", "3", "4", "5").
					Return([]entity.Video{
						{
							DmmID:        "test123",
							Title:        "動画1",
							URL:          "https://example.com",
							SampleURL:    "https://example.com/sample",
							ThumbnailURL: "https://example.com/thumb.jpg",
							CreatedAt:    now,
							Price:        1000,
							LikesCount:   500,
						},
					}, nil)
			},
			expected: []entity.Video{
				{
					DmmID:        "test123",
					Title:        "動画1",
					URL:          "https://example.com",
					SampleURL:    "https://example.com/sample",
					ThumbnailURL: "https://example.com/thumb.jpg",
					CreatedAt:    now,
					Price:        1000,
					LikesCount:   500,
				},
			},
			expectError: false,
		},
		{
			name:    "異常系 - 検索エラー",
			request: &pb.SearchVideosRequest{},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					SearchVideos(gomock.Any(), "", "", "", "", "", "").
					Return([]entity.Video{}, errors.New("search error"))
			},
			expected:    nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := &videoServiceServer{videoUsecase: mockUsecase}
			
			tt.setupMock(mockUsecase)
			
			req := connect.NewRequest(tt.request)
			resp, err := handler.SearchVideos(ctx, req)
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotEmpty(t, resp.Msg.Videos)
				assert.Equal(t, tt.expected[0].DmmID, resp.Msg.Videos[0].DmmId)
				assert.Equal(t, tt.expected[0].Title, resp.Msg.Videos[0].Title)
			}
		})
	}
}

func TestGetVideosByID(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	
	tests := []struct {
		name        string
		request     *pb.GetVideosByIDRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    []entity.Video
		expectError bool
	}{
		{
			name: "正常系",
			request: &pb.GetVideosByIDRequest{
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
			},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByID(gomock.Any(), 
						[]string{"1"}, []string{"2"}, []string{"3"}, []string{"4"}, []string{"5"}, 
						int32(10), int32(0), "rank", "2023-01-01", "2023-12-31", "FANZA", "digital", "videoa").
					Return([]entity.Video{
						{
							DmmID:        "test123",
							Title:        "動画1",
							URL:          "https://example.com",
							SampleURL:    "https://example.com/sample",
							ThumbnailURL: "https://example.com/thumb.jpg",
							CreatedAt:    now,
							Price:        1000,
							LikesCount:   500,
						},
					}, nil)
			},
			expected: []entity.Video{
				{
					DmmID:        "test123",
					Title:        "動画1",
					URL:          "https://example.com",
					SampleURL:    "https://example.com/sample",
					ThumbnailURL: "https://example.com/thumb.jpg",
					CreatedAt:    now,
					Price:        1000,
					LikesCount:   500,
				},
			},
			expectError: false,
		},
		{
			name:    "異常系 - IDによる検索エラー",
			request: &pb.GetVideosByIDRequest{},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByID(gomock.Any(), 
						[]string{}, []string{}, []string{}, []string{}, []string{}, 
						int32(0), int32(0), "", "", "", "", "", "").
					Return([]entity.Video{}, errors.New("get by ID error"))
			},
			expected:    nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := &videoServiceServer{videoUsecase: mockUsecase}
			
			tt.setupMock(mockUsecase)
			
			req := connect.NewRequest(tt.request)
			resp, err := handler.GetVideosByID(ctx, req)
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotEmpty(t, resp.Msg.Videos)
				assert.Equal(t, tt.expected[0].DmmID, resp.Msg.Videos[0].DmmId)
				assert.Equal(t, tt.expected[0].Title, resp.Msg.Videos[0].Title)
			}
		})
	}
}

func TestGetVideosByKeyword(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	
	tests := []struct {
		name        string
		request     *pb.GetVideosByKeywordRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    []entity.Video
		expectError bool
	}{
		{
			name: "正常系",
			request: &pb.GetVideosByKeywordRequest{
				Keyword:  "test",
				Hits:     10,
				Offset:   0,
				Sort:     "rank",
				GteDate:  "2023-01-01",
				LteDate:  "2023-12-31",
				Site:     "FANZA",
				Service:  "digital",
				Floor:    "videoa",
			},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByKeyword(gomock.Any(), "test", int32(10), int32(0), "rank", 
						"2023-01-01", "2023-12-31", "FANZA", "digital", "videoa").
					Return([]entity.Video{
						{
							DmmID:        "test123",
							Title:        "動画1",
							URL:          "https://example.com",
							SampleURL:    "https://example.com/sample",
							ThumbnailURL: "https://example.com/thumb.jpg",
							CreatedAt:    now,
							Price:        1000,
							LikesCount:   500,
						},
					}, nil)
			},
			expected: []entity.Video{
				{
					DmmID:        "test123",
					Title:        "動画1",
					URL:          "https://example.com",
					SampleURL:    "https://example.com/sample",
					ThumbnailURL: "https://example.com/thumb.jpg",
					CreatedAt:    now,
					Price:        1000,
					LikesCount:   500,
				},
			},
			expectError: false,
		},
		{
			name:    "異常系 - キーワードによる検索エラー",
			request: &pb.GetVideosByKeywordRequest{},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByKeyword(gomock.Any(), "", int32(0), int32(0), "", "", "", "", "", "").
					Return([]entity.Video{}, errors.New("get by keyword error"))
			},
			expected:    nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			
			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := &videoServiceServer{videoUsecase: mockUsecase}
			
			tt.setupMock(mockUsecase)
			
			req := connect.NewRequest(tt.request)
			resp, err := handler.GetVideosByKeyword(ctx, req)
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotEmpty(t, resp.Msg.Videos)
				assert.Equal(t, tt.expected[0].DmmID, resp.Msg.Videos[0].DmmId)
				assert.Equal(t, tt.expected[0].Title, resp.Msg.Videos[0].Title)
			}
		})
	}
}
