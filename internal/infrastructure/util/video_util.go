package util

import (
	"fmt"
	"net/http"
	"regexp"
)

// GetValidVideoUrl は DMM ID から実際の動画配信URLを検証して取得します
func GetValidVideoUrl(dmmVideoId string) (string, error) {
	var reAlt0 = regexp.MustCompile(`^([A-Za-z0-9_]+?)0(\d+)([A-Za-z])?$`)
	var reAlt00 = regexp.MustCompile(`^([A-Za-z0-9_]+?)00(\d+)([A-Za-z])?$`)
	
	generateUrl := func(id string) string {
		if len(id) < 3 {
			return ""
		}
		firstChar := id[0:1]
		firstThreeChars := id[0:3]
		return fmt.Sprintf("https://cc3001.dmm.co.jp/litevideo/freepv/%s/%s/%s/%smhb.mp4", firstChar, firstThreeChars, id, id)
	}

	originalUrl := generateUrl(dmmVideoId)
	alternativeUrl0 := generateUrl(reAlt0.ReplaceAllString(dmmVideoId, "$1$2$3"))
	alternativeUrl00 := generateUrl(reAlt00.ReplaceAllString(dmmVideoId, "$1$2$3"))

	urls := []string{originalUrl, alternativeUrl0, alternativeUrl00}
        for _, url := range urls {
                resp, err := http.Head(url)
                if err == nil {
                        // HEAD レスポンスは Body を利用しないが、
                        // コネクションリークを防ぐため確実に Close する
                        resp.Body.Close()
                        if resp.StatusCode >= 200 && resp.StatusCode < 300 {
                                return url, nil
                        }
                }
        }
	return "", fmt.Errorf("有効な動画URLが見つかりませんでした: %s", dmmVideoId)
} 