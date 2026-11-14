package interpreter

import (
	"bytes"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/network"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

type networkCallTestCase struct {
	name      string
	method    string
	arguments []types.Value
	expected  types.Value
}

type networkSessionTestCase struct {
	name           string
	address        func(t *testing.T) common.Address
	expectedOutput string
}

const networkTimeout = 2 * time.Second

func accountInterfaces() map[string]*wire.Interface {
	return map[string]*wire.Interface{"Account": wire.NewInterface(types.NewServiceType("Account", []types.ServiceMethod{
		{Name: "balance", Signature: types.Signature{Return: types.Int}},
		{Name: "deposit", Signature: types.Signature{Params: []types.Type{types.Int}, Return: types.Int}},
		{Name: "rename", Signature: types.Signature{Params: []types.Type{types.Str}, Return: types.Str}},
		{Name: "fail", Signature: types.Signature{Return: types.Void}},
	}))}
}

func localAddress(addr net.Addr) common.Address {
	return common.Address{Host: "127.0.0.1", Port: addr.(*net.TCPAddr).Port}
}

func startAccountServer(t *testing.T, interfaces map[string]*wire.Interface) common.Address {
	t.Helper()

	server, err := network.Listen(common.Address{Host: "127.0.0.1", Port: 0}, interfaces["Account"])
	if err != nil {
		t.Fatalf("network.Listen() unexpected error: %v", err)
	}
	balance := int64(0)
	served := make(chan error, 1)
	go func() {
		served <- server.Serve(func(method *wire.Method, arguments []types.Value) (types.Value, error) {
			switch method.Name {
			case "balance":
				return types.NewInteger(balance), nil
			case "deposit":
				amount, _ := arguments[0].(types.Number).IntegerValue()
				balance += amount
				return types.NewInteger(balance), nil
			case "rename":
				return types.NewString("hello " + arguments[0].Display()), nil
			}
			return nil, errors.New("division by zero")
		})
	}()
	t.Cleanup(func() {
		server.Close()
		<-served
	})
	return localAddress(server.Addr())
}

func closedAddress(t *testing.T) common.Address {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() unexpected error: %v", err)
	}
	address := localAddress(listener.Addr())
	listener.Close()
	return address
}

func openAccount(t *testing.T, interfaces map[string]*wire.Interface, address common.Address) channel {
	t.Helper()

	transport := &networkTransport{interfaces: interfaces}
	accountChannel, cause := transport.open("Account", address, networkTimeout)
	if cause != nil {
		t.Fatalf("open() unexpected failure: %v", cause.value())
	}
	t.Cleanup(accountChannel.close)
	return accountChannel
}

func TestEnableNetwork(t *testing.T) {
	interfaces := accountInterfaces()
	interpreter := NewInterpreter(nil, nil, io.Discard)

	interpreter.EnableNetwork(interfaces)

	transport, ok := interpreter.transport.(*networkTransport)
	if !ok {
		t.Fatalf("transport = %T; want *networkTransport", interpreter.transport)
	}
	if transport.interfaces["Account"] != interfaces["Account"] {
		t.Errorf("transport interfaces = %v; want %v", transport.interfaces, interfaces)
	}
}

func TestNetworkChannelCall(t *testing.T) {
	interfaces := accountInterfaces()
	accountChannel := openAccount(t, interfaces, startAccountServer(t, interfaces))
	testCases := []networkCallTestCase{
		{"first deposit", "deposit", []types.Value{types.NewInteger(30)}, types.NewInteger(30)},
		{"second deposit", "deposit", []types.Value{types.NewInteger(12)}, types.NewInteger(42)},
		{"method without parameters", "balance", nil, types.NewInteger(42)},
		{"method with a string", "rename", []types.Value{types.NewString("Ricol")}, types.NewString("hello Ricol")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, cause, err := accountChannel.call(testCase.method, testCase.arguments)

			if cause != nil || err != nil {
				t.Fatalf("call(%s) unexpected failure %v, error %v", testCase.method, cause, err)
			}
			if !result.Equals(testCase.expected) {
				t.Errorf("call(%s) = %v; want %v", testCase.method, result, testCase.expected)
			}
		})
	}
}

