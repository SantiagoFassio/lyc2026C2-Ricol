package types

import (
	"fmt"
)

type Integer int64

func NewInteger(value int64) Integer {
	return Integer(value)
}

func newIntegerFromFloat(value Float) Integer {
	return Integer(value)
}

func (i Integer) Add(other Number) Number {
	return other.addToInteger(i)
}

func (i Integer) Substract(other Number) Number {
	return other.substractFromInteger(i)
}

func (i Integer) Multiply(other Number) Number {
	return other.multiplyToInteger(i)
}

func (i Integer) Divide(other Number) (Float, error) {
	return other.divideFromInteger(i)
}

func (i Integer) DivideInteger(other Number) (Integer, error) {
	return other.divideIntegerFromInteger(i)
}

func (i Integer) Modulo(other Number) (Integer, error) {
	return other.moduloFromInteger(i)
}

func (i Integer) Power(other Number) Number {
	return other.powerFromInteger(i)
}

func (i Integer) Negate() Number {
	return -i
}

func (i Integer) addToInteger(other Integer) Number {
	return other + i
}

func (i Integer) substractFromInteger(other Integer) Number {
	return other - i
}

func (i Integer) multiplyToInteger(other Integer) Number {
	return other * i
}

func (i Integer) divideFromInteger(other Integer) (Float, error) {
	return newFloatFromInt(i).divideFromFloat(newFloatFromInt(other))
}

func (i Integer) divideIntegerFromInteger(other Integer) (Integer, error) {
	if i == 0 {
		return NewInteger(0), fmt.Errorf("Cannot divide by zero: %d / %d", other, i)
	}
	return other / i, nil
}

func (i Integer) moduloFromInteger(other Integer) (Integer, error) {
	if i == 0 {
		return NewInteger(0), fmt.Errorf("Cannot divide by zero: %d %% %d", other, i)
	}
	return other % i, nil
}

func (i Integer) powerFromInteger(other Integer) Number {
	return newFloatFromInt(i).powerFromFloat(newFloatFromInt(other))
}

func (i Integer) addToFloat(other Float) Float {
	return newFloatFromInt(i).addToFloat(other)
}

func (i Integer) substractFromFloat(other Float) Float {
	return newFloatFromInt(i).substractFromFloat(other)
}

func (i Integer) multiplyToFloat(other Float) Float {
	return newFloatFromInt(i).multiplyToFloat(other)
}

func (i Integer) divideFromFloat(other Float) (Float, error) {
	return newFloatFromInt(i).divideFromFloat(other)
}

func (i Integer) divideIntegerFromFloat(other Float) (Integer, error) {
	return i.divideIntegerFromInteger(newIntegerFromFloat(other))
}

func (i Integer) moduloFromFloat(other Float) (Integer, error) {
	return i.moduloFromInteger(newIntegerFromFloat(other))
}

func (i Integer) powerFromFloat(other Float) Float {
	return newFloatFromInt(i).powerFromFloat(other)
}
