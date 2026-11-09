package wire_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

type methodLookupTestCase struct {
	name          string
	methodName    string
	expectedFound bool
	expectedIndex int
}

type methodAtTestCase struct {
	name          string
	index         int
	expectedFound bool
	expectedName  string
}

type callTestCase struct {
	name             string
	methodName       string
	arguments        []types.Value
	encodedArguments []byte
	result           types.Value
	encodedResult    []byte
}

type invalidMessageTestCase struct {
	name       string
	methodName string
	encoded    []byte
}

type wrongArgumentsTestCase struct {
	name       string
	methodName string
	arguments  []types.Value
}

type failingWriter struct{}

var errConnectionReset = errors.New("connection reset")

func (failingWriter) Write([]byte) (int, error) {
	return 0, errConnectionReset
}

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

func accountService(name string, depositName string) *types.ServiceType {
	return types.NewServiceType(name, []types.ServiceMethod{
		{Name: "balance", Signature: types.Signature{Return: types.Int}},
		{Name: depositName, Signature: types.Signature{Params: []types.Type{types.Int}, Return: movementType}},
		{Name: "withdraw", Signature: types.Signature{Params: []types.Type{types.Int}, Return: withdrawResultType}},
		{Name: "reset", Signature: types.Signature{Return: types.Void}},
	})
}

func accountInterface() *wire.Interface {
	return wire.NewInterface(accountService("Account", "deposit"))
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

func readerWithExtraByte(data []byte) *bytes.Reader {
	return bytes.NewReader(append(bytes.Clone(data), 0xFF))
}

func assertBytes(t *testing.T, description string, result []byte, expected []byte) {
	t.Helper()

	if !bytes.Equal(result, expected) {
		t.Errorf("%s = % X; want % X", description, result, expected)
	}
}

func assertOnlyExtraByteLeft(t *testing.T, description string, reader *bytes.Reader) {
	t.Helper()

	if reader.Len() != 1 {
		t.Errorf("%s left %d bytes unread; want 1", description, reader.Len())
	}
}

func assertInvalidData(t *testing.T, description string, err error) {
	t.Helper()

	if !errors.Is(err, wire.ErrInvalidData) {
		t.Errorf("%s error = %v; want %v", description, err, wire.ErrInvalidData)
	}
}

func assertPanics(t *testing.T, description string, function func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Errorf("%s did not panic", description)
		}
	}()

	function()
}

func TestNewInterface(t *testing.T) {
	account := accountInterface()
	expectedCanonical := strings.Join([]string{
		"ricol-interface-v1",
		"service Account",
		"method balance() -> Int",
		"method deposit(Int) -> Movement",
		"method withdraw(Int) -> WithdrawResult",
		"method reset() -> Void",
		"struct Movement { amount: Int; finalBalance: Int; }",
		"enum WithdrawResult { Ok(movement: Movement); InsufficientFunds(available: Int); }",
	}, "\n")

	if account.Name != "Account" {
		t.Errorf("Name = %q; want %q", account.Name, "Account")
	}
	if account.Canonical != expectedCanonical {
		t.Errorf("Canonical =\n%s\nwant\n%s", account.Canonical, expectedCanonical)
	}
	if expectedHash := sha256.Sum256([]byte(expectedCanonical)); account.Hash != expectedHash {
		t.Errorf("Hash = %x; want %x", account.Hash, expectedHash)
	}
}

