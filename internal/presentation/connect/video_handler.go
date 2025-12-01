package connect

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tikfack/server/gen/video"
	videoconnect "github.com/tikfack/server/gen/video/videoconnect"

	video "github.com/tikfack/server/internal/application/usecase/video"
	"github.com/tikfack/server/internal/middleware/logger"
)

// VideoServiceServer は Connect のサーバー実装です。
type VideoServiceServer struct {
	videoUsecase video.VideoUsecase
	presenter    videoPresenter
	logger       *slog.Logger
	handlerOpts  []connect.HandlerOption
}

// NewVideoServiceHandler はユースケースを受け取り Connect ハンドラを構築する。
func NewVideoServiceHandler(vu video.VideoUsecase, opts ...connect.HandlerOption) *VideoServiceServer {
	return newVideoServiceServer(vu, nil, opts...)
}

func newVideoServiceServer(vu video.VideoUsecase, presenter videoPresenter, opts ...connect.HandlerOption) *VideoServiceServer {
	if vu == nil {
		panic("video usecase must be provided")
	}
	if presenter == nil {
		presenter = newVideoPresenter()
	}
	return &VideoServiceServer{
		videoUsecase: vu,
		presenter:    presenter,
		logger:       slog.Default().With(slog.String("component", "video_handler")),
		handlerOpts:  append([]connect.HandlerOption{connect.WithCompressMinBytes(0)}, opts...),
	}
}

// GetHandler は、Connect サービスのパターンとハンドラーを返します。
func (s *VideoServiceServer) GetHandler() (string, http.Handler) {
	pattern, handler := videoconnect.NewVideoServiceHandler(s, s.handlerOpts...)
	return pattern, handler
}

func (s *VideoServiceServer) loggerWithCtx(ctx context.Context) *slog.Logger {
	return s.logger.With(
		slog.String("user_id", logger.UserIDFromContext(ctx)),
		slog.String("trace_id", logger.TraceIDFromContext(ctx)),
		slog.String("token_id", logger.TokenIDFromContext(ctx)),
	)
}

// GetVideosByDate は、動画一覧を取得するエンドポイントの実装。
func (s *VideoServiceServer) GetVideosByDate(ctx context.Context, req *connect.Request[pb.GetVideosByDateRequest]) (*connect.Response[pb.GetVideosByDateResponse], error) {
	logger := s.loggerWithCtx(ctx)
	logger.Debug("API: GetVideosByDate", "date", req.Msg.Date, "hits", req.Msg.Hits, "offset", req.Msg.Offset)

	targetDate, err := parseDate(req.Msg.Date)
	if err != nil {
		logger.Error("invalid date supplied", "date", req.Msg.Date, "error", err)
		return nil, status.Error(codes.InvalidArgument, "不正な日付形式です")
	}

	hits := clampHits(req.Msg.Hits)
	offset := clampOffset(req.Msg.Offset)

	videos, metadata, err := s.videoUsecase.GetVideosByDate(ctx, targetDate, hits, offset)
	if err != nil {
		logger.Error("動画の取得に失敗", "date", targetDate.Format("2006-01-02"), "hits", hits, "offset", offset, "error", err)
		return nil, status.Errorf(codes.Internal, "動画の取得に失敗しました: %v", err)
	}

	pbVideos := s.presenter.Videos(ctx, videos)
	pbMetadata := s.presenter.Metadata(metadata)
	logger.Debug("GetVideosByDate completed", "count", len(pbVideos), "hits", hits, "offset", offset)
	res := &pb.GetVideosByDateResponse{Videos: pbVideos, Metadata: pbMetadata}
	return connect.NewResponse(res), nil
}

// GetVideoById は、ID で動画を取得するエンドポイントの実装。
func (s *VideoServiceServer) GetVideoById(ctx context.Context, req *connect.Request[pb.GetVideoByIdRequest]) (*connect.Response[pb.GetVideoByIdResponse], error) {
	logger := s.loggerWithCtx(ctx)
	logger.Debug("API: GetVideoById", "dmmId", req.Msg.DmmId)

	video, err := s.videoUsecase.GetVideoById(ctx, req.Msg.DmmId)
	if err != nil {
		logger.Error("動画の取得に失敗", "dmmId", req.Msg.DmmId, "error", err)
		return nil, status.Errorf(codes.Internal, "動画の取得に失敗しました: %v", err)
	}
	if video == nil {
		logger.Info("動画が見つかりません", "dmmId", req.Msg.DmmId)
		return nil, status.Error(codes.NotFound, "video not found")
	}

	logger.Debug("GetVideoById completed", "dmmId", req.Msg.DmmId, "title", video.Title)
	res := &pb.GetVideoByIdResponse{Video: s.presenter.Video(ctx, video)}
	return connect.NewResponse(res), nil
}

