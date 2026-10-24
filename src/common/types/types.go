package types

import "strings"

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
	Name string
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
