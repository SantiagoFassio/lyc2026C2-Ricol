package wire

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"runtime"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type encodingTestCase struct {
	name      string
	valueType types.Type
	value     types.Value
	encoded   []byte
}

type roundTripTestCase struct {
	name      string
	valueType types.Type
	value     types.Value
}

type invalidDataTestCase struct {
	name      string
	valueType types.Type
	encoded   []byte
}

type readerErrorTestCase struct {
	name      string
	valueType types.Type
	received  []byte
}

type unexpectedValueTestCase struct {
	name      string
	valueType types.Type
	value     types.Value
}

type unsupportedTypeTestCase struct {
	name      string
	valueType types.Type
}

type failingWriter struct{}

var errConnectionReset = errors.New("connection reset")

func (failingWriter) Write([]byte) (int, error) {
	return 0, errConnectionReset
}

func integer(value int64) types.Value {
	return types.NewInteger(value)
}

func float(value float64) types.Value {
	return types.NewFloat(value)
}

func boolean(value bool) types.Value {
	return types.NewBoolean(value)
}

func str(value string) types.Value {
	return types.NewString(value)
}

func hexBytes(text string) []byte {
	data, err := hex.DecodeString(strings.ReplaceAll(text, " ", ""))
	if err != nil {
		panic(err)
	}
	return data
}

func uint64Bytes(value uint64) []byte {
	return []byte{
		byte(value >> 56), byte(value >> 48), byte(value >> 40), byte(value >> 32),
		byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value),
	}
}

func sameValue(left types.Value, right types.Value) bool {
	leftNumber, ok := left.(types.Number)
	if !ok {
		if left == nil {
			return right == nil
		}
		return left.Equals(right)
	}
	rightNumber, ok := right.(types.Number)
	if !ok {
		return false
	}
	if leftInteger, ok := leftNumber.IntegerValue(); ok {
		rightInteger, ok := rightNumber.IntegerValue()
		return ok && leftInteger == rightInteger
	}
	leftFloat, _ := leftNumber.FloatValue()
	rightFloat, ok := rightNumber.FloatValue()
	return ok && math.Float64bits(leftFloat) == math.Float64bits(rightFloat)
}

func encode(t *testing.T, valueType types.Type, value types.Value) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := newCodec(valueType).encode(&buffer, value, 0); err != nil {
		t.Fatalf("encode(%#v) unexpected error: %v", value, err)
	}
	return buffer.Bytes()
}

func decode(t *testing.T, valueType types.Type, data []byte) types.Value {
	t.Helper()

	reader := bytes.NewReader(append(bytes.Clone(data), 0xFF))
	value, err := newCodec(valueType).decode(reader, 0)
	if err != nil {
		t.Fatalf("decode(% X) unexpected error: %v", data, err)
	}
	if reader.Len() != 1 {
		t.Errorf("decode(% X) left %d bytes unread; want 1", data, reader.Len())
	}
	return value
}

func assertSameValue(t *testing.T, description string, result types.Value, expected types.Value) {
	t.Helper()

	if !sameValue(result, expected) {
		t.Errorf("%s = %#v; want %#v", description, result, expected)
	}
}

