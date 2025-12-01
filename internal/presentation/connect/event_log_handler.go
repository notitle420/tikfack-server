package connect

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/bufbuild/connect-go"

	pb "github.com/tikfack/server/gen/event_log"
	eventlogconnect "github.com/tikfack/server/gen/event_log/event_logconnect"
	eventloguc "github.com/tikfack/server/internal/application/usecase/event_log"
)

// EventLogServiceServer はイベントログ用のConnectハンドラ実装。
type EventLogServiceServer struct {
	eventLogUsecase eventloguc.EventLogUsecase
	presenter       eventLogPresenter
	logger          *slog.Logger
	handlerOpts     []connect.HandlerOption
}

// NewEventLogServiceHandler はユースケースを受け取り Connect ハンドラを構築する。
func NewEventLogServiceHandler(uc eventloguc.EventLogUsecase, opts ...connect.HandlerOption) *EventLogServiceServer {
	return newEventLogServiceServer(uc, nil, opts...)
}

func newEventLogServiceServer(uc eventloguc.EventLogUsecase, presenter eventLogPresenter, opts ...connect.HandlerOption) *EventLogServiceServer {
	if uc == nil {
		panic("event log usecase must be provided")
	}
	if presenter == nil {
		presenter = newEventLogPresenter()
	}
	return &EventLogServiceServer{
		eventLogUsecase: uc,
		presenter:       presenter,
		logger:          slog.Default().With(slog.String("component", "event_log_handler")),
		handlerOpts:     append([]connect.HandlerOption{connect.WithCompressMinBytes(0)}, opts...),
	}
}

// GetHandler は Connect サービスのパターンとハンドラーを返します。
func (s *EventLogServiceServer) GetHandler() (string, http.Handler) {
	pattern, handler := eventlogconnect.NewEventLogServiceHandler(s, s.handlerOpts...)
	return pattern, handler
}

// Record は単一イベントログを処理する。
func (s *EventLogServiceServer) Record(
	ctx context.Context,
	req *connect.Request[pb.RecordRequest],
) (*connect.Response[pb.RecordResponse], error) {
	domainEvent, err := s.presenter.ToDomain(ctx, req.Msg.Event)
	if err != nil {
		s.logger.Error("failed to marshal props", slog.String("error", err.Error()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := s.eventLogUsecase.Record(ctx, domainEvent); err != nil {
		s.logger.Error("failed to record event", slog.String("error", err.Error()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&pb.RecordResponse{}), nil
}

// RecordBatch は複数イベントを一括処理する。
func (s *EventLogServiceServer) RecordBatch(
	ctx context.Context,
	req *connect.Request[pb.RecordBatchRequest],
) (*connect.Response[pb.RecordResponse], error) {
	events, err := s.presenter.ToDomainBatch(ctx, req.Msg.Events)
	if err != nil {
		s.logger.Error("failed to marshal props", slog.String("error", err.Error()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := s.eventLogUsecase.RecordBatch(ctx, events); err != nil {
		s.logger.Error("failed to record batch events", slog.String("error", err.Error()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&pb.RecordResponse{}), nil
}
