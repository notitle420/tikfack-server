package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

type mockRequest struct {
	connect.AnyRequest
	header http.Header
	spec   connect.Spec
}

func (m *mockRequest) Header() http.Header {
	return m.header
}

func (m *mockRequest) Spec() connect.Spec {
	return m.spec
}

type mockResponse struct {
	connect.AnyResponse
}

func TestLoggingInterceptor(t *testing.T) {
	tests := []struct {
		name            string
		userID          string
		nextError       error
		expectError     bool
		checkLogs       func(t *testing.T, logs string)
	}{
		{
			name:        "Success - With User ID",
			userID:      "user-123",
			nextError:   nil,
			expectError: false,
			checkLogs: func(t *testing.T, logs string) {
				assert.Contains(t, logs, "request started")
				assert.Contains(t, logs, "request completed")
				assert.Contains(t, logs, "user_id=user-123")
				assert.Contains(t, logs, "trace_id=")
				assert.NotContains(t, logs, "request error")
			},
		},
		{
			name:        "Success - Without User ID",
			userID:      "",
			nextError:   nil,
			expectError: false,
			checkLogs: func(t *testing.T, logs string) {
				assert.Contains(t, logs, "request started")
				assert.Contains(t, logs, "request completed")
				assert.Contains(t, logs, "trace_id=")
				assert.NotContains(t, logs, "request error")
			},
		},
		{
			name:        "Error - Request Failed",
			userID:      "user-456",
			nextError:   errors.New("processing failed"),
			expectError: true,
			checkLogs: func(t *testing.T, logs string) {
				assert.Contains(t, logs, "request started")
				assert.Contains(t, logs, "request error")
				assert.Contains(t, logs, "processing failed")
				assert.Contains(t, logs, "user_id=user-456")
				assert.Contains(t, logs, "trace_id=")
				assert.NotContains(t, logs, "request completed")
			},
		},
		{
			name:        "Error - Connect Error",
			userID:      "user-789",
			nextError:   connect.NewError(connect.CodeInternal, errors.New("internal error")),
			expectError: true,
			checkLogs: func(t *testing.T, logs string) {
				assert.Contains(t, logs, "request started")
				assert.Contains(t, logs, "request error")
				assert.Contains(t, logs, "internal error")
				assert.Contains(t, logs, "user_id=user-789")
				assert.NotContains(t, logs, "request completed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture logs
			var buf bytes.Buffer
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
			oldLogger := slog.Default()
			slog.SetDefault(slog.New(handler))
			defer slog.SetDefault(oldLogger)

			interceptor := LoggingInterceptor()

			nextCalled := false
			var capturedTraceID string
			next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				nextCalled = true
				
				// Capture trace ID from context
				if v := ctx.Value(ctxkeys.TraceIDKey); v != nil {
					capturedTraceID = v.(string)
				}
				
				if tt.nextError != nil {
					return nil, tt.nextError
				}
				return &mockResponse{}, nil
			}

			// Prepare context
			ctx := context.Background()
			if tt.userID != "" {
				ctx = context.WithValue(ctx, ctxkeys.SubKey, tt.userID)
			}

			req := &mockRequest{}

			unaryFunc := interceptor.WrapUnary(next)
			resp, err := unaryFunc(ctx, req)

			// Verify
			assert.True(t, nextCalled)
			
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}

			// Verify trace ID was generated
			assert.NotEmpty(t, capturedTraceID)
			
			// Check UUID format (simple check)
			assert.Equal(t, 36, len(capturedTraceID)) // UUID v4 length with hyphens
			assert.Contains(t, capturedTraceID, "-")

			// Check logs
			logs := buf.String()
			tt.checkLogs(t, logs)
		})
	}
}

func TestLoggingInterceptor_TraceIDFormat(t *testing.T) {
	interceptor := LoggingInterceptor()

	var traceIDs []string
	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if v := ctx.Value(ctxkeys.TraceIDKey); v != nil {
			traceIDs = append(traceIDs, v.(string))
		}
		return &mockResponse{}, nil
	}

	unaryFunc := interceptor.WrapUnary(next)

	// Generate multiple trace IDs
	for i := 0; i < 5; i++ {
		_, err := unaryFunc(context.Background(), &mockRequest{})
		require.NoError(t, err)
	}

	// Verify all trace IDs are unique
	assert.Equal(t, 5, len(traceIDs))
	seen := make(map[string]bool)
	for _, id := range traceIDs {
		assert.False(t, seen[id], "Duplicate trace ID found: %s", id)
		seen[id] = true
		
		// Verify UUID format
		assert.Equal(t, 36, len(id))
		parts := strings.Split(id, "-")
		assert.Equal(t, 5, len(parts))
	}
}

func TestLoggingInterceptor_ContextPropagation(t *testing.T) {
	interceptor := LoggingInterceptor()

	var capturedCtx context.Context
	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		capturedCtx = ctx
		return &mockResponse{}, nil
	}

	// Create context with existing values
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.SubKey, "existing-user")
	ctx = context.WithValue(ctx, "custom-key", "custom-value")

	unaryFunc := interceptor.WrapUnary(next)
	_, err := unaryFunc(ctx, &mockRequest{})
	require.NoError(t, err)

	// Verify existing values are preserved
	assert.Equal(t, "existing-user", capturedCtx.Value(ctxkeys.SubKey))
	assert.Equal(t, "custom-value", capturedCtx.Value("custom-key"))
	
	// Verify trace ID was added
	assert.NotNil(t, capturedCtx.Value(ctxkeys.TraceIDKey))
}

func TestLoggingInterceptor_PanicRecovery(t *testing.T) {
	// Note: This test is to ensure the interceptor doesn't panic
	// In production, you might want to add panic recovery
	
	interceptor := LoggingInterceptor()

	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Simulate a panic in the handler
		panic("test panic")
	}

	unaryFunc := interceptor.WrapUnary(next)
	
	// This should panic as the interceptor doesn't handle panics
	assert.Panics(t, func() {
		_, _ = unaryFunc(context.Background(), &mockRequest{})
	})
}