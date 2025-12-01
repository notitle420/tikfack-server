package connect

import (
	"context"
	"time"

	"github.com/tikfack/server/internal/application/model"
	"github.com/tikfack/server/internal/infrastructure/util"

	pb "github.com/tikfack/server/gen/video"
)

// videoPresenter はドメインモデルをtransport層のpbメッセージへ変換する責務を担う。
type videoPresenter interface {
	Video(ctx context.Context, video *model.Video) *pb.Video
	Videos(ctx context.Context, videos []model.Video) []*pb.Video
	Metadata(md *model.SearchMetadata) *pb.SearchMetadata
}

// videoURLResolver は動画のDirectURLを検証・取得する動作を抽象化する。
type videoURLResolver interface {
	Resolve(ctx context.Context, dmmID string) (string, error)
}

type videoURLResolverFunc func(ctx context.Context, dmmID string) (string, error)

func (f videoURLResolverFunc) Resolve(ctx context.Context, dmmID string) (string, error) {
	return f(ctx, dmmID)
}

// pbVideoPresenter は videoPresenter のデフォルト実装。
type pbVideoPresenter struct {
	urlResolver videoURLResolver
}

func newVideoPresenter() videoPresenter {
	return &pbVideoPresenter{
		urlResolver: videoURLResolverFunc(func(_ context.Context, dmmID string) (string, error) {
			return util.GetValidVideoUrl(dmmID)
		}),
	}
}

func (p *pbVideoPresenter) Video(ctx context.Context, video *model.Video) *pb.Video {
	if video == nil {
		return nil
	}
	// コピーを作ってDirectURLの補完による副作用を避ける。
	copyVideo := *video
	if copyVideo.DirectURL == "" {
		if directURL, err := p.urlResolver.Resolve(ctx, copyVideo.DmmID); err == nil {
			copyVideo.DirectURL = directURL
		}
	}
	return convertToPbVideo(copyVideo)
}

func (p *pbVideoPresenter) Videos(ctx context.Context, videos []model.Video) []*pb.Video {
	if len(videos) == 0 {
		return nil
	}
	converted := make([]*pb.Video, 0, len(videos))
	for i := range videos {
		converted = append(converted, p.Video(ctx, &videos[i]))
	}
	return converted
}

func (p *pbVideoPresenter) Metadata(md *model.SearchMetadata) *pb.SearchMetadata {
	if md == nil {
		return nil
	}
	return &pb.SearchMetadata{
		ResultCount:   int32(md.ResultCount),
		TotalCount:    int32(md.TotalCount),
		FirstPosition: int32(md.FirstPosition),
	}
}

// convertToPbVideo はモデルからpb.Videoへ変換するヘルパー。
func convertToPbVideo(v model.Video) *pb.Video {
	actresses := make([]*pb.Actress, 0, len(v.Actresses))
	for _, a := range v.Actresses {
		actresses = append(actresses, &pb.Actress{Id: a.ID, Name: a.Name})
	}

	genres := make([]*pb.Genre, 0, len(v.Genres))
	for _, g := range v.Genres {
		genres = append(genres, &pb.Genre{Id: g.ID, Name: g.Name})
	}

	makers := make([]*pb.Maker, 0, len(v.Makers))
	for _, m := range v.Makers {
		makers = append(makers, &pb.Maker{Id: m.ID, Name: m.Name})
	}

	series := make([]*pb.Series, 0, len(v.Series))
	for _, s := range v.Series {
		series = append(series, &pb.Series{Id: s.ID, Name: s.Name})
	}

	directors := make([]*pb.Director, 0, len(v.Directors))
	for _, d := range v.Directors {
		directors = append(directors, &pb.Director{Id: d.ID, Name: d.Name})
	}

	review := &pb.Review{Count: int32(v.Review.Count), Average: v.Review.Average}

	return &pb.Video{
		DmmId:        v.DmmID,
		Title:        v.Title,
		DirectUrl:    v.DirectURL,
		Url:          v.URL,
		SampleUrl:    v.SampleURL,
		ThumbnailUrl: v.ThumbnailURL,
		CreatedAt:    v.CreatedAt.Format(time.RFC3339),
		Price:        int32(v.Price),
		LikesCount:   int32(v.LikesCount),
		Actresses:    actresses,
		Genres:       genres,
		Makers:       makers,
		Series:       series,
		Directors:    directors,
		Review:       review,
	}
}
