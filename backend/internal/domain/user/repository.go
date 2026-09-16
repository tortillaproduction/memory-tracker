package user

import (
	"context"
	"errors"
)

// ErrNotFound はリポジトリが該当ユーザーを見つけられなかったことを表すドメインエラー。
// usecase層はこれを見て「新規ユーザーとして作成する」等の分岐を行う。
var ErrNotFound = errors.New("user not found")

type Repository interface {
	// Save はgoogle_idを基準にUPSERTする。戻り値の*Userは実際にDBへ確定した内容
	// （並行リクエストと競合した場合は相手が確定させた行）であり、boolは
	// この呼び出し自身が新規INSERTを行ったかどうかを表す。
	Save(ctx context.Context, u *User) (*User, bool, error)
	FindByID(ctx context.Context, id ID) (*User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*User, error)
}
