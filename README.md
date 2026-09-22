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

通知バッチは5分ごとに動作し、期限切れサイトが1件でもあればメールを送信します。

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
[goroutine + time.Ticker（5分間隔）]
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

**開発中に期限切れ通知をブラウザで確認する**

通知バッチはデフォルト5分間隔のため、開発中は次の手順で確認します。

1. `.env`に`NOTIFICATION_INTERVAL=1m`を追加して`docker compose up -d --build app`（バッチ間隔を1分に短縮。省略時は5分。appはイメージをビルドして動かしているため、`.env`やコードを変えたら再ビルドが必要です。起動ログの`notification scheduler started`の`interval`が`60000000000`なら反映されています）
2. `http://localhost:5173`でログインし、Notificationsのトグルを**Push**にして通知を許可する（プッシュ購読はブラウザ固有のため、SQLでは作れません）
3. 期限切れサイトを投入する（`you@example.com`はログインに使ったアドレス）

```bash
docker compose exec -T db psql -U postgres memorytracker -v email=you@example.com \
  < backend/scripts/dev_seed_overdue.sql
```

4. 最大1分待つと、期限切れの2サイト分の通知が届きます（期限内のサイトは通知されません）。もう一度確認したいときは手順3のSQLを再実行してください（通知ログが消えて再び対象になります）
5. 通知（またはボタン）をクリックすると、チェックインが記録され、登録したサイトのURLへ移動します。開発時も`/go/*`はViteのプロキシ（`frontend/vite.config.ts`）でバックエンドへ転送されます（本番の`frontend/vercel.json`のrewriteと同じ挙動）
6. 終わったら`docker compose exec -T db psql -U postgres memorytracker < backend/scripts/dev_seed_cleanup.sql`でシードを削除します

**通知が届く条件・表示のされ方（タブを閉じた場合、PWAの場合など）**: [docs/push-notification-spec.md](docs/push-notification-spec.md)を参照してください。

## 本番デプロイ

コストと運用の手間を抑えるため、以下の構成でデプロイします。

