package interpreter

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

const (
	defaultSessionTimeout = 5000 * time.Millisecond
	maxSessionTimeout     = math.MaxInt64 / int64(time.Millisecond)
)

type transport interface {
	open(service string, address common.Address, timeout time.Duration) (channel, *failure)
	listen(service string, address common.Address) (listener, error)
}

type channel interface {
	call(method string, arguments []types.Value) (types.Value, *failure, error)
	close()
}

type listener interface {
	serve(handler methodHandler) error
}

type methodHandler func(method string, arguments []types.Value) (types.Value, error)

type failure struct {
	variant string
	message string
}

func (f *failure) value() types.Enum {
	variant, ok := types.SessionError.Variant(f.variant)
	if !ok {
		panic(fmt.Sprintf("Unknown session failure '%s'", f.variant))
	}
	if len(variant.Fields) == 0 {
		return types.NewEnum(types.SessionError.Name, variant.Name, nil)
	}
	return types.NewEnum(types.SessionError.Name, variant.Name, []types.Value{types.NewString(f.message)})
}

type sessionFailure struct {
	session *types.Session
	failure *failure
}

func (*sessionFailure) Error() string { return "session failure signal" }

type unavailableTransport struct{}

func (unavailableTransport) open(service string, address common.Address, _ time.Duration) (channel, *failure) {
	return nil, &failure{
		variant: "ConnectionFailed",
		message: fmt.Sprintf("cannot connect to %s at %v: network runtime not available", service, address),
	}
}

func (unavailableTransport) listen(string, common.Address) (listener, error) {
	return nil, errors.New("network runtime not available")
}
