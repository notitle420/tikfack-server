package di

import (
	"github.com/bufbuild/connect-go"

	favoriteuc "github.com/tikfack/server/internal/application/usecase/favorite"
	accountrepo "github.com/tikfack/server/internal/infrastructure/repository/account"
	favoriterepo "github.com/tikfack/server/internal/infrastructure/repository/favorite"
	favoritehandler "github.com/tikfack/server/internal/presentation/connect"
)

func provideFavoriteUsecase() (favoriteuc.FavoriteUsecase, error) {
	db, err := provideDatabase()
	if err != nil {
		return nil, err
	}
	accountRepository := accountrepo.NewPostgresAccountRepository(db)
	videoRepository := favoriterepo.NewPostgresFavoriteVideoRepository(db)
	actorRepository := favoriterepo.NewPostgresFavoriteActorRepository(db)
	return favoriteuc.NewFavoriteUsecase(accountRepository, videoRepository, actorRepository), nil
}

func provideFavoriteHandler(uc favoriteuc.FavoriteUsecase, opts []connect.HandlerOption) *favoritehandler.FavoriteServiceServer {
	return favoritehandler.NewFavoriteServiceHandler(uc, opts...)
}
