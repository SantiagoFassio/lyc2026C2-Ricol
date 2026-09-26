package types_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type operation func(types.Number, types.Number) types.Number

type failibleOperation func(types.Number, types.Number) (types.Number, error)

type operationTestCase struct {
	name     string
	left     types.Number
	right    types.Number
	expected types.Number
}

type operationErrorTestCase struct {
	name            string
	left            types.Number
	right           types.Number
	expectedMessage string
}

type negationTestCase struct {
	name     string
	number   types.Number
	expected types.Number
}

type stringTestCase struct {
	name     string
	number   types.Number
	expected string
}

func integer(value int64) types.Number {
	return types.NewInteger(value)
}

func float(value float64) types.Number {
	return types.NewFloat(value)
}

func checkFloorDivide(t *testing.T, left types.Number, right types.Number) types.Number {
	t.Helper()

	result, err := left.FloorDivide(right)

	if err != nil {
		t.Fatalf("(%v).FloorDivide(%v) unexpected error: %v", left, right, err)
	}
	return result
}

func checkModulo(t *testing.T, left types.Number, right types.Number) types.Number {
	t.Helper()

	result, err := left.Modulo(right)

	if err != nil {
		t.Fatalf("(%v).Modulo(%v) unexpected error: %v", left, right, err)
	}
	return result
}

func checkPower(t *testing.T, left types.Number, right types.Number) types.Number {
	t.Helper()

	result, err := left.Power(right)

	if err != nil {
		t.Fatalf("(%v).Power(%v) unexpected error: %v", left, right, err)
	}
	return result
}

func assertNumber(t *testing.T, description string, result types.Number, expected types.Number) {
	t.Helper()

	if result != expected {
		t.Errorf("%s = %#v; want %#v", description, result, expected)
	}
}

func describe(left types.Number, operationName string, right types.Number) string {
	return fmt.Sprintf("(%v).%s(%v)", left, operationName, right)
}

func runOperationTestCases(t *testing.T, operationName string, operation operation, testCases []operationTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := operation(testCase.left, testCase.right)

			assertNumber(t, describe(testCase.left, operationName, testCase.right), result, testCase.expected)
		})
	}
}

func runFailibleOperationTestCases(t *testing.T, operationName string, operation failibleOperation, testCases []operationTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			description := describe(testCase.left, operationName, testCase.right)

			result, err := operation(testCase.left, testCase.right)

			if err != nil {
				t.Fatalf("%s unexpected error: %v", description, err)
			}
			assertNumber(t, description, result, testCase.expected)
		})
	}
}

func runOperationErrorTestCases(t *testing.T, operationName string, operation failibleOperation, testCases []operationErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			description := describe(testCase.left, operationName, testCase.right)

			result, err := operation(testCase.left, testCase.right)

			if err == nil {
				t.Fatalf("%s = nil error; want %q", description, testCase.expectedMessage)
			}
			if err.Error() != testCase.expectedMessage {
				t.Errorf("%s error = %q; want %q", description, err, testCase.expectedMessage)
			}
			assertNumber(t, description, result, types.Number{})
		})
	}
}

func runNegationTestCases(t *testing.T, testCases []negationTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.number.Negate()

			assertNumber(t, fmt.Sprintf("(%v).Negate()", testCase.number), result, testCase.expected)
		})
	}
}

func runStringTestCases(t *testing.T, testCases []stringTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.number.String()

			if result != testCase.expected {
				t.Errorf("(%#v).String() = %q; want %q", testCase.number, result, testCase.expected)
			}
		})
	}
}

