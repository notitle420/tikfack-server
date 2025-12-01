package model

import "time"

// Video は外部の動画カタログから取得したデータを表す DTO。
type Video struct {
	DmmID        string
	Title        string
	DirectURL    string
	URL          string
	SampleURL    string
	ThumbnailURL string
	CreatedAt    time.Time
	Price        int
	LikesCount   int

	Actresses []Actress
	Genres    []Genre
	Makers    []Maker
	Series    []Series
	Directors []Director

	Review Review
}

// Actress は出演女優情報。
type Actress struct {
	ID   string
	Name string
}

// Genre はジャンル情報。
type Genre struct {
	ID   string
	Name string
}

// Maker はメーカー情報。
type Maker struct {
	ID   string
	Name string
}

// Series はシリーズ情報。
type Series struct {
	ID   string
	Name string
}

// Director は監督情報。
type Director struct {
	ID   string
	Name string
}

// Review はレビュー情報。
type Review struct {
	Count   int
	Average float32
}

// SearchMetadata は検索結果のメタデータ。
type SearchMetadata struct {
	ResultCount   int
	TotalCount    int
	FirstPosition int
}
