.PHONY: run test migrate-up migrate-down

run:
	go run ./cmd/api

test:
	ginkgo -r --randomize-all --randomize-suites

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1
