package logger

import (
	"context"
	"log/slog"

	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

// UserIDFromContext は ctx から "sub" (ユーザーID) を取り出す

// LoggerWithCtx は slog.Default() に ctx から取得したユーザーIDやトレースIDを付与して返す
func LoggerWithCtx(ctx context.Context) *slog.Logger {
    return slog.Default().With(
        slog.String("user_id",  ctxkeys.UserIDFromContext(ctx)),
        slog.String("trace_id", ctxkeys.TraceIDFromContext(ctx)),
        slog.String("token_id", ctxkeys.TokenIDFromContext(ctx)),
    )
}