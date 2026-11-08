package types

type Session struct {
	service string
}

func NewSession(service string) *Session {
	return &Session{service: service}
}

func (s *Session) Service() string {
	return s.service
}

func (s *Session) Equals(other Value) bool {
	otherSession, ok := other.(*Session)
	return ok && s == otherSession
}

func (s *Session) TypeName() string {
	return "session<" + s.service + ">"
}

func (s *Session) String() string {
	return s.TypeName()
}

func (s *Session) Display() string {
	return s.String()
}

func (*Session) isValue() {}
