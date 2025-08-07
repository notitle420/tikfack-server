package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

func TestUserIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected string
	}{
		{
			name: "Valid User ID",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.SubKey, "user-123")
			},
			expected: "user-123",
		},
		{
			name: "Empty User ID",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.SubKey, "")
			},
			expected: "",
		},
		{
			name: "No User ID in Context",
			setup: func() context.Context {
				return context.Background()
			},
			expected: "",
		},
		{
			name: "Wrong Type in Context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.SubKey, 123)
			},
			expected: "",
		},
		{
			name: "Nil Value in Context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.SubKey, nil)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := UserIDFromContext(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTraceIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected string
	}{
		{
			name: "Valid Trace ID",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.TraceIDKey, "trace-abc-123")
			},
			expected: "trace-abc-123",
		},
		{
			name: "UUID Format Trace ID",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.TraceIDKey, "550e8400-e29b-41d4-a716-446655440000")
			},
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name: "Empty Trace ID",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.TraceIDKey, "")
			},
			expected: "",
		},
		{
			name: "No Trace ID in Context",
			setup: func() context.Context {
				return context.Background()
			},
			expected: "",
		},
		{
			name: "Wrong Type in Context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.TraceIDKey, 456)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := TraceIDFromContext(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected string
	}{
		{
			name: "Valid Token - Long",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.TokenKey, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.payload.signature")
			},
			expected: "eyJhbGciOi",
		},
		{
			name: "Valid Token - Exact 10 chars",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.TokenKey, "1234567890")
			},
			expected: "1234567890",
		},
		{
			name: "Valid Token - Less than 10 chars",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.TokenKey, "short")
			},
			expected: "",
		},
		{
			name: "Empty Token",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.TokenKey, "")
			},
			expected: "",
		},
		{
			name: "No Token in Context",
			setup: func() context.Context {
				return context.Background()
			},
			expected: "",
		},
		{
			name: "Wrong Type in Context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.TokenKey, 789)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := TokenIDFromContext(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoggerWithCtx(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() context.Context
		checkLogs func(t *testing.T, logs string)
	}{
		{
			name: "All Values Present",
			setup: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, ctxkeys.SubKey, "user-123")
				ctx = context.WithValue(ctx, ctxkeys.TraceIDKey, "trace-456")
				ctx = context.WithValue(ctx, ctxkeys.TokenKey, "token-789-abcdef")
				return ctx
			},
			checkLogs: func(t *testing.T, logs string) {
				assert.Contains(t, logs, "user_id=user-123")
				assert.Contains(t, logs, "trace_id=trace-456")
				assert.Contains(t, logs, "token_id=token-789-")
				assert.Contains(t, logs, "test message")
			},
		},
		{
			name: "Partial Values Present",
			setup: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, ctxkeys.SubKey, "user-456")
				// No trace ID
				// No token
				return ctx
			},
			checkLogs: func(t *testing.T, logs string) {
				assert.Contains(t, logs, "user_id=user-456")
				assert.Contains(t, logs, "trace_id=\"\"")
				assert.Contains(t, logs, "token_id=\"\"")
				assert.Contains(t, logs, "test message")
			},
		},
		{
			name: "No Values in Context",
			setup: func() context.Context {
				return context.Background()
			},
			checkLogs: func(t *testing.T, logs string) {
				assert.Contains(t, logs, "user_id=\"\"")
				assert.Contains(t, logs, "trace_id=\"\"")
				assert.Contains(t, logs, "token_id=\"\"")
				assert.Contains(t, logs, "test message")
			},
		},
		{
			name: "Token Too Short",
			setup: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, ctxkeys.TokenKey, "short")
				return ctx
			},
			checkLogs: func(t *testing.T, logs string) {
				assert.Contains(t, logs, "token_id=\"\"")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture logs
			var buf bytes.Buffer
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})
			oldLogger := slog.Default()
			slog.SetDefault(slog.New(handler))
			defer slog.SetDefault(oldLogger)

			ctx := tt.setup()
			logger := LoggerWithCtx(ctx)
			
			// Log a test message
			logger.Info("test message", "extra_field", "extra_value")

			// Check logs
			logs := buf.String()
			tt.checkLogs(t, logs)
			assert.Contains(t, logs, "extra_field=extra_value")
		})
	}
}

func TestLoggerWithCtx_MultipleLogLevels(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(oldLogger)

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.SubKey, "user-test")
	ctx = context.WithValue(ctx, ctxkeys.TraceIDKey, "trace-test")
	ctx = context.WithValue(ctx, ctxkeys.TokenKey, "token-test-123456")

	logger := LoggerWithCtx(ctx)

	// Test different log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message", "error_code", 500)

	logs := buf.String()
	
	// Verify all log levels are present
	assert.Contains(t, logs, "level=DEBUG")
	assert.Contains(t, logs, "level=INFO")
	assert.Contains(t, logs, "level=WARN")
	assert.Contains(t, logs, "level=ERROR")
	
	// Verify context values are in all logs
	lines := strings.Split(logs, "\n")
	for _, line := range lines {
		if line != "" {
			assert.Contains(t, line, "user_id=user-test")
			assert.Contains(t, line, "trace_id=trace-test")
			assert.Contains(t, line, "token_id=token-test")
		}
	}
}

func TestLoggerWithCtx_ConcurrentAccess(t *testing.T) {
	// Test that LoggerWithCtx is safe for concurrent use
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.SubKey, "concurrent-user")
	ctx = context.WithValue(ctx, ctxkeys.TraceIDKey, "concurrent-trace")

	logger := LoggerWithCtx(ctx)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			logger.Info("concurrent log", "goroutine", id)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestTokenIDFromContext_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "Token with exactly 10 characters",
			token:    "0123456789",
			expected: "0123456789",
		},
		{
			name:     "Token with 11 characters",
			token:    "01234567890",
			expected: "0123456789",
		},
		{
			name:     "Token with 9 characters",
			token:    "012345678",
			expected: "",
		},
		{
			name:     "Token with special characters",
			token:    "!@#$%^&*()_+-=",
			expected: "!@#$%^&*()",
		},
		{
			name:     "Token with ASCII only - exactly 10",
			token:    "0123456789",
			expected: "0123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), ctxkeys.TokenKey, tt.token)
			result := TokenIDFromContext(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}