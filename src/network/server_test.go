package network

import (
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

type serverCallTestCase struct {
	name       string
	methodName string
	arguments  []types.Value
	expected   types.Value
}

type mismatchTestCase struct {
	name             string
	serviceInterface *wire.Interface
}

type brokenConnectionTestCase struct {
	name             string
	sent             []byte
	expectedResponse []byte
}

type invalidRequestResponseTestCase struct {
	name            string
	request         []byte
	expectedMessage string
}

func listenLocally(t *testing.T, serviceInterface *wire.Interface) *Server {
	t.Helper()

	server, err := Listen(common.Address{Host: "127.0.0.1", Port: 0}, serviceInterface)
	if err != nil {
		t.Fatalf("Listen() unexpected error: %v", err)
	}
	return server
}

func startServer(t *testing.T, serviceInterface *wire.Interface, handler Handler) common.Address {
	t.Helper()

	server := listenLocally(t, serviceInterface)
	served := make(chan error, 1)
	go func() { served <- server.Serve(handler) }()
	t.Cleanup(func() {
		server.Close()
		if err := <-served; err != nil {
			t.Errorf("Serve() unexpected error: %v", err)
		}
	})
	return addressOf(server.Addr())
}

func integerArgument(value types.Value) int64 {
	number, _ := value.(types.Number).IntegerValue()
	return number
}

func accountHandler() Handler {
	balance := int64(0)
	return func(method *wire.Method, arguments []types.Value) (types.Value, error) {
		switch method.Name {
		case "balance":
			return integer(balance), nil
		case "deposit":
			amount := integerArgument(arguments[0])
			balance += amount
			return movement(amount, balance), nil
		case "withdraw":
			amount := integerArgument(arguments[0])
			if amount > balance {
				return withdrawResult("InsufficientFunds", integer(balance)), nil
			}
			balance -= amount
			return withdrawResult("Ok", movement(-amount, balance)), nil
		case "rename":
			return types.NewString("hello " + arguments[0].Display()), nil
		case "reset":
			balance = 0
			return nil, nil
		}
		return nil, fmt.Errorf("unknown method '%s'", method.Name)
	}
}

func mustCall(t *testing.T, client *Client, serviceInterface *wire.Interface, methodName string, arguments ...types.Value) types.Value {
	t.Helper()

	result, failure, err := client.Call(methodOf(t, serviceInterface, methodName), arguments)
	if failure != nil || err != nil {
		t.Fatalf("Call(%s) unexpected failure %v, error %v", methodName, failure, err)
	}
	return result
}

func assertBalance(t *testing.T, client *Client, expected int64) {
	t.Helper()

	if balance := mustCall(t, client, accountInterface(), "balance"); !sameValue(balance, integer(expected)) {
		t.Errorf("balance = %v; want %d", balance, expected)
	}
}

func exchange(t *testing.T, address common.Address, sent []byte) []byte {
	t.Helper()

	conn, err := net.Dial("tcp", address.String())
	if err != nil {
		t.Fatalf("net.Dial() unexpected error: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(testTimeout))
	if _, err := conn.Write(sent); err != nil {
		t.Fatalf("Write() unexpected error: %v", err)
	}
	conn.(*net.TCPConn).CloseWrite()
	received, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("ReadAll() unexpected error: %v", err)
	}
	return received
}

func TestServeKeepsTheStateBetweenCalls(t *testing.T) {
	account := accountInterface()
	client := dial(t, startServer(t, account, accountHandler()), account, testTimeout)
	testCases := []serverCallTestCase{
		{"deposit", "deposit", []types.Value{integer(100)}, movement(100, 100)},
		{"withdraw", "withdraw", []types.Value{integer(30)}, withdrawResult("Ok", movement(-30, 70))},
		{"withdraw more than the balance", "withdraw", []types.Value{integer(500)}, withdrawResult("InsufficientFunds", integer(70))},
		{"balance", "balance", nil, integer(70)},
		{"string", "rename", []types.Value{types.NewString("Ricol")}, types.NewString("hello Ricol")},
		{"void", "reset", nil, nil},
		{"balance after the reset", "balance", nil, integer(0)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if result := mustCall(t, client, account, testCase.methodName, testCase.arguments...); !sameValue(result, testCase.expected) {
				t.Errorf("Call(%s) = %v; want %v", testCase.methodName, result, testCase.expected)
			}
		})
	}
}

func TestServeRecursiveValues(t *testing.T) {
	lists := listInterface()
	address := startServer(t, lists, func(method *wire.Method, arguments []types.Value) (types.Value, error) {
		length := 0
		for node := arguments[0].(types.Enum); node.Variant() == "Node"; node = node.Fields()[1].(types.Enum) {
			length++
		}
		return integer(int64(length)), nil
	})
	client := dial(t, address, lists, testTimeout)

	if result := mustCall(t, client, lists, "length", list(1000)); !sameValue(result, integer(1000)) {
		t.Errorf("length = %v; want 1000", result)
	}
}

func TestServeSharesTheStateBetweenSessions(t *testing.T) {
	account := accountInterface()
	address := startServer(t, account, accountHandler())
	first := dial(t, address, account, testTimeout)
	second := dial(t, address, account, testTimeout)

	mustCall(t, first, account, "deposit", integer(10))
	mustCall(t, second, account, "deposit", integer(5))
	assertBalance(t, first, 15)
	first.Close()

	assertBalance(t, dial(t, address, account, testTimeout), 15)
}

func TestServeRunsOneCallAtATime(t *testing.T) {
	account := accountInterface()
	var running atomic.Int32
	var overlapped atomic.Bool
	calls := 0
	address := startServer(t, account, func(*wire.Method, []types.Value) (types.Value, error) {
		if running.Add(1) > 1 {
			overlapped.Store(true)
		}
		time.Sleep(5 * time.Millisecond)
		calls++
		running.Add(-1)
		return integer(int64(calls)), nil
	})
	clients := make([]*Client, 4)
	for i := range clients {
		clients[i] = dial(t, address, account, testTimeout)
	}
	balance := methodOf(t, account, "balance")
	var mutex sync.Mutex
	var results []int64
	var group sync.WaitGroup

	for _, client := range clients {
		group.Go(func() {
			for range 5 {
				result, failure, err := client.Call(balance, nil)
				if failure != nil || err != nil {
					t.Errorf("Call() unexpected failure %v, error %v", failure, err)
					return
				}
				mutex.Lock()
				results = append(results, integerArgument(result))
				mutex.Unlock()
			}
		})
	}
	group.Wait()

	if overlapped.Load() {
		t.Errorf("the handler ran two calls at the same time")
	}
	slices.Sort(results)
	expected := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	if !slices.Equal(results, expected) {
		t.Errorf("results = %v; want each call to see the previous ones: %v", results, expected)
	}
}

func TestServeHandlerError(t *testing.T) {
	account := accountInterface()
	handler := accountHandler()
	address := startServer(t, account, func(method *wire.Method, arguments []types.Value) (types.Value, error) {
		if method.Name == "withdraw" {
			return nil, errors.New("[server.ric, line 3, column 20] Division by zero")
		}
		return handler(method, arguments)
	})
	client := dial(t, address, account, testTimeout)

	_, failure, err := client.Call(methodOf(t, account, "withdraw"), []types.Value{integer(1)})

	if err != nil {
		t.Fatalf("Call() unexpected error: %v", err)
	}
	assertFailure(t, "Call() failure", failure, &Failure{Kind: ServerError, Message: "[server.ric, line 3, column 20] Division by zero"})
	assertBalance(t, client, 0)
}

func TestServeResultThatCannotBeEncoded(t *testing.T) {
	account := accountInterface()
	handler := accountHandler()
	address := startServer(t, account, func(method *wire.Method, arguments []types.Value) (types.Value, error) {
		if method.Name == "rename" {
			return types.NewString("\xFF"), nil
		}
		return handler(method, arguments)
	})
	client := dial(t, address, account, testTimeout)

	_, failure, err := client.Call(methodOf(t, account, "rename"), []types.Value{types.NewString("x")})

	if err != nil {
		t.Fatalf("Call() unexpected error: %v", err)
	}
	assertFailure(t, "Call() failure", failure,
		&Failure{Kind: ServerError, Message: "cannot send the result of 'rename': invalid data: String is not valid UTF-8"})
	assertBalance(t, client, 0)
}

func TestServeRejectsOtherInterfaces(t *testing.T) {
	testCases := []mismatchTestCase{
		{"other service", listInterface()},
		{"other version of the service", wire.NewInterface(types.NewServiceType("Account", []types.ServiceMethod{
			{Name: "balance", Signature: types.Signature{Return: types.Int}},
		}))},
	}
	account := accountInterface()
	address := startServer(t, account, accountHandler())

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, failure := Dial(address, testCase.serviceInterface, testTimeout)

			assertFailure(t, "Dial() failure", failure, &Failure{
				Kind: ProtocolError,
				Message: fmt.Sprintf("cannot connect to %s at %v: interface mismatch with the server, which implements service 'Account'",
					testCase.serviceInterface.Name, address),
			})
			assertBalance(t, dial(t, address, account, testTimeout), 0)
		})
	}
}

