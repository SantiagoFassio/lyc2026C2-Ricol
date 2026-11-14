package network

import (
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

type dialFailureTestCase struct {
	name          string
	timeout       time.Duration
	handle        func(conn net.Conn)
	expectedKind  Kind
	expectedCause string
}

type callTestCase struct {
	name       string
	methodName string
	arguments  []types.Value
	result     types.Value
}

type receivedCall struct {
	methodName string
	arguments  []types.Value
}

type callFailureTestCase struct {
	name            string
	timeout         time.Duration
	respond         func(conn net.Conn)
	expectedKind    Kind
	expectedMessage string
}

type unencodableArgumentsTestCase struct {
	name             string
	serviceInterface *wire.Interface
	methodName       string
	arguments        []types.Value
	expectedError    error
}

const testTimeout = 2 * time.Second

func fakeServer(t *testing.T, handle func(conn net.Conn)) common.Address {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() unexpected error: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		handle(conn)
	}()
	t.Cleanup(func() {
		listener.Close()
		<-done
	})
	return addressOf(listener.Addr())
}

func sessionServer(t *testing.T, handle func(conn net.Conn)) common.Address {
	t.Helper()

	return fakeServer(t, func(conn net.Conn) {
		if _, err := readHello(conn); err != nil {
			t.Errorf("readHello() unexpected error: %v", err)
			return
		}
		conn.Write(encodeWelcome())
		handle(conn)
	})
}

func addressOf(address net.Addr) common.Address {
	return common.Address{Host: "127.0.0.1", Port: address.(*net.TCPAddr).Port}
}

func closedAddress(t *testing.T) common.Address {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() unexpected error: %v", err)
	}
	address := addressOf(listener.Addr())
	listener.Close()
	return address
}

func dial(t *testing.T, address common.Address, serviceInterface *wire.Interface, timeout time.Duration) *Client {
	t.Helper()

	client, failure := Dial(address, serviceInterface, timeout)
	if failure != nil {
		t.Fatalf("Dial(%v) unexpected failure: %v", address, failure)
	}
	t.Cleanup(client.Close)
	return client
}

func waitForClose(conn net.Conn) {
	io.Copy(io.Discard, conn)
}

func ignoreRequest(conn net.Conn) {
	readFrame(conn)
}

func respondWith(response []byte) func(conn net.Conn) {
	return func(conn net.Conn) {
		ignoreRequest(conn)
		conn.Write(response)
	}
}

func listInterface() *wire.Interface {
	listType := &types.EnumType{Name: "List"}
	listType.Variants = []types.EnumVariant{
		{Name: "Empty"},
		{Name: "Node", Fields: []types.StructField{{Name: "value", Type: types.Int}, {Name: "rest", Type: listType}}},
	}
	return wire.NewInterface(types.NewServiceType("Lists", []types.ServiceMethod{
		{Name: "length", Signature: types.Signature{Params: []types.Type{listType}, Return: types.Int}},
	}))
}

func list(length int) types.Value {
	value := types.NewEnum("List", "Empty", nil)
	for i := range length {
		value = types.NewEnum("List", "Node", []types.Value{integer(int64(i)), value})
	}
	return value
}

func TestDial(t *testing.T) {
	account := accountInterface()
	received := make(chan []byte, 1)
	address := fakeServer(t, func(conn net.Conn) {
		conn.Write(encodeWelcome())
		data, _ := io.ReadAll(conn)
		received <- data
	})

	client, failure := Dial(address, account, testTimeout)
	if failure != nil {
		t.Fatalf("Dial() unexpected failure: %v", failure)
	}
	client.Close()

	assertBytes(t, "bytes sent by the client", <-received, encodeHello(account.Hash))
}

