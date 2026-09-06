package migration

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // "file://"スキームを使うためのside-effect import
)

// Run は指定ディレクトリ内の未適用マイグレーションを実行する。
// 既に最新の場合は migrate.ErrNoChange が返るが、これはエラーとして扱わない。
//
// 本番運用では複数インスタンスが同時にこれを呼ぶと競合するリスクがあるため、
// 開発時（main.goからの自動実行）専用の使い方を想定している。
// 本番では docker-compose.prod.yml の migrate サービス（CLIイメージ）で
// アプリ起動前に一度だけ明示的に実行する運用に切り替える。
func Run(db *sql.DB, migrationPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
