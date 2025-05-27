package ctxkeys

import "context"

type ContextKey string

const (
	SubKey     ContextKey = "sub"
	TokenKey   ContextKey = "token"
	TraceIDKey ContextKey = "trace_id"
)

func UserIDFromContext(ctx context.Context) string {
    v := ctx.Value(SubKey)
    if userID, ok := v.(string); ok {
        return userID
    }
    return ""
}

// TraceIDFromContext は ctx から "trace_id" を取り出す
func TraceIDFromContext(ctx context.Context) string {
    v := ctx.Value(TraceIDKey)
    if traceID, ok := v.(string); ok {
        return traceID
    }
    return ""
}
func TokenIDFromContext(ctx context.Context) string {
    v := ctx.Value(TokenKey)
    if TokenKey, ok := v.(string); ok {
        return TokenKey[0:10]
    }
    return ""
}