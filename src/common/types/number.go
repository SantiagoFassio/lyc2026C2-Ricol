package types

import (
	"fmt"
	"math"
	"strconv"
)

type numberKind uint8

const (
	integerKind numberKind = iota
	floatKind
)

type Number struct {
	kind         numberKind
	integerValue int64
	floatValue   float64
}

func NewInteger(value int64) Number {
	return Number{kind: integerKind, integerValue: value}
}

func NewFloat(value float64) Number {
	return Number{kind: floatKind, floatValue: value}
}

func (n Number) Add(other Number) Number {
	if n.isInteger() && other.isInteger() {
		return NewInteger(n.integerValue + other.integerValue)
	}
	return NewFloat(n.asFloat() + other.asFloat())
}

func (n Number) Substract(other Number) Number {
	if n.isInteger() && other.isInteger() {
		return NewInteger(n.integerValue - other.integerValue)
	}
	return NewFloat(n.asFloat() - other.asFloat())
}

func (n Number) Multiply(other Number) Number {
	if n.isInteger() && other.isInteger() {
		return NewInteger(n.integerValue * other.integerValue)
	}
	return NewFloat(n.asFloat() * other.asFloat())
}

func (n Number) Divide(other Number) (Number, error) {
	if other.asFloat() == 0 {
		return Number{}, fmt.Errorf("Cannot divide by zero: %v / %v", n.asFloat(), other.asFloat())
	}
	return NewFloat(n.asFloat() / other.asFloat()), nil
}

func (n Number) DivideInteger(other Number) (Number, error) {
	if other.asInteger() == 0 {
		return Number{}, fmt.Errorf("Cannot divide by zero: %d / %d", n.asInteger(), other.asInteger())
	}
	return NewInteger(n.asInteger() / other.asInteger()), nil
}

func (n Number) Modulo(other Number) (Number, error) {
	if other.asInteger() == 0 {
		return Number{}, fmt.Errorf("Cannot divide by zero: %d %% %d", n.asInteger(), other.asInteger())
	}
	return NewInteger(n.asInteger() % other.asInteger()), nil
}

func (n Number) Power(other Number) Number {
	return NewFloat(math.Pow(n.asFloat(), other.asFloat()))
}

func (n Number) Negate() Number {
	if n.isInteger() {
		return NewInteger(-n.integerValue)
	}
	return NewFloat(-n.floatValue)
}

func (n Number) String() string {
	if n.isInteger() {
		return strconv.FormatInt(n.integerValue, 10)
	}
	return strconv.FormatFloat(n.floatValue, 'f', -1, 64)
}

func (n Number) isInteger() bool {
	return n.kind == integerKind
}

func (n Number) asFloat() float64 {
	if n.isInteger() {
		return float64(n.integerValue)
	}
	return n.floatValue
}

func (n Number) asInteger() int64 {
	if n.isInteger() {
		return n.integerValue
	}
	return int64(n.floatValue)
}
