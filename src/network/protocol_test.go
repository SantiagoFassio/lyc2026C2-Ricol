package network

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"testing/iotest"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

type readHelloErrorTestCase struct {
	name          string
	received      []byte
	expectedError error
}

type welcomeTestCase struct {
	name            string
	payload         []byte
	expectedFailure *Failure
}

type requestTestCase struct {
	name       string
	methodName string
	arguments  []types.Value
	frame      []byte
}

type invalidRequestTestCase struct {
	name            string
	payload         []byte
	expectedMessage string
}

type resultTestCase struct {
	name       string
	methodName string
	result     types.Value
	frame      []byte
}

type invalidResultTestCase struct {
	name            string
	methodName      string
	payload         []byte
	expectedFailure *Failure
}

type failureFrameTestCase struct {
	name    string
	status  status
	message string
	frame   []byte
}

type readFrameTestCase struct {
	name    string
	payload []byte
}

type readFrameErrorTestCase struct {
	name          string
	received      io.Reader
	expectedError error
}

type rootCauseTestCase struct {
	name     string
	err      error
	expected error
}

var errConnectionReset = errors.New("connection reset")

var (
	movementType = types.NewStructType("Movement", []types.StructField{
		{Name: "amount", Type: types.Int},
		{Name: "finalBalance", Type: types.Int},
	})
	withdrawResultType = &types.EnumType{Name: "WithdrawResult", Variants: []types.EnumVariant{
		{Name: "Ok", Fields: []types.StructField{{Name: "movement", Type: movementType}}},
		{Name: "InsufficientFunds", Fields: []types.StructField{{Name: "available", Type: types.Int}}},
	}}
)

func accountService() *types.ServiceType {
	return types.NewServiceType("Account", []types.ServiceMethod{
		{Name: "balance", Signature: types.Signature{Return: types.Int}},
		{Name: "deposit", Signature: types.Signature{Params: []types.Type{types.Int}, Return: movementType}},
		{Name: "withdraw", Signature: types.Signature{Params: []types.Type{types.Int}, Return: withdrawResultType}},
		{Name: "rename", Signature: types.Signature{Params: []types.Type{types.Str}, Return: types.Str}},
		{Name: "reset", Signature: types.Signature{Return: types.Void}},
	})
}

func accountInterface() *wire.Interface {
	return wire.NewInterface(accountService())
}

func integer(value int64) types.Value {
	return types.NewInteger(value)
}

func movement(amount int64, finalBalance int64) types.Value {
	return types.NewStruct("Movement", []types.FieldValue{
		{Name: "amount", Value: integer(amount)},
		{Name: "finalBalance", Value: integer(finalBalance)},
	})
}

func withdrawResult(variant string, fields ...types.Value) types.Value {
	return types.NewEnum("WithdrawResult", variant, fields)
}

func hexBytes(text string) []byte {
	data, err := hex.DecodeString(strings.ReplaceAll(text, " ", ""))
	if err != nil {
		panic(err)
	}
	return data
}

func concat(parts ...[]byte) []byte {
	return bytes.Join(parts, nil)
}

func sameValue(left types.Value, right types.Value) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equals(right)
}

func methodOf(t *testing.T, serviceInterface *wire.Interface, name string) *wire.Method {
	t.Helper()

	method, ok := serviceInterface.Method(name)
	if !ok {
		t.Fatalf("Method(%q) not found", name)
	}
	return method
}

func assertBytes(t *testing.T, description string, result []byte, expected []byte) {
	t.Helper()

	if !bytes.Equal(result, expected) {
		t.Errorf("%s = % X; want % X", description, result, expected)
	}
}

func assertFailure(t *testing.T, description string, failure *Failure, expected *Failure) {
	t.Helper()

	if failure == nil || expected == nil {
		if failure != expected {
			t.Errorf("%s = %v; want %v", description, failure, expected)
		}
		return
	}
	if *failure != *expected {
		t.Errorf("%s = %q; want %q", description, failure, expected)
	}
}