func TestMethodByName(t *testing.T) {
	testCases := []methodLookupTestCase{
		{"first method", "balance", true, 0},
		{"method with parameters", "deposit", true, 1},
		{"method returning an enum", "withdraw", true, 2},
		{"last method", "reset", true, 3},
		{"unknown method", "transfer", false, 0},
		{"method with other capitalization", "Deposit", false, 0},
		{"empty name", "", false, 0},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			method, found := account.Method(testCase.methodName)

			if found != testCase.expectedFound {
				t.Fatalf("Method(%q) found = %v; want %v", testCase.methodName, found, testCase.expectedFound)
			}
			if !found {
				return
			}
			if method.Name != testCase.methodName || method.Index != testCase.expectedIndex {
				t.Errorf("Method(%q) = {Name: %q, Index: %d}; want {Name: %q, Index: %d}",
					testCase.methodName, method.Name, method.Index, testCase.methodName, testCase.expectedIndex)
			}
			if byIndex, _ := account.MethodAt(method.Index); byIndex != method {
				t.Errorf("MethodAt(%d) is not the method found by name", method.Index)
			}
		})
	}
}

func TestMethodAt(t *testing.T) {
	testCases := []methodAtTestCase{
		{"first index", 0, true, "balance"},
		{"middle index", 2, true, "withdraw"},
		{"last index", 3, true, "reset"},
		{"negative index", -1, false, ""},
		{"index equal to the number of methods", 4, false, ""},
		{"index far out of range", 1000, false, ""},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			method, found := account.MethodAt(testCase.index)

			if found != testCase.expectedFound {
				t.Fatalf("MethodAt(%d) found = %v; want %v", testCase.index, found, testCase.expectedFound)
			}
			if found && (method.Name != testCase.expectedName || method.Index != testCase.index) {
				t.Errorf("MethodAt(%d) = {Name: %q, Index: %d}; want {Name: %q, Index: %d}",
					testCase.index, method.Name, method.Index, testCase.expectedName, testCase.index)
			}
		})
	}
}

func TestCalls(t *testing.T) {
	testCases := []callTestCase{
		{"method without parameters", "balance",
			nil, []byte{},
			integer(130), hexBytes("00 00 00 00 00 00 00 82")},
		{"method returning a struct", "deposit",
			[]types.Value{integer(30)}, hexBytes("00 00 00 00 00 00 00 1E"),
			movement(30, 130), hexBytes("00 00 00 00 00 00 00 1E 00 00 00 00 00 00 00 82")},
		{"method returning a variant with a struct", "withdraw",
			[]types.Value{integer(30)}, hexBytes("00 00 00 00 00 00 00 1E"),
			withdrawResult("Ok", movement(-30, 100)), hexBytes("00 00 00 00 FF FF FF FF FF FF FF E2 00 00 00 00 00 00 00 64")},
		{"method returning a variant with an integer", "withdraw",
			[]types.Value{integer(500)}, hexBytes("00 00 00 00 00 00 01 F4"),
			withdrawResult("InsufficientFunds", integer(100)), hexBytes("00 00 00 01 00 00 00 00 00 00 00 64")},
		{"method without parameters or result", "reset",
			nil, []byte{},
			nil, []byte{}},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			method := methodOf(t, account, testCase.methodName)

			var arguments bytes.Buffer
			if err := method.EncodeArguments(&arguments, testCase.arguments); err != nil {
				t.Fatalf("EncodeArguments() unexpected error: %v", err)
			}
			assertBytes(t, "EncodeArguments()", arguments.Bytes(), testCase.encodedArguments)

			argumentsReader := readerWithExtraByte(testCase.encodedArguments)
			decodedArguments, err := method.DecodeArguments(argumentsReader)
			if err != nil {
				t.Fatalf("DecodeArguments() unexpected error: %v", err)
			}
			assertOnlyExtraByteLeft(t, "DecodeArguments()", argumentsReader)
			if len(decodedArguments) != len(testCase.arguments) {
				t.Fatalf("DecodeArguments() = %v; want %v", decodedArguments, testCase.arguments)
			}
			for position, argument := range decodedArguments {
				if !sameValue(argument, testCase.arguments[position]) {
					t.Errorf("DecodeArguments()[%d] = %v; want %v", position, argument, testCase.arguments[position])
				}
			}

			var result bytes.Buffer
			if err := method.EncodeResult(&result, testCase.result); err != nil {
				t.Fatalf("EncodeResult() unexpected error: %v", err)
			}
			assertBytes(t, "EncodeResult()", result.Bytes(), testCase.encodedResult)

			resultReader := readerWithExtraByte(testCase.encodedResult)
			decodedResult, err := method.DecodeResult(resultReader)
			if err != nil {
				t.Fatalf("DecodeResult() unexpected error: %v", err)
			}
			assertOnlyExtraByteLeft(t, "DecodeResult()", resultReader)
			if !sameValue(decodedResult, testCase.result) {
				t.Errorf("DecodeResult() = %v; want %v", decodedResult, testCase.result)
			}
		})
	}
}

