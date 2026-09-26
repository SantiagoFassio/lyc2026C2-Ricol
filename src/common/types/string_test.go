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
