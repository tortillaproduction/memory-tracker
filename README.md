# Memory Tracker

学習意欲が高いときに登録したWebサイト・サービスを、一定期間訪問しなかったら通知して忘却を防ぐサービス。

## 技術スタック

- **Backend**: Go, DDD構成, 標準`net/http`, Ginkgo/Gomega（テスト）
- **Frontend**: React (Vite), TailwindCSS v4, daisyUI v5, TanStack Query
- **DB**: PostgreSQL
- **API**: REST（gRPCは今回のスコープではオーバースペックのため見送り）
- **認証**: Google OAuth + Cookieベースセッション
- **通知**: メール（SendGrid）、LINE（Phase 2、UI上は現状トグルをグレーアウト）

## ディレクトリ構成

詳細は [DIRECTORY_STRUCTURE.md](./DIRECTORY_STRUCTURE.md) を参照。

```
memory-tracker/
├── backend/    # Go + DDD API
├── frontend/   # React SPA
└── docker-compose.yml
```

## 起動方法

```bash
cp .env.example .env
# .envにGOOGLE_CLIENT_ID/SECRETを設定（Google OAuthの設定を参照）
docker compose up
```

- フロントエンド: http://localhost:5173
- バックエンドAPI: http://localhost:8080
- Postgres: localhost:5432

## マイグレーション

**開発時（デフォルト）**: `AUTO_MIGRATE=true`（`.env`のデフォルト）の場合、`docker compose up`のたびにバックエンドが起動時に自動で未適用のマイグレーションを実行します。手動での実行は不要です。

**本番相当の検証時**: 複数インスタンス起動時の競合を避けるため、専用の`migrate`コンテナで事前に一度だけ実行し、アプリ側は`AUTO_MIGRATE=false`にします。

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build
```

手元のCLIで直接実行したい場合は`golang-migrate`をインストールした上でこちらも使えます。

```bash
migrate -path ./backend/migrations \
  -database "postgres://postgres:postgres@localhost:5432/memorytracker?sslmode=disable" \
  up
```

## Google OAuthの設定

1. [Google Cloud Console](https://console.cloud.google.com/) でプロジェクトを作成
2. 「APIとサービス」→「認証情報」から OAuth 2.0 クライアントID を作成（アプリケーションの種類: ウェブアプリケーション）
3. 承認済みのリダイレクトURIに `http://localhost:8080/api/auth/google/callback` を追加
4. 発行された クライアントID・クライアントシークレット を `.env` の `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` に設定
5. `docker compose up` で再起動すればログインボタンから一連の流れが動作します

### 認証フローの流れ

```
[フロント] Googleでログインボタン
    ↓ window.location.href
[バックエンド] GET /api/auth/google/login
    → state生成 → Cookie保存 → Googleの同意画面へリダイレクト
    ↓
[Google] ユーザーが許可
    ↓
[バックエンド] GET /api/auth/google/callback?code=...&state=...
    → state検証 → 認可コードをユーザー情報に交換
    → ユーザーをDBで検索/新規作成（usecase/auth）
    → セッション発行（sessionsテーブル）→ Cookie(session_id)をセット
    → フロントエンドへリダイレクト
    ↓
[フロント] 起動時に GET /api/auth/me を呼び、Cookie付きで認証確認
    → 200ならダッシュボード表示、401ならログイン画面表示
```

## メール通知（Resend）の設定

通知バッチは15分ごとに動作し、期限切れサイトが1件でもあればメールを送信します。

1. [resend.com](https://resend.com) でアカウントを作成（無料枠: 3,000通/月）
2. API Keys → Create API Key で`re_`から始まるAPIキーを発行
3. `.env`の`RESEND_API_KEY`に設定

```env
RESEND_API_KEY=re_xxxxxxxxxxxx
EMAIL_FROM_ADDRESS=onboarding@resend.dev
EMAIL_FROM_NAME=Memory Tracker
```

**開発・動作確認時**: `EMAIL_FROM_ADDRESS`を`onboarding@resend.dev`のままにすると、ドメイン認証なしで自分のメールアドレスに送信できます。

**本番運用時**: Resendのダッシュボードで独自ドメインを認証し、`EMAIL_FROM_ADDRESS`をそのドメインのアドレスに変更してください。

**`RESEND_API_KEY`が未設定の場合**: 通知バッチは自動的に無効化されます。APIサーバーとしては通常通り動作するため、開発時はキーなしで起動できます。

### 通知の仕組み

```
[goroutine + time.Ticker（15分間隔）]
    ↓
期限切れ かつ 前回通知からinterval_hours以上経過 のサイトを検出
    ↓
ユーザーごとにまとめて1通送信（複数サイトが期限切れでも1メール）
    ↓
notification_logsに記録（二重送信防止）
```

## トラブルシューティング

**フロントエンドの `npm install` で ERESOLVE エラーが出る場合**

`package-lock.json` をコミットせずに `npm install` を実行すると、環境やnpmキャッシュの状態によって解決されるパッケージバージョンがぶれ、依存関係の競合が起きることがあります。このリポジトリでは `package-lock.json` を同梱し、Dockerfile側も `npm ci`（lockfileを厳密に再現するインストール）を使う構成にしているため、通常は発生しません。

もし発生した場合は、キャッシュを使わず再ビルドしてください。

```bash
docker compose build --no-cache frontend
```

それでも解決しない場合は、ローカルの `frontend/node_modules` や `frontend/package-lock.json` が古い状態で残っていないか確認し、削除してから再度試してください。

```bash
rm -rf frontend/node_modules frontend/package-lock.json
docker compose build --no-cache frontend
```

## テスト（バックエンド）

```bash
cd backend
make test
```

## 現在のスコープ（MVP）

- [x] ドメインモデル設計（User / Plan / Site / CheckIn）
- [x] プロジェクト雛形・ディレクトリ構成
- [x] Google OAuth実装（標準ライブラリのみで実装、`golang.org/x/oauth2`は不使用）
- [x] Cookieセッション実装（Postgresの`sessions`テーブルで管理、即時失効可能）
- [x] 認証ミドルウェア（`RequireAuth`）とフロントエンドの認証状態分岐（ログイン画面⇔ダッシュボード）
- [x] サイト登録API（登録=初回チェックイン）
- [x] 経由リンク方式のチェックインAPI（`/go/:siteId`、認証必須）
- [x] マイグレーション自動化（開発時: main.go起動時に自動実行 / 本番相当: 専用migrateコンテナで事前実行）
- [x] サイト一覧取得API（`GET /api/sites`）とフロントの実データ反映
- [x] ストリーク計算・ダッシュボード表示（ユーザー全体 / サイト別）
- [x] メール通知バッチ（SendGrid、15分ごとに期限切れサイトを検出・通知）
- [ ] LINE通知（UIはトグル用意、実装完了までグレーアウト）
- [ ] Stripe連携（Phase 2、`plan_type`と`subscriptions`テーブルは用意済み）

## 将来の拡張

- **Phase 2（判定精度向上）**: ブラウザ拡張機能で実アクセスを直接検知
- **Phase 2（マネタイズ）**: Stripe連携、無料プランはサイト5件まで（`domain/plan`にルール集約済み）
- **Phase 2（通知）**: LINE通知（Messaging API連携）