func TestServeIgnoresBrokenConnections(t *testing.T) {
	account := accountInterface()
	hello := encodeHello(account.Hash)
	testCases := []brokenConnectionTestCase{
		{"closed before the hello", nil, nil},
		{"hello cut in the middle", []byte("RIC"), nil},
		{"other program", []byte("GET / "), nil},
		{"other protocol version", []byte("RICOL\x02"), encodeRejection("unsupported protocol version 2, expected 1")},
		{"closed after the handshake", hello, encodeWelcome()},
		{"request cut in the middle", concat(hello, hexBytes("00 00 00 08 00")), encodeWelcome()},
	}
	address := startServer(t, account, accountHandler())

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertBytes(t, "response of the server", exchange(t, address, testCase.sent), testCase.expectedResponse)
			assertBalance(t, dial(t, address, account, testTimeout), 0)
		})
	}
}

func TestServeInvalidRequests(t *testing.T) {
	testCases := []invalidRequestResponseTestCase{
		{"unknown method", hexBytes("00 00 00 04 00 00 00 09"), "unknown method index 9"},
		{"invalid arguments", hexBytes("00 00 00 05 00 00 00 01 00"), "invalid arguments of 'deposit': invalid data: truncated"},
		{"data after the arguments", hexBytes("00 00 00 05 00 00 00 00 FF"), "unexpected data after the arguments of 'balance'"},
	}
	account := accountInterface()
	handler := accountHandler()
	var calls atomic.Int32
	address := startServer(t, account, func(method *wire.Method, arguments []types.Value) (types.Value, error) {
		calls.Add(1)
		return handler(method, arguments)
	})

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			received := exchange(t, address, concat(encodeHello(account.Hash), testCase.request))

			assertBytes(t, "response of the server", received, concat(encodeWelcome(), encodeFailure(statusProtocolError, testCase.expectedMessage)))
			if count := calls.Load(); count != 0 {
				t.Errorf("the handler was called %d times; want 0", count)
			}
		})
	}
}

