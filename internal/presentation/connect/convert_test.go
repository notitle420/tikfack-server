package connect

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tikfack/server/internal/application/model"
)

func TestConvertToPbVideo(t *testing.T) {
	created := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	v := model.Video{
		DmmID:        "id1",
		Title:        "title1",
		DirectURL:    "durl",
		URL:          "url",
		SampleURL:    "surl",
		ThumbnailURL: "turl",
		CreatedAt:    created,
		Price:        1000,
		LikesCount:   3,
		Actresses:    []model.Actress{{ID: "a1", Name: "A"}},
		Genres:       []model.Genre{{ID: "g1", Name: "G"}},
		Makers:       []model.Maker{{ID: "m1", Name: "M"}},
		Series:       []model.Series{{ID: "s1", Name: "S"}},
		Directors:    []model.Director{{ID: "d1", Name: "D"}},
		Review:       model.Review{Count: 5, Average: 4.5},
	}

	pbVideo := convertToPbVideo(v)
	require.Equal(t, v.DmmID, pbVideo.DmmId)
	require.Equal(t, v.Title, pbVideo.Title)
	require.Equal(t, v.DirectURL, pbVideo.DirectUrl)
	require.Equal(t, v.URL, pbVideo.Url)
	require.Equal(t, v.SampleURL, pbVideo.SampleUrl)
	require.Equal(t, v.ThumbnailURL, pbVideo.ThumbnailUrl)
	require.Equal(t, created.Format(time.RFC3339), pbVideo.CreatedAt)
	require.Equal(t, int32(v.Price), pbVideo.Price)
	require.Equal(t, int32(v.LikesCount), pbVideo.LikesCount)
	require.Equal(t, int32(v.Review.Count), pbVideo.Review.Count)
	require.Equal(t, v.Review.Average, pbVideo.Review.Average)
	require.Len(t, pbVideo.Actresses, 1)
	require.Equal(t, v.Actresses[0].ID, pbVideo.Actresses[0].Id)
	require.Len(t, pbVideo.Genres, 1)
	require.Equal(t, v.Genres[0].ID, pbVideo.Genres[0].Id)
}

func TestPresenterVideos(t *testing.T) {
	ctx := context.Background()
	stubResolver := videoURLResolverFunc(func(_ context.Context, dmmID string) (string, error) {
		return "resolved-" + dmmID, nil
	})
	presenter := &pbVideoPresenter{urlResolver: stubResolver}
	videos := []model.Video{
		{DmmID: "1"},
		{DmmID: "2", DirectURL: "existing"},
	}
	result := presenter.Videos(ctx, videos)
	require.Len(t, result, 2)
	require.Equal(t, "resolved-1", result[0].DirectUrl)
	require.Equal(t, "existing", result[1].DirectUrl)
	// 元のスライスが書き換えられていないことを確認
	require.Equal(t, "", videos[0].DirectURL)
	require.Equal(t, "existing", videos[1].DirectURL)
}

func TestPresenterMetadata(t *testing.T) {
	presenter := newVideoPresenter()
	require.Nil(t, presenter.Metadata(nil))
	pbMeta := presenter.Metadata(&model.SearchMetadata{ResultCount: 1, TotalCount: 2, FirstPosition: 3})
	require.Equal(t, int32(1), pbMeta.ResultCount)
	require.Equal(t, int32(2), pbMeta.TotalCount)
	require.Equal(t, int32(3), pbMeta.FirstPosition)
}
