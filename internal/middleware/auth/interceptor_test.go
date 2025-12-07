package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bufbuild/connect-go"
	gocloak "github.com/mviniciusgc/gocloak/v13"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tikfack/server/internal/middleware/auth/mock"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

// mockIDToken implements IDTokenInterface for testing
type mockIDToken struct {
	sub string
	err error
}

func (m *mockIDToken) Claims(v interface{}) error {
	if m.err != nil {
		return m.err
	}
	if claims, ok := v.(*struct {
		Sub string `json:"sub"`
	}); ok {
		claims.Sub = m.sub
	}
	return nil
}

// mockTokenVerifier implements IDTokenVerifierInterface for testing
type mockTokenVerifier struct {
	token *mockIDToken
	err   error
}

func (tv *mockTokenVerifier) Verify(ctx context.Context, rawIDToken string) (mock.IDTokenInterface, error) {
	if tv.err != nil {
		return nil, tv.err
	}
	return tv.token, nil
}

// mockGocloakClient implements GocloakClientInterface for testing
type mockGocloakClient struct {
	introspectResult *gocloak.IntroSpectTokenResult
	introspectError  error
	rptResult        *gocloak.JWT
	rptError         error
}

func (m *mockGocloakClient) RetrospectToken(ctx context.Context, accessToken, clientID, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error) {
	return m.introspectResult, m.introspectError
}

func (m *mockGocloakClient) GetRequestingPartyToken(ctx context.Context, token, realm string, options gocloak.RequestingPartyTokenOptions) (*gocloak.JWT, error) {
	return m.rptResult, m.rptError
}

// mockRequest implements connect.AnyRequest for testing
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

func TestIntrospectionInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		setupVerifier func() mock.IDTokenVerifierInterface
		setupClient   func() mock.GocloakClientInterface
		expectError   bool
		expectedCode  connect.Code
		expectedSub   string
		expectedToken string
	}{
		{
			name:       "Valid Active Token",
			authHeader: "Bearer valid-token",
			setupVerifier: func() mock.IDTokenVerifierInterface {
				return &mockTokenVerifier{
					token: &mockIDToken{sub: "test-sub"},
				}
			},
			setupClient: func() mock.GocloakClientInterface {
				activeTrue := true
				return &mockGocloakClient{
					introspectResult: &gocloak.IntroSpectTokenResult{Active: &activeTrue},
				}
			},
			expectError:   false,
			expectedSub:   "test-sub",
			expectedToken: "valid-token",
		},
		{
			name:       "Missing Authorization Header",
			authHeader: "",
			setupVerifier: func() mock.IDTokenVerifierInterface {
				return &mockTokenVerifier{}
			},
			setupClient: func() mock.GocloakClientInterface {
				return &mockGocloakClient{}
			},
			expectError:   false,
			expectedSub:   "",
			expectedToken: "",
		},
		{
			name:       "Invalid Authorization Format",
			authHeader: "InvalidFormat",
			setupVerifier: func() mock.IDTokenVerifierInterface {
				return &mockTokenVerifier{}
			},
			setupClient: func() mock.GocloakClientInterface {
				return &mockGocloakClient{}
			},
			expectError:  true,
			expectedCode: connect.CodeUnauthenticated,
		},
		{
			name:       "Token Verification Failed",
			authHeader: "Bearer invalid-token",
			setupVerifier: func() mock.IDTokenVerifierInterface {
				return &mockTokenVerifier{
					err: errors.New("verification failed"),
				}
			},
			setupClient: func() mock.GocloakClientInterface {
				return &mockGocloakClient{}
			},
			expectError:  true,
			expectedCode: connect.CodeUnauthenticated,
		},
		{
			name:       "Claims Extraction Failed",
			authHeader: "Bearer valid-token",
			setupVerifier: func() mock.IDTokenVerifierInterface {
				return &mockTokenVerifier{
					token: &mockIDToken{err: errors.New("claims failed")},
				}
			},
			setupClient: func() mock.GocloakClientInterface {
				return &mockGocloakClient{}
			},
			expectError:  true,
			expectedCode: connect.CodeUnauthenticated,
		},
		{
			name:       "Introspection Failed",
			authHeader: "Bearer valid-token",
			setupVerifier: func() mock.IDTokenVerifierInterface {
				return &mockTokenVerifier{
					token: &mockIDToken{sub: "test-sub"},
				}
			},
			setupClient: func() mock.GocloakClientInterface {
				return &mockGocloakClient{
					introspectError: errors.New("introspection failed"),
				}
			},
			expectError:  true,
			expectedCode: connect.CodeUnauthenticated,
		},
		{
			name:       "Token Not Active",
			authHeader: "Bearer valid-token",
			setupVerifier: func() mock.IDTokenVerifierInterface {
				return &mockTokenVerifier{
					token: &mockIDToken{sub: "test-sub"},
				}
			},
			setupClient: func() mock.GocloakClientInterface {
				activeFalse := false
				return &mockGocloakClient{
					introspectResult: &gocloak.IntroSpectTokenResult{Active: &activeFalse},
				}
			},
			expectError:  true,
			expectedCode: connect.CodeUnauthenticated,
		},
		{
			name:       "Token Active is nil",
			authHeader: "Bearer valid-token",
			setupVerifier: func() mock.IDTokenVerifierInterface {
				return &mockTokenVerifier{
					token: &mockIDToken{sub: "test-sub"},
				}
			},
			setupClient: func() mock.GocloakClientInterface {
				return &mockGocloakClient{
					introspectResult: &gocloak.IntroSpectTokenResult{Active: nil},
				}
			},
			expectError:  true,
			expectedCode: connect.CodeUnauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := tt.setupVerifier()
			client := tt.setupClient()

			interceptor := IntrospectionInterceptorWithInterfaces(
				verifier,
				client,
				"test-realm",
				"test-client",
				"test-secret",
			)

			nextCalled := false
			var ctxSub, ctxToken string
			next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				nextCalled = true
				if v := ctx.Value(ctxkeys.SubKey); v != nil {
					ctxSub = v.(string)
				}
				if v := ctx.Value(ctxkeys.TokenKey); v != nil {
					ctxToken = v.(string)
				}
				return nil, nil
			}

			req := &mockRequest{
				header: http.Header{
					"Authorization": []string{tt.authHeader},
				},
			}

			unaryFunc := interceptor.WrapUnary(next)
			_, err := unaryFunc(context.Background(), req)

			if tt.expectError {
				require.Error(t, err)
				connectErr, ok := err.(*connect.Error)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, connectErr.Code())
				assert.False(t, nextCalled)
			} else {
				require.NoError(t, err)
				assert.True(t, nextCalled)
				assert.Equal(t, tt.expectedSub, ctxSub)
				assert.Equal(t, tt.expectedToken, ctxToken)
			}
		})
	}
}

func TestCheckPermissionFunc(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func() *mockGocloakClient
		expectError bool
	}{
		{
			name: "Valid RPT Token",
			setupMock: func() *mockGocloakClient {
				return &mockGocloakClient{
					rptResult: &gocloak.JWT{AccessToken: "valid-rpt-token"},
				}
			},
			expectError: false,
		},
		{
			name: "RPT Request Failed",
			setupMock: func() *mockGocloakClient {
				return &mockGocloakClient{
					rptError: errors.New("rpt request failed"),
				}
			},
			expectError: true,
		},
		{
			name: "Empty RPT Token",
			setupMock: func() *mockGocloakClient {
				return &mockGocloakClient{
					rptResult: &gocloak.JWT{AccessToken: ""},
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.setupMock()

			err := CheckPermissionFuncWithInterface(
				context.Background(),
				mockClient,
				"user-token",
				"test-resource",
				"test-realm",
				"test-client",
			)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
