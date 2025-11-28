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
	testTime  = time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	testVideo = entity.Video{
		DmmID:        "test123",
		Title:        "Test Video",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    testTime,
		Price:        1000,
		LikesCount:   500,
		Actresses:    []entity.Actress{{ID: "a1", Name: "Actress"}},
		Genres:       []entity.Genre{{ID: "g1", Name: "Genre"}},
		Makers:       []entity.Maker{{ID: "m1", Name: "Maker"}},
		Series:       []entity.Series{{ID: "s1", Name: "Series"}},
		Directors:    []entity.Director{{ID: "d1", Name: "Director"}},
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
	uc := NewVideoUsecase(mockRepo)
	require.NotNil(t, uc)
}

func TestGetVideosByDate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		input      GetVideosByDateInput
		now        func() time.Time
		expectCall func(m *mockrepo.MockVideoRepository)
		wantErr    error
		wantHits   int32
		wantOffset int32
		wantDate   time.Time
	}{
		{
			name: "指定日付で取得",
			input: GetVideosByDateInput{
				Date:   "2024-01-01",
				Hits:   20,
				Offset: 0,
			},
			expectCall: func(m *mockrepo.MockVideoRepository) {
				targetDate, _ := time.Parse("2006-01-02", "2024-01-01")
				m.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate, int32(20), int32(0)).
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			wantHits:   20,
			wantOffset: 0,
			wantDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "日付未指定は現在時刻",
			input: GetVideosByDateInput{
				Hits:   20,
				Offset: 10,
			},
			now: func() time.Time { return testTime },
			expectCall: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideosByDate(gomock.Any(), testTime, int32(20), int32(10)).
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			wantHits:   20,
			wantOffset: 10,
			wantDate:   testTime,
		},
		{
			name: "ヒット数は最大値で丸め",
			input: GetVideosByDateInput{
				Date:   "2024-01-01",
				Hits:   200,
				Offset: -5,
			},
			expectCall: func(m *mockrepo.MockVideoRepository) {
				targetDate, _ := time.Parse("2006-01-02", "2024-01-01")
				m.EXPECT().
					GetVideosByDate(gomock.Any(), targetDate, int32(100), int32(0)).
					Return([]entity.Video{testVideo}, testMetadata, nil)
			},
			wantHits:   100,
			wantOffset: 0,
			wantDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "日付フォーマットエラー",
			input:      GetVideosByDateInput{Date: "invalid"},
			expectCall: func(m *mockrepo.MockVideoRepository) {},
			wantErr:    ErrInvalidDateFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mockrepo.NewMockVideoRepository(ctrl)
			if tt.expectCall != nil {
				tt.expectCall(mockRepo)
			}

			uc := NewVideoUsecaseWithDeps(mockRepo, tt.now)
			output, err := uc.GetVideosByDate(ctx, tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, output)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)
			require.Equal(t, tt.wantHits, output.Hits)
			require.Equal(t, tt.wantOffset, output.Offset)
			require.Equal(t, tt.wantDate, output.TargetDate)
			require.Equal(t, []entity.Video{testVideo}, output.Videos)
			require.Equal(t, testMetadata, output.Metadata)
		})
	}
}

func TestGetVideoById(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          string
		setupMock   func(m *mockrepo.MockVideoRepository)
		expected    *entity.Video
		expectError error
	}{
		{
			name: "正常系",
			id:   "test123",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideoById(gomock.Any(), "test123").
					Return(&testVideo, nil)
			},
			expected: &testVideo,
		},
		{
			name: "異常系",
			id:   "notfound",
			setupMock: func(m *mockrepo.MockVideoRepository) {
				m.EXPECT().
					GetVideoById(gomock.Any(), "notfound").
					Return(nil, errors.New("repository error"))
			},
			expectError: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mockrepo.NewMockVideoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := NewVideoUsecase(mockRepo)
			video, err := uc.GetVideoById(ctx, tt.id)

			if tt.expectError != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, video)
		})
	}
}

func TestSearchVideos(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockVideoRepository(ctrl)
	mockRepo.EXPECT().
		SearchVideos(gomock.Any(), "keyword", "a", "g", "m", "s", "d").
		Return([]entity.Video{testVideo}, testMetadata, nil)

	uc := NewVideoUsecase(mockRepo)
	videos, metadata, err := uc.SearchVideos(ctx, "keyword", "a", "g", "m", "s", "d")
	require.NoError(t, err)
	require.Equal(t, []entity.Video{testVideo}, videos)
	require.Equal(t, testMetadata, metadata)
}

func TestGetVideosByID(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockVideoRepository(ctrl)
	mockRepo.EXPECT().
		GetVideosByID(
			gomock.Any(),
			[]string{"1"},
			[]string{"2"},
			[]string{"3"},
			[]string{"4"},
			[]string{"5"},
			int32(100),
			int32(0),
			"rank",
			"2023-01-01",
			"2023-12-31",
			"FANZA",
			"digital",
			"videoa",
		).
		Return([]entity.Video{testVideo}, testMetadata, nil)

	uc := NewVideoUsecase(mockRepo)
	output, err := uc.GetVideosByID(ctx, GetVideosByIDInput{
		ActressIDs:  []string{"1"},
		GenreIDs:    []string{"2"},
		MakerIDs:    []string{"3"},
		SeriesIDs:   []string{"4"},
		DirectorIDs: []string{"5"},
		Hits:        200, // clamp to 100
		Offset:      -10, // clamp to 0
		Sort:        "rank",
		GteDate:     "2023-01-01",
		LteDate:     "2023-12-31",
		Site:        "FANZA",
		Service:     "digital",
		Floor:       "videoa",
	})
	require.NoError(t, err)
	require.Equal(t, int32(100), output.Hits)
	require.Equal(t, int32(0), output.Offset)
	require.Equal(t, []entity.Video{testVideo}, output.Videos)
	require.Equal(t, testMetadata, output.Metadata)
}

func TestGetVideosByKeyword(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockVideoRepository(ctrl)
	mockRepo.EXPECT().
		GetVideosByKeyword(
			gomock.Any(),
			"keyword",
			int32(100),
			int32(0),
			"rank",
			"",
			"",
			"FANZA",
			"digital",
			"videoa",
		).
		Return([]entity.Video{testVideo}, testMetadata, nil)

	uc := NewVideoUsecase(mockRepo)
	output, err := uc.GetVideosByKeyword(ctx, GetVideosByKeywordInput{
		Keyword: "keyword",
		Hits:    150, // clamp to 100
		Offset:  -5,  // clamp to 0
		Sort:    "rank",
		Site:    "FANZA",
		Service: "digital",
		Floor:   "videoa",
	})
	require.NoError(t, err)
	require.Equal(t, int32(100), output.Hits)
	require.Equal(t, int32(0), output.Offset)
	require.Equal(t, []entity.Video{testVideo}, output.Videos)
	require.Equal(t, testMetadata, output.Metadata)
}
