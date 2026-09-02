# study-tracker ディレクトリ構成

```
study-tracker/
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
│   │   │   │   ├── repository.go     # CheckInRepositoryインターフェース
│   │   │   │   └── streak.go         # ストリーク計算ロジック
│   │   │   └── notification/
│   │   │       └── setting.go        # NotificationSettingエンティティ
│   │   │
│   │   ├── usecase/                  # アプリケーション層（ユースケース）
│   │   │   ├── register_site/
│   │   │   │   └── register_site.go  # サイト登録（登録=チェックイン込み）
│   │   │   ├── checkin_site/
│   │   │   │   └── checkin_site.go   # 経由リンク踏破時のチェックイン処理
│   │   │   ├── calculate_streak/
│   │   │   │   └── calculate_streak.go
│   │   │   └── auth/
│   │   │       └── google_login.go   # Googleログイン処理
│   │   │
│   │   ├── infrastructure/           # インフラ層（外部技術の実装詳細）
│   │   │   ├── persistence/
│   │   │   │   └── postgres/
│   │   │   │       ├── user_repository.go
│   │   │   │       ├── site_repository.go
│   │   │   │       └── checkin_repository.go
│   │   │   ├── notification/
│   │   │   │   ├── email_sender.go   # メール送信実装
│   │   │   │   └── line_sender.go    # LINE通知実装（初期はスタブ/無効化）
│   │   │   └── auth/
│   │   │       ├── google_oauth.go   # Google OAuthクライアント
│   │   │       └── session_store.go  # Cookieセッションストア（Postgres）
│   │   │
│   │   └── interface/                # インターフェース層（外部との接点）
│   │       └── http/
│   │           ├── handler/
│   │           │   ├── site_handler.go
│   │           │   ├── checkin_handler.go   # /go/:siteId のリダイレクトもここ
│   │           │   └── auth_handler.go
│   │           ├── middleware/
│   │           │   ├── auth_middleware.go
│   │           │   └── cors.go
│   │           └── router.go
│   │
│   ├── migrations/                   # golang-migrate用SQL
│   │   └── 0001_init.up.sql
│   │
│   ├── test/                         # Ginkgo統合テスト（testcontainers-go利用）
│   │
│   ├── go.mod
│   ├── Dockerfile
│   └── Makefile
│
├── frontend/                         # React + Vite + Tailwind SPA
│   ├── src/
│   │   ├── api/                      # orval等で自動生成されるAPIクライアント置き場
│   │   ├── components/               # ボタン、サイトカード、ストリーク表示等
│   │   ├── hooks/                    # useSites, useCheckin等（React Query）
│   │   ├── pages/
│   │   │   └── Dashboard.tsx         # 1画面構成のメインページ
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── index.html
│   ├── package.json
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   ├── vite.config.ts
│   └── Dockerfile
│
├── docker-compose.yml                # app / frontend / db をまとめて起動
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