func runDisplayTestCases(t *testing.T, testCases []stringTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.number.Display()

			if result != testCase.expected {
				t.Errorf("(%#v).Display() = %q; want %q", testCase.number, result, testCase.expected)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	runOperationTestCases(t, "Add", types.Number.Add, []operationTestCase{
		{"two integers", integer(3), integer(5), integer(8)},
		{"two negative integers", integer(-3), integer(-5), integer(-8)},
		{"integer and float", integer(3), float(0.5), float(3.5)},
		{"float and integer", float(0.5), integer(3), float(3.5)},
		{"two floats", float(1.5), float(2.25), float(3.75)},
		{"integer zeros", integer(0), integer(0), integer(0)},
		{"float result keeps the float kind", float(1.5), float(-1.5), float(0)},
	})
}

func TestSubstract(t *testing.T) {
	runOperationTestCases(t, "Substract", types.Number.Substract, []operationTestCase{
		{"two integers", integer(8), integer(5), integer(3)},
		{"negative result", integer(5), integer(8), integer(-3)},
		{"integer and float", integer(3), float(0.5), float(2.5)},
		{"float and integer", float(3.5), integer(3), float(0.5)},
		{"two floats", float(3.75), float(1.5), float(2.25)},
		{"float result is not simplified to an integer", float(3.5), float(0.5), float(3)},
	})
}

func TestMultiply(t *testing.T) {
	runOperationTestCases(t, "Multiply", types.Number.Multiply, []operationTestCase{
		{"two integers", integer(3), integer(5), integer(15)},
		{"negative integer", integer(-3), integer(5), integer(-15)},
		{"integer and float", integer(3), float(0.5), float(1.5)},
		{"float and integer", float(0.5), integer(3), float(1.5)},
		{"two floats", float(1.5), float(2.5), float(3.75)},
		{"by integer zero", integer(3), integer(0), integer(0)},
	})
}

func TestDivide(t *testing.T) {
	runFailibleOperationTestCases(t, "Divide", types.Number.Divide, []operationTestCase{
		{"two integers always return a float", integer(10), integer(2), float(5)},
		{"integer operands with remainder", integer(10), integer(4), float(2.5)},
		{"two floats", float(7.5), float(2.5), float(3)},
		{"integer and float", integer(5), float(2.5), float(2)},
		{"negative result", integer(-10), integer(4), float(-2.5)},
		{"zero dividend", integer(0), integer(5), float(0)},
	})
}

func TestDivideByZero(t *testing.T) {
	runOperationErrorTestCases(t, "Divide", types.Number.Divide, []operationErrorTestCase{
		{"integer by integer zero", integer(10), integer(0), "Cannot divide by zero: 10 / 0"},
		{"float by float zero", float(2.5), float(0), "Cannot divide by zero: 2.5 / 0"},
		{"zero by zero", integer(0), integer(0), "Cannot divide by zero: 0 / 0"},
	})
}

func TestFloorDivide(t *testing.T) {
	runFailibleOperationTestCases(t, "FloorDivide", types.Number.FloorDivide, []operationTestCase{
		{"exact division", integer(10), integer(5), integer(2)},
		{"division with remainder truncates", integer(10), integer(3), integer(3)},
		{"float operands divide in float", float(10.9), float(3.9), float(2)},
		{"float divisor", integer(10), float(2.9), float(3)},
		{"divisor between zero and one", integer(10), float(0.5), float(20)},
		{"float divisor without remainder", integer(10), float(4), float(2)},
		{"negative float dividend rounds towards minus infinity", float(-7.5), integer(2), float(-4)},
		{"float dividend with negative divisor", float(7.5), integer(-2), float(-4)},
		{"negative float dividend with float divisor", float(-10.9), float(3.9), float(-3)},
		{"a quotient that rounds up to an exact integer is not rounded up", float(0.5), float(0.1), float(4)},
		{"a quotient that rounds up to an exact integer, negative divisor", float(0.9), float(-0.3), float(-4)},
		{"a quotient just below an integer is snapped to it", float(-165.48739011576902), float(0.4106209292141962), float(-404)},
		{"dividend smaller than the divisor", float(1.5), float(3), float(0)},
		{"dividend smaller than the divisor, both negative", float(-1.5), float(-3), float(0)},
		{"negative dividend rounds towards minus infinity", integer(-7), integer(2), integer(-4)},
		{"negative divisor rounds towards minus infinity", integer(7), integer(-2), integer(-4)},
		{"both operands negative", integer(-7), integer(-2), integer(3)},
		{"negative dividend without remainder", integer(-8), integer(4), integer(-2)},
		{"negative dividend smaller than the divisor", integer(-1), integer(4), integer(-1)},
		{"zero dividend", integer(0), integer(5), integer(0)},
		{"min int64 divided by minus one overflows into itself", integer(math.MinInt64), integer(-1), integer(math.MinInt64)},
	})
}

func TestFloorDivideByZero(t *testing.T) {
	runOperationErrorTestCases(t, "FloorDivide", types.Number.FloorDivide, []operationErrorTestCase{
		{"integer by integer zero", integer(10), integer(0), "Cannot divide by zero: 10 // 0"},
		{"integer by float zero", integer(10), float(0), "Cannot divide by zero: 10 // 0"},
		{"float by integer zero", float(10.5), integer(0), "Cannot divide by zero: 10.5 // 0"},
		{"float by float zero", float(10.5), float(0), "Cannot divide by zero: 10.5 // 0"},
	})
}

func TestModulo(t *testing.T) {
	runFailibleOperationTestCases(t, "Modulo", types.Number.Modulo, []operationTestCase{
		{"with remainder", integer(7), integer(4), integer(3)},
		{"without remainder", integer(8), integer(4), integer(0)},
		{"negative dividend takes the sign of the divisor", integer(-7), integer(4), integer(1)},
		{"negative divisor makes the remainder negative", integer(7), integer(-4), integer(-1)},
		{"both operands negative", integer(-7), integer(-2), integer(-1)},
		{"negative dividend without remainder", integer(-8), integer(4), integer(0)},
		{"negative dividend smaller than the divisor", integer(-1), integer(4), integer(3)},
		{"float operands operate in float", float(7.9), float(4.9), float(3)},
		{"float operands with remainder", float(10.9), float(3.9), float(3.1000000000000005)},
		{"divisor between zero and one", integer(5), float(0.5), float(0)},
		{"float divisor", integer(10), float(2.9), float(1.3000000000000003)},
		{"negative float dividend takes the sign of the divisor", float(-7.5), integer(2), float(0.5)},
		{"float dividend with negative divisor", float(7.5), integer(-2), float(-0.5)},
		{"negative float dividend with float divisor", float(-10.9), float(3.9), float(0.7999999999999994)},
		{"remainder of a quotient that rounds up to an exact integer", float(0.5), float(0.1), float(0.09999999999999998)},
		{"zero dividend", integer(0), integer(5), integer(0)},
	})
}

func TestModuloByZero(t *testing.T) {
	runOperationErrorTestCases(t, "Modulo", types.Number.Modulo, []operationErrorTestCase{
		{"integer by integer zero", integer(7), integer(0), "Cannot divide by zero: 7 % 0"},
		{"integer by float zero", integer(5), float(0), "Cannot divide by zero: 5 % 0"},
		{"float by integer zero", float(7.5), integer(0), "Cannot divide by zero: 7.5 % 0"},
	})
}

func TestFloorDivideZeroQuotientTakesTheSignOfTheDivision(t *testing.T) {
	runStringTestCases(t, []stringTestCase{
		{"positive divisor", checkFloorDivide(t, float(0), float(5)), "0"},
		{"negative divisor", checkFloorDivide(t, float(0), float(-5)), "-0"},
	})
}

func TestModuloZeroRemainderTakesTheSignOfTheDivisor(t *testing.T) {
	runStringTestCases(t, []stringTestCase{
		{"positive divisor", checkModulo(t, float(-4), float(2)), "0"},
		{"negative divisor", checkModulo(t, float(4), float(-2)), "-0"},
	})
}

func TestPower(t *testing.T) {
	runFailibleOperationTestCases(t, "Power", types.Number.Power, []operationTestCase{
		{"two integers return an integer", integer(2), integer(3), integer(8)},
		{"zero exponent", integer(5), integer(0), integer(1)},
		{"zero to the zero", integer(0), integer(0), integer(1)},
		{"negative base with odd exponent", integer(-2), integer(3), integer(-8)},
		{"negative base with even exponent", integer(-2), integer(4), integer(16)},
		{"one to a huge exponent", integer(1), integer(1000000), integer(1)},
		{"exact result above the float mantissa", integer(3), integer(34), integer(16677181699666569)},
		{"exact result close to max int64", integer(11), integer(17), integer(505447028499293771)},
		{"max power of two that fits in int64", integer(2), integer(62), integer(4611686018427387904)},
		{"float base returns a float", float(2.5), integer(2), float(6.25)},
		{"float base with an integer value still returns a float", float(2), integer(3), float(8)},
		{"float exponent with an integer value still returns a float", integer(2), float(3), float(8)},
		{"fractional exponent", integer(9), float(0.5), float(3)},
		{"negative float exponent returns a float", integer(2), float(-1), float(0.5)},
		{"float base with a negative exponent returns a float", float(2), integer(-1), float(0.5)},
		{"integer overflow wraps around", integer(2), integer(64), integer(0)},
		{"integer overflow with a huge exponent", integer(2), integer(100000), integer(0)},
		{"float overflow returns infinity", float(2), integer(100000), float(math.Inf(1))},
	})
}

func TestPowerWithANegativeIntegerExponent(t *testing.T) {
	runOperationErrorTestCases(t, "Power", types.Number.Power, []operationErrorTestCase{
		{
			"negative exponent",
			integer(2), integer(-1),
			"Cannot raise an integer to a negative power: 2 ** -1",
		},
		{
			"negative base and negative exponent",
			integer(-2), integer(-3),
			"Cannot raise an integer to a negative power: -2 ** -3",
		},
		{"one to a negative exponent", integer(1), integer(-1), "Cannot raise an integer to a negative power: 1 ** -1"},
		{"zero to a negative exponent", integer(0), integer(-1), "Cannot raise an integer to a negative power: 0 ** -1"},
	})
}

func TestPowerWithoutRealResult(t *testing.T) {
	result := checkPower(t, integer(-8), float(0.5))

	if result.String() != "NaN" {
		t.Errorf("(-8).Power(0.5) = %#v; want NaN", result)
	}
}

func TestNumberEquals(t *testing.T) {
	testCases := []struct {
		name     string
		left     types.Number
		right    types.Number
		expected bool
	}{
		{"equal integers", integer(3), integer(3), true},
		{"different integers", integer(3), integer(5), false},
		{"equal floats", float(2.5), float(2.5), true},
		{"different floats", float(2.5), float(2.75), false},
		{"integer and float with the same value", integer(3), float(3), true},
		{"float and integer with the same value", float(3), integer(3), true},
		{"integer and float with different values", integer(3), float(3.5), false},
		{"integer zero and float zero", integer(0), float(0), true},
		{"float zero and negated float zero", float(0), float(0).Negate(), true},
		{"negative numbers", integer(-3), float(-3), true},
		{"not a number is never equal to itself", checkPower(t, integer(-8), float(0.5)), checkPower(t, integer(-8), float(0.5)), false},
		{"a big integer loses precision against a float", integer(math.MaxInt64), float(math.MaxInt64), true},
		{"the integer just below the max is also equal to that float", integer(math.MaxInt64 - 1), float(math.MaxInt64), true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left.Equals(testCase.right)
			expected := types.NewBoolean(testCase.expected)

			if result != expected {
				t.Errorf("%s = %v; want %v", describe(testCase.left, "Equals", testCase.right), result, expected)
			}
		})
	}
}

