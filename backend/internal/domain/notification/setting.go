package notification

import (
	"errors"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type ID string

// Mode はユーザーが選んだ通知の受け取り方。
type Mode string

const (
	// ModeEmail は期限切れサイトをメールで通知する(既定)。
	ModeEmail Mode = "email"
	// ModeGate はAndroidアプリ(ゲート)で提示する。ゲート端末が未接続の間はメールも併用する。
	ModeGate Mode = "gate"
)

var ErrInvalidMode = errors.New("invalid notification mode")

// ParseMode は外部入力の文字列をModeに変換する。
func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeEmail, ModeGate:
		return Mode(s), nil
	}
	return "", ErrInvalidMode
}

type Setting struct {
	id           ID
	userID       user.ID
	emailEnabled bool
	lineEnabled  bool
	lineUserID   string
	mode         Mode
}

func NewSetting(id ID, userID user.ID) *Setting {
	return &Setting{
		id:           id,
		userID:       userID,
		emailEnabled: true,
		lineEnabled:  false,
		mode:         ModeEmail,
	}
}

func (s *Setting) ID() ID             { return s.id }
func (s *Setting) UserID() user.ID    { return s.userID }
func (s *Setting) EmailEnabled() bool { return s.emailEnabled }
func (s *Setting) LineEnabled() bool  { return s.lineEnabled }
func (s *Setting) LineUserID() string { return s.lineUserID }
func (s *Setting) Mode() Mode         { return s.mode }

func (s *Setting) SetEmailEnabled(v bool) { s.emailEnabled = v }
func (s *Setting) SetLineEnabled(enabled bool, lineUserID string) {
	s.lineEnabled = enabled
	s.lineUserID = lineUserID
}
func (s *Setting) SetMode(m Mode) { s.mode = m }
