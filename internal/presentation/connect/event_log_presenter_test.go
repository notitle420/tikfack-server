package connect

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	pb "github.com/tikfack/server/gen/event_log"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEventLogPresenter_ToDomain(t *testing.T) {
	ctx := context.Background()
	presenter := newEventLogPresenter()
	props, _ := structpb.NewStruct(map[string]any{"foo": "bar"})
	evtTime := timestamppb.New(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC))
	evt := &pb.Event{
		UserId:      "user",
		SessionId:   "sess",
		VideoDmmId:  "dmm",
		ActressIds:  []string{"a"},
		DirectorIds: []string{"d"},
		GenreIds:    []string{"g"},
		MakerIds:    []string{"m"},
		SeriesIds:   []string{"s"},
		EventType:   "start",
		EventTime:   evtTime,
		Props:       props,
	}

	domain, err := presenter.ToDomain(ctx, evt)
	require.NoError(t, err)
	require.Equal(t, "user", domain.UserID)
	require.Equal(t, evtTime.AsTime(), domain.EventTime)
	require.Equal(t, []string{"a"}, domain.ActressIDs)
	require.NotEmpty(t, domain.EventLogID)
	require.NotEmpty(t, domain.Props)
}

func TestEventLogPresenter_Batch(t *testing.T) {
	presenter := newEventLogPresenter()
	ctx := context.Background()
	events := []*pb.Event{{EventTime: timestamppb.Now(), Props: &structpb.Struct{}}}
	domain, err := presenter.ToDomainBatch(ctx, events)
	require.NoError(t, err)
	require.Len(t, domain, 1)
}
