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

func checkModulo(t *testing.T, left types.Number, right types.Number) types.Number {
	t.Helper()

	result, err := left.Modulo(right)

	if err != nil {
		t.Fatalf("(%v).Modulo(%v) unexpected error: %v", left, right, err)
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
		{"integer division with remainder", integer(10), integer(4), float(2.5)},
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

func TestDivideInteger(t *testing.T) {
	runFailibleOperationTestCases(t, "DivideInteger", types.Number.DivideInteger, []operationTestCase{
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
		{"negative dividend rounds towards minus infinity", integer(-7), integer(2), integer(-4)},
		{"negative divisor rounds towards minus infinity", integer(7), integer(-2), integer(-4)},
		{"both operands negative", integer(-7), integer(-2), integer(3)},
		{"negative dividend without remainder", integer(-8), integer(4), integer(-2)},
		{"negative dividend smaller than the divisor", integer(-1), integer(4), integer(-1)},
		{"zero dividend", integer(0), integer(5), integer(0)},
		{"min int64 divided by minus one overflows into itself", integer(math.MinInt64), integer(-1), integer(math.MinInt64)},
	})
}

func TestDivideIntegerByZero(t *testing.T) {
	runOperationErrorTestCases(t, "DivideInteger", types.Number.DivideInteger, []operationErrorTestCase{
		{"integer by integer zero", integer(10), integer(0), "Cannot divide by zero: 10 / 0"},
		{"integer by float zero", integer(10), float(0), "Cannot divide by zero: 10 / 0"},
		{"float by integer zero", float(10.5), integer(0), "Cannot divide by zero: 10.5 / 0"},
		{"float by float zero", float(10.5), float(0), "Cannot divide by zero: 10.5 / 0"},
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

func TestModuloZeroRemainderTakesTheSignOfTheDivisor(t *testing.T) {
	runStringTestCases(t, []stringTestCase{
		{"positive divisor", checkModulo(t, float(-4), float(2)), "0"},
		{"negative divisor", checkModulo(t, float(4), float(-2)), "-0"},
	})
}

func TestPower(t *testing.T) {
	runOperationTestCases(t, "Power", types.Number.Power, []operationTestCase{
		{"two integers always return a float", integer(2), integer(3), float(8)},
		{"zero exponent", integer(5), integer(0), float(1)},
		{"negative exponent", integer(2), integer(-1), float(0.5)},
		{"float base", float(2.5), integer(2), float(6.25)},
		{"fractional exponent", integer(9), float(0.5), float(3)},
		{"zero to the zero", integer(0), integer(0), float(1)},
		{"negative base with integer exponent", integer(-2), integer(3), float(-8)},
	})
}

func TestPowerWithoutRealResult(t *testing.T) {
	result := integer(-8).Power(float(0.5))

	if result.String() != "NaN" {
		t.Errorf("(-8).Power(0.5) = %#v; want NaN", result)
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
