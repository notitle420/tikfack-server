package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tikfack/server/internal/application/model"
	mockcatalog "github.com/tikfack/server/internal/application/port/mock"
	"go.uber.org/mock/gomock"
)

var (
	testTime  = time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	testVideo = model.Video{
		DmmID:        "test123",
		Title:        "動画1",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    testTime,
		Price:        1000,
		LikesCount:   500,
		Actresses:    []model.Actress{{ID: "a1", Name: "女優A"}},
		Genres:       []model.Genre{{ID: "g1", Name: "ジャンルA"}},
		Makers:       []model.Maker{{ID: "m1", Name: "メーカーA"}},
		Series:       []model.Series{{ID: "s1", Name: "シリーズA"}},
		Directors:    []model.Director{{ID: "d1", Name: "監督A"}},
		Review:       model.Review{Count: 100, Average: 4.5},
	}
	testMetadata = &model.SearchMetadata{ResultCount: 10, TotalCount: 100, FirstPosition: 1}
)

func TestNewVideoUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCatalog := mockcatalog.NewMockVideoCatalog(ctrl)
	uc := NewVideoUsecase(mockCatalog)
	require.NotNil(t, uc)
}

func TestGetVideosByDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	catalog := mockcatalog.NewMockVideoCatalog(ctrl)
	uc := NewVideoUsecase(catalog)

	cases := []struct {
		name         string
		hits         int32
		offset       int32
		expectHits   int32
		expectOffset int32
		expectErr    error
	}{
		{
			name:         "normal request",
			hits:         10,
			offset:       0,
			expectHits:   10,
			expectOffset: 0,
		},
		{
			name:         "clamp hits and offset",
			hits:         500,
			offset:       -5,
			expectHits:   maxHits,
			expectOffset: 0,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			catalog.EXPECT().
				GetVideosByDate(gomock.Any(), testTime, tt.expectHits, tt.expectOffset).
				Return([]model.Video{testVideo}, testMetadata, tt.expectErr)

			videos, md, err := uc.GetVideosByDate(context.Background(), testTime, tt.hits, tt.offset)
			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
				require.Nil(t, videos)
				require.Nil(t, md)
				return
			}

			require.NoError(t, err)
			require.Equal(t, []model.Video{testVideo}, videos)
			require.Equal(t, testMetadata, md)
		})
	}
}

func TestGetVideoById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	catalog := mockcatalog.NewMockVideoCatalog(ctrl)
	uc := NewVideoUsecase(catalog)

	catalog.EXPECT().
		GetVideoById(gomock.Any(), "abc123").
		Return(&testVideo, nil)

	video, err := uc.GetVideoById(context.Background(), "abc123")
	require.NoError(t, err)
	require.Equal(t, &testVideo, video)
}

func TestSearchVideos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	catalog := mockcatalog.NewMockVideoCatalog(ctrl)
	uc := NewVideoUsecase(catalog)

	catalog.EXPECT().
		SearchVideos(gomock.Any(), "keyword", "a", "g", "m", "s", "d").
		Return([]model.Video{testVideo}, testMetadata, nil)

	videos, md, err := uc.SearchVideos(context.Background(), "keyword", "a", "g", "m", "s", "d")
	require.NoError(t, err)
	require.Equal(t, []model.Video{testVideo}, videos)
	require.Equal(t, testMetadata, md)
}

func TestGetVideosByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	catalog := mockcatalog.NewMockVideoCatalog(ctrl)
	uc := NewVideoUsecase(catalog)

	actress := []string{"a1"}
	genre := []string{"g1"}
	maker := []string{"m1"}
	series := []string{"s1"}
	director := []string{"d1"}

	catalog.EXPECT().
		GetVideosByID(gomock.Any(), actress, genre, maker, series, director, maxHits, int32(0), "popular", "", "", "", "", "").
		Return([]model.Video{testVideo}, testMetadata, nil)

	videos, md, err := uc.GetVideosByID(context.Background(), actress, genre, maker, series, director, 1000, -10, "popular", "", "", "", "", "")
	require.NoError(t, err)
	require.Equal(t, []model.Video{testVideo}, videos)
	require.Equal(t, testMetadata, md)
}

func TestGetVideosByKeyword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	catalog := mockcatalog.NewMockVideoCatalog(ctrl)
	uc := NewVideoUsecase(catalog)

	catalog.EXPECT().
		GetVideosByKeyword(gomock.Any(), "hello", int32(50), int32(20), "date", "2024-01-01", "", "FANZA", "digital", "videoa").
		Return([]model.Video{testVideo}, testMetadata, nil)

	videos, md, err := uc.GetVideosByKeyword(context.Background(), "hello", 50, 20, "date", "2024-01-01", "", "FANZA", "digital", "videoa")
	require.NoError(t, err)
	require.Equal(t, []model.Video{testVideo}, videos)
	require.Equal(t, testMetadata, md)
}

func TestGetVideosByKeywordClamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	catalog := mockcatalog.NewMockVideoCatalog(ctrl)
	uc := NewVideoUsecase(catalog)

	catalog.EXPECT().
		GetVideosByKeyword(gomock.Any(), "hello", maxHits, int32(0), "", "", "", "", "", "").
		Return(nil, nil, errors.New("boom"))

	videos, md, err := uc.GetVideosByKeyword(context.Background(), "hello", 200, -30, "", "", "", "", "", "")
	require.Nil(t, videos)
	require.Nil(t, md)
	require.Error(t, err)
}
