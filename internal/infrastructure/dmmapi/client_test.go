package dmmapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	os.Setenv("BASE_URL", "http://example.com")
	os.Setenv("DMM_API_ID", "id")
	os.Setenv("DMM_API_AFFILIATE_ID", "aff")
	c, err := NewClient()
	require.NoError(t, err)
	require.Equal(t, "http://example.com", c.BaseURL)
	require.Equal(t, "id", c.APIID)
	require.Equal(t, "aff", c.AffiliateID)
}

func TestNewClientMissingEnv(t *testing.T) {
	os.Unsetenv("BASE_URL")
	os.Unsetenv("DMM_API_ID")
	os.Unsetenv("DMM_API_AFFILIATE_ID")
	_, err := NewClient()
	require.Error(t, err)
}

func TestClientCall(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "api_id=id") || !strings.Contains(r.URL.RawQuery, "affiliate_id=aff") {
			t.Fatalf("missing query params: %s", r.URL.RawQuery)
		}
		io.WriteString(w, `{"key":"value"}`)
	}))
	defer ts.Close()

	c := &Client{BaseURL: ts.URL, APIID: "id", AffiliateID: "aff", HTTPClient: ts.Client()}
	var v struct {
		Key string `json:"key"`
	}
	err := c.Call("/path?x=1", &v)
	require.NoError(t, err)
	require.Equal(t, "value", v.Key)
}
