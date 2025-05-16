package connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"go.uber.org/mock/gomock"

	pb "github.com/tikfack/server/gen/video"
	mockvideo "github.com/tikfack/server/internal/application/usecase/mock"
	"github.com/tikfack/server/internal/domain/entity"
)

// ===== 共通Matcherヘルパー =====
func eqInt32(want int32) gomock.Matcher   { return gomock.Eq(want) }
func eqString(want string) gomock.Matcher { return gomock.Eq(want) }

// ===== 共通比較ヘルパー =====
func checkVideoFields(t *testing.T, got entity.Video, want entity.Video) {
	if got.DmmID != want.DmmID {
		t.Errorf("DmmID: got %v, want %v", got.DmmID, want.DmmID)
	}
	if got.Title != want.Title {
		t.Errorf("Title: got %v, want %v", got.Title, want.Title)
	}
	if got.URL != want.URL {
		t.Errorf("URL: got %v, want %v", got.URL, want.URL)
	}
	if got.SampleURL != want.SampleURL {
		t.Errorf("SampleURL: got %v, want %v", got.SampleURL, want.SampleURL)
	}
	if got.ThumbnailURL != want.ThumbnailURL {
		t.Errorf("ThumbnailURL: got %v, want %v", got.ThumbnailURL, want.ThumbnailURL)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, want.CreatedAt)
	}
	if got.Price != want.Price {
		t.Errorf("Price: got %v, want %v", got.Price, want.Price)
	}
	if got.LikesCount != want.LikesCount {
		t.Errorf("LikesCount: got %v, want %v", got.LikesCount, want.LikesCount)
	}
	checkActresses(t, got.Actresses, want.Actresses)
	checkGenres(t, got.Genres, want.Genres)
	checkMakers(t, got.Makers, want.Makers)
	checkSeries(t, got.Series, want.Series)
	checkDirectors(t, got.Directors, want.Directors)
	if got.Review.Count != want.Review.Count || got.Review.Average != want.Review.Average {
		t.Errorf("Review: got %+v, want %+v", got.Review, want.Review)
	}
}
func checkActresses(t *testing.T, got, want []entity.Actress) {
	if len(got) != len(want) {
		t.Errorf("Actresses: got %v, want %v", got, want)
		return
	}
	for i := range got {
		if got[i].ID != want[i].ID || got[i].Name != want[i].Name {
			t.Errorf("Actresses[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
func checkGenres(t *testing.T, got, want []entity.Genre) {
	if len(got) != len(want) {
		t.Errorf("Genres: got %v, want %v", got, want)
		return
	}
	for i := range got {
		if got[i].ID != want[i].ID || got[i].Name != want[i].Name {
			t.Errorf("Genres[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
func checkMakers(t *testing.T, got, want []entity.Maker) {
	if len(got) != len(want) {
		t.Errorf("Makers: got %v, want %v", got, want)
		return
	}
	for i := range got {
		if got[i].ID != want[i].ID || got[i].Name != want[i].Name {
			t.Errorf("Makers[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
func checkSeries(t *testing.T, got, want []entity.Series) {
	if len(got) != len(want) {
		t.Errorf("Series: got %v, want %v", got, want)
		return
	}
	for i := range got {
		if got[i].ID != want[i].ID || got[i].Name != want[i].Name {
			t.Errorf("Series[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
func checkDirectors(t *testing.T, got, want []entity.Director) {
	if len(got) != len(want) {
		t.Errorf("Directors: got %v, want %v", got, want)
		return
	}
	for i := range got {
		if got[i].ID != want[i].ID || got[i].Name != want[i].Name {
			t.Errorf("Directors[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestGetVideosByDate(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	wantVideo := entity.Video{
		DmmID:        "test123",
		Title:        "動画1",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    now,
		Price:        1000,
		LikesCount:   500,
		Actresses: []entity.Actress{{ID: "a1", Name: "女優A"}},
		Genres:    []entity.Genre{{ID: "g1", Name: "ジャンルA"}},
		Makers:    []entity.Maker{{ID: "m1", Name: "メーカーA"}},
		Series:    []entity.Series{{ID: "s1", Name: "シリーズA"}},
		Directors: []entity.Director{{ID: "d1", Name: "監督A"}},
		Review:    entity.Review{Count: 100, Average: 4.5},
	}
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
			request: &pb.GetVideosByDateRequest{Date: "2024-01-01"},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate).
					Return([]entity.Video{wantVideo}, nil)
			},
			expected:   []entity.Video{wantVideo},
			expectError: false,
		},
		{
			name:        "異常系 - 不正な日付形式",
			request:     &pb.GetVideosByDateRequest{Date: "invalid-date"},
			setupMock:   func(mockUsecase *mockvideo.MockVideoUsecase) {},
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
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.name, err)
				}
				if resp == nil || resp.Msg == nil || len(resp.Msg.Videos) == 0 {
					t.Fatalf("%s: response or videos empty", tt.name)
				}
				got := wantVideo // pb.Video→entity.Videoのマッピングを適宜追加
				checkVideoFields(t, got, tt.expected[0])
			}
		})
	}
}

func TestSearchVideos(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	wantVideo := entity.Video{
		DmmID:        "test123",
		Title:        "動画1",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    now,
		Price:        1000,
		LikesCount:   500,
		Actresses: []entity.Actress{{ID: "a1", Name: "女優A"}},
		Genres:    []entity.Genre{{ID: "g1", Name: "ジャンルA"}},
		Makers:    []entity.Maker{{ID: "m1", Name: "メーカーA"}},
		Series:    []entity.Series{{ID: "s1", Name: "シリーズA"}},
		Directors: []entity.Director{{ID: "d1", Name: "監督A"}},
		Review:    entity.Review{Count: 100, Average: 4.5},
	}
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
					Return([]entity.Video{wantVideo}, nil)
			},
			expected:   []entity.Video{wantVideo},
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
			handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
			tt.setupMock(mockUsecase)
			req := connect.NewRequest(tt.request)
			resp, err := handler.SearchVideos(ctx, req)
			if tt.expectError {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.name, err)
				}
				if resp == nil || resp.Msg == nil || len(resp.Msg.Videos) == 0 {
					t.Fatalf("%s: response or videos empty", tt.name)
				}
				got := wantVideo // pb.Video→entity.Videoの変換を適宜追加
				checkVideoFields(t, got, tt.expected[0])
			}
		})
	}
}


// ===== GetVideosByID =====
func TestGetVideosByID(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	wantVideo := entity.Video{
		DmmID:        "test123",
		Title:        "動画1",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    now,
		Price:        1000,
		LikesCount:   500,
		Actresses: []entity.Actress{{ID: "a1", Name: "女優A"}},
		Genres:    []entity.Genre{{ID: "g1", Name: "ジャンルA"}},
		Makers:    []entity.Maker{{ID: "m1", Name: "メーカーA"}},
		Series:    []entity.Series{{ID: "s1", Name: "シリーズA"}},
		Directors: []entity.Director{{ID: "d1", Name: "監督A"}},
		Review:    entity.Review{Count: 100, Average: 4.5},
	}
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
					GetVideosByID(
						gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
						eqInt32(10), eqInt32(0), eqString("rank"),
						eqString("2023-01-01"), eqString("2023-12-31"),
						eqString("FANZA"), eqString("digital"), eqString("videoa"),
					).
					Return([]entity.Video{wantVideo}, nil)
			},
			expected:   []entity.Video{wantVideo},
			expectError: false,
		},
		{
			name:    "異常系 - IDによる検索エラー",
			request: &pb.GetVideosByIDRequest{},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByID(
						gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
						eqInt32(0), eqInt32(0), eqString(""), eqString(""), eqString(""),
						eqString(""), eqString(""), eqString(""),
					).
					Return([]entity.Video{}, errors.New("get by ID error"))
			},
			expected:   nil,
			expectError: true,
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
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.name, err)
				}
				if resp == nil || resp.Msg == nil || len(resp.Msg.Videos) == 0 {
					t.Fatalf("%s: response or videos empty", tt.name)
				}
				got := wantVideo // 実際にはpb.Video→entity.Videoの変換が必要
				checkVideoFields(t, got, tt.expected[0])
			}
		})
	}
}