func readOneFrame(t *testing.T, frame []byte) []byte {
	t.Helper()

	payload, err := readFrame(bytes.NewReader(frame))
	if err != nil {
		t.Fatalf("readFrame(% X) unexpected error: %v", frame, err)
	}
	return payload
}

func TestKindsAreSessionErrorVariants(t *testing.T) {
	kinds := []Kind{ConnectionFailed, ConnectionLost, Timeout, ProtocolError, ServerError}

	if len(kinds) != len(types.SessionError.Variants) {
		t.Errorf("there are %d kinds; want one for each of the %d variants of SessionError", len(kinds), len(types.SessionError.Variants))
	}
	for _, kind := range kinds {
		if _, ok := types.SessionError.Variant(string(kind)); !ok {
			t.Errorf("kind %q is not a variant of SessionError", kind)
		}
	}
}

func TestFailureError(t *testing.T) {
	if message := (&Failure{Kind: ConnectionLost, Message: "reset by peer"}).Error(); message != "ConnectionLost: reset by peer" {
		t.Errorf("Error() = %q; want %q", message, "ConnectionLost: reset by peer")
	}
	if message := (&Failure{Kind: Timeout}).Error(); message != "Timeout" {
		t.Errorf("Error() = %q; want %q", message, "Timeout")
	}
}

func TestHello(t *testing.T) {
	hash := sha256.Sum256([]byte("Account"))

	encoded := encodeHello(hash)

	assertBytes(t, "encodeHello()", encoded, concat(hexBytes("52 49 43 4F 4C 01"), hash[:]))
	reader := bytes.NewReader(append(encoded, 0xFF))
	received, err := readHello(reader)
	if err != nil {
		t.Fatalf("readHello() unexpected error: %v", err)
	}
	if received != hash {
		t.Errorf("readHello() = %x; want %x", received, hash)
	}
	if reader.Len() != 1 {
		t.Errorf("readHello() left %d bytes unread; want 1", reader.Len())
	}
}

func TestReadHelloErrors(t *testing.T) {
	hash := make([]byte, 32)
	testCases := []readHelloErrorTestCase{
		{"without data", nil, io.EOF},
		{"truncated header", hexBytes("52 49 43"), io.ErrUnexpectedEOF},
		{"other magic", concat([]byte("GET /\x01"), hash), errNotRicol},
		{"magic with other capitalization", concat([]byte("ricol\x01"), hash), errNotRicol},
		{"other version", concat(hexBytes("52 49 43 4F 4C 02"), hash), errUnsupportedVersion},
		{"version zero", concat(hexBytes("52 49 43 4F 4C 00"), hash), errUnsupportedVersion},
		{"without hash", hexBytes("52 49 43 4F 4C 01"), io.EOF},
		{"truncated hash", concat(hexBytes("52 49 43 4F 4C 01"), hash[:10]), io.ErrUnexpectedEOF},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := readHello(bytes.NewReader(testCase.received))

			if !errors.Is(err, testCase.expectedError) {
				t.Errorf("readHello() error = %v; want %v", err, testCase.expectedError)
			}
		})
	}
}

func TestReadHelloReturnsReaderErrors(t *testing.T) {
	_, err := readHello(io.MultiReader(bytes.NewReader(hexBytes("52 49 43 4F 4C 01 00")), iotest.ErrReader(errConnectionReset)))

	if !errors.Is(err, errConnectionReset) {
		t.Errorf("readHello() error = %v; want %v", err, errConnectionReset)
	}
}

func TestUnsupportedVersionMessage(t *testing.T) {
	err := readHeader(bytes.NewReader(hexBytes("52 49 43 4F 4C 07")))

	if expected := "unsupported protocol version 7, expected 1"; err == nil || err.Error() != expected {
		t.Errorf("readHeader() error = %v; want %q", err, expected)
	}
}

