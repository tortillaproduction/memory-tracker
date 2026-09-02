module github.com/tortillaproduction/study-tracker

go 1.23

require (
	github.com/lib/pq v1.10.9
	github.com/oklog/ulid/v2 v2.1.0
	github.com/onsi/ginkgo/v2 v2.20.0
	github.com/onsi/gomega v1.34.1
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20240727154555-813a5fbdbec8 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.23.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// 開発時に追加予定:
// github.com/testcontainers/testcontainers-go        (統合テスト用にPostgresコンテナを起動)
// github.com/testcontainers/testcontainers-go/modules/postgres
// github.com/pressly/goose/v3                         (マイグレーション)
// github.com/deepmap/oapi-codegen                     (OpenAPIからのコード生成)