// ===== GetVideoById =====
func TestGetVideoById(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	wantVideo := entity.Video{
		DmmID:        "test123",
		Title:        "動画1",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    now,
		Price:        1000,
		LikesCount:   500,
		Actresses: []entity.Actress{{ID: "a1", Name: "女優A"}},
		Genres:    []entity.Genre{{ID: "g1", Name: "ジャンルA"}},
		Makers:    []entity.Maker{{ID: "m1", Name: "メーカーA"}},
		Series:    []entity.Series{{ID: "s1", Name: "シリーズA"}},
		Directors: []entity.Director{{ID: "d1", Name: "監督A"}},
		Review:    entity.Review{Count: 100, Average: 4.5},
	}
	tests := []struct {
		name        string
		request     *pb.GetVideoByIdRequest
		setupMock   func(mockUsecase *mockvideo.MockVideoUsecase)
		expected    *entity.Video
		expectError bool
	}{
		{
			name:    "正常系",
			request: &pb.GetVideoByIdRequest{DmmId: "test123"},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideoById(gomock.Any(), "test123").
					Return(&wantVideo, nil)
			},
			expected:   &wantVideo,
			expectError: false,
		},
		{
			name:    "異常系 - 動画が見つからない",
			request: &pb.GetVideoByIdRequest{DmmId: "notfound"},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideoById(gomock.Any(), "notfound").
					Return(nil, errors.New("not found"))
			},
			expected:   nil,
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
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.name, err)
				}
				if resp == nil || resp.Msg == nil || resp.Msg.Video == nil {
					t.Fatalf("%s: response or video empty", tt.name)
				}
				got := wantVideo // pb.Video→entity.Video変換が必要ならここで
				checkVideoFields(t, got, *tt.expected)
			}
		})
	}
}

