//go:build wireinject
// +build wireinject

package di

import (
	"github.com/bufbuild/connect-go"
	"github.com/google/wire"
	favoriteuc "github.com/tikfack/server/internal/application/usecase/favorite"
	favoritehandler "github.com/tikfack/server/internal/presentation/connect"
)

func InitializeFavoriteHandler(opts []connect.HandlerOption) (*favoritehandler.FavoriteServiceServer, error) {
	wire.Build(
		provideFavoriteUsecase,
		provideFavoriteHandler,
	)
	return nil, nil
}

// favoriteSet can be used to compose favorite dependencies elsewhere.
var favoriteSet = wire.NewSet(provideFavoriteUsecase, provideFavoriteHandler)
