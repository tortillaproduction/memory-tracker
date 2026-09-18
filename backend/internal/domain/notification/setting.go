package notification

import "github.com/tortillaproduction/memory-tracker/internal/domain/user"

type ID string

type Setting struct {
	id                            ID
	userID                        user.ID
	emailEnabled                  bool
	lineEnabled                   bool
	lineUserID                    string
	pushEnabled                   bool
	disableEmailWhenPushAvailable bool
}

func NewSetting(id ID, userID user.ID) *Setting {
	return &Setting{
		id:                            id,
		userID:                        userID,
		emailEnabled:                  true,
		lineEnabled:                   false,
		pushEnabled:                   false,
		disableEmailWhenPushAvailable: true,
	}
}

func (s *Setting) ID() ID             { return s.id }
func (s *Setting) UserID() user.ID    { return s.userID }
func (s *Setting) EmailEnabled() bool { return s.emailEnabled }
func (s *Setting) LineEnabled() bool  { return s.lineEnabled }
func (s *Setting) LineUserID() string { return s.lineUserID }
func (s *Setting) PushEnabled() bool  { return s.pushEnabled }

// DisableEmailWhenPushAvailable は、有効なプッシュ購読が存在する間はメール通知を
// 送らないようにするかどうかのユーザー設定。trueがデフォルトで、インストール後に
// メールとプッシュが二重に届く煩わしさを避ける。
func (s *Setting) DisableEmailWhenPushAvailable() bool { return s.disableEmailWhenPushAvailable }

func (s *Setting) SetPushEnabled(v bool)                   { s.pushEnabled = v }
func (s *Setting) SetDisableEmailWhenPushAvailable(v bool) { s.disableEmailWhenPushAvailable = v }
func (s *Setting) SetEmailEnabled(v bool)                  { s.emailEnabled = v }
func (s *Setting) SetLineEnabled(enabled bool, lineUserID string) {
	s.lineEnabled = enabled
	s.lineUserID = lineUserID
}
