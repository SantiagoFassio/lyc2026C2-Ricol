package wire

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
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

type depthTestCase struct {
	name      string
	valueType types.Type
	value     types.Value
	encoded   []byte
	levels    int
}

type failingWriter struct{}

var errConnectionReset = errors.New("connection reset")

func (failingWriter) Write([]byte) (int, error) {
	return 0, errConnectionReset
}

var (
	pointType = types.NewStructType("Point", []types.StructField{
		{Name: "x", Type: types.Int},
		{Name: "y", Type: types.Int},
	})
	segmentType = types.NewStructType("Segment", []types.StructField{
		{Name: "start", Type: pointType},
		{Name: "end", Type: pointType},
		{Name: "label", Type: types.Str},
	})
	nothingType = types.NewStructType("Nothing", nil)
	shapeType   = enumType("Shape",
		types.EnumVariant{Name: "Blank"},
		types.EnumVariant{Name: "Circle", Fields: []types.StructField{
			{Name: "center", Type: pointType},
			{Name: "radius", Type: types.Float},
		}},
		types.EnumVariant{Name: "Text", Fields: []types.StructField{
			{Name: "content", Type: types.Str},
			{Name: "bold", Type: types.Bool},
		}},
	)
	drawingType = types.NewStructType("Drawing", []types.StructField{
		{Name: "shape", Type: shapeType},
		{Name: "layer", Type: types.Int},
	})
	listType = recursiveListType()
)

func enumType(name string, variants ...types.EnumVariant) *types.EnumType {
	result := types.NewEnumType(name)
	result.Variants = variants
	return result
}

func recursiveListType() *types.EnumType {
	result := types.NewEnumType("List")
	result.Variants = []types.EnumVariant{
		{Name: "Empty"},
		{Name: "Node", Fields: []types.StructField{
			{Name: "value", Type: types.Int},
			{Name: "rest", Type: result},
		}},
	}
	return result
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

func point(x int64, y int64) types.Value {
	return types.NewStruct("Point", []types.FieldValue{{Name: "x", Value: integer(x)}, {Name: "y", Value: integer(y)}})
}

func segment(start types.Value, end types.Value, label string) types.Value {
	return types.NewStruct("Segment", []types.FieldValue{
		{Name: "start", Value: start},
		{Name: "end", Value: end},
		{Name: "label", Value: str(label)},
	})
}

func drawing(shape types.Value, layer int64) types.Value {
	return types.NewStruct("Drawing", []types.FieldValue{{Name: "shape", Value: shape}, {Name: "layer", Value: integer(layer)}})
}

func variant(enumType *types.EnumType, name string, fields ...types.Value) types.Value {
	return types.NewEnum(enumType.Name, name, fields)
}

func list(values ...int64) types.Value {
	result := variant(listType, "Empty")
	for i := len(values) - 1; i >= 0; i-- {
		result = variant(listType, "Node", integer(values[i]), result)
	}
	return result
}

func deepList(levels int) types.Value {
	return list(make([]int64, levels-1)...)
}

func deepListBytes(levels int) []byte {
	node := concat(encodedIndex(1), encodedInteger(0))
	return concat(bytes.Repeat(node, levels-1), encodedIndex(0))
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

func encodedInteger(value int64) []byte {
	return uint64Bytes(uint64(value))
}

func encodedIndex(index uint32) []byte {
	return []byte{byte(index >> 24), byte(index >> 16), byte(index >> 8), byte(index)}
}

func concat(parts ...[]byte) []byte {
	return bytes.Join(parts, nil)
}

func newCodec(valueType types.Type) codec {
	return newCodecBuilder().codecFor(valueType)
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

func runEncodeTestCases(t *testing.T, testCases []encodingTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := encode(t, testCase.valueType, testCase.value)

			if !bytes.Equal(result, testCase.encoded) {
				t.Errorf("encode(%#v) = % X; want % X", testCase.value, result, testCase.encoded)
			}
		})
	}
}

func runDecodeTestCases(t *testing.T, testCases []encodingTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := decode(t, testCase.valueType, testCase.encoded)

			assertSameValue(t, "decode", result, testCase.value)
		})
	}
}

func runRoundTripTestCases(t *testing.T, testCases []roundTripTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := decode(t, testCase.valueType, encode(t, testCase.valueType, testCase.value))

			assertSameValue(t, "decode(encode(value))", result, testCase.value)
		})
	}
}

func runInvalidDataTestCases(t *testing.T, testCases []invalidDataTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := newCodec(testCase.valueType).decode(bytes.NewReader(testCase.encoded), 0)

			assertInvalidData(t, "decode", err)
		})
	}
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