| 役割 | サービス |
|---|---|
| フロントエンド | [Vercel](https://vercel.com)（Hobbyプラン、$0） |
| バックエンド | [Render](https://render.com) Web Service（Docker、無料枠。無アクセス時はスリープ） |
| DB | [Neon](https://neon.tech)（サーバーレスPostgres、無料枠） |
| メール | [Resend](https://resend.com) |

### 全体像とリクエストの流れ

ブラウザはVercelのドメインだけにアクセスします。`frontend/vercel.json`のrewriteで`/api/*`と`/go/*`をRenderのバックエンドへプロキシするため、ブラウザから見ると**同一オリジン**の通信になり、サードパーティCookieブロックの影響を受けずにセッションCookieを扱えます。

```
ブラウザ ──> https://<Vercelドメイン>/            → Vercel（SPA配信）
         ──> https://<Vercelドメイン>/api/*, /go/* → Vercelがrewriteで https://<Renderドメイン>/... へ転送
                                                    └─> Neon(Postgres)
```

このため、本番では次の2点が重要です。

- フロントエンドの`VITE_API_BASE_URL`は**設定しない**（未設定なら本番ビルドは相対パスを使う）。
- OAuthのリダイレクトURIは**Renderではなく、Vercelのドメイン**にする（`state` CookieがVercelドメインに保存されるため、コールバックも同じドメインで受ける必要がある）。

### 環境変数の一覧（バックエンド / Render）

| 変数 | 必須 | 本番での設定値 |
|---|---|---|
| `PORT` | 必須 | `8080`（アプリは8080固定で待ち受けるため、Renderにポートを教える） |
| `DATABASE_URL` | 必須 | Neonの接続文字列（`?sslmode=require`付き） |
| `AUTO_MIGRATE` | 必須 | `true`（起動時に未適用のマイグレーションを実行） |
| `FRONTEND_URL` | 必須 | VercelのURL（例: `https://memory-tracker.vercel.app`）。末尾スラッシュ・スキーム省略は不可。CORS許可オリジン、OAuth後のリダイレクト先、通知内のチェックインリンクに使われる |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | 必須 | Google Cloud Consoleで発行した値 |
| `GOOGLE_REDIRECT_URL` | 必須 | `https://<Vercelドメイン>/api/auth/google/callback` |
| `SESSION_SECRET` | 必須 | 開発用とは別の強いランダム値（例: `openssl rand -hex 32`）。チェックインリンクの署名鍵も兼ねる。未設定だと起動に失敗する |
| `RESEND_API_KEY` | 任意 | 未設定ならメール通知は無効 |
| `EMAIL_FROM_ADDRESS` / `EMAIL_FROM_NAME` | 任意 | 認証済み独自ドメインのアドレス / 表示名 |
| `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY` | 任意 | 本番用に新規生成した鍵ペア。未設定ならブラウザ通知は無効 |
| `VAPID_SUBJECT` | 任意 | `mailto:<連絡先メールアドレス>` |

`LINE_*`と`STRIPE_*`は未実装のため設定不要です。

### 環境変数の一覧（フロントエンド / Vercel）

| 変数 | 設定値 |
|---|---|
| `VITE_API_BASE_URL` | **設定しない**（上記の理由。ローカル開発時のみ既定で`http://localhost:8080`が使われる） |

### 手順

デプロイ先のURLが互いに必要になるため、次の順で進めます。

#### 1. Neon（DB）

1. Neonでプロジェクトを作成する
2. ダッシュボードの Connection string（`postgres://...neon.tech/...?sslmode=require`）を控える → `DATABASE_URL`

#### 2. VAPID鍵・SESSION_SECRETを生成する（ローカル）

```bash
# SESSION_SECRET
openssl rand -hex 32

# VAPID鍵ペア（ブラウザ通知を使う場合）
npx web-push generate-vapid-keys
```

#### 3. Render（バックエンド）

1. New → Web Service でGitHubリポジトリを連携する
2. 次のとおり設定する
   - Language: `Docker`
   - Root Directory: `backend`（`backend/Dockerfile`が使われる）
   - Instance Type: Free
   - Branch: `main`（pushで自動デプロイ）
3. Environment Variablesに、上の一覧のバックエンド用変数を登録する。この時点でVercelのURLが未確定なら、`FRONTEND_URL`と`GOOGLE_REDIRECT_URL`は仮の値にして、手順4の後に更新する
4. デプロイ後、発行されたURL（`https://<name>.onrender.com`）を控える
5. `https://<name>.onrender.com/api/auth/me` を開き、`401`が返れば起動成功（未ログインのため）

> **マイグレーション**: Renderの無料枠はPre-Deploy Commandを使えず、ランタイムイメージ（alpine）にも`migrate` CLIは入っていません。そのため`AUTO_MIGRATE=true`にして、アプリ起動時に`backend/migrations`を自動適用します（`backend/Dockerfile`で同梱済み）。複数インスタンスに増やす場合は、`docker-compose.prod.yml`の`migrate`サービスのように、事前に1回だけ実行する方式に切り替えてください。

#### 4. Vercel（フロントエンド）

1. **先に`frontend/vercel.json`のrewrite先を自分のRenderのURLに書き換えてコミット・pushする**（現状は`memory-tracker-lamg.onrender.com`が直書きされている）

   ```json
   {
     "rewrites": [
       { "source": "/api/(.*)", "destination": "https://<name>.onrender.com/api/$1" },
       { "source": "/go/(.*)", "destination": "https://<name>.onrender.com/go/$1" },
       { "source": "/(.*)", "destination": "/index.html" }
     ]
   }
   ```

2. Add New → Project でGitHubリポジトリを連携する
3. Root Directoryを`frontend`に設定する（Viteは自動検出。Build Command: `npm run build`、Output Directory: `dist`）
4. 環境変数は設定せずにデプロイする
5. 発行されたVercelのURL（`https://<project>.vercel.app`）を控える

#### 5. RenderにVercelのURLを反映する

Renderの環境変数を更新し、再デプロイする。

- `FRONTEND_URL` = `https://<project>.vercel.app`
- `GOOGLE_REDIRECT_URL` = `https://<project>.vercel.app/api/auth/google/callback`

#### 6. Google OAuthの本番設定

[Google Cloud Console](https://console.cloud.google.com/) → 認証情報 → OAuthクライアントIDで次を追加する。

- 承認済みのリダイレクトURI: `https://<project>.vercel.app/api/auth/google/callback`
- （必要に応じて）承認済みのJavaScript生成元: `https://<project>.vercel.app`

`GOOGLE_REDIRECT_URL`とここの値は**完全一致**させてください（不一致だと`redirect_uri_mismatch`になります）。

#### 7. メールの本番設定（Resend）

1. Resendのダッシュボードで独自ドメインを追加し、表示されたDNSレコード（SPF/DKIM）を設定して認証する
2. `EMAIL_FROM_ADDRESS`をそのドメインのアドレスに変更する（`onboarding@resend.dev`のままだと、自分のアドレス宛にしか送れない）

#### 8. 動作確認

1. `https://<project>.vercel.app`を開き、Googleでログインできる
2. サイトを登録するとダッシュボードに表示される
3. スマホでホーム画面に追加（PWA）し、ナビバーのユーザーメニューからPushを有効にできる（VAPID鍵を設定した場合。iOSはPWAとしてインストールが必要）

### 運用上の注意

- **URLを変えたとき**: Vercel/RenderのURL（カスタムドメイン含む）を変更したら、`vercel.json`のrewrite先、`FRONTEND_URL`、`GOOGLE_REDIRECT_URL`、Google Cloud Consoleのリダイレクト URIをすべて揃えて更新してください。
- **`SESSION_SECRET`を変更すると**、発行済みのチェックインリンク（メール/Push内のリンク）が無効になります。

### CI

`.github/workflows/ci.yml`により、push（main）/PR時にbackendの`go build` / `go vet` / `go test`とfrontendの`npm run build`（型チェック含む）を実行します。デプロイはVercel/Renderそれぞれのgit連携（mainへのpushで自動反映）に任せています。

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
- [x] メール通知バッチ（Resend、5分ごとに期限切れサイトを検出・通知）
- [x] PWA化・ブラウザ通知（Web Push、購読状況に応じてメールと自動的に切り替え）
- [ ] LINE通知（UIはトグル用意、実装完了までグレーアウト）
- [ ] Stripe連携（Phase 2、`plan_type`と`subscriptions`テーブルは用意済み）

## 将来の拡張

- **Phase 2（判定精度向上）**: ブラウザ拡張機能で実アクセスを直接検知
- **Phase 2（マネタイズ）**: Stripe連携、無料プランはサイト5件まで（`domain/plan`にルール集約済み）
- **Phase 2（通知）**: LINE通知（Messaging API連携）