func TestWelcome(t *testing.T) {
	encoded := encodeWelcome()

	assertBytes(t, "encodeWelcome()", encoded, hexBytes("52 49 43 4F 4C 01 00 00 00 01 00"))
	reader := bytes.NewReader(encoded)
	if err := readHeader(reader); err != nil {
		t.Fatalf("readHeader() unexpected error: %v", err)
	}
	payload, err := readFrame(reader)
	if err != nil {
		t.Fatalf("readFrame() unexpected error: %v", err)
	}
	assertFailure(t, "decodeWelcome()", decodeWelcome(payload), nil)
}

func TestRejection(t *testing.T) {
	encoded := encodeRejection("wrong interface")

	assertBytes(t, "encodeRejection()", encoded, concat(hexBytes("52 49 43 4F 4C 01 00 00 00 10 02"), []byte("wrong interface")))
	reader := bytes.NewReader(encoded)
	if err := readHeader(reader); err != nil {
		t.Fatalf("readHeader() unexpected error: %v", err)
	}
	payload, err := readFrame(reader)
	if err != nil {
		t.Fatalf("readFrame() unexpected error: %v", err)
	}
	assertFailure(t, "decodeWelcome()", decodeWelcome(payload), &Failure{Kind: ProtocolError, Message: "wrong interface"})
}

func TestDecodeInvalidWelcome(t *testing.T) {
	testCases := []welcomeTestCase{
		{"empty payload", nil, &Failure{Kind: ProtocolError, Message: "empty response"}},
		{"ok with data", hexBytes("00 FF"), &Failure{Kind: ProtocolError, Message: "unexpected data in the welcome of the server"}},
		{"unknown status", hexBytes("07"), &Failure{Kind: ProtocolError, Message: "unknown response status 0x07"}},
		{"server error", concat(hexBytes("01"), []byte("busy")), &Failure{Kind: ServerError, Message: "busy"}},
		{"message that is not utf-8", hexBytes("02 FF"), &Failure{Kind: ProtocolError, Message: "the error message of the server is not valid UTF-8"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertFailure(t, "decodeWelcome()", decodeWelcome(testCase.payload), testCase.expectedFailure)
		})
	}
}

func TestRequests(t *testing.T) {
	testCases := []requestTestCase{
		{"method without parameters", "balance", nil, hexBytes("00 00 00 04 00 00 00 00")},
		{"method with an integer", "deposit", []types.Value{integer(30)},
			hexBytes("00 00 00 0C 00 00 00 01 00 00 00 00 00 00 00 1E")},
		{"method in the middle", "withdraw", []types.Value{integer(500)},
			hexBytes("00 00 00 0C 00 00 00 02 00 00 00 00 00 00 01 F4")},
		{"method with a string", "rename", []types.Value{types.NewString("hola")},
			hexBytes("00 00 00 0C 00 00 00 03 00 00 00 04 68 6F 6C 61")},
		{"last method", "reset", nil, hexBytes("00 00 00 04 00 00 00 04")},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			method := methodOf(t, account, testCase.methodName)

			frame, err := encodeRequest(method, testCase.arguments)
			if err != nil {
				t.Fatalf("encodeRequest() unexpected error: %v", err)
			}
			assertBytes(t, "encodeRequest()", frame, testCase.frame)

			decodedMethod, arguments, err := decodeRequest(readOneFrame(t, frame), account)
			if err != nil {
				t.Fatalf("decodeRequest() unexpected error: %v", err)
			}
			if decodedMethod != method {
				t.Errorf("decodeRequest() method = %q; want %q", decodedMethod.Name, method.Name)
			}
			if len(arguments) != len(testCase.arguments) {
				t.Fatalf("decodeRequest() arguments = %v; want %v", arguments, testCase.arguments)
			}
			for position, argument := range arguments {
				if !sameValue(argument, testCase.arguments[position]) {
					t.Errorf("decodeRequest() argument %d = %v; want %v", position, argument, testCase.arguments[position])
				}
			}
		})
	}
}

func TestEncodeRequestWithInvalidArguments(t *testing.T) {
	_, err := encodeRequest(methodOf(t, accountInterface(), "rename"), []types.Value{types.NewString("\xFF")})

	if !errors.Is(err, wire.ErrInvalidData) {
		t.Errorf("encodeRequest() error = %v; want %v", err, wire.ErrInvalidData)
	}
}