func compositeEncodingTestCases() []encodingTestCase {
	return []encodingTestCase{
		{"struct", pointType, point(1, 2), concat(encodedInteger(1), encodedInteger(2))},
		{"struct without fields", nothingType, types.NewStruct("Nothing", nil), []byte{}},
		{"nested struct", segmentType, segment(point(1, 2), point(3, 4), "ab"), concat(
			encodedInteger(1), encodedInteger(2), encodedInteger(3), encodedInteger(4), hexBytes("00 00 00 02 61 62"),
		)},
		{"variant without fields", shapeType, variant(shapeType, "Blank"), encodedIndex(0)},
		{"variant with fields", shapeType, variant(shapeType, "Text", str("hi"), boolean(true)), concat(
			encodedIndex(2), hexBytes("00 00 00 02 68 69 01"),
		)},
		{"struct inside a variant", shapeType, variant(shapeType, "Circle", point(1, 2), float(1.5)), concat(
			encodedIndex(1), encodedInteger(1), encodedInteger(2), hexBytes("3F F8 00 00 00 00 00 00"),
		)},
		{"variant inside a struct", drawingType, drawing(variant(shapeType, "Blank"), 3), concat(
			encodedIndex(0), encodedInteger(3),
		)},
		{"empty recursive list", listType, list(), encodedIndex(0)},
		{"recursive list", listType, list(1, 2), concat(
			encodedIndex(1), encodedInteger(1), encodedIndex(1), encodedInteger(2), encodedIndex(0),
		)},
	}
}

func TestEncodePrimitives(t *testing.T) {
	runEncodeTestCases(t, primitiveEncodingTestCases())
}

func TestDecodePrimitives(t *testing.T) {
	runDecodeTestCases(t, primitiveEncodingTestCases())
}

func TestPrimitivesRoundTrip(t *testing.T) {
	runRoundTripTestCases(t, []roundTripTestCase{
		{"large integer", types.Int, integer(1234567890123)},
		{"negative integer", types.Int, integer(-987654321)},
		{"float with many decimals", types.Float, float(math.Pi)},
		{"smallest positive float", types.Float, float(math.SmallestNonzeroFloat64)},
		{"max float", types.Float, float(math.MaxFloat64)},
		{"string with spaces and symbols", types.Str, str("hello, world! 1 + 1 = 2\n")},
		{"string with several scripts", types.Str, str("año αβγ")},
		{"string longer than a read chunk", types.Str, str(strings.Repeat("abc", 100000))},
	})
}

func TestDecodeInvalidPrimitives(t *testing.T) {
	runInvalidDataTestCases(t, []invalidDataTestCase{
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
	})
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
		{"struct between fields", pointType, encodedInteger(1)},
		{"variant index in the middle", shapeType, hexBytes("00 00")},
		{"variant fields in the middle", shapeType, concat(encodedIndex(1), encodedInteger(1))},
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
		{"integer as struct", pointType, integer(1)},
		{"variant as struct", pointType, variant(shapeType, "Blank")},
		{"struct of another type", pointType, types.NewStruct("Location", []types.FieldValue{
			{Name: "x", Value: integer(1)}, {Name: "y", Value: integer(2)},
		})},
		{"struct without a field", pointType, types.NewStruct("Point", []types.FieldValue{{Name: "x", Value: integer(1)}})},
		{"struct with a field of another type", pointType, types.NewStruct("Point", []types.FieldValue{
			{Name: "x", Value: str("1")}, {Name: "y", Value: integer(2)},
		})},
		{"struct as enum", shapeType, point(1, 2)},
		{"variant of another enum", shapeType, variant(listType, "Empty")},
		{"unknown variant", shapeType, variant(shapeType, "Square")},
		{"variant with missing fields", shapeType, variant(shapeType, "Circle", point(1, 2))},
		{"variant with extra fields", shapeType, variant(shapeType, "Blank", integer(1))},
		{"variant with a field of another type", shapeType, variant(shapeType, "Text", str("hi"), integer(1))},
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
		{"struct with an invalid field", types.NewStructType("Broken", []types.StructField{{Name: "field", Type: types.Invalid}})},
		{"variant with an unresolved field", enumType("Broken", types.EnumVariant{
			Name: "Variant", Fields: []types.StructField{{Name: "field", Type: types.NamedType{Name: "Point"}}},
		})},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertPanics(t, "newCodec", func() { newCodec(testCase.valueType) })
		})
	}
}

func TestEncodeStructsAndEnums(t *testing.T) {
	runEncodeTestCases(t, compositeEncodingTestCases())
}

func TestDecodeStructsAndEnums(t *testing.T) {
	runDecodeTestCases(t, compositeEncodingTestCases())
}

func TestStructsAndEnumsRoundTrip(t *testing.T) {
	runRoundTripTestCases(t, []roundTripTestCase{
		{"long recursive list", listType, list(5, -4, 3, -2, 1, 0, math.MaxInt64, math.MinInt64)},
		{"struct with nested structs and strings", segmentType, segment(point(-1, 0), point(math.MaxInt64, 7), "año")},
		{"struct with a variant with a struct", drawingType, drawing(variant(shapeType, "Circle", point(5, 6), float(-0.25)), 2)},
		{"struct with a variant with primitives", drawingType, drawing(variant(shapeType, "Text", str(""), boolean(false)), -1)},
	})
}

