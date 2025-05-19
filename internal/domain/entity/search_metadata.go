package entity

// SearchMetadata は検索結果のメタデータを表す
type SearchMetadata struct {
	ResultCount    int // 取得件数
	TotalCount     int // 全体件数
	FirstPosition  int // 検索開始位置
} 