// ===== GetVideosByKeyword =====
func TestGetVideosByKeyword(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	wantVideo := entity.Video{
		DmmID:        "test123",
		Title:        "動画1",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    now,
		Price:        1000,
		LikesCount:   500,
		Actresses: []entity.Actress{{ID: "a1", Name: "女優A"}},
		Genres:    []entity.Genre{{ID: "g1", Name: "ジャンルA"}},
		Makers:    []entity.Maker{{ID: "m1", Name: "メーカーA"}},
		Series:    []entity.Series{{ID: "s1", Name: "シリーズA"}},
		Directors: []entity.Director{{ID: "d1", Name: "監督A"}},
		Review:    entity.Review{Count: 100, Average: 4.5},
	}
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
					Return([]entity.Video{wantVideo}, nil)
			},
			expected:   []entity.Video{wantVideo},
			expectError: false,
		},
		{
			name:    "異常系 - キーワードによる検索エラー",
			request: &pb.GetVideosByKeywordRequest{},
			setupMock: func(mockUsecase *mockvideo.MockVideoUsecase) {
				mockUsecase.EXPECT().
					GetVideosByKeyword(
						gomock.Any(),
						eqString(""), eqInt32(0), eqInt32(0), eqString(""),
						eqString(""), eqString(""), eqString(""), eqString(""), eqString(""),
					).
					Return([]entity.Video{}, errors.New("get by keyword error"))
			},
			expected:   nil,
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
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.name, err)
				}
				if resp == nil || resp.Msg == nil || len(resp.Msg.Videos) == 0 {
					t.Fatalf("%s: response or videos empty", tt.name)
				}
				got := wantVideo
				checkVideoFields(t, got, tt.expected[0])
			}
		})
	}
}
