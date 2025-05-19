package connect

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tikfack/server/gen/video" // ※生成コードのパッケージ名に合わせて調整してください
	videoconnect "github.com/tikfack/server/gen/video/videoconnect"

	video "github.com/tikfack/server/internal/application/usecase/video"
	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/infrastructure/repository"
	"github.com/tikfack/server/internal/infrastructure/util"
)

// videoServiceServer は Connect のサーバー実装です。
type videoServiceServer struct {
	videoUsecase video.VideoUsecase
	logger       *slog.Logger
}

// NewVideoServiceHandler はハンドラーの初期化を行います。
func NewVideoServiceHandler() *videoServiceServer {
	// repository.NewDMMVideoRepository() の実装を渡す
	repo, err := repository.NewVideoRepository()
	if err != nil {
		// TODO: Consider how to handle initialization error
		panic(err)
	}
	vu := video.NewVideoUsecase(repo)
	return &videoServiceServer{
		videoUsecase: vu,
		logger:       slog.Default().With(slog.String("component", "video_handler")),
	}
}

func NewVideoServiceHandlerWithUsecase(vu video.VideoUsecase) *videoServiceServer {
	return &videoServiceServer{
		videoUsecase: vu,
		logger:       slog.Default().With(slog.String("component", "video_handler")),
	}
}

// GetHandler は、Connect サービスのパターンとハンドラーを返します。
func (s *videoServiceServer) GetHandler() (string, http.Handler) {
	// 生成されたコードの関数名に合わせてください
	// 例: pb.NewVideoServiceHandler または v1connect.NewVideoServiceHandler
	pattern, handler := videoconnect.NewVideoServiceHandler(s, connect.WithCompressMinBytes(0))
	return pattern, handler
}

// GetVideosByDate は、動画一覧を取得するエンドポイントの実装例です。
func (s *videoServiceServer) GetVideosByDate(ctx context.Context, req *connect.Request[pb.GetVideosByDateRequest]) (*connect.Response[pb.GetVideosByDateResponse], error) {
	s.logger.Debug("API: GetVideosByDate", "date", req.Msg.Date)
	
	var targetDate time.Time
	if req.Msg.Date == "" {
		targetDate = time.Now()
	} else {
		t, err := time.Parse("2006-01-02", req.Msg.Date)
		if err != nil {
			s.logger.Error("不正な日付形式", "date", req.Msg.Date, "error", err)
			return nil, status.Error(codes.InvalidArgument, "不正な日付形式です")
		}
		targetDate = t
	}
	
	// ユースケースから動画リストを取得
	videos, err := s.videoUsecase.GetVideosByDate(ctx, targetDate)
	if err != nil {
		s.logger.Error("動画の取得に失敗", "date", targetDate.Format("2006-01-02"), "error", err)
		return nil, status.Errorf(codes.Internal, "動画の取得に失敗しました: %v", err)
	}
	
	// ハンドラー層で各動画のURL検証を行う
	for i := range videos {
		directURL, err := util.GetValidVideoUrl(videos[i].DmmID)
		if err == nil {
			videos[i].DirectURL = directURL
		}
	}
	
	pbVideos := convertVideosToPb(videos)
	s.logger.Debug("GetVideosByDate completed", "count", len(pbVideos))
	res := &pb.GetVideosByDateResponse{Videos: pbVideos}
	return connect.NewResponse(res), nil
}

// GetVideoById は、ID で動画を取得するエンドポイントの実装例です。
func (s *videoServiceServer) GetVideoById(ctx context.Context, req *connect.Request[pb.GetVideoByIdRequest]) (*connect.Response[pb.GetVideoByIdResponse], error) {
	s.logger.Debug("API: GetVideoById", "dmmId", req.Msg.DmmId)
	
	video, err := s.videoUsecase.GetVideoById(ctx, req.Msg.DmmId)
	if err != nil {
		s.logger.Error("動画の取得に失敗", "dmmId", req.Msg.DmmId, "error", err)
		return nil, status.Errorf(codes.Internal, "動画の取得に失敗しました: %v", err)
	}
	if video == nil {
		s.logger.Info("動画が見つかりません", "dmmId", req.Msg.DmmId)
		return nil, status.Error(codes.NotFound, "video not found")
	}
	
	// ハンドラー層でURL検証を行う
	directURL, err := util.GetValidVideoUrl(video.DmmID)
	if err == nil {
		video.DirectURL = directURL
	}
	
	s.logger.Debug("GetVideoById completed", "dmmId", req.Msg.DmmId, "title", video.Title)
	res := &pb.GetVideoByIdResponse{Video: convertToPbVideo(*video)}
	return connect.NewResponse(res), nil
}

