package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"log/slog"

	"github.com/google/uuid"

	connect "github.com/bufbuild/connect-go"
	"github.com/segmentio/kafka-go"
	pb "github.com/tikfack/server/gen/event_log"
	eventlogconnect "github.com/tikfack/server/gen/event_log/event_logconnect"
	eventloguc "github.com/tikfack/server/internal/application/usecase/event_log"
	"github.com/tikfack/server/internal/domain/entity"
	repo "github.com/tikfack/server/internal/infrastructure/repository/event_log"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

// eventLogServiceServer implements the Connect-Go gRPC service for event logs.
type eventLogServiceServer struct {
	eventLogUsecase eventloguc.EventLogUsecase
	logger          *slog.Logger
	handlerOpts     []connect.HandlerOption
}

// NewEventLogServiceHandler initializes the handler with default Kafka repository implementation.
func NewEventLogServiceHandler(opts ...connect.HandlerOption) *eventLogServiceServer {
	var brokers []string
	raw := strings.TrimSpace(os.Getenv("KAFKA_BROKER_ADDRESSES"))
	if raw != "" {
		for _, addr := range strings.Split(raw, ",") {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				brokers = append(brokers, addr)
			}
		}
	}
	if len(brokers) == 0 {
		brokers = []string{"localhost:9094"}
	}

	// Initialize Kafka writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   "event-logs",
	})
	slog.Info("Kafka writer initialized", slog.Any("brokers", brokers))
	// Initialize repository implementing domain EventLogRepository
	repository := repo.NewKafkaEventLogRepository(writer)
	// Initialize usecase
	eu := eventloguc.NewEventLogService(repository)
	return &eventLogServiceServer{
		eventLogUsecase: eu,
		logger:          slog.Default().With(slog.String("component", "event_log_handler")),
		handlerOpts:     append([]connect.HandlerOption{connect.WithCompressMinBytes(0)}, opts...),
	}
}

// NewEventLogServiceHandlerWithUsecase allows injecting a custom usecase (e.g., for testing).
func NewEventLogServiceHandlerWithUsecase(
	eu eventloguc.EventLogUsecase,
	handlerOpts ...connect.HandlerOption,
) *eventLogServiceServer {
	return &eventLogServiceServer{
		eventLogUsecase: eu,
		logger:          slog.Default().With(slog.String("component", "event_log_handler")),
		handlerOpts:     append([]connect.HandlerOption{connect.WithCompressMinBytes(0)}, handlerOpts...),
	}
}

// GetHandler returns the HTTP pattern and handler for Connect-Go server registration.
func (s *eventLogServiceServer) GetHandler() (string, http.Handler) {
	pattern, handler := eventlogconnect.NewEventLogServiceHandler(s, s.handlerOpts...)
	return pattern, handler
}

// Record implements the Record RPC, mapping from pb to domain entity and delegating to usecase.
func (s *eventLogServiceServer) Record(
	ctx context.Context,
	req *connect.Request[pb.RecordRequest],
) (*connect.Response[pb.RecordResponse], error) {
	// Map pb.Event to domain entity
	e := req.Msg.Event
	// Convert Struct props to JSON
	rawProps, err := json.Marshal(e.Props.AsMap())
	if err != nil {
		s.logger.Error("failed to marshal props", slog.Any("props", e.Props), slog.String("error", err.Error()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	domainEvt := &entity.EventLog{
		EventLogID:  uuid.New().String(),
		TraceID:     ctxkeys.TraceIDFromContext(ctx),
		UserID:      e.UserId,
		SessionID:   e.SessionId,
		VideoDmmID:  e.VideoDmmId,
		ActressIDs:  e.ActressIds,
		DirectorIDs: e.DirectorIds,
		GenreIDs:    e.GenreIds,
		MakerIDs:    e.MakerIds,
		SeriesIDs:   e.SeriesIds,
		EventType:   e.EventType,
		EventTime:   e.EventTime.AsTime(),
		Props:       rawProps,
	}
	// Delegate to usecase
	if err := s.eventLogUsecase.Record(ctx, domainEvt); err != nil {
		s.logger.Error("failed to record event", slog.String("error", err.Error()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&pb.RecordResponse{}), nil
}

// RecordBatch implements the RecordBatch RPC
func (s *eventLogServiceServer) RecordBatch(
	ctx context.Context,
	req *connect.Request[pb.RecordBatchRequest],
) (*connect.Response[pb.RecordResponse], error) {
	events := make([]*entity.EventLog, len(req.Msg.Events))
	for i, e := range req.Msg.Events {
		// Convert Struct props to JSON
		rawProps, err := json.Marshal(e.Props.AsMap())
		if err != nil {
			s.logger.Error("failed to marshal props", slog.Any("props", e.Props), slog.String("error", err.Error()))
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		events[i] = &entity.EventLog{
			EventLogID:  uuid.New().String(),
			TraceID:     ctxkeys.TraceIDFromContext(ctx),
			UserID:      e.UserId,
			SessionID:   e.SessionId,
			VideoDmmID:  e.VideoDmmId,
			ActressIDs:  e.ActressIds,
			DirectorIDs: e.DirectorIds,
			GenreIDs:    e.GenreIds,
			MakerIDs:    e.MakerIds,
			SeriesIDs:   e.SeriesIds,
			EventType:   e.EventType,
			EventTime:   e.EventTime.AsTime(),
			Props:       rawProps,
		}
	}
	// Delegate to usecase
	if err := s.eventLogUsecase.RecordBatch(ctx, events); err != nil {
		s.logger.Error("failed to record batch events", slog.String("error", err.Error()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&pb.RecordResponse{}), nil
}
