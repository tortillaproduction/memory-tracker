package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

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
	"github.com/tortillaproduction/memory-tracker/internal/usecase/push_delivery"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/register_site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/subscribe_push"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/unsubscribe_push"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/update_notification_preferences"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DATABASE_URL")
	// Neonのpooled接続(PgBouncer transaction pooling)はサーバーサイドprepared statementの
	// 状態をコネクション間で保持できないため、simple query protocolを強制する。
	// これをしないと "unnamed prepared statement does not exist" や
	// "bind message has N result formats but query has M columns" のようなエラーが
	// 同時アクセス時に散発する。
	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		logger.Error("failed to parse DATABASE_URL", "error", err)
		os.Exit(1)
	}
	connConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	db := stdlib.OpenDB(*connConfig)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close db", "error", err)
		} else {
			logger.Info("db connection closed")
		}
	}()

	frontendURL, err := normalizeFrontendURL(getEnvOrDefault("FRONTEND_URL", "http://localhost:5173"))
	if err != nil {
		logger.Error("invalid FRONTEND_URL", "error", err)
		os.Exit(1)
	}

	// SESSION_SECRETはメールのチェックインリンク(/go/{siteId}?token=...)の署名鍵として使う。
	// この値が漏洩・推測可能だと誰でもチェックイントークンを偽造できてしまうため、
	// 空文字のまま起動することを許さない。
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		logger.Error("SESSION_SECRET must be set")
		os.Exit(1)
	}

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
	notificationRepo := pg.NewNotificationRepository(db)
	pushSubRepo := pg.NewPushSubscriptionRepository(db)

	sessionStore := infraauth.NewSessionStore(db)
	checkinTokenIssuer := infraauth.NewCheckinTokenIssuer(sessionSecret)
	googleClient := infraauth.NewGoogleOAuthClient(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GOOGLE_REDIRECT_URL"),
	)

	registerSiteUC := register_site.NewUsecase(userRepo, siteRepo, checkinRepo, idGen)
	listSitesUC := list_sites.NewUsecase(siteRepo, checkinRepo)
	deleteSiteUC := delete_site.NewUsecase(siteRepo)
	checkinSiteUC := checkin_site.NewUsecase(siteRepo, checkinRepo, idGen)
	googleLoginUC := usecaseauth.NewUsecase(userRepo, notificationRepo, idGen)

	authHandler := handler.NewAuthHandler(googleClient, sessionStore, googleLoginUC, userRepo, frontendURL, logger)

	// --- プッシュ通知のセットアップ ---
	// VAPID_PUBLIC_KEY/VAPID_PRIVATE_KEYが未設定の場合、プッシュ送信はno-op実装にフォールバックする。
	// これにより設定/購読APIやDB周りは常に動作しつつ、実際のプッシュ配信だけが無効化される
	// (=既存のメール専用環境でも安全に起動できる)。
	vapidPublicKey := os.Getenv("VAPID_PUBLIC_KEY")
	vapidPrivateKey := os.Getenv("VAPID_PRIVATE_KEY")
	var pushSender notification.PushSender
	if vapidPublicKey != "" && vapidPrivateKey != "" {
		pushSender = notification.NewWebPushSender(vapidPublicKey, vapidPrivateKey, getEnvOrDefault("VAPID_SUBJECT", "mailto:support@example.com"), logger)
	} else {
		logger.Warn("VAPID_PUBLIC_KEY/VAPID_PRIVATE_KEY is not set, push notifications are disabled")
		pushSender = notification.NewNoopPushSender()
	}

	subscribePushUC := subscribe_push.NewUsecase(pushSubRepo, notificationRepo, idGen)
	unsubscribePushUC := unsubscribe_push.NewUsecase(pushSubRepo, notificationRepo)
	notificationPreferencesUC := update_notification_preferences.NewUsecase(notificationRepo)
	pushTracker := push_delivery.NewTracker(infraauth.NewPushAckTokenIssuer(sessionSecret), logger)
	pushHandler := handler.NewPushHandler(vapidPublicKey, subscribePushUC, unsubscribePushUC, notificationPreferencesUC, pushTracker, logger)

	// --- 通知バッチのセットアップ ---
	// RESEND_API_KEYが設定されていない場合はメール送信をno-opにする(プッシュだけの運用も許容する)。
	// 開発時はキーなしで起動してもAPIサーバーとしては動作する。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var emailSender notification.EmailSender
	if apiKey := os.Getenv("RESEND_API_KEY"); apiKey != "" {
		emailSender = notification.NewResendEmailSender(
			apiKey,
			getEnvOrDefault("EMAIL_FROM_ADDRESS", "onboarding@resend.dev"),
			getEnvOrDefault("EMAIL_FROM_NAME", "Memory Tracker"),
		)
	} else {
		logger.Warn("RESEND_API_KEY is not set, email notifications are disabled")
		emailSender = notification.NewNoopEmailSender()
	}

	notifyUC := notify_overdue_sites.NewUsecase(db, emailSender, pushSender, pushSubRepo, notificationRepo, logger, frontendURL, checkinTokenIssuer).WithPushDeliveryTracker(pushTracker)
	scheduler := batch.NewNotificationScheduler(notifyUC, notificationInterval(logger), logger)
	schedulerDone := scheduler.Start(ctx)

	router := httpinterface.NewRouter(httpinterface.Dependencies{
		RegisterSiteUsecase:  registerSiteUC,
		ListSiteUsecase:      listSitesUC,
		DeleteSiteUsecase:    deleteSiteUC,
		CheckinSiteUsecase:   checkinSiteUC,
		AuthHandler:          authHandler,
		PushHandler:          pushHandler,
		SessionStore:         sessionStore,
		CheckinTokenVerifier: checkinTokenIssuer,
		FrontendURL:          frontendURL,
		Logger:               logger,
	})

	addr := ":8080"
	srv := &http.Server{Addr: addr, Handler: router}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown error", "error", err)
	} else {
		logger.Info("http server stopped")
	}

	<-schedulerDone
	logger.Info("graceful shutdown complete")
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

const defaultNotificationInterval = 5 * time.Minute

// notificationInterval returns the batch interval, overridable via
// NOTIFICATION_INTERVAL (e.g. "1m") so notifications can be verified quickly
// in development. Invalid or non-positive values fall back to the default.
func notificationInterval(logger *slog.Logger) time.Duration {
	raw := os.Getenv("NOTIFICATION_INTERVAL")
	if raw == "" {
		return defaultNotificationInterval
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		logger.Warn("invalid NOTIFICATION_INTERVAL, using default", "value", raw, "default", defaultNotificationInterval)
		return defaultNotificationInterval
	}
	return d
}

// normalizeFrontendURL trims whitespace and a trailing slash, then verifies
// the result is an absolute URL with a scheme (e.g. "https://example.com").
// FRONTEND_URL is used both as the CORS Access-Control-Allow-Origin value
// and the post-login OAuth redirect target, so a malformed value (such as a
// bare domain missing "https://") must fail fast at startup rather than
// silently breaking both at runtime.
func normalizeFrontendURL(raw string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")

	u, err := url.ParseRequestURI(trimmed)
	if err != nil {
		return "", fmt.Errorf("FRONTEND_URL %q is not a valid URL: %w", raw, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("FRONTEND_URL %q must be an absolute URL with a scheme (e.g. https://example.com)", raw)
	}

	return trimmed, nil
}
