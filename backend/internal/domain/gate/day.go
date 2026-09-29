package gate

import "time"

// Location はゲートの「今日」を判定するタイムゾーン(日本時間)。
// 日本には夏時間がないため固定オフセットで表し、実行環境のtzdataに依存しない。
var Location = time.FixedZone("Asia/Tokyo", 9*60*60)

// StartOfDay は now が属する日本時間の日の0時を返す。
func StartOfDay(now time.Time) time.Time {
	t := now.In(Location)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, Location)
}
