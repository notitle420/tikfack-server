package connect

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	pb "github.com/tikfack/server/gen/event_log"
	"github.com/tikfack/server/internal/domain/entity"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

// eventLogPresenter は pb.Event をドメインの EventLog に変換する責務を持つ。
type eventLogPresenter interface {
	ToDomain(ctx context.Context, evt *pb.Event) (*entity.EventLog, error)
	ToDomainBatch(ctx context.Context, events []*pb.Event) ([]*entity.EventLog, error)
}

type defaultEventLogPresenter struct{}

func newEventLogPresenter() eventLogPresenter {
	return &defaultEventLogPresenter{}
}

func (p *defaultEventLogPresenter) ToDomain(ctx context.Context, evt *pb.Event) (*entity.EventLog, error) {
	if evt == nil {
		return nil, nil
	}
	props := map[string]any{}
	if evt.GetProps() != nil {
		props = evt.GetProps().AsMap()
	}
	rawProps, err := json.Marshal(props)
	if err != nil {
		return nil, err
	}
	return &entity.EventLog{
		EventLogID:  uuid.New().String(),
		TraceID:     ctxkeys.TraceIDFromContext(ctx),
		UserID:      evt.GetUserId(),
		SessionID:   evt.GetSessionId(),
		VideoDmmID:  evt.GetVideoDmmId(),
		ActressIDs:  evt.GetActressIds(),
		DirectorIDs: evt.GetDirectorIds(),
		GenreIDs:    evt.GetGenreIds(),
		MakerIDs:    evt.GetMakerIds(),
		SeriesIDs:   evt.GetSeriesIds(),
		EventType:   evt.GetEventType(),
		EventTime:   evt.GetEventTime().AsTime(),
		Props:       rawProps,
	}, nil
}

func (p *defaultEventLogPresenter) ToDomainBatch(ctx context.Context, events []*pb.Event) ([]*entity.EventLog, error) {
	result := make([]*entity.EventLog, len(events))
	for i, evt := range events {
		domainEvt, err := p.ToDomain(ctx, evt)
		if err != nil {
			return nil, err
		}
		result[i] = domainEvt
	}
	return result, nil
}
