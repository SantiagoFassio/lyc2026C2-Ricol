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

func (n Number) FloorDivide(other Number) (Number, error) {
	if other.asFloat() == 0 {
		return Number{}, fmt.Errorf("Cannot divide by zero: %v // %v", n, other)
	}
	if n.isInteger() && other.isInteger() {
		return NewInteger(floorDivision(n.integerValue, other.integerValue)), nil
	}
	return NewFloat(floorFloatDivision(n.asFloat(), other.asFloat())), nil
}

func (n Number) Modulo(other Number) (Number, error) {
	if other.asFloat() == 0 {
		return Number{}, fmt.Errorf("Cannot divide by zero: %v %% %v", n, other)
	}
	if n.isInteger() && other.isInteger() {
		return NewInteger(floorModulo(n.integerValue, other.integerValue)), nil
	}
	return NewFloat(floorFloatModulo(n.asFloat(), other.asFloat())), nil
}

func (n Number) Power(other Number) Number {
	if n.isInteger() && other.isInteger() && other.integerValue >= 0 {
		return NewInteger(integerPower(n.integerValue, other.integerValue))
	}
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

func integerPower(base int64, exponent int64) int64 {
	result := int64(1)
	for exponent > 0 {
		if exponent%2 == 1 {
			result *= base
		}
		base *= base
		exponent /= 2
	}
	return result
}

func floorDivision(dividend int64, divisor int64) int64 {
	quotient := dividend / divisor
	if dividend%divisor != 0 && (dividend < 0) != (divisor < 0) {
		quotient--
	}
	return quotient
}

func floorModulo(dividend int64, divisor int64) int64 {
	remainder := dividend % divisor
	if remainder != 0 && (remainder < 0) != (divisor < 0) {
		remainder += divisor
	}
	return remainder
}

func floorFloatDivision(dividend float64, divisor float64) float64 {
	remainder := math.Mod(dividend, divisor)
	quotient := (dividend - remainder) / divisor
	if remainder != 0 && (remainder < 0) != (divisor < 0) {
		quotient--
	}
	if quotient == 0 {
		return math.Copysign(0, dividend/divisor)
	}
	floorQuotient := math.Floor(quotient)
	if quotient-floorQuotient > 0.5 {
		floorQuotient++
	}
	return floorQuotient
}

func floorFloatModulo(dividend float64, divisor float64) float64 {
	remainder := math.Mod(dividend, divisor)
	if remainder == 0 {
		return math.Copysign(0, divisor)
	}
	if (remainder < 0) != (divisor < 0) {
		remainder += divisor
	}
	return remainder
}
