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

// NamedType es un tipo escrito por su nombre en el lugar donde se usa un tipo.
// Ejemplo: let v: Vector = ... => NamedType{Name: "Vector"}
// El parser no sabe a qué declaración se refiere.
// El typechecker resuelve la declaración según el scope.
type NamedType struct {
	Name string
}

func (NamedType) isType() {}

func (n NamedType) String() string { return n.Name }

// Tipo de un struct declarado, con sus campos en orden de declaración.
// Se identifica por puntero: dos declaraciones con el mismo nombre en scopes
// distintos son tipos distintos.
// Ejemplo: struct Vector { x: Int; y: Int; }
// => &StructType{Name: "Vector", Fields: []StructField{{Name: "x", Type: Int}, {Name: "y", Type: Int}}}
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
