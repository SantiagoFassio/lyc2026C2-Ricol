package types

type Number interface {
	Add(Number) Number
	Substract(Number) Number
	Multiply(Number) Number
	Divide(Number) (Float, error)
	DivideInteger(Number) (Integer, error)
	Modulo(Number) (Integer, error)
	Power(Number) Number
	Negate() Number

	String() string

	addToInteger(Integer) Number
	substractFromInteger(Integer) Number
	multiplyToInteger(Integer) Number
	divideFromInteger(Integer) (Float, error)
	divideIntegerFromInteger(Integer) (Integer, error)
	moduloFromInteger(Integer) (Integer, error)
	powerFromInteger(Integer) Number

	addToFloat(Float) Float
	substractFromFloat(Float) Float
	multiplyToFloat(Float) Float
	divideFromFloat(Float) (Float, error)
	divideIntegerFromFloat(Float) (Integer, error)
	moduloFromFloat(Float) (Integer, error)
	powerFromFloat(Float) Float
}
