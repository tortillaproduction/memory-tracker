# memory-tracker ディレクトリ構成

```
memory-tracker/
├── .devcontainer/
│   └── devcontainer.json             # 開発コンテナ設定
├── .github/
│   └── workflows/
│       └── ci.yml                    # CI（GitHub Actions）
├── .vscode/                          # エディタ設定・推奨拡張機能
│
├── backend/                          # Go + DDD バックエンド
│   ├── cmd/
│   │   └── api/
│   │       └── main.go               # エントリーポイント（DI、サーバー起動）
│   │
│   ├── internal/
│   │   ├── domain/                   # ドメイン層（ビジネスルールの核）
│   │   │   ├── user/
│   │   │   │   ├── user.go           # Userエンティティ
│   │   │   │   └── repository.go     # UserRepositoryインターフェース
│   │   │   ├── plan/
│   │   │   │   └── plan.go           # Plan値オブジェクト（無料/有料、件数制限ルール）
│   │   │   ├── site/
│   │   │   │   ├── site.go           # Siteエンティティ
│   │   │   │   └── repository.go     # SiteRepositoryインターフェース
│   │   │   ├── checkin/
│   │   │   │   ├── checkin.go        # CheckInエンティティ
│   │   │   │   ├── streak.go         # ストリーク計算ロジック
│   │   │   │   └── *_test.go         # Ginkgoテスト
│   │   │   ├── gate/                 # ゲート(Androidアプリ)のトークン・脱出口の記録・日本時間の日付
│   │   │   └── notification/
│   │   │       ├── setting.go        # NotificationSettingエンティティ（通知モード email / gate を含む）
│   │   │       └── repository.go     # 通知関連のリポジトリインターフェース
│   │   │
│   │   ├── usecase/                  # アプリケーション層（ユースケース）
│   │   │   ├── auth/
│   │   │   │   └── google_login.go   # Googleログイン処理
│   │   │   ├── register_site/        # サイト登録（登録=チェックイン込み）
│   │   │   ├── list_sites/           # サイト一覧取得
│   │   │   ├── delete_site/          # サイト削除
│   │   │   ├── checkin_site/         # 経由リンク・ゲートからのチェックイン処理（所有者確認・5分以内の重複防止）
│   │   │   ├── overdue/              # 期限切れサイトの読み取りモデル（通知とゲートで共有）
│   │   │   ├── get_gate_candidates/  # ゲートに出す候補（最大3件）と今日済みの判定
│   │   │   ├── dismiss_gate/         # ゲートの脱出口の記録
│   │   │   ├── issue_gate_token/     # ゲート用トークンの発行・状態
│   │   │   ├── authenticate_gate_token/  # Bearerトークンの照合
│   │   │   ├── update_notification_mode/ # 通知モード(email / gate)の切り替え
│   │   │   └── notify_overdue_sites/ # 未チェックインサイトのメール通知
│   │   │       ├── notify_overdue_sites.go
│   │   │       └── email_template.go
│   │   │
│   │   ├── infrastructure/           # インフラ層（外部技術の実装詳細）
│   │   │   ├── persistence/
│   │   │   │   └── postgres/
│   │   │   │       ├── user_repository.go
│   │   │   │       ├── site_repository.go
│   │   │   │       ├── checkin_repository.go
│   │   │   │       ├── notification_repository.go
│   │   │   │       ├── overdue_site_query.go   # 期限切れ判定SQL（唯一の定義）
│   │   │   │       ├── gate_token_repository.go
│   │   │   │       ├── gate_dismissal_repository.go
│   │   │   │       └── id_generator.go
│   │   │   ├── notification/
│   │   │   │   ├── email_sender.go       # メール送信実装
│   │   │   │   └── fake_email_sender.go  # 開発/テスト用
│   │   │   ├── auth/
│   │   │   │   ├── google_oauth.go   # Google OAuthクライアント
│   │   │   │   ├── session_store.go  # Cookieセッションストア（Postgres）
│   │   │   │   ├── checkin_token.go  # チェックイン用トークン
│   │   │   │   └── gate_token.go     # ゲート用トークンの生成・ハッシュ
│   │   │   ├── batch/
│   │   │   │   └── notification_scheduler.go  # 通知の定期実行
│   │   │   └── migration/
│   │   │       └── migration.go      # マイグレーション実行
│   │   │
│   │   └── interface/                # インターフェース層（外部との接点）
│   │       └── http/
│   │           ├── handler/
│   │           │   ├── site_handler.go
│   │           │   ├── checkin_handler.go   # /go/:siteId のリダイレクトもここ
│   │           │   ├── auth_handler.go
│   │           │   ├── gate_handler.go          # /api/gate/*
│   │           │   └── notification_settings_handler.go
│   │           ├── middleware/
│   │           │   ├── auth_middleware.go
│   │           │   └── gate_token_middleware.go # Authorization: Bearer 認証
│   │           └── router.go
│   │
│   ├── migrations/                   # golang-migrate用SQL（up/downのペア、0001〜0008）
│   │
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   └── Makefile
│
├── frontend/                         # React + Vite + Tailwind + daisyUI SPA（PWA）
│   ├── public/                       # favicon、PWAアイコン、logo.svg
│   ├── src/
│   │   ├── api/                      # APIクライアント（auth / client / sites）
│   │   ├── components/               # AddSiteModal、DeleteSiteModal、Footer、
│   │   │                             # GoogleSignInButton、ThemeSwitcher
│   │   ├── contexts/
│   │   │   └── ToastContext.tsx      # トースト通知
│   │   ├── hooks/                    # useAuth、useSites、useInstallPrompt
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx         # メインページ
│   │   │   ├── Login.tsx
│   │   │   ├── Privacy.tsx
│   │   │   └── Terms.tsx
│   │   ├── App.tsx                   # ルーティング、ヘッダー/ユーザーメニュー
│   │   ├── main.tsx
│   │   ├── sw.ts                     # Service Worker（PWAのインストール要件用）
│   │   ├── index.css                 # Tailwind読み込み
│   │   └── vite-env.d.ts
│   ├── index.html
│   ├── package.json
│   ├── tsconfig*.json                # app / node / worker
│   ├── vite.config.ts
│   ├── vercel.json                   # Vercelデプロイ設定
│   ├── .prettierrc.json
│   └── Dockerfile
│
├── android/                          # ゲート（Androidアプリ、Kotlin + Jetpack Compose）
│   ├── build.sh                      # Dockerでのビルド（APK + 単体テスト）
│   └── app/src/
│       ├── main/java/.../gate/
│       │   ├── policy/               # Androidに依存しない判定ロジック（除外リスト・発動制限など）
│       │   ├── api/                  # APIクライアント（タイムアウト2秒）とセットアップコード
│       │   ├── ui/                   # ゲート画面・セットアップ・設定画面（Compose）
│       │   ├── GateAccessibilityService.kt  # 前面アプリの検知
│       │   └── GateLauncher.kt       # ゲートの起動と自動終了のタイマー
│       ├── debug/                    # デバッグビルドのみ平文HTTPを許可
│       └── test/                     # JVMの単体テスト
│
├── docs/
│   ├── notification-operations.md    # 通知の運用(リトライ・ログ・切り分け)
│   ├── gate-setup.md                 # ゲートのセットアップ手順（使う人向け）
│   └── gate-device-testing.md        # ゲートの実機での動作確認手順
│
├── docker-compose.yml                # 開発用: backend / frontend / db をまとめて起動
├── docker-compose.prod.yml           # 本番用
├── .env.example
└── README.md
```

## レイヤー間の依存方向（DDDの鉄則）

```
interface  →  usecase  →  domain
                              ↑
infrastructure ───────────────┘
（domainのインターフェースを実装する形で依存）
```

- `domain` 層は他のどの層にも依存しない（Goの標準ライブラリのみに依存する想定）
- `infrastructure` 層は `domain` のリポジトリインターフェースを実装する（依存性逆転）
- `main.go` で全ての依存を組み立てる（DIコンテナは使わず、手動DIで十分な規模）