func TestDialFailures(t *testing.T) {
	afterHello := func(respond func(conn net.Conn)) func(conn net.Conn) {
		return func(conn net.Conn) {
			readHello(conn)
			respond(conn)
		}
	}
	write := func(data []byte) func(conn net.Conn) {
		return afterHello(func(conn net.Conn) { conn.Write(data) })
	}
	testCases := []dialFailureTestCase{
		{"server that closes the connection", testTimeout, afterHello(func(net.Conn) {}),
			ConnectionFailed, "the connection was closed"},
		{"header cut in the middle", testTimeout, write([]byte("RIC")),
			ConnectionFailed, "the connection was closed in the middle of a message"},
		{"welcome cut in the middle", testTimeout, write(concat(encodeHeader(), hexBytes("00 00 00 05 00"))),
			ConnectionFailed, "the connection was closed in the middle of a message"},
		{"server that does not answer", 100 * time.Millisecond, afterHello(waitForClose),
			ConnectionFailed, "timed out after 100ms"},
		{"server that is not a ricol program", testTimeout, write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n")),
			ProtocolError, "the server is not a Ricol program"},
		{"other protocol version", testTimeout, write(concat([]byte("RICOL\x02"), hexBytes("00 00 00 01 00"))),
			ProtocolError, "unsupported protocol version 2, expected 1"},
		{"rejected interface", testTimeout, write(encodeRejection("the server implements a different interface of service 'Counter'")),
			ProtocolError, "the server implements a different interface of service 'Counter'"},
		{"invalid welcome", testTimeout, write(concat(encodeHeader(), hexBytes("00 00 00 01 07"))),
			ProtocolError, "unknown response status 0x07"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			address := fakeServer(t, testCase.handle)

			client, failure := Dial(address, accountInterface(), testCase.timeout)

			if client != nil {
				client.Close()
				t.Fatalf("Dial() = a client; want a failure")
			}
			expected := &Failure{Kind: testCase.expectedKind, Message: "cannot connect to Account at " + address.String() + ": " + testCase.expectedCause}
			assertFailure(t, "Dial() failure", failure, expected)
		})
	}
}

func TestDialRefused(t *testing.T) {
	address := closedAddress(t)

	_, failure := Dial(address, accountInterface(), testTimeout)

	assertFailure(t, "Dial() failure", failure,
		&Failure{Kind: ConnectionFailed, Message: "cannot connect to Account at " + address.String() + ": connection refused"})
}

func TestDialUnknownHost(t *testing.T) {
	address := common.Address{Host: "nonexistent.invalid", Port: 3825}

	_, failure := Dial(address, accountInterface(), testTimeout)

	if failure == nil || failure.Kind != ConnectionFailed || !strings.HasPrefix(failure.Message, "cannot connect to Account at nonexistent.invalid:3825: ") {
		t.Errorf("Dial() failure = %v; want a ConnectionFailed to nonexistent.invalid:3825", failure)
	}
}

func TestDialStopsWaitingAtTheTimeout(t *testing.T) {
	address := fakeServer(t, func(conn net.Conn) {
		readHello(conn)
		waitForClose(conn)
	})
	start := time.Now()

	_, failure := Dial(address, accountInterface(), 200*time.Millisecond)

	elapsed := time.Since(start)
	if failure == nil || failure.Kind != ConnectionFailed {
		t.Errorf("Dial() failure = %v; want a ConnectionFailed", failure)
	}
	if elapsed < 200*time.Millisecond || elapsed > time.Second {
		t.Errorf("Dial() took %v; want about 200ms", elapsed)
	}
}

func TestCall(t *testing.T) {
	account := accountInterface()
	results := map[string]types.Value{
		"balance":  integer(130),
		"deposit":  movement(30, 130),
		"withdraw": withdrawResult("InsufficientFunds", integer(100)),
		"rename":   types.NewString("chau"),
		"reset":    nil,
	}
	received := make(chan receivedCall, 1)
	address := sessionServer(t, func(conn net.Conn) {
		for {
			payload, err := readFrame(conn)
			if err != nil {
				return
			}
			method, arguments, err := decodeRequest(payload, account)
			if err != nil {
				t.Errorf("decodeRequest() unexpected error: %v", err)
				return
			}
			received <- receivedCall{methodName: method.Name, arguments: arguments}
			frame, _ := encodeResult(method, results[method.Name])
			conn.Write(frame)
		}
	})
	client := dial(t, address, account, testTimeout)
	testCases := []callTestCase{
		{"method without parameters", "balance", nil, results["balance"]},
		{"method returning a struct", "deposit", []types.Value{integer(30)}, results["deposit"]},
		{"method returning an enum", "withdraw", []types.Value{integer(500)}, results["withdraw"]},
		{"method with a string", "rename", []types.Value{types.NewString("hola")}, results["rename"]},
		{"method without result", "reset", nil, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, failure, err := client.Call(methodOf(t, account, testCase.methodName), testCase.arguments)

			if failure != nil || err != nil {
				t.Fatalf("Call() unexpected failure %v, error %v", failure, err)
			}
			if !sameValue(result, testCase.result) {
				t.Errorf("Call() = %v; want %v", result, testCase.result)
			}
			call := <-received
			if call.methodName != testCase.methodName || len(call.arguments) != len(testCase.arguments) {
				t.Fatalf("server received %s%v; want %s%v", call.methodName, call.arguments, testCase.methodName, testCase.arguments)
			}
			for position, argument := range call.arguments {
				if !sameValue(argument, testCase.arguments[position]) {
					t.Errorf("server received argument %d = %v; want %v", position, argument, testCase.arguments[position])
				}
			}
		})
	}
}

