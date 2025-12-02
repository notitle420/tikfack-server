package di

import (
	"github.com/bufbuild/connect-go"

	favoriteuc "github.com/tikfack/server/internal/application/usecase/favorite"
	accountrepo "github.com/tikfack/server/internal/infrastructure/repository/account"
	favoriterepo "github.com/tikfack/server/internal/infrastructure/repository/favorite"
	favoritehandler "github.com/tikfack/server/internal/presentation/connect"
)

func provideFavoriteUsecase() favoriteuc.FavoriteUsecase {
	accountRepository := accountrepo.NewMemoryAccountRepository()
	videoRepository := favoriterepo.NewMemoryFavoriteVideoRepository()
	actorRepository := favoriterepo.NewMemoryFavoriteActorRepository()
	return favoriteuc.NewFavoriteUsecase(accountRepository, videoRepository, actorRepository)
}

func provideFavoriteHandler(uc favoriteuc.FavoriteUsecase, opts []connect.HandlerOption) *favoritehandler.FavoriteServiceServer {
	return favoritehandler.NewFavoriteServiceHandler(uc, opts...)
}
