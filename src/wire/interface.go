package wire

import (
	"fmt"
	"io"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type Interface struct {
	Name      string
	Canonical string
	Hash      [32]byte
	methods   []*Method
}

type Method struct {
	Name      string
	Index     int
	arguments []codec
	result    codec
}

func NewInterface(service *types.ServiceType) *Interface {
	canonical := canonicalForm(service)
	builder := newCodecBuilder()
	methods := make([]*Method, len(service.Methods))
	for index, method := range service.Methods {
		arguments := make([]codec, len(method.Signature.Params))
		for position, param := range method.Signature.Params {
			arguments[position] = builder.codecFor(param)
		}
		methods[index] = &Method{
			Name:      method.Name,
			Index:     index,
			arguments: arguments,
			result:    builder.codecFor(method.Signature.Return),
		}
	}
	return &Interface{Name: service.Name, Canonical: canonical, Hash: hashOf(canonical), methods: methods}
}

func (i *Interface) Method(name string) (*Method, bool) {
	for _, method := range i.methods {
		if method.Name == name {
			return method, true
		}
	}
	return nil, false
}

func (i *Interface) MethodAt(index int) (*Method, bool) {
	if index < 0 || index >= len(i.methods) {
		return nil, false
	}
	return i.methods[index], true
}

func (m *Method) EncodeArguments(writer io.Writer, arguments []types.Value) error {
	if len(arguments) != len(m.arguments) {
		panic(fmt.Sprintf("Method '%s' expects %d arguments, got %d", m.Name, len(m.arguments), len(arguments)))
	}
	for position, argument := range arguments {
		if err := m.arguments[position].encode(writer, argument, 0); err != nil {
			return err
		}
	}
	return nil
}

func (m *Method) DecodeArguments(reader io.Reader) ([]types.Value, error) {
	arguments := make([]types.Value, len(m.arguments))
	for position, argumentCodec := range m.arguments {
		argument, err := argumentCodec.decode(reader, 0)
		if err != nil {
			return nil, err
		}
		arguments[position] = argument
	}
	return arguments, nil
}

func (m *Method) EncodeResult(writer io.Writer, result types.Value) error {
	return m.result.encode(writer, result, 0)
}

func (m *Method) DecodeResult(reader io.Reader) (types.Value, error) {
	return m.result.decode(reader, 0)
}
