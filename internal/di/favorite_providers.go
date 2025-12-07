package di

import (
	"github.com/bufbuild/connect-go"

	favoriteuc "github.com/tikfack/server/internal/application/usecase/favorite"
	favoriterepo "github.com/tikfack/server/internal/infrastructure/repository/favorite"
	userrepo "github.com/tikfack/server/internal/infrastructure/repository/user"
	favoritehandler "github.com/tikfack/server/internal/presentation/connect"
)

func provideFavoriteUsecase() (favoriteuc.FavoriteUsecase, error) {
	db, err := provideDatabase()
	if err != nil {
		return nil, err
	}
	userRepository := userrepo.NewPostgresUserRepository(db)
	videoRepository := favoriterepo.NewPostgresFavoriteVideoRepository(db)
	actorRepository := favoriterepo.NewPostgresFavoriteActorRepository(db)
	return favoriteuc.NewFavoriteUsecase(userRepository, videoRepository, actorRepository), nil
}

func provideFavoriteHandler(uc favoriteuc.FavoriteUsecase, opts []connect.HandlerOption) *favoritehandler.FavoriteServiceServer {
	return favoritehandler.NewFavoriteServiceHandler(uc, opts...)
}
