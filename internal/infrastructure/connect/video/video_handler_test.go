package connect

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/require"
	pb "github.com/tikfack/server/gen/video"
	mockvideo "github.com/tikfack/server/internal/application/usecase/mock"
	video "github.com/tikfack/server/internal/application/usecase/video"
	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	handlerTestTime  = time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	handlerTestVideo = entity.Video{
		DmmID:        "test123",
		Title:        "動画1",
		URL:          "https://example.com",
		SampleURL:    "https://example.com/sample",
		ThumbnailURL: "https://example.com/thumb.jpg",
		CreatedAt:    handlerTestTime,
		Price:        1000,
		LikesCount:   500,
		Actresses:    []entity.Actress{{ID: "a1", Name: "女優A"}},
		Genres:       []entity.Genre{{ID: "g1", Name: "ジャンルA"}},
		Makers:       []entity.Maker{{ID: "m1", Name: "メーカーA"}},
		Series:       []entity.Series{{ID: "s1", Name: "シリーズA"}},
		Directors:    []entity.Director{{ID: "d1", Name: "監督A"}},
		Review:       entity.Review{Count: 100, Average: 4.5},
	}
	handlerTestMetadata = &entity.SearchMetadata{
		ResultCount:   10,
		TotalCount:    100,
		FirstPosition: 1,
	}
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func stubHTTPHead(t *testing.T, status int) func() {
	t.Helper()
	originalClient := http.DefaultClient
	http.DefaultClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: status,
				Header:     make(http.Header),
				Body:       http.NoBody,
			}, nil
		}),
	}
	return func() { http.DefaultClient = originalClient }
}

func TestGetVideosByDate_Success(t *testing.T) {
	defer stubHTTPHead(t, http.StatusNotFound)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
	mockUsecase.EXPECT().
		GetVideosByDate(gomock.Any(), video.GetVideosByDateInput{Date: "2024-01-01", Hits: 20, Offset: 5}).
		Return(&video.GetVideosByDateOutput{
			Videos:     []entity.Video{handlerTestVideo},
			Metadata:   handlerTestMetadata,
			TargetDate: handlerTestTime,
			Hits:       20,
			Offset:     5,
		}, nil)

	handler := NewVideoServiceHandlerWithUsecase(mockUsecase)

	req := connect.NewRequest(&pb.GetVideosByDateRequest{Date: "2024-01-01", Hits: 20, Offset: 5})
	resp, err := handler.GetVideosByDate(withCtxValues(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Msg.Videos, 1)
	require.Equal(t, handlerTestVideo.DmmID, resp.Msg.Videos[0].DmmId)
	require.NotNil(t, resp.Msg.Metadata)
	require.Equal(t, int32(handlerTestMetadata.ResultCount), resp.Msg.Metadata.ResultCount)
}

func TestGetVideosByDate_InvalidDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
	mockUsecase.EXPECT().
		GetVideosByDate(gomock.Any(), video.GetVideosByDateInput{Date: "invalid", Hits: 0, Offset: 0}).
		Return(nil, video.ErrInvalidDateFormat)

	handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
	req := connect.NewRequest(&pb.GetVideosByDateRequest{Date: "invalid"})

	_, err := handler.GetVideosByDate(withCtxValues(), req)
	require.Error(t, err)
	s, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, s.Code())
}

func TestGetVideosByDate_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
	mockUsecase.EXPECT().
		GetVideosByDate(gomock.Any(), video.GetVideosByDateInput{Date: "2024-01-01", Hits: 10, Offset: 0}).
		Return(nil, errors.New("repository error"))

	handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
	req := connect.NewRequest(&pb.GetVideosByDateRequest{Date: "2024-01-01", Hits: 10})

	_, err := handler.GetVideosByDate(withCtxValues(), req)
	require.Error(t, err)
	s, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, s.Code())
}

