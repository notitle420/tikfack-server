package connect

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/bufbuild/connect-go"
	pb "github.com/tikfack/server/gen/favorite"
	favoriteconnect "github.com/tikfack/server/gen/favorite/favoriteconnect"
	"github.com/tikfack/server/internal/application/usecase/favorite"
	"github.com/tikfack/server/internal/domain/repository"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
	"github.com/tikfack/server/internal/middleware/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FavoriteServiceServer is the Connect handler implementing FavoriteService.
type FavoriteServiceServer struct {
	usecase     favorite.FavoriteUsecase
	presenter   favoritePresenter
	logger      *slog.Logger
	handlerOpts []connect.HandlerOption
}

// NewFavoriteServiceHandler constructs a new handler.
func NewFavoriteServiceHandler(uc favorite.FavoriteUsecase, opts ...connect.HandlerOption) *FavoriteServiceServer {
	if uc == nil {
		panic("favorite usecase must be provided")
	}
	return &FavoriteServiceServer{
		usecase:     uc,
		presenter:   newFavoritePresenter(),
		logger:      slog.Default().With(slog.String("component", "favorite_handler")),
		handlerOpts: append([]connect.HandlerOption{connect.WithCompressMinBytes(0)}, opts...),
	}
}

// GetHandler exposes the Connect handler pair.
func (s *FavoriteServiceServer) GetHandler() (string, http.Handler) {
	pattern, handler := favoriteconnect.NewFavoriteServiceHandler(s, s.handlerOpts...)
	return pattern, handler
}

func (s *FavoriteServiceServer) loggerWithCtx(ctx context.Context) *slog.Logger {
	return s.logger.With(
		slog.String("user_id", logger.UserIDFromContext(ctx)),
		slog.String("trace_id", logger.TraceIDFromContext(ctx)),
		slog.String("token_id", logger.TokenIDFromContext(ctx)),
	)
}

func (s *FavoriteServiceServer) AddFavoriteVideo(ctx context.Context, req *connect.Request[pb.AddFavoriteVideoRequest]) (*connect.Response[pb.AddFavoriteVideoResponse], error) {
	log := s.loggerWithCtx(ctx)
	userID := ctxkeys.UserIDFromContext(ctx)
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user id missing in context"))
	}
	if req.Msg.VideoId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("video_id is required"))
	}

	favoriteVideo, err := s.usecase.AddFavoriteVideo(ctx, userID, req.Msg.VideoId)
	if err != nil {
		log.Error("failed to add favorite video", "video_id", req.Msg.VideoId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to add favorite video: %v", err)
	}

	resp := &pb.AddFavoriteVideoResponse{FavoriteVideo: s.presenter.FavoriteVideo(*favoriteVideo)}
	return connect.NewResponse(resp), nil
}

func (s *FavoriteServiceServer) RemoveFavoriteVideo(ctx context.Context, req *connect.Request[pb.RemoveFavoriteVideoRequest]) (*connect.Response[pb.RemoveFavoriteVideoResponse], error) {
	log := s.loggerWithCtx(ctx)
	userID := ctxkeys.UserIDFromContext(ctx)
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user id missing in context"))
	}
	if req.Msg.VideoId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("video_id is required"))
	}
	removed, err := s.usecase.RemoveFavoriteVideo(ctx, userID, req.Msg.VideoId)
	if err != nil {
		if errors.Is(err, repository.ErrFavoriteVideoNotFound) {
			return nil, status.Error(codes.NotFound, "favorite video not found")
		}
		log.Error("failed to remove favorite video", "video_id", req.Msg.VideoId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to remove favorite video: %v", err)
	}

	resp := &pb.RemoveFavoriteVideoResponse{FavoriteVideoUuid: removed.FavoriteVideoUUID}
	return connect.NewResponse(resp), nil
}

func (s *FavoriteServiceServer) ListFavoriteVideos(ctx context.Context, req *connect.Request[pb.ListFavoriteVideosRequest]) (*connect.Response[pb.ListFavoriteVideosResponse], error) {
	log := s.loggerWithCtx(ctx)
	userID := ctxkeys.UserIDFromContext(ctx)
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user id missing in context"))
	}

	favorites, err := s.usecase.ListFavoriteVideos(ctx, userID)
	if err != nil {
		log.Error("failed to list favorite videos", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list favorite videos: %v", err)
	}

	resp := &pb.ListFavoriteVideosResponse{FavoriteVideos: s.presenter.FavoriteVideos(favorites)}
	return connect.NewResponse(resp), nil
}

func (s *FavoriteServiceServer) AddFavoriteActor(ctx context.Context, req *connect.Request[pb.AddFavoriteActorRequest]) (*connect.Response[pb.AddFavoriteActorResponse], error) {
	log := s.loggerWithCtx(ctx)
	userID := ctxkeys.UserIDFromContext(ctx)
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user id missing in context"))
	}
	if req.Msg.ActorId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("actor_id is required"))
	}

	favoriteActor, err := s.usecase.AddFavoriteActor(ctx, userID, req.Msg.ActorId)
	if err != nil {
		log.Error("failed to add favorite actor", "actor_id", req.Msg.ActorId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to add favorite actor: %v", err)
	}

	resp := &pb.AddFavoriteActorResponse{FavoriteActor: s.presenter.FavoriteActor(*favoriteActor)}
	return connect.NewResponse(resp), nil
}

func (s *FavoriteServiceServer) RemoveFavoriteActor(ctx context.Context, req *connect.Request[pb.RemoveFavoriteActorRequest]) (*connect.Response[pb.RemoveFavoriteActorResponse], error) {
	log := s.loggerWithCtx(ctx)
	userID := ctxkeys.UserIDFromContext(ctx)
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user id missing in context"))
	}
	if req.Msg.ActorId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("actor_id is required"))
	}

	removed, err := s.usecase.RemoveFavoriteActor(ctx, userID, req.Msg.ActorId)
	if err != nil {
		if errors.Is(err, repository.ErrFavoriteActorNotFound) {
			return nil, status.Error(codes.NotFound, "favorite actor not found")
		}
		log.Error("failed to remove favorite actor", "actor_id", req.Msg.ActorId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to remove favorite actor: %v", err)
	}

	resp := &pb.RemoveFavoriteActorResponse{FavoriteActorUuid: removed.FavoriteActorUUID}
	return connect.NewResponse(resp), nil
}

func (s *FavoriteServiceServer) ListFavoriteActors(ctx context.Context, req *connect.Request[pb.ListFavoriteActorsRequest]) (*connect.Response[pb.ListFavoriteActorsResponse], error) {
	log := s.loggerWithCtx(ctx)
	userID := ctxkeys.UserIDFromContext(ctx)
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user id missing in context"))
	}

	favorites, err := s.usecase.ListFavoriteActors(ctx, userID)
	if err != nil {
		log.Error("failed to list favorite actors", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list favorite actors: %v", err)
	}

	resp := &pb.ListFavoriteActorsResponse{FavoriteActors: s.presenter.FavoriteActors(favorites)}
	return connect.NewResponse(resp), nil
}
