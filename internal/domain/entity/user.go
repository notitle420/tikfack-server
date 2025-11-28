package entity

import (
	"time"
	"errors"
	"strings"
)

// User はユーザー情報を表すエンティティです
type User struct {
	ID            string    // システム内部で使用するID
	KeycloakID    string    // KeycloakのユーザーID
	AccountName   string    // アカウント名
	Email         string    // メールアドレス
	FolderIDs    []string  // お気に入りフォルダのID一覧
	CreatedAt    time.Time // 作成日時
	UpdatedAt    time.Time // 更新日時
}

// ユーザー関連のドメインエラー
var (
	ErrInvalidAccountName = errors.New("account name must be between 3 and 32 characters")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrEmptyKeycloakID  = errors.New("keycloak ID cannot be empty")
)

// Validate はユーザーエンティティのバリデーションを行います
func (u *User) Validate() error {
	// アカウント名の検証
	if len(u.AccountName) < 3 || len(u.AccountName) > 32 {
		return ErrInvalidAccountName
	}

	// メールアドレスの検証
	if !validateEmail(u.Email) {
		return ErrInvalidEmail
	}

	// KeycloakIDの検証
	if u.KeycloakID == "" {
		return ErrEmptyKeycloakID
	}

	return nil
}

// validateEmail はメールアドレスの形式を検証します
// 実際のプロジェクトではより厳密な検証が必要かもしれません
func validateEmail(email string) bool {
	// 簡易的な検証
	return len(email) > 3 && strings.Contains(email, "@") && strings.Contains(email, ".")
}

// AddFolderID は新しいフォルダIDを追加します
func (u *User) AddFolderID(folderID string) {
	// 重複チェック
	for _, id := range u.FolderIDs {
		if id == folderID {
			return
		}
	}
	u.FolderIDs = append(u.FolderIDs, folderID)
	u.UpdatedAt = time.Now()
}

// RemoveFolderID はフォルダIDを削除します
func (u *User) RemoveFolderID(folderID string) {
	for i, id := range u.FolderIDs {
		if id == folderID {
			u.FolderIDs = append(u.FolderIDs[:i], u.FolderIDs[i+1:]...)
			u.UpdatedAt = time.Now()
			return
		}
	}
}

// UpdateEmail はメールアドレスを更新します
func (u *User) UpdateEmail(email string) error {
	if !validateEmail(email) {
		return ErrInvalidEmail
	}
	u.Email = email
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateAccountName はアカウント名を更新します
func (u *User) UpdateAccountName(name string) error {
	if len(name) < 3 || len(name) > 32 {
		return ErrInvalidAccountName
	}
	u.AccountName = name
	u.UpdatedAt = time.Now()
	return nil
}
