package wire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"unicode/utf8"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

const maxDepth = 100000
const readChunkSize = 64 * 1024

var (
	ErrInvalidData = errors.New("invalid data")
	ErrTooDeep     = errors.New("value too deep")
)

type codec interface {
	encode(writer io.Writer, value types.Value, depth int) error
	decode(reader io.Reader, depth int) (types.Value, error)
}

func newCodec(valueType types.Type) codec {
	switch valueType {
	case types.Int:
		return intCodec{}
	case types.Float:
		return floatCodec{}
	case types.Bool:
		return boolCodec{}
	case types.Str:
		return stringCodec{}
	case types.Void:
		return voidCodec{}
	}
	panic(fmt.Sprintf("Cannot build a codec for type %v (%T)", valueType, valueType))
}

type intCodec struct{}

func (intCodec) encode(writer io.Writer, value types.Value, _ int) error {
	number, isNumber := value.(types.Number)
	integer, ok := number.IntegerValue()
	if !isNumber || !ok {
		panic(unexpectedValue("Int", value))
	}
	return writeUint64(writer, uint64(integer))
}

func (intCodec) decode(reader io.Reader, _ int) (types.Value, error) {
	bits, err := readUint64(reader)
	if err != nil {
		return nil, err
	}
	return types.NewInteger(int64(bits)), nil
}

type floatCodec struct{}

func (floatCodec) encode(writer io.Writer, value types.Value, _ int) error {
	number, isNumber := value.(types.Number)
	float, ok := number.FloatValue()
	if !isNumber || !ok {
		panic(unexpectedValue("Float", value))
	}
	return writeUint64(writer, math.Float64bits(float))
}

func (floatCodec) decode(reader io.Reader, _ int) (types.Value, error) {
	bits, err := readUint64(reader)
	if err != nil {
		return nil, err
	}
	return types.NewFloat(math.Float64frombits(bits)), nil
}

type boolCodec struct{}

func (boolCodec) encode(writer io.Writer, value types.Value, _ int) error {
	boolean, ok := value.(types.Boolean)
	if !ok {
		panic(unexpectedValue("Bool", value))
	}
	encoded := byte(0x00)
	if boolean.IsTrue() {
		encoded = 0x01
	}
	_, err := writer.Write([]byte{encoded})
	return err
}

func (boolCodec) decode(reader io.Reader, _ int) (types.Value, error) {
	data, err := readFull(reader, 1)
	if err != nil {
		return nil, err
	}
	switch data[0] {
	case 0x00:
		return types.NewBoolean(false), nil
	case 0x01:
		return types.NewBoolean(true), nil
	}
	return nil, fmt.Errorf("%w: invalid Bool byte 0x%02X", ErrInvalidData, data[0])
}

type stringCodec struct{}

func (stringCodec) encode(writer io.Writer, value types.Value, _ int) error {
	str, ok := value.(types.String)
	if !ok {
		panic(unexpectedValue("String", value))
	}
	text := str.Display()
	if !utf8.ValidString(text) {
		return fmt.Errorf("%w: String is not valid UTF-8", ErrInvalidData)
	}
	if uint64(len(text)) > math.MaxUint32 {
		return fmt.Errorf("%w: String of %d bytes is too long", ErrInvalidData, len(text))
	}
	data := binary.BigEndian.AppendUint32(make([]byte, 0, 4+len(text)), uint32(len(text)))
	_, err := writer.Write(append(data, text...))
	return err
}

func (stringCodec) decode(reader io.Reader, _ int) (types.Value, error) {
	header, err := readFull(reader, 4)
	if err != nil {
		return nil, err
	}
	data, err := readFull(reader, int(binary.BigEndian.Uint32(header)))
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("%w: String is not valid UTF-8", ErrInvalidData)
	}
	return types.NewString(string(data)), nil
}

type voidCodec struct{}

func (voidCodec) encode(_ io.Writer, value types.Value, _ int) error {
	if value != nil {
		panic(unexpectedValue("Void", value))
	}
	return nil
}

func (voidCodec) decode(io.Reader, int) (types.Value, error) {
	return nil, nil
}

func unexpectedValue(expectedType string, value types.Value) string {
	return fmt.Sprintf("Cannot encode %#v as %s", value, expectedType)
}

func writeUint64(writer io.Writer, value uint64) error {
	_, err := writer.Write(binary.BigEndian.AppendUint64(nil, value))
	return err
}

func readUint64(reader io.Reader) (uint64, error) {
	data, err := readFull(reader, 8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(data), nil
}

func readFull(reader io.Reader, size int) ([]byte, error) {
	data := make([]byte, 0, min(size, readChunkSize))
	for len(data) < size {
		start := len(data)
		end := start + min(size-start, readChunkSize)
		data = slices.Grow(data, end-start)[:end]
		if _, err := io.ReadFull(reader, data[start:]); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return nil, fmt.Errorf("%w: truncated", ErrInvalidData)
			}
			return nil, fmt.Errorf("cannot read value: %w", err)
		}
	}
	return data, nil
}
