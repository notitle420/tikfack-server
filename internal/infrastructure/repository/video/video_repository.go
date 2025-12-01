package repository

import (
	"github.com/tikfack/server/internal/application/port"
	"github.com/tikfack/server/internal/infrastructure/dmmapi"
)

// NewVideoRepository は VideoCatalog ポートの実装を返す。
// 現状 DMM API 実装をそのまま返すだけなのでテスト不要。
func NewVideoRepository() (port.VideoCatalog, error) {
	return dmmapi.NewRepository()
}
