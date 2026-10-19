package types

import (
	"slices"
	"strings"
)

type Enum struct {
	name    string
	variant string
	fields  []Value
}

func NewEnum(name string, variant string, fields []Value) Enum {
	return Enum{name: name, variant: variant, fields: slices.Clone(fields)}
}

func (e Enum) Variant() string {
	return e.variant
}

func (e Enum) Fields() []Value {
	return slices.Clone(e.fields)
}

func (e Enum) Equals(other Enum) Boolean {
	if e.name != other.name || e.variant != other.variant || len(e.fields) != len(other.fields) {
		return NewBoolean(false)
	}
	for i, field := range e.fields {
		if !field.equalsValue(other.fields[i]) {
			return NewBoolean(false)
		}
	}
	return NewBoolean(true)
}

func (e Enum) TypeName() string {
	return e.name
}

func (e Enum) String() string {
	if len(e.fields) == 0 {
		return e.name + "." + e.variant
	}
	fields := make([]string, len(e.fields))
	for i, field := range e.fields {
		fields[i] = field.String()
	}
	return e.name + "." + e.variant + "(" + strings.Join(fields, ", ") + ")"
}

func (e Enum) Display() string {
	return e.String()
}

func (e Enum) equalsValue(other Value) bool {
	otherEnum, ok := other.(Enum)
	return ok && e.Equals(otherEnum).IsTrue()
}

func (Enum) isValue() {}