func TestNumberOrder(t *testing.T) {
	notANumber := checkPower(t, integer(-8), float(0.5))
	testCases := []struct {
		name           string
		left           types.Number
		right          types.Number
		lessThan       bool
		lessOrEqual    bool
		greaterThan    bool
		greaterOrEqual bool
	}{
		{"smaller integer", integer(1), integer(2), true, true, false, false},
		{"bigger integer", integer(2), integer(1), false, false, true, true},
		{"equal integers", integer(2), integer(2), false, true, false, true},
		{"negative against positive", integer(-1), integer(1), true, true, false, false},
		{"two negative integers", integer(-2), integer(-1), true, true, false, false},
		{"smaller float", float(2.5), float(2.75), true, true, false, false},
		{"equal floats", float(2.5), float(2.5), false, true, false, true},
		{"integer against a bigger float", integer(2), float(2.5), true, true, false, false},
		{"integer against a float with the same value", integer(2), float(2), false, true, false, true},
		{"float against a smaller integer", float(2.5), integer(2), false, false, true, true},
		{"float zero against negated float zero", float(0), float(0).Negate(), false, true, false, true},
		{"not a number is neither smaller nor bigger", notANumber, integer(1), false, false, false, false},
		{"not a number against itself", notANumber, notANumber, false, false, false, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertOrder(t, "LessThan", testCase.left.LessThan(testCase.right), testCase.lessThan, testCase.left, testCase.right)
			assertOrder(t, "LessOrEqualThan", testCase.left.LessOrEqualThan(testCase.right), testCase.lessOrEqual, testCase.left, testCase.right)
			assertOrder(t, "GreaterThan", testCase.left.GreaterThan(testCase.right), testCase.greaterThan, testCase.left, testCase.right)
			assertOrder(t, "GreaterOrEqualThan", testCase.left.GreaterOrEqualThan(testCase.right), testCase.greaterOrEqual, testCase.left, testCase.right)
		})
	}
}