func TestDecodeInvalidRequest(t *testing.T) {
	testCases := []invalidRequestTestCase{
		{"empty payload", nil, "request without a method index"},
		{"truncated method index", hexBytes("00 00 00"), "request without a method index"},
		{"method index out of range", hexBytes("00 00 00 05"), "unknown method index 5"},
		{"largest method index", hexBytes("FF FF FF FF"), "unknown method index 4294967295"},
		{"missing argument", hexBytes("00 00 00 01"), "invalid arguments of 'deposit': invalid data: truncated"},
		{"truncated argument", hexBytes("00 00 00 01 00 00 00"), "invalid arguments of 'deposit': invalid data: truncated"},
		{"string that is not utf-8", hexBytes("00 00 00 03 00 00 00 01 FF"),
			"invalid arguments of 'rename': invalid data: String is not valid UTF-8"},
		{"data after a method without parameters", hexBytes("00 00 00 00 FF"), "unexpected data after the arguments of 'balance'"},
		{"data after the arguments", hexBytes("00 00 00 01 00 00 00 00 00 00 00 1E 00"),
			"unexpected data after the arguments of 'deposit'"},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := decodeRequest(testCase.payload, account)

			if err == nil || err.Error() != testCase.expectedMessage {
				t.Errorf("decodeRequest() error = %v; want %q", err, testCase.expectedMessage)
			}
		})
	}
}

func TestDecodeRequestKeepsTheWireError(t *testing.T) {
	_, _, err := decodeRequest(hexBytes("00 00 00 01 00"), accountInterface())

	if !errors.Is(err, wire.ErrInvalidData) {
		t.Errorf("decodeRequest() error = %v; want %v", err, wire.ErrInvalidData)
	}
}

func TestResults(t *testing.T) {
	testCases := []resultTestCase{
		{"integer", "balance", integer(130), hexBytes("00 00 00 09 00 00 00 00 00 00 00 00 82")},
		{"struct", "deposit", movement(30, 130),
			hexBytes("00 00 00 11 00 00 00 00 00 00 00 00 1E 00 00 00 00 00 00 00 82")},
		{"enum", "withdraw", withdrawResult("InsufficientFunds", integer(100)),
			hexBytes("00 00 00 0D 00 00 00 00 01 00 00 00 00 00 00 00 64")},
		{"string", "rename", types.NewString("chau"), hexBytes("00 00 00 09 00 00 00 00 04 63 68 61 75")},
		{"void", "reset", nil, hexBytes("00 00 00 01 00")},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			method := methodOf(t, account, testCase.methodName)

			frame, err := encodeResult(method, testCase.result)
			if err != nil {
				t.Fatalf("encodeResult() unexpected error: %v", err)
			}
			assertBytes(t, "encodeResult()", frame, testCase.frame)

			result, failure := decodeResult(readOneFrame(t, frame), method)
			assertFailure(t, "decodeResult() failure", failure, nil)
			if !sameValue(result, testCase.result) {
				t.Errorf("decodeResult() = %v; want %v", result, testCase.result)
			}
		})
	}
}

func TestEncodeResultWithInvalidValue(t *testing.T) {
	_, err := encodeResult(methodOf(t, accountInterface(), "rename"), types.NewString("\xFF"))

	if !errors.Is(err, wire.ErrInvalidData) {
		t.Errorf("encodeResult() error = %v; want %v", err, wire.ErrInvalidData)
	}
}