func assertInvalidData(t *testing.T, description string, err error) {
	t.Helper()

	if !errors.Is(err, ErrInvalidData) {
		t.Errorf("%s error = %v; want %v", description, err, ErrInvalidData)
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

func primitiveEncodingTestCases() []encodingTestCase {
	return []encodingTestCase{
		{"integer zero", types.Int, integer(0), hexBytes("00 00 00 00 00 00 00 00")},
		{"integer one", types.Int, integer(1), hexBytes("00 00 00 00 00 00 00 01")},
		{"integer thirty", types.Int, integer(30), hexBytes("00 00 00 00 00 00 00 1E")},
		{"integer minus one", types.Int, integer(-1), hexBytes("FF FF FF FF FF FF FF FF")},
		{"max int64", types.Int, integer(math.MaxInt64), hexBytes("7F FF FF FF FF FF FF FF")},
		{"min int64", types.Int, integer(math.MinInt64), hexBytes("80 00 00 00 00 00 00 00")},
		{"float zero", types.Float, float(0), hexBytes("00 00 00 00 00 00 00 00")},
		{"negated float zero", types.Float, float(math.Copysign(0, -1)), hexBytes("80 00 00 00 00 00 00 00")},
		{"float with decimals", types.Float, float(1.5), hexBytes("3F F8 00 00 00 00 00 00")},
		{"float without exact representation", types.Float, float(0.1), hexBytes("3F B9 99 99 99 99 99 9A")},
		{"negative float", types.Float, float(-2.5), hexBytes("C0 04 00 00 00 00 00 00")},
		{"positive infinity", types.Float, float(math.Inf(1)), hexBytes("7F F0 00 00 00 00 00 00")},
		{"negative infinity", types.Float, float(math.Inf(-1)), hexBytes("FF F0 00 00 00 00 00 00")},
		{"not a number", types.Float, float(math.NaN()), uint64Bytes(math.Float64bits(math.NaN()))},
		{"false", types.Bool, boolean(false), hexBytes("00")},
		{"true", types.Bool, boolean(true), hexBytes("01")},
		{"empty string", types.Str, str(""), hexBytes("00 00 00 00")},
		{"ASCII string", types.Str, str("hello"), hexBytes("00 00 00 05 68 65 6C 6C 6F")},
		{"two-byte character", types.Str, str("ñ"), hexBytes("00 00 00 02 C3 B1")},
		{"three-byte character", types.Str, str("€"), hexBytes("00 00 00 03 E2 82 AC")},
		{"four-byte character", types.Str, str("😀"), hexBytes("00 00 00 04 F0 9F 98 80")},
		{"void", types.Void, nil, []byte{}},
	}
}

func TestEncodePrimitives(t *testing.T) {
	for _, testCase := range primitiveEncodingTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			result := encode(t, testCase.valueType, testCase.value)

			if !bytes.Equal(result, testCase.encoded) {
				t.Errorf("encode(%#v) = % X; want % X", testCase.value, result, testCase.encoded)
			}
		})
	}
}

func TestDecodePrimitives(t *testing.T) {
	for _, testCase := range primitiveEncodingTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			result := decode(t, testCase.valueType, testCase.encoded)

			assertSameValue(t, "decode", result, testCase.value)
		})
	}
}

func TestPrimitivesRoundTrip(t *testing.T) {
	testCases := []roundTripTestCase{
		{"large integer", types.Int, integer(1234567890123)},
		{"negative integer", types.Int, integer(-987654321)},
		{"float with many decimals", types.Float, float(math.Pi)},
		{"smallest positive float", types.Float, float(math.SmallestNonzeroFloat64)},
		{"max float", types.Float, float(math.MaxFloat64)},
		{"string with spaces and symbols", types.Str, str("hello, world! 1 + 1 = 2\n")},
		{"string with several scripts", types.Str, str("año αβγ")},
		{"string longer than a read chunk", types.Str, str(strings.Repeat("abc", 100000))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := decode(t, testCase.valueType, encode(t, testCase.valueType, testCase.value))

			assertSameValue(t, "decode(encode(value))", result, testCase.value)
		})
	}
}

