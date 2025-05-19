package connect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tikfack/server/internal/domain/entity"
)

func TestConvertToPbVideo(t *testing.T) {
	created := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	v := entity.Video{
		DmmID:        "id1",
		Title:        "title1",
		DirectURL:    "durl",
		URL:          "url",
		SampleURL:    "surl",
		ThumbnailURL: "turl",
		CreatedAt:    created,
		Price:        1000,
		LikesCount:   3,
		Actresses:    []entity.Actress{{ID: "a1", Name: "A"}},
		Genres:       []entity.Genre{{ID: "g1", Name: "G"}},
		Makers:       []entity.Maker{{ID: "m1", Name: "M"}},
		Series:       []entity.Series{{ID: "s1", Name: "S"}},
		Directors:    []entity.Director{{ID: "d1", Name: "D"}},
		Review:       entity.Review{Count: 5, Average: 4.5},
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

func TestConvertVideosToPb(t *testing.T) {
	videos := []entity.Video{{DmmID: "1"}, {DmmID: "2"}}
	pbVideos := convertVideosToPb(videos)
	require.Equal(t, 2, len(pbVideos))
	require.Equal(t, "1", pbVideos[0].DmmId)
	require.Equal(t, "2", pbVideos[1].DmmId)
}

func TestConvertHelpers(t *testing.T) {
	acts := convertActressesToPb([]entity.Actress{{ID: "a"}})
	require.Equal(t, "a", acts[0].Id)
	gens := convertGenresToPb([]entity.Genre{{ID: "g"}})
	require.Equal(t, "g", gens[0].Id)
	mak := convertMakersToPb([]entity.Maker{{ID: "m"}})
	require.Equal(t, "m", mak[0].Id)
	ser := convertSeriesToPb([]entity.Series{{ID: "s"}})
	require.Equal(t, "s", ser[0].Id)
	dir := convertDirectorsToPb([]entity.Director{{ID: "d"}})
	require.Equal(t, "d", dir[0].Id)
	rev := convertReviewToPb(entity.Review{Count: 2, Average: 1.2})
	require.Equal(t, int32(2), rev.Count)
	require.Equal(t, float32(1.2), rev.Average)
}