func TestCloseWithOpenConnections(t *testing.T) {
	account := accountInterface()
	server := listenLocally(t, account)
	served := make(chan error, 1)
	go func() { served <- server.Serve(accountHandler()) }()
	address := addressOf(server.Addr())
	client := dial(t, address, account, testTimeout)
	assertBalance(t, client, 0)

	if err := server.Close(); err != nil {
		t.Fatalf("Close() unexpected error: %v", err)
	}

	if err := <-served; err != nil {
		t.Errorf("Serve() after Close() = %v; want nil", err)
	}
	if _, failure, _ := client.Call(methodOf(t, account, "balance"), nil); failure == nil || failure.Kind != ConnectionLost {
		t.Errorf("Call() after Close() failure = %v; want a ConnectionLost", failure)
	}
	if _, failure := Dial(address, account, testTimeout); failure == nil || failure.Kind != ConnectionFailed {
		t.Errorf("Dial() after Close() failure = %v; want a ConnectionFailed", failure)
	}
	if err := server.Close(); err != nil {
		t.Errorf("second Close() unexpected error: %v", err)
	}
}

func TestCloseWhileACallRuns(t *testing.T) {
	account := accountInterface()
	server := listenLocally(t, account)
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	served := make(chan error, 1)
	go func() {
		served <- server.Serve(func(*wire.Method, []types.Value) (types.Value, error) {
			if calls.Add(1) == 1 {
				close(started)
				<-release
			}
			return integer(0), nil
		})
	}()
	address := addressOf(server.Addr())
	running := dial(t, address, account, testTimeout)
	waiting := dial(t, address, account, testTimeout)
	balance := methodOf(t, account, "balance")
	failures := make(chan *Failure, 2)

	go func() {
		_, failure, _ := running.Call(balance, nil)
		failures <- failure
	}()
	<-started
	go func() {
		_, failure, _ := waiting.Call(balance, nil)
		failures <- failure
	}()
	time.Sleep(50 * time.Millisecond)
	server.Close()
	close(release)

	if err := <-served; err != nil {
		t.Errorf("Serve() after Close() = %v; want nil", err)
	}
	for range 2 {
		if failure := <-failures; failure == nil || failure.Kind != ConnectionLost {
			t.Errorf("Call() failure = %v; want a ConnectionLost", failure)
		}
	}
	if count := calls.Load(); count != 1 {
		t.Errorf("the handler ran %d calls; want only the one that was running when the server was closed", count)
	}
}

func TestServeAfterClose(t *testing.T) {
	server := listenLocally(t, accountInterface())
	server.Close()

	if err := server.Serve(accountHandler()); err != nil {
		t.Errorf("Serve() after Close() = %v; want nil", err)
	}
}

func TestServeReturnsAcceptErrors(t *testing.T) {
	server := listenLocally(t, accountInterface())
	server.listener.Close()

	if err := server.Serve(accountHandler()); !errors.Is(err, net.ErrClosed) {
		t.Errorf("Serve() = %v; want %v", err, net.ErrClosed)
	}
}

func TestListenOnAPortInUse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() unexpected error: %v", err)
	}
	defer listener.Close()

	_, err = Listen(addressOf(listener.Addr()), accountInterface())

	if !errors.Is(err, syscall.EADDRINUSE) || err.Error() != "address already in use" {
		t.Errorf("Listen() error = %v; want %v", err, syscall.EADDRINUSE)
	}
}
