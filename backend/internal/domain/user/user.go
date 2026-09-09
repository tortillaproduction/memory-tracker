package user

import (
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/plan"
)

// User はドメインのエンティティ。IDで同一性を判定する。
type User struct {
	id         ID
	googleID   string
	email      string
	name       string
	pictureURL string
	plan       plan.Plan
	createdAt  time.Time
}

type ID string

func NewUser(id ID, googleID, email, name, pictureURL string) *User {
	return &User{
		id:         id,
		googleID:   googleID,
		email:      email,
		name:       name,
		pictureURL: pictureURL,
		plan:       plan.Free(), // デフォルトは無料プラン
		createdAt:  time.Now(),
	}
}

func (u *User) ID() ID             { return u.id }
func (u *User) GoogleID() string   { return u.googleID }
func (u *User) Email() string      { return u.email }
func (u *User) Name() string       { return u.name }
func (u *User) PictureURL() string { return u.pictureURL }
func (u *User) Plan() plan.Plan    { return u.plan }

// UpdatePicture はGoogle側でプロフィール画像が変わった場合に、再ログイン時に反映するための操作。
func (u *User) UpdatePicture(pictureURL string) {
	u.pictureURL = pictureURL
}

// UpgradeTo はプラン変更のドメイン操作。将来Stripe連携時にusecaseから呼ばれる。
func (u *User) UpgradeTo(p plan.Plan) {
	u.plan = p
}
