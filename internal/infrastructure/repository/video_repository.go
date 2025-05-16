package repository

import (
	"github.com/tikfack/server/internal/domain/repository"
	"github.com/tikfack/server/internal/infrastructure/dmmapi"
)

// NewVideoRepository は domain.VideoRepository の実装を返す
// dmmapi.repositoryを返すだけなのでテスト不要
func NewVideoRepository() (repository.VideoRepository, error) {
    // 今は DMM API 実装を返す
    return dmmapi.NewRepository()
}