func TestNetworkChannelFailure(t *testing.T) {
	interfaces := accountInterfaces()
	accountChannel := openAccount(t, interfaces, startAccountServer(t, interfaces))

	result, cause, err := accountChannel.call("fail", nil)

	if result != nil || err != nil {
		t.Fatalf("call(fail) = %v, error %v; want only a failure", result, err)
	}
	if cause == nil {
		t.Fatalf("call(fail) failure = nil; want a ServerError")
	}
	if expected := types.NewEnum("SessionError", "ServerError", []types.Value{types.NewString("division by zero")}); !cause.value().Equals(expected) {
		t.Errorf("call(fail) failure = %v; want %v", cause.value(), expected)
	}
}

func TestNetworkChannelLocalError(t *testing.T) {
	interfaces := accountInterfaces()
	accountChannel := openAccount(t, interfaces, startAccountServer(t, interfaces))

	result, cause, err := accountChannel.call("rename", []types.Value{types.NewString("\xFF")})

	if result != nil || cause != nil {
		t.Errorf("call(rename) = %v, failure %v; want only a local error", result, cause)
	}
	if !errors.Is(err, wire.ErrInvalidData) {
		t.Errorf("call(rename) error = %v; want %v", err, wire.ErrInvalidData)
	}
}

func TestNetworkTransportOpenFailure(t *testing.T) {
	address := closedAddress(t)
	transport := &networkTransport{interfaces: accountInterfaces()}

	accountChannel, cause := transport.open("Account", address, networkTimeout)

	if accountChannel != nil {
		accountChannel.close()
		t.Fatalf("open() = a channel; want a failure")
	}
	expected := types.NewEnum("SessionError", "ConnectionFailed",
		[]types.Value{types.NewString("cannot connect to Account at " + address.String() + ": connection refused")})
	if cause == nil || !cause.value().Equals(expected) {
		t.Errorf("open() failure = %v; want %v", cause, expected)
	}
}

func TestNetworkTransportInvariantPanics(t *testing.T) {
	interfaces := accountInterfaces()
	accountChannel := openAccount(t, interfaces, startAccountServer(t, interfaces))
	transport := &networkTransport{interfaces: interfaces}

	assertPanics(t, "open() of a service without interface", func() { transport.open("Other", closedAddress(t), networkTimeout) })
	assertPanics(t, "call() of a method that is not in the service", func() { accountChannel.call("transfer", nil) })
}

func TestSessionsOverTheNetwork(t *testing.T) {
	interfaces := accountInterfaces()
	testCases := []networkSessionTestCase{
		{"calls until a failure of the server", func(t *testing.T) common.Address { return startAccountServer(t, interfaces) },
			`3042SessionError.ServerError("division by zero")`},
		{"server that is not listening", closedAddress,
			`SessionError.ConnectionFailed("cannot connect to Account at {address}: connection refused")`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			address := testCase.address(t)
			d := distanceMap{}
			program := []common.Statement{
				varDeclaration("conn", types.NamedConnectionType{Service: types.NamedType{Name: "Account"}}, connectExpression("Account", address.String())),
				d.accountSession(0, []common.Statement{
					printStatement(d.remoteCall("c", 0, "deposit", integerLiteral(30))),
					printStatement(d.remoteCall("c", 0, "deposit", integerLiteral(12))),
					common.NewExpressionStatement(d.remoteCall("c", 0, "fail")),
					printStatement(stringLiteral("after the failure")),
				}, printStatement(d.variable("e", 0))),
			}
			output := &bytes.Buffer{}
			interpreter := NewInterpreter(program, d, output)
			interpreter.EnableNetwork(interfaces)

			if err := interpreter.Interpret(); err != nil {
				t.Fatalf("Interpret() unexpected error: %v", err)
			}

			if expected := strings.ReplaceAll(testCase.expectedOutput, "{address}", address.String()); output.String() != expected {
				t.Errorf("Interpret() output = %q; want %q", output, expected)
			}
		})
	}
}