func TestDecodeInvalidResult(t *testing.T) {
	protocolError := func(message string) *Failure { return &Failure{Kind: ProtocolError, Message: message} }
	testCases := []invalidResultTestCase{
		{"empty payload", "balance", nil, protocolError("empty response")},
		{"unknown status", "balance", hexBytes("03"), protocolError("unknown response status 0x03")},
		{"missing result", "balance", hexBytes("00"), protocolError("invalid result of 'balance': invalid data: truncated")},
		{"truncated result", "balance", hexBytes("00 00 00 00"), protocolError("invalid result of 'balance': invalid data: truncated")},
		{"invalid variant", "withdraw", hexBytes("00 00 00 00 07"),
			protocolError("invalid result of 'withdraw': invalid data: variant index 7 out of range for WithdrawResult")},
		{"data after a void result", "reset", hexBytes("00 FF"), protocolError("unexpected data after the result of 'reset'")},
		{"data after the result", "balance", hexBytes("00 00 00 00 00 00 00 00 82 00"),
			protocolError("unexpected data after the result of 'balance'")},
		{"server error", "balance", concat(hexBytes("01"), []byte("division by zero")),
			&Failure{Kind: ServerError, Message: "division by zero"}},
		{"server error without message", "balance", hexBytes("01"), &Failure{Kind: ServerError}},
		{"protocol error", "balance", concat(hexBytes("02"), []byte("unknown method index 9")), protocolError("unknown method index 9")},
		{"error message that is not utf-8", "balance", hexBytes("01 61 FF"),
			protocolError("the error message of the server is not valid UTF-8")},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, failure := decodeResult(testCase.payload, methodOf(t, account, testCase.methodName))

			assertFailure(t, "decodeResult() failure", failure, testCase.expectedFailure)
			if result != nil {
				t.Errorf("decodeResult() = %v; want nil", result)
			}
		})
	}
}

func TestEncodeFailure(t *testing.T) {
	testCases := []failureFrameTestCase{
		{"server error", statusServerError, "division by zero", concat(hexBytes("00 00 00 11 01"), []byte("division by zero"))},
		{"protocol error", statusProtocolError, "bad", concat(hexBytes("00 00 00 04 02"), []byte("bad"))},
		{"empty message", statusServerError, "", hexBytes("00 00 00 01 01")},
		{"message that is not utf-8", statusServerError, "a\xFFb", hexBytes("00 00 00 06 01 61 EF BF BD 62")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertBytes(t, "encodeFailure()", encodeFailure(testCase.status, testCase.message), testCase.frame)
		})
	}
}

func TestEncodeFailureTruncatesLongMessages(t *testing.T) {
	message := strings.Repeat("a", maxMessageLength) + "end"

	payload := readOneFrame(t, encodeFailure(statusServerError, message))

	assertBytes(t, "encodeFailure() payload", payload, concat(hexBytes("01"), []byte(message[:maxMessageLength])))
}

func TestEncodeFailureTruncatesInsideACharacter(t *testing.T) {
	message := strings.Repeat("a", maxMessageLength-1) + "ñ"

	payload := readOneFrame(t, encodeFailure(statusServerError, message))

	_, failure := decodeResult(payload, methodOf(t, accountInterface(), "balance"))
	if failure == nil || failure.Kind != ServerError || failure.Message != strings.Repeat("a", maxMessageLength-1)+"�" {
		t.Errorf("decodeResult() failure = %v; want a ServerError ending in a replacement character", failure)
	}
}

func TestReadFrame(t *testing.T) {
	testCases := []readFrameTestCase{
		{"empty payload", nil},
		{"small payload", []byte("hola")},
		{"payload of one chunk", bytes.Repeat([]byte{0xAB}, readChunkSize)},
		{"payload of several chunks", bytes.Repeat([]byte{0xCD}, 3*readChunkSize+5)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			frame := concat(binary.BigEndian.AppendUint32(nil, uint32(len(testCase.payload))), testCase.payload, hexBytes("FF"))
			reader := bytes.NewReader(frame)

			payload, err := readFrame(iotest.HalfReader(reader))

			if err != nil {
				t.Fatalf("readFrame() unexpected error: %v", err)
			}
			if !bytes.Equal(payload, testCase.payload) {
				t.Errorf("readFrame() = %d bytes; want the %d bytes of the payload", len(payload), len(testCase.payload))
			}
			if reader.Len() != 1 {
				t.Errorf("readFrame() left %d bytes unread; want 1", reader.Len())
			}
		})
	}
}

