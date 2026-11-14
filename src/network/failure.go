package network

type Kind string

const (
	ConnectionFailed Kind = "ConnectionFailed"
	ConnectionLost   Kind = "ConnectionLost"
	Timeout          Kind = "Timeout"
	ProtocolError    Kind = "ProtocolError"
	ServerError      Kind = "ServerError"
)

type Failure struct {
	Kind    Kind
	Message string
}

func (f *Failure) Error() string {
	if f.Message == "" {
		return string(f.Kind)
	}
	return string(f.Kind) + ": " + f.Message
}
