package types

import "strconv"

type String struct {
	value string
}

func NewString(value string) String {
	return String{value: value}
}

func (s String) Concatenate(other String) String {
	return NewString(s.value + other.value)
}

func (s String) Equals(other String) Boolean {
	return NewBoolean(s.value == other.value)
}

func (s String) LessThan(other String) Boolean {
	return NewBoolean(s.value < other.value)
}

func (s String) LessOrEqualThan(other String) Boolean {
	return NewBoolean(s.value <= other.value)
}

func (s String) GreaterThan(other String) Boolean {
	return NewBoolean(s.value > other.value)
}

func (s String) GreaterOrEqualThan(other String) Boolean {
	return NewBoolean(s.value >= other.value)
}

func (s String) TypeName() string {
	return "String"
}

func (s String) String() string {
	return strconv.Quote(s.value)
}

func (s String) Display() string {
	return s.value
}

func (String) isValue() {}
