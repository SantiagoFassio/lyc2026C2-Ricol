package types

import (
	"fmt"
	"math"
	"strconv"
)

type Float float64

func NewFloat(value float64) Float {
	return Float(value)
}

func newFloatFromInt(value Integer) Float {
	return Float(value)
}

func (f Float) Add(other Number) Number {
	return other.addToFloat(f)
}

func (f Float) Substract(other Number) Number {
	return other.substractFromFloat(f)
}

func (f Float) Multiply(other Number) Number {
	return other.multiplyToFloat(f)
}

func (f Float) Divide(other Number) (Float, error) {
	return other.divideFromFloat(f)
}

func (f Float) DivideInteger(other Number) (Integer, error) {
	return other.divideIntegerFromFloat(f)
}

func (f Float) Modulo(other Number) (Integer, error) {
	return other.moduloFromFloat(f)
}

func (f Float) Power(other Number) Number {
	return other.powerFromFloat(f)
}

func (f Float) Negate() Number {
	return -f
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f Float) addToFloat(other Float) Float {
	return other + f
}

func (f Float) substractFromFloat(other Float) Float {
	return other - f
}

func (f Float) multiplyToFloat(other Float) Float {
	return other * f
}

func (f Float) divideFromFloat(other Float) (Float, error) {
	if f == 0 {
		return NewFloat(0), fmt.Errorf("Cannot divide by zero: %v / %v", float64(other), float64(f))
	}
	return other / f, nil
}

func (f Float) divideIntegerFromFloat(other Float) (Integer, error) {
	return newIntegerFromFloat(f).divideIntegerFromInteger(newIntegerFromFloat(other))
}

func (f Float) moduloFromFloat(other Float) (Integer, error) {
	return newIntegerFromFloat(f).moduloFromInteger(newIntegerFromFloat(other))
}

func (f Float) powerFromFloat(other Float) Float {
	return NewFloat(math.Pow(float64(other), float64(f)))
}

func (f Float) addToInteger(other Integer) Number {
	return f.addToFloat(newFloatFromInt(other))
}

func (f Float) substractFromInteger(other Integer) Number {
	return f.substractFromFloat(newFloatFromInt(other))
}

func (f Float) multiplyToInteger(other Integer) Number {
	return f.multiplyToFloat(newFloatFromInt(other))
}

func (f Float) divideFromInteger(other Integer) (Float, error) {
	return f.divideFromFloat(newFloatFromInt(other))
}

func (f Float) divideIntegerFromInteger(other Integer) (Integer, error) {
	return newIntegerFromFloat(f).divideIntegerFromInteger(other)
}

func (f Float) moduloFromInteger(other Integer) (Integer, error) {
	return newIntegerFromFloat(f).moduloFromInteger(other)
}

func (f Float) powerFromInteger(other Integer) Number {
	return f.powerFromFloat(newFloatFromInt(other))
}
