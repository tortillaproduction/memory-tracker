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
│   │   │   └── notification/
│   │   │       ├── setting.go        # NotificationSettingエンティティ
│   │   │       ├── push_subscription.go  # PushSubscriptionエンティティ
│   │   │       └── repository.go     # 通知関連のリポジトリインターフェース
│   │   │
│   │   ├── usecase/                  # アプリケーション層（ユースケース）
│   │   │   ├── auth/
│   │   │   │   └── google_login.go   # Googleログイン処理
│   │   │   ├── register_site/        # サイト登録（登録=チェックイン込み）
│   │   │   ├── list_sites/           # サイト一覧取得
│   │   │   ├── delete_site/          # サイト削除
│   │   │   ├── checkin_site/         # 経由リンク踏破時のチェックイン処理
│   │   │   ├── notify_overdue_sites/ # 未チェックインサイトの通知（メール/Push）
│   │   │   │   ├── notify_overdue_sites.go
│   │   │   │   ├── email_template.go
│   │   │   │   └── push_payload.go
│   │   │   ├── subscribe_push/       # Push購読の登録
│   │   │   ├── unsubscribe_push/     # Push購読の解除
│   │   │   └── update_notification_preferences/  # 通知チャネル設定の更新
│   │   │
│   │   ├── infrastructure/           # インフラ層（外部技術の実装詳細）
│   │   │   ├── persistence/
│   │   │   │   └── postgres/
│   │   │   │       ├── user_repository.go
│   │   │   │       ├── site_repository.go
│   │   │   │       ├── checkin_repository.go
│   │   │   │       ├── notification_repository.go
│   │   │   │       ├── push_subscription_repository.go
│   │   │   │       └── id_generator.go
│   │   │   ├── notification/
│   │   │   │   ├── email_sender.go       # メール送信実装
│   │   │   │   ├── push_sender.go        # Web Push送信実装
│   │   │   │   ├── fake_email_sender.go  # 開発/テスト用
│   │   │   │   └── fake_push_sender.go   # 開発/テスト用
│   │   │   ├── auth/
│   │   │   │   ├── google_oauth.go   # Google OAuthクライアント
│   │   │   │   ├── session_store.go  # Cookieセッションストア（Postgres）
│   │   │   │   └── checkin_token.go  # チェックイン用トークン
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
│   │           │   └── push_handler.go
│   │           ├── middleware/
│   │           │   └── auth_middleware.go
│   │           └── router.go
│   │
│   ├── migrations/                   # golang-migrate用SQL（up/downのペア、0001〜0007）
│   │
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   └── Makefile
│
├── frontend/                         # React + Vite + Tailwind + daisyUI SPA（PWA）
│   ├── public/                       # favicon、PWAアイコン、logo.svg
│   ├── src/
│   │   ├── api/                      # APIクライアント（auth / client / push / sites）
│   │   ├── components/               # AddSiteModal、DeleteSiteModal、Footer、
│   │   │                             # GoogleSignInButton、ThemeSwitcher
│   │   ├── contexts/
│   │   │   └── ToastContext.tsx      # トースト通知
│   │   ├── hooks/                    # useAuth、useSites、usePushSubscription、
│   │   │                             # useInstallPrompt
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx         # メインページ
│   │   │   ├── Login.tsx
│   │   │   ├── Privacy.tsx
│   │   │   └── Terms.tsx
│   │   ├── App.tsx                   # ルーティング、ヘッダー/ユーザーメニュー
│   │   ├── main.tsx
│   │   ├── sw.ts                     # Service Worker（Push受信）
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
├── docs/
│   └── push-notification-spec.md     # Push通知の仕様
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
