package get_gate_candidates

import (
	"context"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/memory-tracker/internal/domain/gate"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/overdue"
)

// maxCandidates はゲートに一度に提示するサイトの最大数。
const maxCandidates = 3

type Candidate struct {
	SiteID    string
	Name      string
	URL       string
	OverdueBy time.Duration
}

type Result struct {
	// AlreadyDoneToday がtrueなら、今日(日本時間)は既にチェックインしたか脱出口で解除済み。
	// このときCandidatesは空で、アプリはゲートを出さない。
	AlreadyDoneToday bool
	Candidates       []Candidate
}

type Usecase struct {
	finder        overdue.Finder
	checkinRepo   checkin.Repository
	dismissalRepo gate.DismissalRepository
}

func NewUsecase(finder overdue.Finder, checkinRepo checkin.Repository, dismissalRepo gate.DismissalRepository) *Usecase {
	return &Usecase{finder: finder, checkinRepo: checkinRepo, dismissalRepo: dismissalRepo}
}

// Execute はゲートに提示する期限切れサイトを、超過時間の長い順に最大3件返す。
// 通知の二重送信防止(notification_logs)は考慮しない。通知済みのサイトも候補に出す。
func (uc *Usecase) Execute(ctx context.Context, userID user.ID, now time.Time) (Result, error) {
	done, err := uc.doneToday(ctx, userID, now)
	if err != nil {
		return Result{}, err
	}
	if done {
		return Result{AlreadyDoneToday: true, Candidates: []Candidate{}}, nil
	}

	sites, err := uc.finder.ListOverdue(ctx, now, overdue.Filter{UserID: string(userID), Limit: maxCandidates})
	if err != nil {
		return Result{}, err
	}

	candidates := make([]Candidate, 0, len(sites))
	for _, s := range sites {
		candidates = append(candidates, Candidate{
			SiteID:    s.SiteID,
			Name:      s.SiteName,
			URL:       s.SiteURL,
			OverdueBy: s.OverdueBy(now),
		})
	}
	return Result{Candidates: candidates}, nil
}

// doneToday は、今日(日本時間)に経路を問わずチェックインしたか、脱出口で解除したかを返す。
// 登録時の初回チェックイン(is_initial)は数えない。
func (uc *Usecase) doneToday(ctx context.Context, userID user.ID, now time.Time) (bool, error) {
	start := gate.StartOfDay(now)

	checkIns, err := uc.checkinRepo.FindAllByUserID(ctx, userID, start)
	if err != nil {
		return false, err
	}
	for _, c := range checkIns {
		if !c.CheckedAt().After(now) {
			return true, nil
		}
	}

	return uc.dismissalRepo.ExistsOn(ctx, userID, start)
}
