package dmmapi

//go:generate mockgen -destination=mock_mapper.go -package=dmmapi github.com/tikfack/server/internal/infrastructure/dmmapi MapperInterface

import (
	"strconv"
	"strings"
	"time"

	"github.com/tikfack/server/internal/domain/entity"
)

type MapperInterface interface {
    ConvertItem(item Item) entity.Video
}

// ConvertItem は dmmapi.Item を entity.Video に変換する
func ConvertItem(item Item) entity.Video {
    // 価格や日付のパースなどロジックをここに集約
    price := parsePrice(item)
    created := parseDate(item.Date)

    // actor/genre/maker などの変換
    var actresses []entity.Actress
    for _, a := range item.ItemInfo.Actress {
        actresses = append(actresses, entity.Actress{ID: strconv.Itoa(a.ID), Name: a.Name})
    }
    
    // genres変換
    var genres []entity.Genre
    for _, g := range item.ItemInfo.Genre {
        genres = append(genres, entity.Genre{ID: strconv.Itoa(g.ID), Name: g.Name})
    }
    
    // makers変換
    var makers []entity.Maker
    for _, m := range item.ItemInfo.Maker {
        makers = append(makers, entity.Maker{ID: strconv.Itoa(m.ID), Name: m.Name})
    }
    
    // series変換
    var series []entity.Series
    for _, s := range item.ItemInfo.Series {
        series = append(series, entity.Series{ID: strconv.Itoa(s.ID), Name: s.Name})
    }
    
    // directors変換
    var directors []entity.Director
    for _, d := range item.ItemInfo.Director {
        directors = append(directors, entity.Director{ID: strconv.Itoa(d.ID), Name: d.Name})
    }
    
    // サンプル動画URL
    var sampleURL string
    if item.SampleMovieURL != nil {
        sampleURL = item.SampleMovieURL.Size720480
    }

    return entity.Video{
        DmmID:        item.ContentID,
        Title:        item.Title,
        DirectURL:    sampleURL, // サンプル動画URLをDirectURLとして使用
        URL:          item.URL,
        SampleURL:    sampleURL,
        ThumbnailURL: item.ImageURL.Large,
        CreatedAt:    created,
        Price:        price,
        LikesCount:   0, // APIからは取得できないのでデフォルト値0を設定
        Actresses:    actresses,
        Genres:       genres,
        Makers:       makers,
        Series:       series,
        Directors:    directors,
    }
}

func parsePrice(item Item) int {
    // サンプル実装
    s := strings.ReplaceAll(item.Prices.Price, "円", "")
    p, _ := strconv.Atoi(strings.ReplaceAll(s, "~", ""))
    return p
}

func parseDate(dateStr string) time.Time {
    t, _ := time.Parse("2006-01-02 15:04:05", dateStr)
    return t
}