func TestGetVideoById_Success(t *testing.T) {
	defer stubHTTPHead(t, http.StatusNotFound)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
	mockUsecase.EXPECT().
		GetVideoById(gomock.Any(), "test123").
		Return(&handlerTestVideo, nil)

	handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
	req := connect.NewRequest(&pb.GetVideoByIdRequest{DmmId: "test123"})

	resp, err := handler.GetVideoById(withCtxValues(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, handlerTestVideo.DmmID, resp.Msg.Video.DmmId)
}

func TestSearchVideos_Success(t *testing.T) {
	defer stubHTTPHead(t, http.StatusNotFound)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
	mockUsecase.EXPECT().
		SearchVideos(gomock.Any(), "keyword", "1", "2", "3", "4", "5").
		Return([]entity.Video{handlerTestVideo}, handlerTestMetadata, nil)

	handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
	req := connect.NewRequest(&pb.SearchVideosRequest{
		Keyword:    "keyword",
		ActressId:  "1",
		GenreId:    "2",
		MakerId:    "3",
		SeriesId:   "4",
		DirectorId: "5",
	})

	resp, err := handler.SearchVideos(withCtxValues(), req)
	require.NoError(t, err)
	require.Len(t, resp.Msg.Videos, 1)
}

func TestGetVideosByID_Success(t *testing.T) {
	defer stubHTTPHead(t, http.StatusNotFound)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
	mockUsecase.EXPECT().
		GetVideosByID(gomock.Any(), video.GetVideosByIDInput{
			ActressIDs:  []string{"1"},
			GenreIDs:    []string{"2"},
			MakerIDs:    []string{"3"},
			SeriesIDs:   []string{"4"},
			DirectorIDs: []string{"5"},
			Hits:        10,
			Offset:      20,
			Sort:        "rank",
			GteDate:     "2023-01-01",
			LteDate:     "2023-12-31",
			Site:        "FANZA",
			Service:     "digital",
			Floor:       "videoa",
		}).
		Return(&video.GetVideosOutput{
			Videos:   []entity.Video{handlerTestVideo},
			Metadata: handlerTestMetadata,
			Hits:     10,
			Offset:   20,
		}, nil)

	handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
	req := connect.NewRequest(&pb.GetVideosByIDRequest{
		ActressId:  []string{"1"},
		GenreId:    []string{"2"},
		MakerId:    []string{"3"},
		SeriesId:   []string{"4"},
		DirectorId: []string{"5"},
		Hits:       10,
		Offset:     20,
		Sort:       "rank",
		GteDate:    "2023-01-01",
		LteDate:    "2023-12-31",
		Site:       "FANZA",
		Service:    "digital",
		Floor:      "videoa",
	})

	resp, err := handler.GetVideosByID(withCtxValues(), req)
	require.NoError(t, err)
	require.Len(t, resp.Msg.Videos, 1)
}

func TestGetVideosByKeyword_Success(t *testing.T) {
	defer stubHTTPHead(t, http.StatusNotFound)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mockvideo.NewMockVideoUsecase(ctrl)
	mockUsecase.EXPECT().
		GetVideosByKeyword(gomock.Any(), video.GetVideosByKeywordInput{
			Keyword: "keyword",
			Hits:    15,
			Offset:  5,
			Sort:    "rank",
			Site:    "FANZA",
			Service: "digital",
			Floor:   "videoa",
		}).
		Return(&video.GetVideosOutput{
			Videos:   []entity.Video{handlerTestVideo},
			Metadata: handlerTestMetadata,
			Hits:     15,
			Offset:   5,
		}, nil)

	handler := NewVideoServiceHandlerWithUsecase(mockUsecase)
	req := connect.NewRequest(&pb.GetVideosByKeywordRequest{
		Keyword: "keyword",
		Hits:    15,
		Offset:  5,
		Sort:    "rank",
		Site:    "FANZA",
		Service: "digital",
		Floor:   "videoa",
	})

	resp, err := handler.GetVideosByKeyword(withCtxValues(), req)
	require.NoError(t, err)
	require.Len(t, resp.Msg.Videos, 1)
}

func withCtxValues() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.SubKey, "user")
	ctx = context.WithValue(ctx, ctxkeys.TraceIDKey, "trace")
	ctx = context.WithValue(ctx, ctxkeys.TokenKey, "tokentoken")
	return ctx
}