// SearchVideos は、動画を検索するエンドポイントの実装です。
func (s *videoServiceServer) SearchVideos(ctx context.Context, req *connect.Request[pb.SearchVideosRequest]) (*connect.Response[pb.SearchVideosResponse], error) {
	s.logger.Debug("API: SearchVideos", 
		"keyword", req.Msg.Keyword, 
		"actressId", req.Msg.ActressId, 
		"genreId", req.Msg.GenreId, 
		"makerId", req.Msg.MakerId,
		"seriesId", req.Msg.SeriesId,
		"directorId", req.Msg.DirectorId)
	
	// 実装が完了したら、こちらでもURL検証を行うように修正
	videos, err := s.videoUsecase.SearchVideos(ctx, 
		req.Msg.Keyword, 
		req.Msg.ActressId, 
		req.Msg.GenreId, 
		req.Msg.MakerId, 
		req.Msg.SeriesId, 
		req.Msg.DirectorId)
	
	if err != nil {
		s.logger.Error("動画の検索に失敗", "keyword", req.Msg.Keyword, "error", err)
		return nil, status.Errorf(codes.Internal, "動画の検索に失敗しました: %v", err)
	}
	
	// ハンドラー層で各動画のURL検証を行う
	for i := range videos {
		directURL, err := util.GetValidVideoUrl(videos[i].DmmID)
		if err == nil {
			videos[i].DirectURL = directURL
		}
	}
	
	pbVideos := convertVideosToPb(videos)
	s.logger.Debug("SearchVideos completed", "count", len(pbVideos))
	return connect.NewResponse(&pb.SearchVideosResponse{
		Videos: pbVideos,
	}), nil
}

// GetVideosByID は、複数のIDで動画を検索するエンドポイントの実装です。
func (s *videoServiceServer) GetVideosByID(ctx context.Context, req *connect.Request[pb.GetVideosByIDRequest]) (*connect.Response[pb.GetVideosByIDResponse], error) {
	s.logger.Debug("API: GetVideosByID", 
		"actressId_count", len(req.Msg.ActressId),
		"genreId_count", len(req.Msg.GenreId),
		"makerId_count", len(req.Msg.MakerId),
		"seriesId_count", len(req.Msg.SeriesId),
		"directorId_count", len(req.Msg.DirectorId),
		"hits", req.Msg.Hits,
		"offset", req.Msg.Offset)
	
	videos, err := s.videoUsecase.GetVideosByID(ctx, 
		req.Msg.ActressId, 
		req.Msg.GenreId, 
		req.Msg.MakerId, 
		req.Msg.SeriesId, 
		req.Msg.DirectorId,
		req.Msg.Hits,
		req.Msg.Offset,
		req.Msg.Sort,
		req.Msg.GteDate,
		req.Msg.LteDate,
		req.Msg.Site,
		req.Msg.Service,
		req.Msg.Floor)

	if err != nil {
		s.logger.Error("動画の検索に失敗", "error", err)
		return nil, status.Errorf(codes.Internal, "動画の検索に失敗しました: %v", err)
	}
	
	// ハンドラー層で各動画のURL検証を行う
	for i := range videos {
		directURL, err := util.GetValidVideoUrl(videos[i].DmmID)
		if err == nil {
			videos[i].DirectURL = directURL
		}
	}
	
	pbVideos := convertVideosToPb(videos)
	s.logger.Debug("GetVideosByID completed", "count", len(pbVideos))
	return connect.NewResponse(&pb.GetVideosByIDResponse{
		Videos: pbVideos,
	}), nil
}

