package notification

import "github.com/tortillaproduction/memory-tracker/internal/domain/user"

type ID string

type Setting struct {
	id           ID
	userID       user.ID
	emailEnabled bool
	lineEnabled  bool
	lineUserID   string
}

func NewSetting(id ID, userID user.ID) *Setting {
	return &Setting{
		id:           id,
		userID:       userID,
		emailEnabled: true,
		lineEnabled:  false,
	}
}

func (s *Setting) ID() ID             { return s.id }
func (s *Setting) UserID() user.ID    { return s.userID }
func (s *Setting) EmailEnabled() bool { return s.emailEnabled }
func (s *Setting) LineEnabled() bool  { return s.lineEnabled }
func (s *Setting) LineUserID() string { return s.lineUserID }
