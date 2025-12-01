//go:build wireinject
// +build wireinject

package di

import (
	"github.com/bufbuild/connect-go"
	"github.com/google/wire"
	video "github.com/tikfack/server/internal/application/usecase/video"
	videorepo "github.com/tikfack/server/internal/infrastructure/repository/video"
	connecthandler "github.com/tikfack/server/internal/presentation/connect"
)

func InitializeVideoHandler(opts []connect.HandlerOption) (*connecthandler.VideoServiceServer, error) {
	wire.Build(
		videorepo.NewVideoRepository,
		video.NewVideoUsecase,
		provideVideoHandler,
	)
	return nil, nil
}

func InitializeEventLogHandler(opts []connect.HandlerOption) (*connecthandler.EventLogServiceServer, error) {
	wire.Build(
		provideKafkaWriter,
		provideEventLogUsecase,
		provideEventLogHandler,
	)
	return nil, nil
}
