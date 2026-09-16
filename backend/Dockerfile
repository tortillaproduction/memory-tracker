# --- build stage ---
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

# --- runtime stage ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/api .
# AUTO_MIGRATE=true(開発時)でmain.goがこのディレクトリを読むため同梱する。
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./api"]
