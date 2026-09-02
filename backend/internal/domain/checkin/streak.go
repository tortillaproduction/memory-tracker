package checkin

import (
	"sort"
	"time"
)

// CalculateUserStreak は「登録サイトのどれか1つでも期限内にチェックインした日」の
// 連続日数を計算する（ユーザー全体の日次ストリーク）。
// checkIns は対象ユーザーの全チェックインログ（新しい順である必要はない）。
func CalculateUserStreak(checkIns []*CheckIn, now time.Time) int {
	if len(checkIns) == 0 {
		return 0
	}

	// checkinがあった日付の集合を作る（同じ日の複数チェックインは1日として扱う）
	daySet := make(map[string]struct{})
	for _, c := range checkIns {
		daySet[c.CheckedAt().Format("2006-01-02")] = struct{}{}
	}

	streak := 0
	cursor := now
	for {
		key := cursor.Format("2006-01-02")
		if _, ok := daySet[key]; !ok {
			break
		}
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak
}

// CalculateSiteStreak は特定サイトについて、通知間隔を守り続けた連続回数を計算する。
// siteCheckIns は対象サイトのチェックインのみ、古い順にソートされている前提はなく内部でソートする。
func CalculateSiteStreak(siteCheckIns []*CheckIn, intervalHours int, now time.Time) int {
	if len(siteCheckIns) == 0 {
		return 0
	}
	sorted := make([]*CheckIn, len(siteCheckIns))
	copy(sorted, siteCheckIns)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CheckedAt().After(sorted[j].CheckedAt())
	})

	interval := time.Duration(intervalHours) * time.Hour
	streak := 1
	prev := sorted[0].CheckedAt()

	// 直近のチェックインが既に期限切れなら、ストリークは0（即座に0リセットのルール）
	if now.After(prev.Add(interval)) {
		return 0
	}

	for i := 1; i < len(sorted); i++ {
		curr := sorted[i].CheckedAt()
		if prev.Sub(curr) <= interval {
			streak++
			prev = curr
		} else {
			break
		}
	}
	return streak
}
