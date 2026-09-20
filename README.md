# Memory Tracker

学習意欲が高いときに登録したWebサイト・サービスを、一定期間訪問しなかったら通知して忘却を防ぐサービス。

## 技術スタック

- **Backend**: Go, DDD構成, 標準`net/http`, Ginkgo/Gomega（テスト）
- **Frontend**: React (Vite), TailwindCSS v4, daisyUI v5, TanStack Query
- **DB**: PostgreSQL
- **API**: REST（gRPCは今回のスコープではオーバースペックのため見送り）
- **認証**: Google OAuth + Cookieベースセッション
- **通知**: メール（Resend）、ブラウザ通知（Web Push、PWAインストール時）、LINE（Phase 2、UI上は現状トグルをグレーアウト）
- **PWA**: `vite-plugin-pwa`（Service Worker、installable manifest）

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
ユーザーごとにまとめて通知（複数サイトが期限切れでも1回にまとめる）
    ↓
有効なプッシュ購読があればプッシュを優先、無ければ（または設定次第で）メールを送信
    ↓
notification_logsに記録（二重送信防止）
```

## ブラウザ通知（Web Push / PWA）の設定

アプリをホーム画面にインストール（PWA化）すると、期限切れサイトの通知をブラウザ通知（Web Push）で受け取れます。

**通知チャネルの自動切替（二重通知の防止）**

ナビバーのユーザーアイコン→ドロップダウンメニューの「Notifications」に、Email/Pushを切り替える1本のスライドトグルがあります（`frontend/src/App.tsx`、状態管理は`frontend/src/hooks/usePushSubscription.ts`）。トグルをPush側にするとその場でプッシュ購読を作成し、Email側に戻すと購読を解除します。ブラウザがPush APIに対応していない場合（PWAとしてインストールするまでプッシュが使えないiOS Safari等）はトグルが無効化され、ツールチップで案内されます。

- 有効なプッシュ購読があれば、その端末にはプッシュ通知のみを送信し、メールは送りません（デフォルト）。トグルはEmail/Pushのどちらか一方を選ぶ形なので、通常は両方同時に届くことはありません。
- プッシュの購読が失効している場合（ブラウザ側で通知を許可解除した場合など）は、送信時に自動検知して購読情報を削除し、同じタイミングでメール通知にフォールバックします（手動での切り戻し操作は不要）。
- バックエンドには「プッシュが使える場合でもメールを両方送る」設定（`disableEmailWhenPushAvailable`、`PATCH /api/notification-preferences`）が残っていますが、現在のUIからは変更できず、常にデフォルト値（メールは止める）のまま使われます。

**VAPID鍵の設定**

Web Pushの送信にはVAPID（Voluntary Application Server Identification）鍵ペアが必要です。

1. VAPID鍵ペアを生成する（`webpush-go`同梱のCLIか、[web-push](https://www.npmjs.com/package/web-push)などのツールを利用）
2. `.env`に設定する

```env
VAPID_PUBLIC_KEY=xxxxxxxxxxxx
VAPID_PRIVATE_KEY=xxxxxxxxxxxx
VAPID_SUBJECT=mailto:support@example.com
```

**`VAPID_PUBLIC_KEY`/`VAPID_PRIVATE_KEY`が未設定の場合**: プッシュ送信は自動的に無効化され、購読・設定APIやメール通知バッチは通常通り動作します（開発時は鍵なしで起動できます）。

**ローカルでのプッシュ通知確認**: `localhost`はセキュアコンテキストとして扱われるため、ブラウザでの購読からバックエンドからの実際のプッシュ送信まで、トンネルなしでローカル完結して確認できます。インストールプロンプトのスマートフォン実機確認にはHTTPSが必要なため、Vercelのプレビューデプロイ等を利用してください。

**通知が届く条件・表示のされ方（タブを閉じた場合、PWAの場合など）**: [docs/push-notification-spec.md](docs/push-notification-spec.md)を参照してください。

## 本番デプロイ

コストと運用の手間を抑えるため、以下のハイブリッド構成でデプロイする。

| 役割 | サービス |
|---|---|
| フロントエンド | [Vercel](https://vercel.com)（Hobbyプラン、$0） |
| バックエンド | [Render](https://render.com) Web Service（無料枠、無アクセス時はスリープ） |
| DB | [Neon](https://neon.tech)（サーバーレスPostgres、無料枠） |
| メール | Resend（既存） |

### 1. Neon（DB）

1. Neonでプロジェクトを作成し、発行された接続文字列（`DATABASE_URL`）を控える

### 2. Render（バックエンド）

1. GitHubリポジトリと連携し、`backend/Dockerfile` を使うWeb Serviceを作成
2. 環境変数を設定:
   - `DATABASE_URL`: Neonの接続文字列
   - `AUTO_MIGRATE`: `false`
   - `FRONTEND_URL`: Vercelの本番URL（CORS許可オリジンとして使用）
   - `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET`
   - `GOOGLE_REDIRECT_URL`: `https://<Renderのバックエンドドメイン>/api/auth/google/callback`
   - `SESSION_SECRET`: 開発用とは別に新規生成した強い値
   - `RESEND_API_KEY` / `EMAIL_FROM_ADDRESS`（独自ドメイン認証後のアドレス） / `EMAIL_FROM_NAME`
   - `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY` / `VAPID_SUBJECT`（本番用に新規生成した鍵ペア。ブラウザ通知を使う場合のみ必須）
3. Pre-Deploy Command にマイグレーション実行コマンドを設定し、アプリ起動前に一度だけ適用されるようにする（`docker-compose.prod.yml` の `migrate` サービスと同じ考え方）:
   ```bash
   migrate -path ./migrations -database "$DATABASE_URL" up
   ```

### 3. Vercel（フロントエンド）

1. GitHubリポジトリと連携し、Root Directoryを `frontend` に設定（Vite構成は自動検出される）
2. 環境変数 `VITE_API_BASE_URL` にRenderのバックエンドURLを設定
3. `frontend/vercel.json` によりSPAのクライアントサイドルーティングが有効になる

### 4. Google OAuthの本番設定

[Google Cloud Console](https://console.cloud.google.com/) の承認済みリダイレクトURIに、本番のRenderバックエンドURL（`https://<Renderドメイン>/api/auth/google/callback`）を追加する。

### 5. CI

`.github/workflows/ci.yml` により、push/PR時にbackendの `go build` / `go vet` / `go test` とfrontendの `npm run build`（型チェック含む）を実行する。実際のデプロイはVercel/Renderそれぞれのgit連携（mainブランチへのpushで自動反映）に任せる。

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
- [x] メール通知バッチ（Resend、15分ごとに期限切れサイトを検出・通知）
- [x] PWA化・ブラウザ通知（Web Push、購読状況に応じてメールと自動的に切り替え）
- [ ] LINE通知（UIはトグル用意、実装完了までグレーアウト）
- [ ] Stripe連携（Phase 2、`plan_type`と`subscriptions`テーブルは用意済み）

## 将来の拡張

- **Phase 2（判定精度向上）**: ブラウザ拡張機能で実アクセスを直接検知
- **Phase 2（マネタイズ）**: Stripe連携、無料プランはサイト5件まで（`domain/plan`にルール集約済み）
- **Phase 2（通知）**: LINE通知（Messaging API連携）
