package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	infraauth "github.com/tortillaproduction/memory-tracker/internal/infrastructure/auth"
	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/batch"
	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/migration"
	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/notification"
	pg "github.com/tortillaproduction/memory-tracker/internal/infrastructure/persistence/postgres"
	httpinterface "github.com/tortillaproduction/memory-tracker/internal/interface/http"
	"github.com/tortillaproduction/memory-tracker/internal/interface/http/handler"
	usecaseauth "github.com/tortillaproduction/memory-tracker/internal/usecase/auth"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/checkin_site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/delete_site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/list_sites"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/notify_overdue_sites"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/register_site"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:5173")

	// AUTO_MIGRATE=true(デフォルト、開発時向け)ならアプリ起動時に自動でマイグレーションを実行する。
	// 本番運用ではdocker-compose.prod.ymlでAUTO_MIGRATE=falseにし、
	// 専用のmigrateコンテナで事前に1回だけ明示的に実行する運用に切り替える想定。
	if getEnvOrDefault("AUTO_MIGRATE", "true") == "true" {
		migrationPath := getEnvOrDefault("MIGRATION_PATH", "./migrations")
		logger.Info("running migrations", "path", migrationPath)
		if err := migration.Run(db, migrationPath); err != nil {
			logger.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
	}

	// --- 依存性の組み立て（手動DI） ---
	userRepo := pg.NewUserRepository(db)
	siteRepo := pg.NewSiteRepository(db)
	checkinRepo := pg.NewCheckInRepository(db)
	idGen := pg.NewULIDGenerator()

	sessionStore := infraauth.NewSessionStore(db)
	googleClient := infraauth.NewGoogleOAuthClient(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GOOGLE_REDIRECT_URL"),
	)

	registerSiteUC := register_site.NewUsecase(userRepo, siteRepo, checkinRepo, idGen)
	listSitesUC := list_sites.NewUsecase(siteRepo, checkinRepo)
	deleteSiteUC := delete_site.NewUsecase(siteRepo)
	checkinSiteUC := checkin_site.NewUsecase(siteRepo, checkinRepo, idGen)
	googleLoginUC := usecaseauth.NewUsecase(userRepo, idGen)

	authHandler := handler.NewAuthHandler(googleClient, sessionStore, googleLoginUC, userRepo, frontendURL, logger)

	// --- 通知バッチのセットアップ ---
	// RESEND_API_KEYが設定されていない場合は通知バッチを無効化する。
	// 開発時はキーなしで起動してもAPIサーバーとしては動作する。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if apiKey := os.Getenv("RESEND_API_KEY"); apiKey != "" {
		emailSender := notification.NewResendEmailSender(
			apiKey,
			getEnvOrDefault("EMAIL_FROM_ADDRESS", "onboarding@resend.dev"),
			getEnvOrDefault("EMAIL_FROM_NAME", "Memory Tracker"),
		)
		notifyUC := notify_overdue_sites.NewUsecase(db, emailSender, logger)
		scheduler := batch.NewNotificationScheduler(notifyUC, 15*time.Minute, logger)
		scheduler.Start(ctx)
	} else {
		logger.Warn("RESEND_API_KEY is not set, notification batch is disabled")
	}

	router := httpinterface.NewRouter(httpinterface.Dependencies{
		RegisterSiteUsecase: registerSiteUC,
		ListSiteUsecase:     listSitesUC,
		DeleteSiteUsecase:   deleteSiteUC,
		CheckinSiteUsecase:  checkinSiteUC,
		AuthHandler:         authHandler,
		SessionStore:        sessionStore,
		FrontendURL:         frontendURL,
		Logger:              logger,
	})

	addr := ":8080"
	logger.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
