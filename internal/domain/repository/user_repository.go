package repository

import (
	"context"
	"errors"

	"github.com/tikfack/server/internal/domain/entity"
)

// UserRepository はユーザー情報の永続化を担当するインターフェースです
type UserRepository interface {
	// Create は新しいユーザーを作成します
	Create(ctx context.Context, user *entity.User) error

	// GetByID はIDでユーザーを取得します
	GetByID(ctx context.Context, id string) (*entity.User, error)

	// GetByKeycloakID はKeycloakIDでユーザーを取得します
	GetByKeycloakID(ctx context.Context, keycloakID string) (*entity.User, error)

	// GetByEmail はメールアドレスでユーザーを取得します
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// GetByAccountName はアカウント名でユーザーを取得します
	GetByAccountName(ctx context.Context, accountName string) (*entity.User, error)

	// Update はユーザー情報を更新します
	Update(ctx context.Context, user *entity.User) error

	// Delete はユーザーを削除します
	Delete(ctx context.Context, id string) error

	// AddFolderToUser はユーザーにフォルダを追加します
	AddFolderToUser(ctx context.Context, userID string, folderID string) error

	// RemoveFolderFromUser はユーザーからフォルダを削除します
	RemoveFolderFromUser(ctx context.Context, userID string, folderID string) error

	// ListUserFolders はユーザーのフォルダ一覧を取得します
	ListUserFolders(ctx context.Context, userID string) ([]string, error)
}

// ユーザーリポジトリのエラー定義
var (
	// ErrUserNotFound はユーザーが見つからない場合のエラー
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists は既に同じユーザーが存在する場合のエラー
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrUserValidation はユーザー情報のバリデーションエラー
	ErrUserValidation = errors.New("user validation failed")
)
