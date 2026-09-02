package plan

// Plan は値オブジェクト。プランごとのビジネスルール（件数制限など）をここに集約する。
// 「無料は5件まで」という数字が変わっても、影響範囲はこのファイルだけに閉じる。
type Plan struct {
	planType    Type
	maxSites    int // -1 は無制限
}

type Type string

const (
	TypeFree    Type = "free"
	TypePremium Type = "premium"
)

func Free() Plan {
	return Plan{planType: TypeFree, maxSites: 5}
}

func Premium() Plan {
	return Plan{planType: TypePremium, maxSites: -1}
}

func (p Plan) Type() Type { return p.planType }

// CanRegisterMoreSites は現在の登録件数から見て、あと1件登録できるかを判定する。
// usecase層はこのメソッドを呼ぶだけでよく、「5件」というマジックナンバーを知らなくてよい。
func (p Plan) CanRegisterMoreSites(currentCount int) bool {
	if p.maxSites == -1 {
		return true
	}
	return currentCount < p.maxSites
}

func (p Plan) MaxSites() int { return p.maxSites }
