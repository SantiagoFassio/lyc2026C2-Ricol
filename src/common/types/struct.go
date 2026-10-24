package types

import (
	"slices"
	"strings"
)

type Struct struct {
	name   string
	fields []FieldValue
}

type FieldValue struct {
	Name  string
	Value Value
}

func NewStruct(name string, fields []FieldValue) Struct {
	return Struct{name: name, fields: slices.Clone(fields)}
}

func (s Struct) Field(name string) (Value, bool) {
	for _, field := range s.fields {
		if field.Name == name {
			return field.Value, true
		}
	}
	return nil, false
}

// Para verificar igualdad se chequean los datos declarados campo a campo.
// El typechecker garantiza que solo se comparen structs del mismo tipo.
func (s Struct) Equals(other Value) bool {
	otherStruct, ok := other.(Struct)
	if !ok || s.name != otherStruct.name || len(s.fields) != len(otherStruct.fields) {
		return false
	}
	for i, field := range s.fields {
		otherField := otherStruct.fields[i]
		if field.Name != otherField.Name || !field.Value.Equals(otherField.Value) {
			return false
		}
	}
	return true
}

func (s Struct) TypeName() string {
	return s.name
}

func (s Struct) String() string {
	if len(s.fields) == 0 {
		return s.name + " {}"
	}
	fields := make([]string, len(s.fields))
	for i, field := range s.fields {
		fields[i] = field.Name + ": " + field.Value.String()
	}
	return s.name + " { " + strings.Join(fields, ", ") + " }"
}

func (s Struct) Display() string {
	return s.String()
}

func (Struct) isValue() {}
