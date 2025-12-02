// Manual connect handler for FavoriteService.
package favoriteconnect

import (
	"context"
	"errors"
	"net/http"

	"github.com/bufbuild/connect-go"
	favorite "github.com/tikfack/server/gen/favorite"
)

const (
	FavoriteServiceName = "favorite.FavoriteService"

	FavoriteServiceAddFavoriteVideoProcedure    = "/favorite.FavoriteService/AddFavoriteVideo"
	FavoriteServiceRemoveFavoriteVideoProcedure = "/favorite.FavoriteService/RemoveFavoriteVideo"
	FavoriteServiceListFavoriteVideosProcedure  = "/favorite.FavoriteService/ListFavoriteVideos"
	FavoriteServiceAddFavoriteActorProcedure    = "/favorite.FavoriteService/AddFavoriteActor"
	FavoriteServiceRemoveFavoriteActorProcedure = "/favorite.FavoriteService/RemoveFavoriteActor"
	FavoriteServiceListFavoriteActorsProcedure  = "/favorite.FavoriteService/ListFavoriteActors"
)

// FavoriteServiceHandler defines the server interface.
type FavoriteServiceHandler interface {
	AddFavoriteVideo(context.Context, *connect.Request[favorite.AddFavoriteVideoRequest]) (*connect.Response[favorite.AddFavoriteVideoResponse], error)
	RemoveFavoriteVideo(context.Context, *connect.Request[favorite.RemoveFavoriteVideoRequest]) (*connect.Response[favorite.RemoveFavoriteVideoResponse], error)
	ListFavoriteVideos(context.Context, *connect.Request[favorite.ListFavoriteVideosRequest]) (*connect.Response[favorite.ListFavoriteVideosResponse], error)
	AddFavoriteActor(context.Context, *connect.Request[favorite.AddFavoriteActorRequest]) (*connect.Response[favorite.AddFavoriteActorResponse], error)
	RemoveFavoriteActor(context.Context, *connect.Request[favorite.RemoveFavoriteActorRequest]) (*connect.Response[favorite.RemoveFavoriteActorResponse], error)
	ListFavoriteActors(context.Context, *connect.Request[favorite.ListFavoriteActorsRequest]) (*connect.Response[favorite.ListFavoriteActorsResponse], error)
}

// NewFavoriteServiceHandler registers unary handlers for FavoriteService.
func NewFavoriteServiceHandler(svc FavoriteServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	addFavoriteVideoHandler := connect.NewUnaryHandler(
		FavoriteServiceAddFavoriteVideoProcedure,
		svc.AddFavoriteVideo,
		opts...,
	)
	removeFavoriteVideoHandler := connect.NewUnaryHandler(
		FavoriteServiceRemoveFavoriteVideoProcedure,
		svc.RemoveFavoriteVideo,
		opts...,
	)
	listFavoriteVideosHandler := connect.NewUnaryHandler(
		FavoriteServiceListFavoriteVideosProcedure,
		svc.ListFavoriteVideos,
		opts...,
	)
	addFavoriteActorHandler := connect.NewUnaryHandler(
		FavoriteServiceAddFavoriteActorProcedure,
		svc.AddFavoriteActor,
		opts...,
	)
	removeFavoriteActorHandler := connect.NewUnaryHandler(
		FavoriteServiceRemoveFavoriteActorProcedure,
		svc.RemoveFavoriteActor,
		opts...,
	)
	listFavoriteActorsHandler := connect.NewUnaryHandler(
		FavoriteServiceListFavoriteActorsProcedure,
		svc.ListFavoriteActors,
		opts...,
	)

	return "/favorite.FavoriteService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case FavoriteServiceAddFavoriteVideoProcedure:
			addFavoriteVideoHandler.ServeHTTP(w, r)
		case FavoriteServiceRemoveFavoriteVideoProcedure:
			removeFavoriteVideoHandler.ServeHTTP(w, r)
		case FavoriteServiceListFavoriteVideosProcedure:
			listFavoriteVideosHandler.ServeHTTP(w, r)
		case FavoriteServiceAddFavoriteActorProcedure:
			addFavoriteActorHandler.ServeHTTP(w, r)
		case FavoriteServiceRemoveFavoriteActorProcedure:
			removeFavoriteActorHandler.ServeHTTP(w, r)
		case FavoriteServiceListFavoriteActorsProcedure:
			listFavoriteActorsHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedFavoriteServiceHandler returns unimplemented errors for all methods.
type UnimplementedFavoriteServiceHandler struct{}

func (UnimplementedFavoriteServiceHandler) AddFavoriteVideo(context.Context, *connect.Request[favorite.AddFavoriteVideoRequest]) (*connect.Response[favorite.AddFavoriteVideoResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("favorite.FavoriteService.AddFavoriteVideo is not implemented"))
}

func (UnimplementedFavoriteServiceHandler) RemoveFavoriteVideo(context.Context, *connect.Request[favorite.RemoveFavoriteVideoRequest]) (*connect.Response[favorite.RemoveFavoriteVideoResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("favorite.FavoriteService.RemoveFavoriteVideo is not implemented"))
}

func (UnimplementedFavoriteServiceHandler) ListFavoriteVideos(context.Context, *connect.Request[favorite.ListFavoriteVideosRequest]) (*connect.Response[favorite.ListFavoriteVideosResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("favorite.FavoriteService.ListFavoriteVideos is not implemented"))
}

func (UnimplementedFavoriteServiceHandler) AddFavoriteActor(context.Context, *connect.Request[favorite.AddFavoriteActorRequest]) (*connect.Response[favorite.AddFavoriteActorResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("favorite.FavoriteService.AddFavoriteActor is not implemented"))
}

func (UnimplementedFavoriteServiceHandler) RemoveFavoriteActor(context.Context, *connect.Request[favorite.RemoveFavoriteActorRequest]) (*connect.Response[favorite.RemoveFavoriteActorResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("favorite.FavoriteService.RemoveFavoriteActor is not implemented"))
}

func (UnimplementedFavoriteServiceHandler) ListFavoriteActors(context.Context, *connect.Request[favorite.ListFavoriteActorsRequest]) (*connect.Response[favorite.ListFavoriteActorsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("favorite.FavoriteService.ListFavoriteActors is not implemented"))
}
