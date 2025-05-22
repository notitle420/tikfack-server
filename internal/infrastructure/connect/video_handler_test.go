package connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tikfack/server/gen/video"
	mockvideo "github.com/tikfack/server/internal/application/usecase/mock"
	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/infrastructure/auth"
)

// ===== 共通テストデータ =====
var (
	testTime  = time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	testVideo = entity.Video{
		DmmID:        "test123",
		Title:        "動画1",
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

// ===== 共通比較ヘルパー =====
func checkVideoFields(t *testing.T, got *pb.Video, want entity.Video) {
	require.Equal(t, want.DmmID, got.DmmId)
	require.Equal(t, want.Title, got.Title)
	require.Equal(t, want.URL, got.Url)
	require.Equal(t, want.SampleURL, got.SampleUrl)
	require.Equal(t, want.ThumbnailURL, got.ThumbnailUrl)
	require.Equal(t, want.CreatedAt.Format(time.RFC3339), got.CreatedAt)
	require.Equal(t, int32(want.Price), got.Price)
	require.Equal(t, int32(want.LikesCount), got.LikesCount)
	require.Equal(t, int32(want.Review.Count), got.Review.Count)
	require.Equal(t, want.Review.Average, got.Review.Average)

	require.Len(t, got.Actresses, len(want.Actresses))
	for i, a := range want.Actresses {
		require.Equal(t, a.ID, got.Actresses[i].Id)
		require.Equal(t, a.Name, got.Actresses[i].Name)
	}

	require.Len(t, got.Genres, len(want.Genres))
	for i, g := range want.Genres {
		require.Equal(t, g.ID, got.Genres[i].Id)
		require.Equal(t, g.Name, got.Genres[i].Name)
	}

	require.Len(t, got.Makers, len(want.Makers))
	for i, m := range want.Makers {
		require.Equal(t, m.ID, got.Makers[i].Id)
		require.Equal(t, m.Name, got.Makers[i].Name)
	}

	require.Len(t, got.Series, len(want.Series))
	for i, s := range want.Series {
		require.Equal(t, s.ID, got.Series[i].Id)
		require.Equal(t, s.Name, got.Series[i].Name)
	}

	require.Len(t, got.Directors, len(want.Directors))
	for i, d := range want.Directors {
		require.Equal(t, d.ID, got.Directors[i].Id)
		require.Equal(t, d.Name, got.Directors[i].Name)
	}
}

func checkMetadata(t *testing.T, got *pb.SearchMetadata, want *entity.SearchMetadata) {
	if want == nil {
		require.Nil(t, got)
		return
	}
	require.NotNil(t, got)
	require.Equal(t, int32(want.ResultCount), got.ResultCount)
	require.Equal(t, int32(want.TotalCount), got.TotalCount)
	require.Equal(t, int32(want.FirstPosition), got.FirstPosition)
}

func TestGetVideosByDate(t *testing.T) {
	tests := []struct {
		name           string
		req            *pb.GetVideosByDateRequest
		mockSetup      func(m *mockvideo.MockVideoUsecase)
		expectedVideos []entity.Video
		expectedMD     *entity.SearchMetadata
		expectedError  error
	}{
		{
			name: "正常系：日付指定で取得",
			req: &pb.GetVideosByDateRequest{
				Date:   "2024-01-01",
				Hits:   20,
				Offset: 0,
			},
			mockSetup: func(m *mockvideo.MockVideoUsecase) {
				targetDate, _ := time.Parse("2006-01-02", "2024-01-01")
				m.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate, int32(20), int32(0)).
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expectedVideos: []entity.Video{testVideo},
			expectedMD:     testMetadata,
			expectedError:  nil,
		},
		{
			name: "正常系：日付未指定で現在日時で取得",
			req: &pb.GetVideosByDateRequest{
				Hits:   20,
				Offset: 0,
			},
			mockSetup: func(m *mockvideo.MockVideoUsecase) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), gomock.Any(), int32(20), int32(0)).
					DoAndReturn(func(_ context.Context, date time.Time, hits, offset int32) ([]entity.Video, *entity.SearchMetadata, error) {
						// 現在時刻との差分が1秒以内であることを確認
						if time.Since(date) > time.Second {
							t.Error("期待される日時との差分が大きすぎます")
						}
						return []entity.Video{testVideo}, testMetadata, nil
					})
			},
			expectedVideos: []entity.Video{testVideo},
			expectedMD:     testMetadata,
			expectedError:  nil,
		},
		{
			name: "正常系：ページネーション",
			req: &pb.GetVideosByDateRequest{
				Date:   "2024-01-01",
				Hits:   10,
				Offset: 20,
			},
			mockSetup: func(m *mockvideo.MockVideoUsecase) {
				targetDate, _ := time.Parse("2006-01-02", "2024-01-01")
				m.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate, int32(10), int32(20)).
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expectedVideos: []entity.Video{testVideo},
			expectedMD:     testMetadata,
			expectedError:  nil,
		},
		{
			name: "正常系：hitsの最大値制限",
			req: &pb.GetVideosByDateRequest{
				Date:   "2024-01-01",
				Hits:   200, // 最大値100を超える
				Offset: 0,
			},
			mockSetup: func(m *mockvideo.MockVideoUsecase) {
				targetDate, _ := time.Parse("2006-01-02", "2024-01-01")
				m.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate, int32(100), int32(0)). // 100に制限される
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expectedVideos: []entity.Video{testVideo},
			expectedMD:     testMetadata,
			expectedError:  nil,
		},
		{
			name: "正常系：offsetの最大値制限",
			req: &pb.GetVideosByDateRequest{
				Date:   "2024-01-01",
				Hits:   20,
				Offset: 60000, // 最大値50000を超える
			},
			mockSetup: func(m *mockvideo.MockVideoUsecase) {
				targetDate, _ := time.Parse("2006-01-02", "2024-01-01")
				m.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate, int32(20), int32(50000)). // 50000に制限される
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expectedVideos: []entity.Video{testVideo},
			expectedMD:     testMetadata,
			expectedError:  nil,
		},
		{
			name: "正常系：offsetの最小値制限",
			req: &pb.GetVideosByDateRequest{
				Date:   "2024-01-01",
				Hits:   20,
				Offset: -10, // 最小値0を下回る
			},
			mockSetup: func(m *mockvideo.MockVideoUsecase) {
				targetDate, _ := time.Parse("2006-01-02", "2024-01-01")
				m.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate, int32(20), int32(0)). // 0に制限される
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expectedVideos: []entity.Video{testVideo},
			expectedMD:     testMetadata,
			expectedError:  nil,
		},
		{
			name: "エラー系：不正な日付形式",
			req: &pb.GetVideosByDateRequest{
				Date:   "invalid-date",
				Hits:   20,
				Offset: 0,
			},
			mockSetup:      func(m *mockvideo.MockVideoUsecase) {},
			expectedVideos: nil,
			expectedMD:     nil,
			expectedError:  status.Error(codes.InvalidArgument, "不正な日付形式です"),
		},
		{
			name: "エラー系：ユースケースでエラー",
			req: &pb.GetVideosByDateRequest{
				Date:   "2024-01-01",
				Hits:   20,
				Offset: 0,
			},
			mockSetup: func(m *mockvideo.MockVideoUsecase) {
				targetDate, _ := time.Parse("2006-01-02", "2024-01-01")
				m.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate, int32(20), int32(0)).
					Return(nil, nil, errors.New("database error"))
			},
			expectedVideos: nil,
			expectedMD:     nil,
			expectedError:  status.Error(codes.Internal, "動画の取得に失敗しました: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
			tt.mockSetup(mockUsecase)

			req := connect.NewRequest(tt.req)
			ctx := context.WithValue(context.Background(), auth.SubKey, "test-user")
			resp, err := handler.GetVideosByDate(ctx, req)

			if tt.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)

			if len(tt.expectedVideos) > 0 {
				require.NotEmpty(t, resp.Msg.Videos)
				checkVideoFields(t, resp.Msg.Videos[0], tt.expectedVideos[0])
			} else {
				require.Empty(t, resp.Msg.Videos)
			}
			checkMetadata(t, resp.Msg.Metadata, tt.expectedMD)
		})
	}
}

