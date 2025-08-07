package logger

import (
	"context"
	"log/slog"

	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

// UserIDFromContext は ctx から "sub" (ユーザーID) を取り出す
<<<<<<< HEAD
=======
func UserIDFromContext(ctx context.Context) string {
    v := ctx.Value(ctxkeys.SubKey)
    if userID, ok := v.(string); ok {
        return userID
    }
    return ""
}

// TraceIDFromContext は ctx から "trace_id" を取り出す
func TraceIDFromContext(ctx context.Context) string {
    v := ctx.Value(ctxkeys.TraceIDKey)
    if traceID, ok := v.(string); ok {
        return traceID
    }
    return ""
}
func TokenIDFromContext(ctx context.Context) string {
    v := ctx.Value(ctxkeys.TokenKey)
    if TokenKey, ok := v.(string); ok {
        if len(TokenKey) >= 10 {
            return TokenKey[0:10]
        }
        return ""
    }
    return ""
}
>>>>>>> dev

// LoggerWithCtx は slog.Default() に ctx から取得したユーザーIDやトレースIDを付与して返す
func LoggerWithCtx(ctx context.Context) *slog.Logger {
    return slog.Default().With(
        slog.String("user_id",  ctxkeys.UserIDFromContext(ctx)),
        slog.String("trace_id", ctxkeys.TraceIDFromContext(ctx)),
        slog.String("token_id", ctxkeys.TokenIDFromContext(ctx)),
    )
}