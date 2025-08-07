package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/tikfack/server/internal/middleware/auth/mock"
)

func createMockResponse(statusCode int, body string) *http.Response {
	status := http.StatusText(statusCode)
	if statusCode != 0 {
		status = fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	}
	return &http.Response{
		StatusCode: statusCode,
		Status:     status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestCallKeycloakPermissionAPIWithClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		setupMock      func() *mock.MockDoer
		expectError    bool
		expectedError  string
		
	}{
		{
			name: "Success - Permission Granted",
			setupMock: func() *mock.MockDoer {
				mockDoer := mock.NewMockDoer(ctrl)
				mockDoer.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					// Verify request details
					assert.Equal(t, "POST", req.Method)
					assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
					assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
					
					// Verify form data
					body, err := io.ReadAll(req.Body)
					require.NoError(t, err)
					formData := string(body)
					assert.Contains(t, formData, "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Auma-ticket")
					assert.Contains(t, formData, "audience=test-audience")
					assert.Contains(t, formData, "permission=test-resource")
					
					return createMockResponse(http.StatusOK, `{"access_token": "rpt-token"}`), nil
				})
				return mockDoer
			},
			expectError: false,
		},
		{
			name: "Permission Denied - 403",
			setupMock: func() *mock.MockDoer {
				mockDoer := mock.NewMockDoer(ctrl)
				mockDoer.EXPECT().Do(gomock.Any()).Return(
					createMockResponse(http.StatusForbidden, `{"error": "permission_denied"}`),
					nil,
				)
				return mockDoer
			},
			expectError:   true,
			expectedError: "permission denied (403 from Keycloak)",
		},
		{
			name: "Unauthorized - 401",
			setupMock: func() *mock.MockDoer {
				mockDoer := mock.NewMockDoer(ctrl)
				mockDoer.EXPECT().Do(gomock.Any()).Return(
					createMockResponse(http.StatusUnauthorized, `{"error": "unauthorized"}`),
					nil,
				)
				return mockDoer
			},
			expectError:   true,
			expectedError: "unexpected status from keycloak: 401 Unauthorized",
		},
		{
			name: "Internal Server Error - 500",
			setupMock: func() *mock.MockDoer {
				mockDoer := mock.NewMockDoer(ctrl)
				mockDoer.EXPECT().Do(gomock.Any()).Return(
					createMockResponse(http.StatusInternalServerError, `{"error": "internal_error"}`),
					nil,
				)
				return mockDoer
			},
			expectError:   true,
			expectedError: "unexpected status from keycloak: 500 Internal Server Error",
		},
		{
			name: "Bad Request - 400",
			setupMock: func() *mock.MockDoer {
				mockDoer := mock.NewMockDoer(ctrl)
				mockDoer.EXPECT().Do(gomock.Any()).Return(
					createMockResponse(http.StatusBadRequest, `{"error": "bad_request"}`),
					nil,
				)
				return mockDoer
			},
			expectError:   true,
			expectedError: "unexpected status from keycloak: 400 Bad Request",
		},
		{
			name: "Network Error",
			setupMock: func() *mock.MockDoer {
				mockDoer := mock.NewMockDoer(ctrl)
				mockDoer.EXPECT().Do(gomock.Any()).Return(
					nil,
					errors.New("network error"),
				)
				return mockDoer
			},
			expectError:   true,
			expectedError: "network error",
		},
		{
			name: "HTTP Client Do Error",
			setupMock: func() *mock.MockDoer {
				mockDoer := mock.NewMockDoer(ctrl)
				mockDoer.EXPECT().Do(gomock.Any()).Return(
					nil,
					errors.New("client error"),
				)
				return mockDoer
			},
			expectError:   true,
			expectedError: "client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.setupMock()
			
			// Test the function
			ctx := context.Background()
			err := CallKeycloakPermissionAPIWithClient(
				ctx,
				mockClient,
				"http://keycloak.example.com/token",
				"test-audience",
				"test-token",
				"test-resource",
			)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				
				// No need for manual request verification with gomock DoAndReturn
			}
		})
	}
}

func TestCallKeycloakPermissionAPIWithClient_RequestCreationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockClient := mock.NewMockDoer(ctrl)
	// No expectation set as the function should fail before calling Do()
	
	// Create a context that will cause NewRequestWithContext to fail
	// Use an invalid URL to cause the error
	ctx := context.Background()
	err := CallKeycloakPermissionAPIWithClient(
		ctx,
		mockClient,
		"://invalid-url", // Invalid URL will cause http.NewRequestWithContext to fail
		"test-audience",
		"test-token",
		"test-resource",
	)
	
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")
}

func TestCallKeycloakPermissionAPIWithClient_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockClient := mock.NewMockDoer(ctrl)
	mockClient.EXPECT().Do(gomock.Any()).Return(
		createMockResponse(http.StatusOK, "{}"),
		nil,
	)
	
	// Create context that is already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	err := CallKeycloakPermissionAPIWithClient(
		ctx,
		mockClient,
		"http://keycloak.example.com/token",
		"test-audience",
		"test-token",
		"test-resource",
	)
	
	// The function should still work as the context cancellation
	// only affects the HTTP request, not the function setup
	require.NoError(t, err)
}

func TestCallKeycloakPermissionAPIWithClient_FormDataValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockClient := mock.NewMockDoer(ctrl)
	mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
		// Verify that special characters are properly URL encoded
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		formData := string(body)
		
		// The & should be URL encoded as %26
		assert.Contains(t, formData, "audience=special-audience%26test%3D1")
		assert.Contains(t, formData, "permission=special-resource%26test%3D3")
		
		return createMockResponse(http.StatusOK, "{}"), nil
	})
	
	ctx := context.Background()
	err := CallKeycloakPermissionAPIWithClient(
		ctx,
		mockClient,
		"http://keycloak.example.com/token",
		"special-audience&test=1",
		"special-token&test=2",
		"special-resource&test=3",
	)
	
	require.NoError(t, err)
}

func TestCallKeycloakPermissionAPI_UsesWrapper(t *testing.T) {
	// This test verifies that the original function uses the wrapper correctly
	// We can't easily mock http.DefaultClient, so we'll just verify the function exists
	// and can be called without panicking
	ctx := context.Background()
	
	// This will fail with a network error since we're using a fake URL,
	// but it should not panic and should return an error
	err := CallKeycloakPermissionAPI(
		ctx,
		"http://nonexistent.example.com:99999/token",
		"test-audience",
		"test-token",
		"test-resource",
	)
	
	// Should get a network error, not a panic
	require.Error(t, err)
}