func TestDecodeInvalidArguments(t *testing.T) {
	testCases := []invalidMessageTestCase{
		{"missing argument", "deposit", hexBytes("")},
		{"truncated argument", "deposit", hexBytes("00 00 00 00 00 00 00")},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := methodOf(t, account, testCase.methodName).DecodeArguments(bytes.NewReader(testCase.encoded))

			assertInvalidData(t, "DecodeArguments()", err)
		})
	}
}

func TestDecodeInvalidResult(t *testing.T) {
	testCases := []invalidMessageTestCase{
		{"missing result", "balance", hexBytes("")},
		{"struct truncated between fields", "deposit", hexBytes("00 00 00 00 00 00 00 1E")},
		{"variant index out of range", "withdraw", hexBytes("00 00 00 02")},
		{"variant without its fields", "withdraw", hexBytes("00 00 00 01")},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := methodOf(t, account, testCase.methodName).DecodeResult(bytes.NewReader(testCase.encoded))

			assertInvalidData(t, "DecodeResult()", err)
		})
	}
}

func TestEncodeArgumentsWithWrongCountPanics(t *testing.T) {
	testCases := []wrongArgumentsTestCase{
		{"missing argument", "deposit", nil},
		{"extra argument", "deposit", []types.Value{integer(1), integer(2)}},
		{"argument for a method without parameters", "balance", []types.Value{integer(1)}},
	}
	account := accountInterface()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			method := methodOf(t, account, testCase.methodName)

			assertPanics(t, "EncodeArguments()", func() {
				_ = method.EncodeArguments(io.Discard, testCase.arguments)
			})
		})
	}
}

func TestEncodeReturnsWriterErrors(t *testing.T) {
	withdraw := methodOf(t, accountInterface(), "withdraw")

	argumentsErr := withdraw.EncodeArguments(failingWriter{}, []types.Value{integer(30)})
	resultErr := withdraw.EncodeResult(failingWriter{}, withdrawResult("InsufficientFunds", integer(100)))

	if !errors.Is(argumentsErr, errConnectionReset) {
		t.Errorf("EncodeArguments() error = %v; want %v", argumentsErr, errConnectionReset)
	}
	if !errors.Is(resultErr, errConnectionReset) {
		t.Errorf("EncodeResult() error = %v; want %v", resultErr, errConnectionReset)
	}
}

func TestSameServiceHasTheSameHash(t *testing.T) {
	first := wire.NewInterface(accountService("Account", "deposit"))
	second := wire.NewInterface(accountService("Account", "deposit"))

	if first.Hash != second.Hash {
		t.Errorf("hashes of the same service differ: %x and %x", first.Hash, second.Hash)
	}
}

func TestDifferentServicesHaveDifferentHashes(t *testing.T) {
	account := wire.NewInterface(accountService("Account", "deposit"))
	renamedService := wire.NewInterface(accountService("Wallet", "deposit"))
	renamedMethod := wire.NewInterface(accountService("Account", "put"))

	if account.Hash == renamedService.Hash {
		t.Errorf("renaming the service did not change the hash %x", account.Hash)
	}
	if account.Hash == renamedMethod.Hash {
		t.Errorf("renaming a method did not change the hash %x", account.Hash)
	}
}
