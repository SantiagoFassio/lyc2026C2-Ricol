package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

const (
	magic            = "RICOL"
	protocolVersion  = 1
	frameHeaderSize  = 4
	readChunkSize    = 64 * 1024
	maxMessageLength = 64 * 1024
)

type status byte

const (
	statusOK            status = 0x00
	statusServerError   status = 0x01
	statusProtocolError status = 0x02
)

var maxFrameLength uint64 = math.MaxUint32

var (
	errNotRicol           = errors.New("the other side is not a Ricol program")
	errUnsupportedVersion = errors.New("unsupported protocol version")
	errTruncatedFrame     = errors.New("the connection was closed in the middle of a message")
	errMessageTooLong     = errors.New("message too long")
)

func encodeHeader() []byte {
	return append([]byte(magic), protocolVersion)
}

func readHeader(reader io.Reader) error {
	header := make([]byte, len(magic)+1)
	if _, err := io.ReadFull(reader, header); err != nil {
		return err
	}
	if string(header[:len(magic)]) != magic {
		return errNotRicol
	}
	if version := header[len(magic)]; version != protocolVersion {
		return fmt.Errorf("%w %d, expected %d", errUnsupportedVersion, version, protocolVersion)
	}
	return nil
}

func encodeHello(hash [32]byte) []byte {
	return append(encodeHeader(), hash[:]...)
}

func readHello(reader io.Reader) ([32]byte, error) {
	var hash [32]byte
	if err := readHeader(reader); err != nil {
		return hash, err
	}
	_, err := io.ReadFull(reader, hash[:])
	return hash, err
}

func encodeWelcome() []byte {
	frame, _ := buildFrame(func(payload *bytes.Buffer) error {
		return payload.WriteByte(byte(statusOK))
	})
	return append(encodeHeader(), frame...)
}

func encodeRejection(message string) []byte {
	return append(encodeHeader(), encodeFailure(statusProtocolError, message)...)
}

func decodeWelcome(payload []byte) *Failure {
	body, failure := decodeStatus(payload)
	if failure != nil && failure.Kind == ServerError {
		return &Failure{Kind: ProtocolError, Message: "unexpected server error in the welcome of the server: " + failure.Message}
	}
	if failure != nil {
		return failure
	}
	if len(body) != 0 {
		return &Failure{Kind: ProtocolError, Message: "unexpected data in the welcome of the server"}
	}
	return nil
}

func encodeRequest(method *wire.Method, arguments []types.Value) ([]byte, error) {
	return buildFrame(func(payload *bytes.Buffer) error {
		payload.Write(binary.BigEndian.AppendUint32(nil, uint32(method.Index)))
		return method.EncodeArguments(payload, arguments)
	})
}

func decodeRequest(payload []byte, serviceInterface *wire.Interface) (*wire.Method, []types.Value, error) {
	if len(payload) < 4 {
		return nil, nil, errors.New("request without a method index")
	}
	index := binary.BigEndian.Uint32(payload)
	method, ok := serviceInterface.MethodAt(int(index))
	if !ok {
		return nil, nil, fmt.Errorf("unknown method index %d", index)
	}
	reader := bytes.NewReader(payload[4:])
	arguments, err := method.DecodeArguments(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid arguments of '%s': %w", method.Name, err)
	}
	if reader.Len() != 0 {
		return nil, nil, fmt.Errorf("unexpected data after the arguments of '%s'", method.Name)
	}
	return method, arguments, nil
}

func encodeResult(method *wire.Method, result types.Value) ([]byte, error) {
	return buildFrame(func(payload *bytes.Buffer) error {
		payload.WriteByte(byte(statusOK))
		return method.EncodeResult(payload, result)
	})
}

func encodeFailure(failureStatus status, message string) []byte {
	if len(message) > maxMessageLength {
		message = message[:maxMessageLength]
	}
	frame, _ := buildFrame(func(payload *bytes.Buffer) error {
		payload.WriteByte(byte(failureStatus))
		_, err := payload.WriteString(strings.ToValidUTF8(message, "�"))
		return err
	})
	return frame
}

func decodeResult(payload []byte, method *wire.Method) (types.Value, *Failure) {
	body, failure := decodeStatus(payload)
	if failure != nil {
		return nil, failure
	}
	reader := bytes.NewReader(body)
	result, err := method.DecodeResult(reader)
	if err != nil {
		return nil, &Failure{Kind: ProtocolError, Message: fmt.Sprintf("invalid result of '%s': %v", method.Name, err)}
	}
	if reader.Len() != 0 {
		return nil, &Failure{Kind: ProtocolError, Message: fmt.Sprintf("unexpected data after the result of '%s'", method.Name)}
	}
	return result, nil
}

func decodeStatus(payload []byte) ([]byte, *Failure) {
	if len(payload) == 0 {
		return nil, &Failure{Kind: ProtocolError, Message: "empty response"}
	}
	body := payload[1:]
	switch status(payload[0]) {
	case statusOK:
		return body, nil
	case statusServerError:
		return nil, failureWithMessage(ServerError, body)
	case statusProtocolError:
		return nil, failureWithMessage(ProtocolError, body)
	}
	return nil, &Failure{Kind: ProtocolError, Message: fmt.Sprintf("unknown response status 0x%02X", payload[0])}
}

func failureWithMessage(kind Kind, message []byte) *Failure {
	if !utf8.Valid(message) {
		return &Failure{Kind: ProtocolError, Message: "the error message of the server is not valid UTF-8"}
	}
	return &Failure{Kind: kind, Message: string(message)}
}

func buildFrame(writePayload func(payload *bytes.Buffer) error) ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, frameHeaderSize))
	if err := writePayload(buffer); err != nil {
		return nil, err
	}
	frame := buffer.Bytes()
	length := uint64(len(frame) - frameHeaderSize)
	if length > maxFrameLength {
		return nil, fmt.Errorf("%w: %d bytes, at most %d", errMessageTooLong, length, maxFrameLength)
	}
	binary.BigEndian.PutUint32(frame, uint32(length))
	return frame, nil
}

func readFrame(reader io.Reader) ([]byte, error) {
	header := make([]byte, frameHeaderSize)
	if _, err := io.ReadFull(reader, header); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, errTruncatedFrame
		}
		return nil, err
	}
	length := uint64(binary.BigEndian.Uint32(header))
	payload := make([]byte, 0, min(length, readChunkSize))
	for uint64(len(payload)) < length {
		start := len(payload)
		size := int(min(length-uint64(start), readChunkSize))
		payload = slices.Grow(payload, size)[:start+size]
		if _, err := io.ReadFull(reader, payload[start:]); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return nil, errTruncatedFrame
			}
			return nil, err
		}
	}
	return payload, nil
}

func rootCause(err error) error {
	for {
		next := errors.Unwrap(err)
		if next == nil {
			return err
		}
		err = next
	}
}
