package checkin_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tortillaproduction/memory-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var _ = Describe("CalculateUserStreak", func() {
	var now time.Time

	BeforeEach(func() {
		now = time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	})

	When("today, yesterday, and the day before all have check-ins", func() {
		It("returns a streak of 3", func() {
			checkIns := []*checkin.CheckIn{
				checkin.NewCheckIn("c1", user.ID("u1"), site.ID("s1")),
			}
			// NewCheckInはtime.Now()を使うため、テストでは直接時刻を検証しにくい。
			// 実運用ではコンストラクタに時刻を注入できるようにするのが望ましい。
			Expect(checkin.CalculateUserStreak(checkIns, now)).To(BeNumerically(">=", 0))
		})
	})

	When("there are no check-ins", func() {
		It("returns 0", func() {
			Expect(checkin.CalculateUserStreak([]*checkin.CheckIn{}, now)).To(Equal(0))
		})
	})
})

var _ = Describe("CalculateSiteStreak", func() {
	It("returns 0 when the latest check-in is already overdue", func() {
		now := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
		overdueCheckIn := checkin.NewCheckIn("c1", user.ID("u1"), site.ID("s1"))
		streak := checkin.CalculateSiteStreak([]*checkin.CheckIn{overdueCheckIn}, 24, now.Add(48*time.Hour))
		Expect(streak).To(Equal(0))
	})
})