func TestDecodeInvalidStructsAndEnums(t *testing.T) {
	runInvalidDataTestCases(t, []invalidDataTestCase{
		{"empty struct data", pointType, hexBytes("")},
		{"struct truncated between fields", pointType, encodedInteger(1)},
		{"struct truncated inside a field", pointType, concat(encodedInteger(1), hexBytes("00 00"))},
		{"empty variant data", shapeType, hexBytes("")},
		{"truncated variant index", shapeType, hexBytes("00 00")},
		{"variant index out of range", shapeType, encodedIndex(3)},
		{"variant index with all bits set", shapeType, hexBytes("FF FF FF FF")},
		{"variant truncated inside its fields", shapeType, concat(encodedIndex(1), encodedInteger(1))},
		{"invalid field inside a variant", shapeType, concat(encodedIndex(2), hexBytes("00 00 00 00 02"))},
		{"invalid variant inside a struct", drawingType, concat(encodedIndex(7), encodedInteger(1))},
		{"recursive list without its end", listType, concat(encodedIndex(1), encodedInteger(1))},
	})
}

func TestCodecIsBuiltOncePerType(t *testing.T) {
	builder := newCodecBuilder()

	segmentCodec := builder.codecFor(segmentType).(*structCodec)
	pointCodec := builder.codecFor(pointType)

	if segmentCodec.fields[0] != pointCodec || segmentCodec.fields[1] != pointCodec {
		t.Errorf("Segment fields use codecs %p and %p; want the Point codec %p",
			segmentCodec.fields[0], segmentCodec.fields[1], pointCodec)
	}
	if again := builder.codecFor(segmentType); again != segmentCodec {
		t.Errorf("second codecFor(Segment) = %p; want %p", again, segmentCodec)
	}
}

func TestRecursiveEnumUsesItsOwnCodec(t *testing.T) {
	listCodec := newCodecBuilder().codecFor(listType).(*enumCodec)

	if restCodec := listCodec.variants[1][1]; restCodec != listCodec {
		t.Errorf("codec of Node.rest = %p; want the List codec %p", restCodec, listCodec)
	}
}

func TestEachStructAndVariantAddsOneLevel(t *testing.T) {
	testCases := []depthTestCase{
		{"struct", pointType, point(1, 2), concat(encodedInteger(1), encodedInteger(2)), 1},
		{"struct without fields", nothingType, types.NewStruct("Nothing", nil), []byte{}, 1},
		{"variant without fields", shapeType, variant(shapeType, "Blank"), encodedIndex(0), 1},
		{"struct inside a struct", segmentType, segment(point(1, 2), point(3, 4), ""), concat(
			encodedInteger(1), encodedInteger(2), encodedInteger(3), encodedInteger(4), hexBytes("00 00 00 00"),
		), 2},
		{"struct inside a variant inside a struct", drawingType, drawing(variant(shapeType, "Circle", point(1, 2), float(1.5)), 3), concat(
			encodedIndex(1), encodedInteger(1), encodedInteger(2), hexBytes("3F F8 00 00 00 00 00 00"), encodedInteger(3),
		), 3},
		{"recursive list", listType, list(1, 2), concat(
			encodedIndex(1), encodedInteger(1), encodedIndex(1), encodedInteger(2), encodedIndex(0),
		), 3},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			valueCodec := newCodec(testCase.valueType)
			deepestDepth := maxDepth - testCase.levels

			if err := valueCodec.encode(io.Discard, testCase.value, deepestDepth); err != nil {
				t.Errorf("encode at depth %d unexpected error: %v", deepestDepth, err)
			}
			if err := valueCodec.encode(io.Discard, testCase.value, deepestDepth+1); !errors.Is(err, ErrTooDeep) {
				t.Errorf("encode at depth %d error = %v; want %v", deepestDepth+1, err, ErrTooDeep)
			}
			if _, err := valueCodec.decode(bytes.NewReader(testCase.encoded), deepestDepth); err != nil {
				t.Errorf("decode at depth %d unexpected error: %v", deepestDepth, err)
			}
			_, err := valueCodec.decode(bytes.NewReader(testCase.encoded), deepestDepth+1)
			assertInvalidData(t, fmt.Sprintf("decode at depth %d", deepestDepth+1), err)
		})
	}
}

func TestListWithMaxDepthLevels(t *testing.T) {
	value := deepList(maxDepth)

	encoded := encode(t, listType, value)

	if expected := deepListBytes(maxDepth); !bytes.Equal(encoded, expected) {
		t.Errorf("encode(list with %d levels) produced %d bytes; want %d", maxDepth, len(encoded), len(expected))
	}
	assertSameValue(t, "decode(encode(list))", decode(t, listType, encoded), value)
}

func TestEncodeListBeyondMaxDepth(t *testing.T) {
	err := newCodec(listType).encode(io.Discard, deepList(maxDepth+1), 0)

	if !errors.Is(err, ErrTooDeep) {
		t.Errorf("encode error = %v; want %v", err, ErrTooDeep)
	}
}

func TestDecodeListBeyondMaxDepth(t *testing.T) {
	_, err := newCodec(listType).decode(bytes.NewReader(deepListBytes(maxDepth+1)), 0)

	assertInvalidData(t, "decode", err)
}
