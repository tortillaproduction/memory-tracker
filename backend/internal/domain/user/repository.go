package user

import (
	"context"
	"errors"
)

// ErrNotFound はリポジトリが該当ユーザーを見つけられなかったことを表すドメインエラー。
// usecase層はこれを見て「新規ユーザーとして作成する」等の分岐を行う。
var ErrNotFound = errors.New("user not found")

type Repository interface {
	Save(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id ID) (*User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*User, error)
}