// SearchVideos は、動画を検索するエンドポイントの実装。
func (s *VideoServiceServer) SearchVideos(ctx context.Context, req *connect.Request[pb.SearchVideosRequest]) (*connect.Response[pb.SearchVideosResponse], error) {
	logger := s.loggerWithCtx(ctx)
	logger.Debug(
		"API: SearchVideos",
		"keyword", req.Msg.Keyword,
		"actressId", req.Msg.ActressId,
		"genreId", req.Msg.GenreId,
		"makerId", req.Msg.MakerId,
		"seriesId", req.Msg.SeriesId,
		"directorId", req.Msg.DirectorId,
	)

	videos, metadata, err := s.videoUsecase.SearchVideos(ctx,
		req.Msg.Keyword,
		req.Msg.ActressId,
		req.Msg.GenreId,
		req.Msg.MakerId,
		req.Msg.SeriesId,
		req.Msg.DirectorId,
	)
	if err != nil {
		logger.Error("動画の検索に失敗", "keyword", req.Msg.Keyword, "error", err)
		return nil, status.Errorf(codes.Internal, "動画の検索に失敗しました: %v", err)
	}

	pbVideos := s.presenter.Videos(ctx, videos)
	pbMetadata := s.presenter.Metadata(metadata)
	logger.Debug("SearchVideos completed", "count", len(pbVideos))
	return connect.NewResponse(&pb.SearchVideosResponse{Videos: pbVideos, Metadata: pbMetadata}), nil
}

// GetVideosByID は、複数IDで動画を検索するエンドポイント。
func (s *VideoServiceServer) GetVideosByID(ctx context.Context, req *connect.Request[pb.GetVideosByIDRequest]) (*connect.Response[pb.GetVideosByIDResponse], error) {
	logger := s.loggerWithCtx(ctx)
	logger.Debug(
		"API: GetVideosByID",
		"actressId_count", len(req.Msg.ActressId),
		"genreId_count", len(req.Msg.GenreId),
		"makerId_count", len(req.Msg.MakerId),
		"seriesId_count", len(req.Msg.SeriesId),
		"directorId_count", len(req.Msg.DirectorId),
		"hits", req.Msg.Hits,
		"offset", req.Msg.Offset,
	)

	hits := clampHits(req.Msg.Hits)
	offset := clampOffset(req.Msg.Offset)

	videos, metadata, err := s.videoUsecase.GetVideosByID(ctx,
		req.Msg.ActressId,
		req.Msg.GenreId,
		req.Msg.MakerId,
		req.Msg.SeriesId,
		req.Msg.DirectorId,
		hits,
		offset,
		req.Msg.Sort,
		req.Msg.GteDate,
		req.Msg.LteDate,
		req.Msg.Site,
		req.Msg.Service,
		req.Msg.Floor,
	)
	if err != nil {
		logger.Error("動画の検索に失敗", "error", err)
		return nil, status.Errorf(codes.Internal, "動画の検索に失敗しました: %v", err)
	}

	pbVideos := s.presenter.Videos(ctx, videos)
	pbMetadata := s.presenter.Metadata(metadata)
	logger.Debug("GetVideosByID completed", "count", len(pbVideos))
	return connect.NewResponse(&pb.GetVideosByIDResponse{Videos: pbVideos, Metadata: pbMetadata}), nil
}

// GetVideosByKeyword は、キーワードで動画を検索するエンドポイント。
func (s *VideoServiceServer) GetVideosByKeyword(ctx context.Context, req *connect.Request[pb.GetVideosByKeywordRequest]) (*connect.Response[pb.GetVideosByKeywordResponse], error) {
	logger := s.loggerWithCtx(ctx)
	logger.Debug("API: GetVideosByKeyword", "keyword", req.Msg.Keyword, "hits", req.Msg.Hits, "offset", req.Msg.Offset, "sort", req.Msg.Sort)

	hits := clampHits(req.Msg.Hits)
	offset := clampOffset(req.Msg.Offset)

	videos, metadata, err := s.videoUsecase.GetVideosByKeyword(ctx,
		req.Msg.Keyword,
		hits,
		offset,
		req.Msg.Sort,
		req.Msg.GteDate,
		req.Msg.LteDate,
		req.Msg.Site,
		req.Msg.Service,
		req.Msg.Floor,
	)
	if err != nil {
		logger.Error("動画の検索に失敗", "keyword", req.Msg.Keyword, "error", err)
		return nil, status.Errorf(codes.Internal, "動画の検索に失敗しました: %v", err)
	}

	pbVideos := s.presenter.Videos(ctx, videos)
	pbMetadata := s.presenter.Metadata(metadata)
	logger.Debug("GetVideosByKeyword completed", "count", len(pbVideos))
	return connect.NewResponse(&pb.GetVideosByKeywordResponse{Videos: pbVideos, Metadata: pbMetadata}), nil
}

func parseDate(date string) (time.Time, error) {
	if date == "" {
		return time.Now(), nil
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func clampHits(hits int32) int32 {
	if hits > 100 {
		return 100
	}
	return hits
}

func clampOffset(offset int32) int32 {
	if offset < 0 {
		return 0
	}
	if offset > 50000 {
		return 50000
	}
	return offset
}
