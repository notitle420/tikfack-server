package connect

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tikfack/server/internal/domain/entity"
)

func TestNewVideoServiceHandler(t *testing.T) {
	os.Setenv("BASE_URL", "http://example.com")
	os.Setenv("DMM_API_ID", "id")
	os.Setenv("DMM_API_AFFILIATE_ID", "aff")
	h := NewVideoServiceHandler()
	require.NotNil(t, h)
}

func TestGetHandler(t *testing.T) {
	vu := &mockVideoUsecase{}
	h := NewVideoServiceHandlerWithUsecase(vu)
	pattern, handler := h.GetHandler()
	require.NotEmpty(t, pattern)
	require.NotNil(t, handler)
}

type mockVideoUsecase struct{}

func (m *mockVideoUsecase) GetVideosByDate(ctx context.Context, d time.Time) ([]entity.Video, error) {
	return nil, nil
}
func (m *mockVideoUsecase) GetVideoById(ctx context.Context, id string) (*entity.Video, error) {
	return nil, nil
}
func (m *mockVideoUsecase) SearchVideos(ctx context.Context, keyword, actressID, genreID, makerID, seriesID, directorID string) ([]entity.Video, error) {
	return nil, nil
}
func (m *mockVideoUsecase) GetVideosByID(ctx context.Context, actressIDs, genreIDs, makerIDs, seriesIDs, directorIDs []string, hits, offset int32, sort, gteDate, lteDate, site, service, floor string) ([]entity.Video, error) {
	return nil, nil
}
func (m *mockVideoUsecase) GetVideosByKeyword(ctx context.Context, keyword string, hits, offset int32, sort, gteDate, lteDate, site, service, floor string) ([]entity.Video, error) {
	return nil, nil
}