func TestDecodeInvalidPrimitives(t *testing.T) {
	testCases := []invalidDataTestCase{
		{"empty integer", types.Int, hexBytes("")},
		{"truncated integer", types.Int, hexBytes("00 00 00 00 00 00 00")},
		{"empty float", types.Float, hexBytes("")},
		{"truncated float", types.Float, hexBytes("3F F8 00")},
		{"empty bool", types.Bool, hexBytes("")},
		{"bool two", types.Bool, hexBytes("02")},
		{"bool with all bits set", types.Bool, hexBytes("FF")},
		{"empty string", types.Str, hexBytes("")},
		{"truncated string length", types.Str, hexBytes("00 00")},
		{"string length bigger than the data", types.Str, hexBytes("00 00 00 05 68 65 6C")},
		{"string with the maximum length", types.Str, hexBytes("FF FF FF FF 68 65 6C")},
		{"string with an invalid byte", types.Str, hexBytes("00 00 00 01 FF")},
		{"string with a truncated character", types.Str, hexBytes("00 00 00 01 C3")},
		{"string with an encoded surrogate", types.Str, hexBytes("00 00 00 03 ED A0 80")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := newCodec(testCase.valueType).decode(bytes.NewReader(testCase.encoded), 0)

			assertInvalidData(t, "decode", err)
		})
	}
}

func TestDecodeStringDoesNotReserveTheReceivedLength(t *testing.T) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	_, err := stringCodec{}.decode(bytes.NewReader(hexBytes("FF FF FF FF 68 65 6C")), 0)

	runtime.ReadMemStats(&after)
	assertInvalidData(t, "decode", err)
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 1<<20 {
		t.Errorf("decode allocated %d bytes; want at most %d", allocated, 1<<20)
	}
}

func TestDecodeReturnsReaderErrors(t *testing.T) {
	testCases := []readerErrorTestCase{
		{"integer without data", types.Int, hexBytes("")},
		{"integer in the middle", types.Int, hexBytes("00 00 00")},
		{"float in the middle", types.Float, hexBytes("3F F8")},
		{"bool without data", types.Bool, hexBytes("")},
		{"string length in the middle", types.Str, hexBytes("00 00")},
		{"string content in the middle", types.Str, hexBytes("00 00 00 05 68 65")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := io.MultiReader(bytes.NewReader(testCase.received), iotest.ErrReader(errConnectionReset))

			_, err := newCodec(testCase.valueType).decode(reader, 0)

			if !errors.Is(err, errConnectionReset) {
				t.Errorf("decode error = %v; want %v", err, errConnectionReset)
			}
			if errors.Is(err, ErrInvalidData) {
				t.Errorf("decode error = %v; want an error that is not %v", err, ErrInvalidData)
			}
		})
	}
}

func TestEncodeReturnsWriterErrors(t *testing.T) {
	err := intCodec{}.encode(failingWriter{}, integer(1), 0)

	if !errors.Is(err, errConnectionReset) {
		t.Errorf("encode error = %v; want %v", err, errConnectionReset)
	}
}

func TestEncodeStringWithInvalidUTF8(t *testing.T) {
	var buffer bytes.Buffer

	err := stringCodec{}.encode(&buffer, str("hello \xFF"), 0)

	assertInvalidData(t, "encode", err)
	if buffer.Len() != 0 {
		t.Errorf("encode wrote % X; want nothing", buffer.Bytes())
	}
}

func TestEncodeValueOfAnotherTypePanics(t *testing.T) {
	testCases := []unexpectedValueTestCase{
		{"string as integer", types.Int, str("1")},
		{"float as integer", types.Int, float(1)},
		{"integer as float", types.Float, integer(1)},
		{"integer as bool", types.Bool, integer(1)},
		{"bool as string", types.Str, boolean(true)},
		{"nothing as integer", types.Int, nil},
		{"integer as void", types.Void, integer(1)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertPanics(t, "encode", func() {
				_ = newCodec(testCase.valueType).encode(io.Discard, testCase.value, 0)
			})
		})
	}
}

func TestNewCodecWithUnsupportedTypePanics(t *testing.T) {
	testCases := []unsupportedTypeTestCase{
		{"invalid type", types.Invalid},
		{"unresolved named type", types.NamedType{Name: "Point"}},
		{"no type", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertPanics(t, "newCodec", func() { newCodec(testCase.valueType) })
		})
	}
}
