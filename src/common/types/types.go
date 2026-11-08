package types

import (
	"slices"
	"strings"
)

type Type interface {
	isType()
	String() string
}

type primitiveType struct {
	name string
}

func (primitiveType) isType() {}

func (p primitiveType) String() string { return p.name }

var (
	Int     Type = primitiveType{name: "Int"}
	Float   Type = primitiveType{name: "Float"}
	Str     Type = primitiveType{name: "String"}
	Bool    Type = primitiveType{name: "Bool"}
	Void    Type = primitiveType{name: "Void"}
	Invalid Type = primitiveType{name: "<Invalid>"}
)

type NamedType struct {
	Name   string
	File   string
	Line   int
	Column int
}

func (NamedType) isType() {}

func (n NamedType) String() string { return n.Name }

type StructType struct {
	Name   string
	Fields []StructField
}

type StructField struct {
	Name string
	Type Type
}

func NewStructType(name string, fields []StructField) *StructType {
	return &StructType{Name: name, Fields: fields}
}

func (*StructType) isType() {}

func (s *StructType) String() string { return s.Name }

func (s *StructType) Field(name string) (Type, bool) {
	for _, field := range s.Fields {
		if field.Name == name {
			return field.Type, true
		}
	}
	return nil, false
}

type EnumType struct {
	Name     string
	Variants []EnumVariant
}

type EnumVariant struct {
	Name   string
	Fields []StructField
}

func NewEnumType(name string) *EnumType {
	return &EnumType{Name: name}
}

func (*EnumType) isType() {}

func (e *EnumType) String() string { return e.Name }

func (e *EnumType) Variant(name string) (EnumVariant, bool) {
	for _, variant := range e.Variants {
		if variant.Name == name {
			return variant, true
		}
	}
	return EnumVariant{}, false
}

type Signature struct {
	Params []Type
	Return Type
}

func (s Signature) String() string {
	params := make([]string, len(s.Params))
	for i, param := range s.Params {
		params[i] = param.String()
	}
	return "(" + strings.Join(params, ", ") + ") -> " + s.Return.String()
}

func (s Signature) Equals(other Signature) bool {
	return slices.Equal(s.Params, other.Params) && s.Return == other.Return
}

type ServiceType struct {
	Name    string
	Methods []ServiceMethod
}

type ServiceMethod struct {
	Name      string
	Signature Signature
}

func NewServiceType(name string, methods []ServiceMethod) *ServiceType {
	return &ServiceType{
		Name:    name,
		Methods: methods,
	}
}

func (s *ServiceType) Method(name string) (Signature, bool) {
	for _, method := range s.Methods {
		if name == method.Name {
			return method.Signature, true
		}
	}
	return Signature{}, false
}

type NamedConnectionType struct {
	Service NamedType
	File    string
	Line    int
	Column  int
}

func (NamedConnectionType) isType() {}

func (n NamedConnectionType) String() string { return "connection<" + n.Service.Name + ">" }

type NamedSessionType struct {
	Service NamedType
	File    string
	Line    int
	Column  int
}

func (NamedSessionType) isType() {}

func (n NamedSessionType) String() string { return "session<" + n.Service.Name + ">" }

type ConnectionType struct {
	Service *ServiceType
}

func (ConnectionType) isType() {}

func (c ConnectionType) String() string { return "connection<" + c.Service.Name + ">" }

type SessionType struct {
	Service *ServiceType
}

func (SessionType) isType() {}

func (s SessionType) String() string { return "session<" + s.Service.Name + ">" }

var SessionError = &EnumType{
	Name: "SessionError",
	Variants: []EnumVariant{
		{Name: "ConnectionFailed", Fields: []StructField{{Name: "message", Type: Str}}},
		{Name: "ConnectionLost", Fields: []StructField{{Name: "message", Type: Str}}},
		{Name: "Timeout"},
		{Name: "ProtocolError", Fields: []StructField{{Name: "message", Type: Str}}},
		{Name: "ServerError", Fields: []StructField{{Name: "message", Type: Str}}},
	},
}
