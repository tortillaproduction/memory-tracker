# Study Tracker

学習意欲が高いときに登録したWebサイト・サービスを、一定期間訪問しなかったら通知して忘却を防ぐサービス。

## 技術スタック

- **Backend**: Go, DDD構成, 標準`net/http`, Ginkgo/Gomega（テスト）
- **Frontend**: React (Vite), TailwindCSS, TanStack Query
- **DB**: PostgreSQL
- **API**: REST + OpenAPI（gRPCは今回のスコープではオーバースペックのため見送り）
- **認証**: Google OAuth + Cookieベースセッション
- **通知**: メール（初期実装）、LINE（Phase 2、UI上は現状トグルをグレーアウト）

## ディレクトリ構成

詳細は [DIRECTORY_STRUCTURE.md](./DIRECTORY_STRUCTURE.md) を参照。

```
study-tracker/
├── backend/    # Go + DDD API
├── frontend/   # React SPA
└── docker-compose.yml
```

## 起動方法

```bash
cp .env.example .env
docker compose up
```

- フロントエンド: http://localhost:5173
- バックエンドAPI: http://localhost:8080
- Postgres: localhost:5432

## マイグレーション

`golang-migrate` の利用を想定（未インストールの場合は別途導入してください）。

```bash
migrate -path ./backend/migrations \
  -database "postgres://postgres:postgres@localhost:5432/studytracker?sslmode=disable" \
  up
```

## テスト（バックエンド）

```bash
cd backend
make test
```

## 現在のスコープ（MVP）

- [x] ドメインモデル設計（User / Plan / Site / CheckIn）
- [x] プロジェクト雛形・ディレクトリ構成
- [ ] Google OAuth実装
- [ ] Cookieセッション実装
- [ ] サイト登録API（登録=初回チェックイン）
- [ ] 経由リンク方式のチェックインAPI（`/go/:siteId`）
- [ ] ストリーク計算API・ダッシュボード表示
- [ ] メール通知バッチ（未訪問サイトの検出）
- [ ] LINE通知（UIはトグル用意、実装完了までグレーアウト）
- [ ] Stripe連携（Phase 2、`plan_type`と`subscriptions`テーブルは用意済み）

## 既知の簡略化点（雛形段階での注意）

- `postgres`パッケージのリポジトリ実装は一部スタブ（TODOコメントあり）。本格実装時は`sqlc`導入を推奨。
- `checkin.NewCheckIn`が内部で`time.Now()`を使っているため、ユニットテストで時刻を固定しづらい。テスト容易性を上げるならコンストラクタに`Clock`インターフェースを注入する設計に変更するのがおすすめ。
- 認証ミドルウェア（`auth_middleware.go`）やGoogle OAuthクライアントは未実装（ディレクトリのみ用意）。
- フロントエンドの`src/api`は空。バックエンドのOpenAPI定義確定後、`orval`等で自動生成する想定。

## 将来の拡張

- **Phase 2（判定精度向上）**: ブラウザ拡張機能で実アクセスを直接検知
- **Phase 2（マネタイズ）**: Stripe連携、無料プランはサイト5件まで（`domain/plan`にルール集約済み）
