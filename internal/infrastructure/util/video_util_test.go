package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValidVideoUrl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/litevideo/freepv/a/abc/abc123/abc123mhb.mp4" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	originalClient := http.DefaultClient
	http.DefaultClient = &http.Client{
		Transport: &mockTransport{
			server: server,
			validPath: "/litevideo/freepv/a/abc/abc123/abc123mhb.mp4",
		},
	}
	defer func() { http.DefaultClient = originalClient }()
	
	tests := []struct {
		name    string
		dmmID   string
		wantErr bool
	}{
		{
			name:    "Valid ID",
			dmmID:   "abc123",
			wantErr: false,
		},
		{
			name:    "Invalid ID",
			dmmID:   "invalid",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := GetValidVideoUrl(tt.dmmID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, url)
			}
		})
	}
}

type mockTransport struct {
	server    *httptest.Server
	validPath string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path == m.validPath || req.URL.Path == "/litevideo/freepv/a/abc/abc123/abc123mhb.mp4" {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       http.NoBody,
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}, nil
}
