package types_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type concatenationTestCase struct {
	name     string
	left     types.String
	right    types.String
	expected types.String
}

type stringRepresentationTestCase struct {
	name     string
	value    types.String
	expected string
}

func stringValue(value string) types.String {
	return types.NewString(value)
}

func TestConcatenate(t *testing.T) {
	testCases := []concatenationTestCase{
		{"two strings", stringValue("Hola, "), stringValue("mundo"), stringValue("Hola, mundo")},
		{"empty left string", stringValue(""), stringValue("mundo"), stringValue("mundo")},
		{"empty right string", stringValue("Hola"), stringValue(""), stringValue("Hola")},
		{"two empty strings", stringValue(""), stringValue(""), stringValue("")},
		{"unicode characters", stringValue("ñan"), stringValue("dú"), stringValue("ñandú")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left.Concatenate(testCase.right)

			if result != testCase.expected {
				t.Errorf("(%v).Concatenate(%v) = %v; want %v", testCase.left, testCase.right, result, testCase.expected)
			}
		})
	}
}

func TestStringEquals(t *testing.T) {
	testCases := []struct {
		name     string
		left     types.String
		right    types.String
		expected bool
	}{
		{"equal strings", stringValue("hola"), stringValue("hola"), true},
		{"different strings", stringValue("hola"), stringValue("adios"), false},
		{"different case", stringValue("hola"), stringValue("Hola"), false},
		{"two empty strings", stringValue(""), stringValue(""), true},
		{"empty against non empty", stringValue(""), stringValue("hola"), false},
		{"unicode characters", stringValue("ñandú"), stringValue("ñandú"), true},
		{"a tab is not two spaces", stringValue("\t"), stringValue("  "), false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left.Equals(testCase.right)

			if result != testCase.expected {
				t.Errorf("(%v).Equals(%v) = %t; want %t", testCase.left, testCase.right, result, testCase.expected)
			}
		})
	}
}

func TestStringOrder(t *testing.T) {
	testCases := []struct {
		name           string
		left           types.String
		right          types.String
		lessThan       bool
		lessOrEqual    bool
		greaterThan    bool
		greaterOrEqual bool
	}{
		{"alphabetical order", stringValue("a"), stringValue("b"), true, true, false, false},
		{"reverse alphabetical order", stringValue("b"), stringValue("a"), false, false, true, true},
		{"equal strings", stringValue("hola"), stringValue("hola"), false, true, false, true},
		{"uppercase before lowercase", stringValue("Z"), stringValue("a"), true, true, false, false},
		{"a prefix is smaller than the whole string", stringValue("hol"), stringValue("hola"), true, true, false, false},
		{"the empty string is the smallest", stringValue(""), stringValue("a"), true, true, false, false},
		{"accented characters go last", stringValue("z"), stringValue("ñ"), true, true, false, false},
		{"digits inside a string are compared as characters", stringValue("10"), stringValue("9"), true, true, false, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertStringOrder(t, "LessThan", testCase.left.LessThan(testCase.right), testCase.lessThan, testCase.left, testCase.right)
			assertStringOrder(t, "LessOrEqualThan", testCase.left.LessOrEqualThan(testCase.right), testCase.lessOrEqual, testCase.left, testCase.right)
			assertStringOrder(t, "GreaterThan", testCase.left.GreaterThan(testCase.right), testCase.greaterThan, testCase.left, testCase.right)
			assertStringOrder(t, "GreaterOrEqualThan", testCase.left.GreaterOrEqualThan(testCase.right), testCase.greaterOrEqual, testCase.left, testCase.right)
		})
	}
}

func assertStringOrder(t *testing.T, operationName string, result bool, expected bool, left types.String, right types.String) {
	t.Helper()

	if result != expected {
		t.Errorf("(%v).%s(%v) = %t; want %t", left, operationName, right, result, expected)
	}
}

func TestStringRepresentation(t *testing.T) {
	testCases := []stringRepresentationTestCase{
		{"simple", stringValue("hola"), `"hola"`},
		{"empty", stringValue(""), `""`},
		{"double quote", stringValue(`dijo "hola"`), `"dijo \"hola\""`},
		{"backslash", stringValue(`a\b`), `"a\\b"`},
		{"line feed", stringValue("a\nb"), `"a\nb"`},
		{"tabulation", stringValue("a\tb"), `"a\tb"`},
		{"unicode characters", stringValue("ñandú"), `"ñandú"`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.String()

			if result != testCase.expected {
				t.Errorf("(%#v).String() = %q; want %q", testCase.value, result, testCase.expected)
			}
		})
	}
}

func TestStringDisplay(t *testing.T) {
	testCases := []stringRepresentationTestCase{
		{"simple", stringValue("hola"), "hola"},
		{"empty", stringValue(""), ""},
		{"double quote", stringValue(`dijo "hola"`), `dijo "hola"`},
		{"backslash", stringValue(`a\b`), `a\b`},
		{"line feed", stringValue("a\nb"), "a\nb"},
		{"tabulation", stringValue("a\tb"), "a\tb"},
		{"unicode characters", stringValue("ñandú"), "ñandú"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.Display()

			if result != testCase.expected {
				t.Errorf("(%#v).Display() = %q; want %q", testCase.value, result, testCase.expected)
			}
		})
	}
}
