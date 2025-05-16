package dmmapi

//go:generate mockgen -destination=mock_client.go -package=dmmapi github.com/tikfack/server/internal/infrastructure/dmmapi ClientInterface

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type ClientInterface interface {
    Call(path string, v interface{}) error
}

// Client は DMM API へのリクエストを行う
type Client struct {
    BaseURL     string
    APIID       string
    AffiliateID string
    HTTPClient  *http.Client
}

// NewClient 環境変数から設定を読み込み、新規 Client を返す
func NewClient() (*Client, error) {
    base := os.Getenv("BASE_URL")
    id := os.Getenv("DMM_API_ID")
    aff := os.Getenv("DMM_API_AFFILIATE_ID")
    if base == "" || id == "" || aff == "" {
        return nil, fmt.Errorf("DMM API credentials not set")
    }
    return &Client{
        BaseURL:     base,
        APIID:       id,
        AffiliateID: aff,
        HTTPClient:  http.DefaultClient,
    }, nil
}

// Call makes a GET request to the specified path and unmarshals into v
func (c *Client) Call(path string, v interface{}) error {
    url := fmt.Sprintf("%s%s&api_id=%s&affiliate_id=%s&output=json", c.BaseURL, path, c.APIID, c.AffiliateID)
    resp, err := c.HTTPClient.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    log.Println(string(body))
    return json.Unmarshal(body, v)
}