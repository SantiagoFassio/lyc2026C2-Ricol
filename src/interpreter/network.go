package interpreter

import (
	"fmt"
	"time"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/network"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

type networkTransport struct {
	interfaces map[string]*wire.Interface
}

type networkChannel struct {
	client           *network.Client
	serviceInterface *wire.Interface
}

func (i *Interpreter) EnableNetwork(interfaces map[string]*wire.Interface) {
	i.transport = &networkTransport{interfaces: interfaces}
}

func (t *networkTransport) open(service string, address common.Address, timeout time.Duration) (channel, *failure) {
	serviceInterface, ok := t.interfaces[service]
	if !ok {
		panic(fmt.Sprintf("No interface for service '%s'", service))
	}
	client, cause := network.Dial(address, serviceInterface, timeout)
	if cause != nil {
		return nil, failureOf(cause)
	}
	return &networkChannel{client: client, serviceInterface: serviceInterface}, nil
}

func (c *networkChannel) call(method string, arguments []types.Value) (types.Value, *failure, error) {
	serviceMethod, ok := c.serviceInterface.Method(method)
	if !ok {
		panic(fmt.Sprintf("Service '%s' has no method '%s'", c.serviceInterface.Name, method))
	}
	result, cause, err := c.client.Call(serviceMethod, arguments)
	if cause != nil {
		return nil, failureOf(cause), nil
	}
	return result, nil, err
}

func (c *networkChannel) close() {
	c.client.Close()
}

func failureOf(cause *network.Failure) *failure {
	return &failure{variant: string(cause.Kind), message: cause.Message}
}