func TestSearchVideos(t *testing.T) {
	ctx := context.WithValue(context.Background(), auth.SubKey, "test-user")

	tests := []struct {
		name        string
		request     *pb.SearchVideosRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    []entity.Video
		expectedMD  *entity.SearchMetadata
		expectError bool
		errorCode   codes.Code
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
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expected:    []entity.Video{testVideo},
			expectedMD:  testMetadata,
			expectError: false,
		},
		{
			name:    "異常系 - 検索エラー",
			request: &pb.SearchVideosRequest{},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					SearchVideos(gomock.Any(), "", "", "", "", "", "").
					Return(nil, nil, errors.New("search error"))
			},
			expected:    nil,
			expectedMD:  nil,
			expectError: true,
			errorCode:   codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
			tt.setupMock(mockUsecase)

			req := connect.NewRequest(tt.request)
			resp, err := handler.SearchVideos(ctx, req)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != 0 {
					s, ok := status.FromError(err)
					require.True(t, ok)
					require.Equal(t, tt.errorCode, s.Code())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)

			if len(tt.expected) > 0 {
				require.NotEmpty(t, resp.Msg.Videos)
				checkVideoFields(t, resp.Msg.Videos[0], tt.expected[0])
			} else {
				require.Empty(t, resp.Msg.Videos)
			}

			checkMetadata(t, resp.Msg.Metadata, tt.expectedMD)
		})
	}
}

