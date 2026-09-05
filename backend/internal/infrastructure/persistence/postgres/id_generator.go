package postgres

import (
	"github.com/tortillaproduction/study-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/study-tracker/internal/domain/site"
	"github.com/tortillaproduction/study-tracker/internal/domain/user"

	"github.com/oklog/ulid/v2"
)

// ULIDGenerator は register_site / checkin_site usecase の IDGenerator インターフェースを満たす。
// ULIDはUUIDと違い辞書順=生成順になるため、DBのインデックス効率が良い。
type ULIDGenerator struct{}

func NewULIDGenerator() *ULIDGenerator {
	return &ULIDGenerator{}
}

func (g *ULIDGenerator) NewSiteID() site.ID {
	return site.ID(ulid.Make().String())
}

func (g *ULIDGenerator) NewCheckInID() checkin.ID {
	return checkin.ID(ulid.Make().String())
}

func (g *ULIDGenerator) NewUserID() user.ID {
	return user.ID(ulid.Make().String())
}
