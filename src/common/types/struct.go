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
func (s Struct) Equals(other Struct) Boolean {
	if s.name != other.name || len(s.fields) != len(other.fields) {
		return NewBoolean(false)
	}
	for i, field := range s.fields {
		otherField := other.fields[i]
		if field.Name != otherField.Name || !field.Value.equalsValue(otherField.Value) {
			return NewBoolean(false)
		}
	}
	return NewBoolean(true)
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

func (s Struct) equalsValue(other Value) bool {
	otherStruct, ok := other.(Struct)
	return ok && s.Equals(otherStruct).IsTrue()
}

func (Struct) isValue() {}
