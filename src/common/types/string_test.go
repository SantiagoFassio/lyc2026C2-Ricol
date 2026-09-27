package types_test

import (
	"fmt"
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

type stringEqualityTestCase struct {
	name     string
	left     types.String
	right    types.String
	expected bool
}

type stringOrderTestCase struct {
	name           string
	left           types.String
	right          types.String
	lessThan       bool
	lessOrEqual    bool
	greaterThan    bool
	greaterOrEqual bool
}

func stringValue(value string) types.String {
	return types.NewString(value)
}

func describeStrings(left types.String, operationName string, right types.String) string {
	return fmt.Sprintf("(%v).%s(%v)", left, operationName, right)
}

func runStringEqualityTestCases(t *testing.T, testCases []stringEqualityTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left.Equals(testCase.right)

			assertBoolean(t, describeStrings(testCase.left, "Equals", testCase.right), result, testCase.expected)
		})
	}
}

func runStringOrderTestCases(t *testing.T, testCases []stringOrderTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			left, right := testCase.left, testCase.right

			assertBoolean(t, describeStrings(left, "LessThan", right), left.LessThan(right), testCase.lessThan)
			assertBoolean(t, describeStrings(left, "LessOrEqualThan", right), left.LessOrEqualThan(right), testCase.lessOrEqual)
			assertBoolean(t, describeStrings(left, "GreaterThan", right), left.GreaterThan(right), testCase.greaterThan)
			assertBoolean(t, describeStrings(left, "GreaterOrEqualThan", right), left.GreaterOrEqualThan(right), testCase.greaterOrEqual)
		})
	}
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
	runStringEqualityTestCases(t, []stringEqualityTestCase{
		{"equal strings", stringValue("hola"), stringValue("hola"), true},
		{"different strings", stringValue("hola"), stringValue("adios"), false},
		{"different case", stringValue("hola"), stringValue("Hola"), false},
		{"two empty strings", stringValue(""), stringValue(""), true},
		{"empty against non empty", stringValue(""), stringValue("hola"), false},
		{"unicode characters", stringValue("ñandú"), stringValue("ñandú"), true},
		{"a tab is not two spaces", stringValue("\t"), stringValue("  "), false},
	})
}

func TestStringOrder(t *testing.T) {
	runStringOrderTestCases(t, []stringOrderTestCase{
		{"alphabetical order", stringValue("a"), stringValue("b"), true, true, false, false},
		{"reverse alphabetical order", stringValue("b"), stringValue("a"), false, false, true, true},
		{"equal strings", stringValue("hola"), stringValue("hola"), false, true, false, true},
		{"uppercase before lowercase", stringValue("Z"), stringValue("a"), true, true, false, false},
		{"a prefix is smaller than the whole string", stringValue("hol"), stringValue("hola"), true, true, false, false},
		{"the empty string is the smallest", stringValue(""), stringValue("a"), true, true, false, false},
		{"accented characters go last", stringValue("z"), stringValue("ñ"), true, true, false, false},
		{"digits inside a string are compared as characters", stringValue("10"), stringValue("9"), true, true, false, false},
	})
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
