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
		} else if r.URL.Path == "/litevideo/freepv/a/abc/abc123/abc123mhb.mp4" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	originalGenerateUrl := generateUrl
	defer func() { generateUrl = originalGenerateUrl }()
	
	generateUrl = func(id string) string {
		if id == "abc123" {
			return server.URL + "/litevideo/freepv/a/abc/abc123/abc123mhb.mp4"
		}
		return server.URL + "/not-found"
	}
	
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
