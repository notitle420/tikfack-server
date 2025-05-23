package ctxkeys

type ContextKey string

const (
	SubKey     ContextKey = "sub"
	TokenKey   ContextKey = "token"
	TraceIDKey ContextKey = "trace_id"
)
