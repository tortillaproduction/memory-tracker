package batch

import (
	"context"
	"log/slog"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/usecase/notify_overdue_sites"
)

// NotificationScheduler はgoroutineで常駐し、一定間隔で通知バッチを実行する。
type NotificationScheduler struct {
	usecase  *notify_overdue_sites.Usecase
	interval time.Duration
	logger   *slog.Logger
}

func NewNotificationScheduler(
	uc *notify_overdue_sites.Usecase,
	interval time.Duration,
	logger *slog.Logger,
) *NotificationScheduler {
	return &NotificationScheduler{usecase: uc, interval: interval, logger: logger}
}

// Start はバックグラウンドでスケジューラーを起動する。
// ctx がキャンセルされると自動的に停止する。
func (s *NotificationScheduler) Start(ctx context.Context) {
	go func() {
		s.logger.Info("notification scheduler started", "interval", s.interval)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("notification scheduler stopped")
				return
			case <-ticker.C:
				if err := s.usecase.Execute(ctx); err != nil {
					s.logger.Error("notification batch failed", "error", err)
				}
			}
		}
	}()
}
