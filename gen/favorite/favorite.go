package favorite

// Data transfer structures for FavoriteService.
type FavoriteVideo struct {
	FavoriteVideoUuid string `json:"favorite_video_uuid"`
	UserId            string `json:"user_id"`
	VideoId           string `json:"video_id"`
	CreatedAt         string `json:"created_at"`
}

type FavoriteActor struct {
	FavoriteActorUuid string `json:"favorite_actor_uuid"`
	UserId            string `json:"user_id"`
	ActorId           string `json:"actor_id"`
	CreatedAt         string `json:"created_at"`
}

type AddFavoriteVideoRequest struct {
	VideoId string `json:"video_id"`
}

type AddFavoriteVideoResponse struct {
	FavoriteVideo *FavoriteVideo `json:"favorite_video"`
}

type RemoveFavoriteVideoRequest struct {
	VideoId string `json:"video_id"`
}

type RemoveFavoriteVideoResponse struct {
	FavoriteVideoUuid string `json:"favorite_video_uuid"`
}

type ListFavoriteVideosRequest struct{}

type ListFavoriteVideosResponse struct {
	FavoriteVideos []*FavoriteVideo `json:"favorite_videos"`
}

type AddFavoriteActorRequest struct {
	ActorId string `json:"actor_id"`
}

type AddFavoriteActorResponse struct {
	FavoriteActor *FavoriteActor `json:"favorite_actor"`
}

type RemoveFavoriteActorRequest struct {
	ActorId string `json:"actor_id"`
}

type RemoveFavoriteActorResponse struct {
	FavoriteActorUuid string `json:"favorite_actor_uuid"`
}

type ListFavoriteActorsRequest struct{}

type ListFavoriteActorsResponse struct {
	FavoriteActors []*FavoriteActor `json:"favorite_actors"`
}
