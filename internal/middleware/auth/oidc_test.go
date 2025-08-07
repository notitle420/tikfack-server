package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerifier(t *testing.T) {
	tests := []struct {
		name            string
		setupServer     func() *httptest.Server
		clientID        string
		expectError     bool
		errorContains   string
	}{
		{
			name: "Success - Valid OIDC Provider",
			setupServer: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/.well-known/openid-configuration" {
						issuerURL := "http://" + r.Host
						config := map[string]interface{}{
							"issuer":                 issuerURL,
							"authorization_endpoint": issuerURL + "/auth",
							"token_endpoint":         issuerURL + "/token",
							"userinfo_endpoint":      issuerURL + "/userinfo",
							"jwks_uri":               issuerURL + "/jwks",
							"response_types_supported": []string{"code", "token", "id_token"},
							"subject_types_supported":  []string{"public"},
							"id_token_signing_alg_values_supported": []string{"RS256"},
						}
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(config)
					} else if r.URL.Path == "/jwks" {
						jwks := map[string]interface{}{
							"keys": []map[string]interface{}{
								{
									"kty": "RSA",
									"use": "sig",
									"kid": "test-key",
									"alg": "RS256",
									"n":   "test-n",
									"e":   "AQAB",
								},
							},
						}
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(jwks)
					}
				}))
				return server
			},
			clientID:    "test-client",
			expectError: false,
		},
		{
			name: "Error - Provider Not Found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			clientID:      "test-client",
			expectError:   true,
			errorContains: "failed to create OIDC provider",
		},
		{
			name: "Error - Invalid JSON Response",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/.well-known/openid-configuration" {
						w.Header().Set("Content-Type", "application/json")
						w.Write([]byte("invalid json"))
					}
				}))
			},
			clientID:      "test-client",
			expectError:   true,
			errorContains: "failed to create OIDC provider",
		},
		{
			name: "Error - Server Error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			clientID:      "test-client",
			expectError:   true,
			errorContains: "failed to create OIDC provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			ctx := context.Background()
			verifier, err := NewVerifier(ctx, server.URL, tt.clientID)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, verifier)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, verifier)
			}
		})
	}
}

func TestNewVerifier_InvalidURL(t *testing.T) {
	ctx := context.Background()
	
	// Test with invalid URL
	verifier, err := NewVerifier(ctx, "://invalid-url", "test-client")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create OIDC provider")
	assert.Nil(t, verifier)
	
	// Test with unreachable URL
	verifier, err = NewVerifier(ctx, "http://localhost:99999", "test-client")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create OIDC provider")
	assert.Nil(t, verifier)
}

func TestNewVerifier_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		<-r.Context().Done()
	}))
	defer server.Close()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start the verifier creation in a goroutine
	done := make(chan struct {
		verifier interface{}
		err      error
	}, 1)
	
	go func() {
		v, e := NewVerifier(ctx, server.URL, "test-client")
		done <- struct {
			verifier interface{}
			err      error
		}{v, e}
	}()
	
	// Cancel the context
	cancel()
	
	// Wait for the result
	result := <-done
	require.Error(t, result.err)
	assert.Contains(t, result.err.Error(), "failed to create OIDC provider")
	assert.Nil(t, result.verifier)
}

func TestNewVerifier_EmptyClientID(t *testing.T) {
	// Setup a valid OIDC provider server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			issuerURL := "http://" + r.Host
			config := map[string]interface{}{
				"issuer":                 issuerURL,
				"authorization_endpoint": issuerURL + "/auth",
				"token_endpoint":         issuerURL + "/token",
				"userinfo_endpoint":      issuerURL + "/userinfo",
				"jwks_uri":               issuerURL + "/jwks",
				"response_types_supported": []string{"code", "token", "id_token"},
				"subject_types_supported":  []string{"public"},
				"id_token_signing_alg_values_supported": []string{"RS256"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(config)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	
	// Test with empty client ID
	verifier, err := NewVerifier(ctx, server.URL, "")
	require.NoError(t, err)
	assert.NotNil(t, verifier)
}