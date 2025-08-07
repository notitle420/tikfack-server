package ctxkeys

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      ContextKey
		value    interface{}
		expected interface{}
	}{
		{
			name:     "SubKey - Store and Retrieve",
			key:      SubKey,
			value:    "user-123",
			expected: "user-123",
		},
		{
			name:     "TokenKey - Store and Retrieve",
			key:      TokenKey,
			value:    "bearer-token-xyz",
			expected: "bearer-token-xyz",
		},
		{
			name:     "TraceIDKey - Store and Retrieve",
			key:      TraceIDKey,
			value:    "trace-456-abc",
			expected: "trace-456-abc",
		},
		{
			name:     "SubKey - Empty String",
			key:      SubKey,
			value:    "",
			expected: "",
		},
		{
			name:     "TokenKey - Complex Token",
			key:      TokenKey,
			value:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
			expected: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with value
			ctx := context.WithValue(context.Background(), tt.key, tt.value)
			
			// Retrieve value
			retrieved := ctx.Value(tt.key)
			
			// Assert
			assert.Equal(t, tt.expected, retrieved)
		})
	}
}

func TestContextKeys_NotFound(t *testing.T) {
	ctx := context.Background()
	
	// Test retrieving non-existent values
	assert.Nil(t, ctx.Value(SubKey))
	assert.Nil(t, ctx.Value(TokenKey))
	assert.Nil(t, ctx.Value(TraceIDKey))
}

func TestContextKeys_TypeAssertion(t *testing.T) {
	tests := []struct {
		name        string
		key         ContextKey
		value       interface{}
		assertType  func(interface{}) (string, bool)
		expectOk    bool
	}{
		{
			name:  "SubKey - Valid String",
			key:   SubKey,
			value: "user-123",
			assertType: func(v interface{}) (string, bool) {
				s, ok := v.(string)
				return s, ok
			},
			expectOk: true,
		},
		{
			name:  "TokenKey - Valid String",
			key:   TokenKey,
			value: "token-abc",
			assertType: func(v interface{}) (string, bool) {
				s, ok := v.(string)
				return s, ok
			},
			expectOk: true,
		},
		{
			name:  "TraceIDKey - Valid String",
			key:   TraceIDKey,
			value: "trace-xyz",
			assertType: func(v interface{}) (string, bool) {
				s, ok := v.(string)
				return s, ok
			},
			expectOk: true,
		},
		{
			name:  "SubKey - Invalid Type (int)",
			key:   SubKey,
			value: 123,
			assertType: func(v interface{}) (string, bool) {
				s, ok := v.(string)
				return s, ok
			},
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), tt.key, tt.value)
			retrieved := ctx.Value(tt.key)
			
			_, ok := tt.assertType(retrieved)
			assert.Equal(t, tt.expectOk, ok)
		})
	}
}

func TestContextKeys_MultipleValues(t *testing.T) {
	// Test storing multiple values in the same context
	ctx := context.Background()
	ctx = context.WithValue(ctx, SubKey, "user-123")
	ctx = context.WithValue(ctx, TokenKey, "token-456")
	ctx = context.WithValue(ctx, TraceIDKey, "trace-789")
	
	// Verify all values are retrievable
	assert.Equal(t, "user-123", ctx.Value(SubKey))
	assert.Equal(t, "token-456", ctx.Value(TokenKey))
	assert.Equal(t, "trace-789", ctx.Value(TraceIDKey))
}

func TestContextKeys_Overwrite(t *testing.T) {
	// Test overwriting values
	ctx := context.Background()
	ctx = context.WithValue(ctx, SubKey, "user-123")
	ctx = context.WithValue(ctx, SubKey, "user-456")
	
	// Verify the latest value is retrieved
	assert.Equal(t, "user-456", ctx.Value(SubKey))
}

func TestContextKeys_Constants(t *testing.T) {
	// Verify the constant values are as expected
	assert.Equal(t, ContextKey("sub"), SubKey)
	assert.Equal(t, ContextKey("token"), TokenKey)
	assert.Equal(t, ContextKey("trace_id"), TraceIDKey)
}

func TestContextKeys_UniqueKeys(t *testing.T) {
	// Verify that each key is unique
	keys := []ContextKey{SubKey, TokenKey, TraceIDKey}
	seen := make(map[ContextKey]bool)
	
	for _, key := range keys {
		assert.False(t, seen[key], "Duplicate key found: %v", key)
		seen[key] = true
	}
}

func TestContextKeys_NestedContext(t *testing.T) {
	// Test with nested contexts
	parentCtx := context.WithValue(context.Background(), SubKey, "parent-user")
	childCtx := context.WithValue(parentCtx, TokenKey, "child-token")
	grandchildCtx := context.WithValue(childCtx, TraceIDKey, "grandchild-trace")
	
	// Verify all values are accessible from grandchild context
	assert.Equal(t, "parent-user", grandchildCtx.Value(SubKey))
	assert.Equal(t, "child-token", grandchildCtx.Value(TokenKey))
	assert.Equal(t, "grandchild-trace", grandchildCtx.Value(TraceIDKey))
	
	// Verify parent context doesn't have child values
	assert.Nil(t, parentCtx.Value(TokenKey))
	assert.Nil(t, parentCtx.Value(TraceIDKey))
}