// GetVideosByKeyword は、キーワードで動画を検索するエンドポイントの実装です。
func (s *videoServiceServer) GetVideosByKeyword(ctx context.Context, req *connect.Request[pb.GetVideosByKeywordRequest]) (*connect.Response[pb.GetVideosByKeywordResponse], error) {
	s.logger.Debug("API: GetVideosByKeyword", 
		"keyword", req.Msg.Keyword,
		"hits", req.Msg.Hits,
		"offset", req.Msg.Offset,
		"sort", req.Msg.Sort)
	
	videos, err := s.videoUsecase.GetVideosByKeyword(ctx, 
		req.Msg.Keyword,
		req.Msg.Hits,
		req.Msg.Offset,
		req.Msg.Sort,
		req.Msg.GteDate,
		req.Msg.LteDate,
		req.Msg.Site,
		req.Msg.Service,
		req.Msg.Floor)
	
	if err != nil {
		s.logger.Error("動画の検索に失敗", "keyword", req.Msg.Keyword, "error", err)
		return nil, status.Errorf(codes.Internal, "動画の検索に失敗しました: %v", err)
	}
	
	// ハンドラー層で各動画のURL検証を行う
	for i := range videos {
		directURL, err := util.GetValidVideoUrl(videos[i].DmmID)
		if err == nil {
			videos[i].DirectURL = directURL
		}
	}
	
	pbVideos := convertVideosToPb(videos)
	s.logger.Debug("GetVideosByKeyword completed", "count", len(pbVideos))
	return connect.NewResponse(&pb.GetVideosByKeywordResponse{
		Videos: pbVideos,
	}), nil
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
	pbVideo := &pb.Video{
		DmmId:        v.DmmID,
		Title:        v.Title,
		DirectUrl:    v.DirectURL,
		Url:          v.URL,
		SampleUrl:    v.SampleURL,
		ThumbnailUrl: v.ThumbnailURL,
		CreatedAt:    v.CreatedAt.Format("2006-01-02 15:04:05"),
		LikesCount:   int32(v.LikesCount),
		Price:        int32(v.Price),
		// Author情報はProtobufモデルに存在しないため削除
		// 代わりに女優リストを使用
		Actresses:    convertActressesToPb(v.Actresses),
		Genres:       convertGenresToPb(v.Genres),
		Makers:       convertMakersToPb(v.Makers),
		Series:       convertSeriesToPb(v.Series),
		Directors:    convertDirectorsToPb(v.Directors),
		Review:       convertReviewToPb(v.Review),
	}
	
	// プロトコルバッファオブジェクトをJSON形式でログに出
	
	return pbVideo
}

// 各エンティティの変換関数
func convertActressesToPb(actresses []entity.Actress) []*pb.Actress {
	var result []*pb.Actress
	for _, a := range actresses {
		result = append(result, &pb.Actress{
			Id:   a.ID,
			Name: a.Name,
		})
	}
	return result
}

func convertGenresToPb(genres []entity.Genre) []*pb.Genre {
	var result []*pb.Genre
	for _, g := range genres {
		result = append(result, &pb.Genre{
			Id:   g.ID,
			Name: g.Name,
		})
	}
	return result
}

func convertMakersToPb(makers []entity.Maker) []*pb.Maker {
	var result []*pb.Maker
	for _, m := range makers {
		result = append(result, &pb.Maker{
			Id:   m.ID,
			Name: m.Name,
		})
	}
	return result
}

func convertSeriesToPb(series []entity.Series) []*pb.Series {
	var result []*pb.Series
	for _, s := range series {
		result = append(result, &pb.Series{
			Id:   s.ID,
			Name: s.Name,
		})
	}
	return result
}

func convertDirectorsToPb(directors []entity.Director) []*pb.Director {
	var result []*pb.Director
	for _, d := range directors {
		result = append(result, &pb.Director{
			Id:   d.ID,
			Name: d.Name,
		})
	}
	return result
}

// レビュー情報の変換関数
func convertReviewToPb(review entity.Review) *pb.Review {
	pbReview := &pb.Review{
		Count:   int32(review.Count),
		Average: review.Average,
	}
	
	// レビュー情報をログに出力
	return pbReview
}