func TestGetVideosByID(t *testing.T) {
	ctx := context.WithValue(context.Background(), auth.SubKey, "test-user")

	tests := []struct {
		name        string
		request     *pb.GetVideosByIDRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    []entity.Video
		expectedMD  *entity.SearchMetadata
		expectError bool
		errorCode   codes.Code
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
					GetVideosByID(
						gomock.Any(),
						[]string{"1"}, []string{"2"}, []string{"3"}, []string{"4"}, []string{"5"},
						int32(10), int32(0), "rank",
						"2023-01-01", "2023-12-31",
						"FANZA", "digital", "videoa",
					).
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expected:    []entity.Video{testVideo},
			expectedMD:  testMetadata,
			expectError: false,
		},
		{
			name:    "異常系 - 検索エラー",
			request: &pb.GetVideosByIDRequest{},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByID(
						gomock.Any(),
						gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
						int32(0), int32(0), "",
						"", "",
						"", "", "",
					).
					Return(nil, nil, errors.New("search error"))
			},
			expected:    nil,
			expectedMD:  nil,
			expectError: true,
			errorCode:   codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
			tt.setupMock(mockUsecase)

			req := connect.NewRequest(tt.request)
			resp, err := handler.GetVideosByID(ctx, req)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != 0 {
					s, ok := status.FromError(err)
					require.True(t, ok)
					require.Equal(t, tt.errorCode, s.Code())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)

			if len(tt.expected) > 0 {
				require.NotEmpty(t, resp.Msg.Videos)
				checkVideoFields(t, resp.Msg.Videos[0], tt.expected[0])
			} else {
				require.Empty(t, resp.Msg.Videos)
			}

			checkMetadata(t, resp.Msg.Metadata, tt.expectedMD)
		})
	}
}

func TestGetVideosByKeyword(t *testing.T) {
	ctx := context.WithValue(context.Background(), auth.SubKey, "test-user")

	tests := []struct {
		name        string
		request     *pb.GetVideosByKeywordRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    []entity.Video
		expectedMD  *entity.SearchMetadata
		expectError bool
		errorCode   codes.Code
	}{
		{
			name: "正常系",
			request: &pb.GetVideosByKeywordRequest{
				Keyword: "test",
				Hits:    10,
				Offset:  0,
				Sort:    "rank",
				GteDate: "2023-01-01",
				LteDate: "2023-12-31",
				Site:    "FANZA",
				Service: "digital",
				Floor:   "videoa",
			},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByKeyword(
						gomock.Any(),
						"test",
						int32(10), int32(0),
						"rank",
						"2023-01-01", "2023-12-31",
						"FANZA", "digital", "videoa",
					).
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			expected:    []entity.Video{testVideo},
			expectedMD:  testMetadata,
			expectError: false,
		},
		{
			name:    "異常系 - 検索エラー",
			request: &pb.GetVideosByKeywordRequest{},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByKeyword(
						gomock.Any(),
						"",
						int32(0), int32(0),
						"",
						"", "",
						"", "", "",
					).
					Return(nil, nil, errors.New("search error"))
			},
			expected:    nil,
			expectedMD:  nil,
			expectError: true,
			errorCode:   codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
			tt.setupMock(mockUsecase)

			req := connect.NewRequest(tt.request)
			resp, err := handler.GetVideosByKeyword(ctx, req)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != 0 {
					s, ok := status.FromError(err)
					require.True(t, ok)
					require.Equal(t, tt.errorCode, s.Code())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)

			if len(tt.expected) > 0 {
				require.NotEmpty(t, resp.Msg.Videos)
				checkVideoFields(t, resp.Msg.Videos[0], tt.expected[0])
			} else {
				require.Empty(t, resp.Msg.Videos)
			}

			checkMetadata(t, resp.Msg.Metadata, tt.expectedMD)
		})
	}
}

func TestGetVideoById(t *testing.T) {
	ctx := context.WithValue(context.Background(), auth.SubKey, "test-user")

	tests := []struct {
		name        string
		request     *pb.GetVideoByIdRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    *entity.Video
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:    "正常系",
			request: &pb.GetVideoByIdRequest{DmmId: "test123"},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideoById(gomock.Any(), "test123").
					Return(&testVideo, nil)
			},
			expected:    &testVideo,
			expectError: false,
		},
		{
			name:    "異常系 - 動画が見つからない",
			request: &pb.GetVideoByIdRequest{DmmId: "notfound"},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideoById(gomock.Any(), "notfound").
					Return(nil, nil)
			},
			expected:    nil,
			expectError: true,
			errorCode:   codes.NotFound,
		},
		{
			name:    "異常系 - ユースケースエラー",
			request: &pb.GetVideoByIdRequest{DmmId: "test123"},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideoById(gomock.Any(), "test123").
					Return(nil, errors.New("usecase error"))
			},
			expected:    nil,
			expectError: true,
			errorCode:   codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
			handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
			tt.setupMock(mockUsecase)

			req := connect.NewRequest(tt.request)
			resp, err := handler.GetVideoById(ctx, req)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != 0 {
					s, ok := status.FromError(err)
					require.True(t, ok)
					require.Equal(t, tt.errorCode, s.Code())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			require.NotNil(t, resp.Msg.Video)

			checkVideoFields(t, resp.Msg.Video, *tt.expected)
		})
	}
}
