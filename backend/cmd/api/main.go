package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	pg "github.com/tortillaproduction/study-tracker/internal/infrastructure/persistence/postgres"
	httpinterface "github.com/tortillaproduction/study-tracker/internal/interface/http"
	"github.com/tortillaproduction/study-tracker/internal/usecase/checkin_site"
	"github.com/tortillaproduction/study-tracker/internal/usecase/register_site"
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

	// --- 依存性の組み立て（手動DI） ---
	userRepo := pg.NewUserRepository(db)
	siteRepo := pg.NewSiteRepository(db)
	checkinRepo := pg.NewCheckInRepository(db)
	idGen := pg.NewULIDGenerator()

	registerSiteUC := register_site.NewUsecase(userRepo, siteRepo, checkinRepo, idGen)
	checkinSiteUC := checkin_site.NewUsecase(siteRepo, checkinRepo, idGen)

	router := httpinterface.NewRouter(httpinterface.Dependencies{
		RegisterSiteUsecase: registerSiteUC,
		CheckinSiteUsecase:  checkinSiteUC,
		Logger:              logger,
	})

	addr := ":8080"
	logger.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
