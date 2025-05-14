package connect

import (
	"context"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tikfack/server/gen/video" // ※生成コードのパッケージ名に合わせて調整してください
	videoconnect "github.com/tikfack/server/gen/video/videoconnect"

	"github.com/tikfack/server/internal/application/usecase/video"
	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/infrastructure/repository"
)

// videoServiceServer は Connect のサーバー実装です。
type videoServiceServer struct {
	videoUsecase video.VideoUsecase
}

// NewVideoServiceHandler はハンドラーの初期化を行います。
func NewVideoServiceHandler() *videoServiceServer {
	// repository.NewDMMVideoRepository() の実装を渡す
	repo := repository.NewDMMVideoRepository()
	vu := video.NewVideoUsecase(repo)
	return &videoServiceServer{
		videoUsecase: vu,
	}
}

// GetHandler は、Connect サービスのパターンとハンドラーを返します。
func (s *videoServiceServer) GetHandler() (string, http.Handler) {
	// 生成されたコードの関数名に合わせてください
	// 例: pb.NewVideoServiceHandler または v1connect.NewVideoServiceHandler
	pattern, handler := videoconnect.NewVideoServiceHandler(s, connect.WithCompressMinBytes(0))
	return pattern, handler
}

// GetVideos は、動画一覧を取得するエンドポイントの実装例です。
func (s *videoServiceServer) GetVideos(ctx context.Context, req *connect.Request[pb.GetVideosRequest]) (*connect.Response[pb.GetVideosResponse], error) {
	var targetDate time.Time
	if req.Msg.Date == "" {
		targetDate = time.Now()
	} else {
		t, err := time.Parse("2006-01-02", req.Msg.Date)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "不正な日付形式です")
		}
		targetDate = t
	}
	videos, err := s.videoUsecase.GetVideos(ctx, targetDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "動画の取得に失敗しました: %v", err)
	}
	pbVideos := convertVideosToPb(videos)
	res := &pb.GetVideosResponse{Videos: pbVideos}
	return connect.NewResponse(res), nil
}

// GetVideoById は、ID で動画を取得するエンドポイントの実装例です。
func (s *videoServiceServer) GetVideoById(ctx context.Context, req *connect.Request[pb.GetVideoByIdRequest]) (*connect.Response[pb.GetVideoByIdResponse], error) {
	video, err := s.videoUsecase.GetVideoById(ctx, req.Msg.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "動画の取得に失敗しました: %v", err)
	}
	if video == nil {
		return nil, status.Error(codes.NotFound, "video not found")
	}
	res := &pb.GetVideoByIdResponse{Video: convertToPbVideo(*video)}
	return connect.NewResponse(res), nil
}

// convertVideosToPb はドメイン層の Video を pb.Video に変換します。
func convertVideosToPb(videos []entity.Video) []*pb.Video {
	var pbVideos []*pb.Video
	for _, v := range videos {
		pbVideos = append(pbVideos, convertToPbVideo(v))
	}
	return pbVideos
}

// convertToPbVideo は、ドメイン層の Video を pb.Video に変換するヘルパーです。
func convertToPbVideo(v entity.Video) *pb.Video {
	return &pb.Video{
		Id:           v.ID,
		Title:        v.Title,
		Description:  v.Description,
		DmmId:        v.DmmVideoId,
		ThumbnailUrl: v.ThumbnailURL,
		CreatedAt:    v.CreatedAt.Format("2006-01-02 15:04:05"),
		LikesCount:   int32(v.LikesCount),
		SampleUrl:    v.SampleURL,
		Url:          v.URL,
		DirectUrl:    v.DirectUrl, // 必要に応じて調整してください
		Author: &pb.User{
			Id:        v.Author.ID,
			Username:  v.Author.Username,
			AvatarUrl: v.Author.AvatarURL,
		},
	}
}
