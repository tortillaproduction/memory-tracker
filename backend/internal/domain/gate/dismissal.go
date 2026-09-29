package gate

import (
	"context"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type DismissalID string

// DismissalRepository はゲートの脱出口(サイトを開かずに解除)を使った記録を扱う。
// 記録は日本時間の1日につき1件。
type DismissalRepository interface {
	// Record は指定日の解除を記録する。同じ日に既に記録があれば何もしない。
	Record(ctx context.Context, id DismissalID, userID user.ID, day time.Time) error
	ExistsOn(ctx context.Context, userID user.ID, day time.Time) (bool, error)
}
