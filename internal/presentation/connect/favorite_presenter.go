package connect

import (
	pb "github.com/tikfack/server/gen/favorite"
	"github.com/tikfack/server/internal/application/model"
)

type favoritePresenter struct{}

func newFavoritePresenter() favoritePresenter {
	return favoritePresenter{}
}

func (p favoritePresenter) FavoriteVideo(v model.FavoriteVideo) *pb.FavoriteVideo {
	return &pb.FavoriteVideo{
		FavoriteVideoUuid: v.FavoriteVideoUUID,
		UserId:            v.UserID,
		VideoId:           v.VideoID,
		CreatedAt:         v.CreatedAt,
	}
}

func (p favoritePresenter) FavoriteActor(a model.FavoriteActor) *pb.FavoriteActor {
	return &pb.FavoriteActor{
		FavoriteActorUuid: a.FavoriteActorUUID,
		UserId:            a.UserID,
		ActorId:           a.ActorID,
		CreatedAt:         a.CreatedAt,
	}
}

func (p favoritePresenter) FavoriteVideos(videos []model.FavoriteVideo) []*pb.FavoriteVideo {
	if len(videos) == 0 {
		return nil
	}
	out := make([]*pb.FavoriteVideo, 0, len(videos))
	for _, v := range videos {
		out = append(out, p.FavoriteVideo(v))
	}
	return out
}

func (p favoritePresenter) FavoriteActors(actors []model.FavoriteActor) []*pb.FavoriteActor {
	if len(actors) == 0 {
		return nil
	}
	out := make([]*pb.FavoriteActor, 0, len(actors))
	for _, a := range actors {
		out = append(out, p.FavoriteActor(a))
	}
	return out
}