func TestReadFrameErrors(t *testing.T) {
	withError := func(received string, err error) io.Reader {
		return io.MultiReader(bytes.NewReader(hexBytes(received)), iotest.ErrReader(err))
	}
	testCases := []readFrameErrorTestCase{
		{"without data", bytes.NewReader(nil), io.EOF},
		{"truncated header", bytes.NewReader(hexBytes("00 00")), errTruncatedFrame},
		{"without payload", bytes.NewReader(hexBytes("00 00 00 05")), errTruncatedFrame},
		{"truncated payload", bytes.NewReader(hexBytes("00 00 00 05 61 62")), errTruncatedFrame},
		{"huge length", bytes.NewReader(hexBytes("FF FF FF FF 61")), errTruncatedFrame},
		{"error before the header", withError("", errConnectionReset), errConnectionReset},
		{"error in the header", withError("00", errConnectionReset), errConnectionReset},
		{"error in the payload", withError("00 00 00 05 61", errConnectionReset), errConnectionReset},
		{"deadline in the payload", withError("00 00 00 05 61", os.ErrDeadlineExceeded), os.ErrDeadlineExceeded},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := readFrame(testCase.received)

			if !errors.Is(err, testCase.expectedError) {
				t.Errorf("readFrame() error = %v; want %v", err, testCase.expectedError)
			}
		})
	}
}

func TestReadFrameDoesNotReserveTheReceivedLength(t *testing.T) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	_, err := readFrame(bytes.NewReader(hexBytes("FF FF FF FF 61 62 63")))

	runtime.ReadMemStats(&after)
	if !errors.Is(err, errTruncatedFrame) {
		t.Errorf("readFrame() error = %v; want %v", err, errTruncatedFrame)
	}
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 1<<20 {
		t.Errorf("readFrame() allocated %d bytes; want at most %d", allocated, 1<<20)
	}
}

func TestFrameLengthLimit(t *testing.T) {
	previous := maxFrameLength
	t.Cleanup(func() { maxFrameLength = previous })
	deposit := methodOf(t, accountInterface(), "deposit")

	maxFrameLength = 12
	if _, err := encodeRequest(deposit, []types.Value{integer(30)}); err != nil {
		t.Errorf("encodeRequest() with a payload of the maximum length unexpected error: %v", err)
	}

	maxFrameLength = 11
	_, err := encodeRequest(deposit, []types.Value{integer(30)})
	if !errors.Is(err, errMessageTooLong) {
		t.Errorf("encodeRequest() error = %v; want %v", err, errMessageTooLong)
	}
	if expected := "message too long: 12 bytes, at most 11"; err == nil || err.Error() != expected {
		t.Errorf("encodeRequest() error = %v; want %q", err, expected)
	}
}

func TestRootCause(t *testing.T) {
	refused := &net.OpError{Op: "dial", Net: "tcp", Err: &os.SyscallError{Syscall: "connect", Err: syscall.ECONNREFUSED}}
	plain := errors.New("plain")
	testCases := []rootCauseTestCase{
		{"error without cause", plain, plain},
		{"wrapped errors", fmt.Errorf("a: %w", fmt.Errorf("b: %w", io.EOF)), io.EOF},
		{"refused connection", refused, syscall.ECONNREFUSED},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if cause := rootCause(testCase.err); cause != testCase.expected {
				t.Errorf("rootCause(%v) = %v; want %v", testCase.err, cause, testCase.expected)
			}
		})
	}
}

func TestRootCauseOfARefusedConnection(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() unexpected error: %v", err)
	}
	address := listener.Addr().String()
	listener.Close()

	_, err = net.Dial("tcp", address)

	if err == nil {
		t.Fatalf("net.Dial(%q) = nil error; want a refused connection", address)
	}
	if cause := rootCause(err).Error(); cause != "connection refused" {
		t.Errorf("rootCause(%v) = %q; want %q", err, cause, "connection refused")
	}
}