func TestCallFailures(t *testing.T) {
	testCases := []callFailureTestCase{
		{"server error", testTimeout, respondWith(encodeFailure(statusServerError, "division by zero")),
			ServerError, "division by zero"},
		{"protocol error", testTimeout, respondWith(encodeFailure(statusProtocolError, "unknown method index 9")),
			ProtocolError, "unknown method index 9"},
		{"invalid result", testTimeout, respondWith(hexBytes("00 00 00 03 00 00 00")),
			ProtocolError, "invalid result of 'balance': invalid data: truncated"},
		{"server that closes the connection", testTimeout, ignoreRequest,
			ConnectionLost, "connection to Account at {address} lost during 'balance': the connection was closed"},
		{"response cut in the middle", testTimeout, respondWith(hexBytes("00 00 00 09 00 00")),
			ConnectionLost, "connection to Account at {address} lost during 'balance': the connection was closed in the middle of a message"},
		{"server that does not answer", 100 * time.Millisecond, waitForClose,
			Timeout, "'balance' did not answer within 100ms"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			account := accountInterface()
			address := sessionServer(t, testCase.respond)
			client := dial(t, address, account, testCase.timeout)

			result, failure, err := client.Call(methodOf(t, account, "balance"), nil)

			if err != nil {
				t.Fatalf("Call() unexpected error: %v", err)
			}
			if result != nil {
				t.Errorf("Call() = %v; want nil", result)
			}
			expectedMessage := strings.ReplaceAll(testCase.expectedMessage, "{address}", address.String())
			assertFailure(t, "Call() failure", failure, &Failure{Kind: testCase.expectedKind, Message: expectedMessage})
		})
	}
}

func TestCallToAServerThatDoesNotRead(t *testing.T) {
	account := accountInterface()
	stop := make(chan struct{})
	defer close(stop)
	address := sessionServer(t, func(conn net.Conn) {
		<-stop
	})
	client := dial(t, address, account, 200*time.Millisecond)

	_, failure, err := client.Call(methodOf(t, account, "rename"), []types.Value{types.NewString(strings.Repeat("a", 32<<20))})

	if err != nil {
		t.Fatalf("Call() unexpected error: %v", err)
	}
	assertFailure(t, "Call() failure", failure, &Failure{Kind: Timeout, Message: "'rename' did not answer within 200ms"})
}

func TestCallWithArgumentsThatCannotBeEncoded(t *testing.T) {
	testCases := []unencodableArgumentsTestCase{
		{"string that is not utf-8", accountInterface(), "rename", []types.Value{types.NewString("\xFF")}, wire.ErrInvalidData},
		{"value too deep", listInterface(), "length", []types.Value{list(100000)}, wire.ErrTooDeep},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			received := make(chan int64, 1)
			address := sessionServer(t, func(conn net.Conn) {
				count, _ := io.Copy(io.Discard, conn)
				received <- count
			})
			client := dial(t, address, testCase.serviceInterface, testTimeout)

			result, failure, err := client.Call(methodOf(t, testCase.serviceInterface, testCase.methodName), testCase.arguments)

			client.Close()
			if !errors.Is(err, testCase.expectedError) {
				t.Errorf("Call() error = %v; want %v", err, testCase.expectedError)
			}
			if result != nil || failure != nil {
				t.Errorf("Call() = %v, %v; want no result and no failure", result, failure)
			}
			if count := <-received; count != 0 {
				t.Errorf("the client sent %d bytes after the hello; want 0", count)
			}
		})
	}
}

func TestEachCallHasItsOwnTimeout(t *testing.T) {
	account := accountInterface()
	address := sessionServer(t, func(conn net.Conn) {
		for range 2 {
			if _, err := readFrame(conn); err != nil {
				return
			}
			time.Sleep(200 * time.Millisecond)
			frame, _ := encodeResult(methodOf(t, account, "balance"), integer(1))
			conn.Write(frame)
		}
	})
	client := dial(t, address, account, 300*time.Millisecond)

	for i := range 2 {
		if _, failure, err := client.Call(methodOf(t, account, "balance"), nil); failure != nil || err != nil {
			t.Errorf("call %d unexpected failure %v, error %v", i+1, failure, err)
		}
	}
}