func assertOrder(t *testing.T, operationName string, result types.Boolean, expected bool, left types.Number, right types.Number) {
	t.Helper()

	if result != types.NewBoolean(expected) {
		t.Errorf("%s = %v; want %v", describe(left, operationName, right), result, types.NewBoolean(expected))
	}
}

func TestNegate(t *testing.T) {
	runNegationTestCases(t, []negationTestCase{
		{"positive integer", integer(3), integer(-3)},
		{"negative integer", integer(-3), integer(3)},
		{"positive float", float(2.5), float(-2.5)},
		{"negative float", float(-2.5), float(2.5)},
		{"integer zero", integer(0), integer(0)},
	})
}

func TestString(t *testing.T) {
	runStringTestCases(t, []stringTestCase{
		{"positive integer", integer(3), "3"},
		{"negative integer", integer(-3), "-3"},
		{"integer zero", integer(0), "0"},
		{"max int64", integer(math.MaxInt64), "9223372036854775807"},
		{"float with decimals", float(2.5), "2.5"},
		{"float without decimals", float(8), "8"},
		{"float keeps its precision", float(3.14159265358979), "3.14159265358979"},
		{"negated float zero", float(0).Negate(), "-0"},
		{"zero value is an integer zero", types.Number{}, "0"},
	})
}

func TestDisplay(t *testing.T) {
	runDisplayTestCases(t, []stringTestCase{
		{"positive integer", integer(3), "3"},
		{"negative integer", integer(-3), "-3"},
		{"integer zero", integer(0), "0"},
		{"float with decimals", float(2.5), "2.5"},
		{"float without decimals", float(8), "8"},
		{"negated float zero", float(0).Negate(), "-0"},
	})
}
