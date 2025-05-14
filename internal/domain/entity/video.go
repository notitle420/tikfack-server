package entity

import "time"

// Video は動画を表すドメインエンティティ
type Video struct {
	DmmID       string
	Title       string
	DirectURL   string
	URL         string
	SampleURL   string
	ThumbnailURL string
	CreatedAt   time.Time
	Price       int
	LikesCount  int

	Actresses []Actress
	Genres    []Genre
	Makers     []Maker
	Series    []Series
	Directors  []Director
}

// Actress は出演女優を表す
type Actress struct {
	ID   string
	Name string
}

// Genre は動画のジャンルを表す
type Genre struct {
	ID   string
	Name string
}

// Maker は動画のメーカーを表す
type Maker struct {
	ID   string
	Name string
}

// Series は動画のシリーズを表す
type Series struct {
	ID   string
	Name string
}

// Director は動画の監督を表す
type Director struct {
	ID   string
	Name string
}