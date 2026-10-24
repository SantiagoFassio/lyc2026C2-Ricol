package types_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type typeNameTestCase struct {
	name     string
	value    types.Value
	expected string
}

type valueEqualityTestCase struct {
	name     string
	left     types.Value
	right    types.Value
	expected bool
}

func TestTypeName(t *testing.T) {
	testCases := []typeNameTestCase{
		{"integer", integer(3), "Int"},
		{"float", float(2.5), "Float"},
		{"float without decimals", float(2), "Float"},
		{"string", stringValue("hola"), "String"},
		{"true", booleanValue(true), "Bool"},
		{"false", booleanValue(false), "Bool"},
		{"struct", punto(1, 2), "Punto"},
		{"empty struct", structValue("Vacio"), "Vacio"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.TypeName()

			if result != testCase.expected {
				t.Errorf("(%#v).TypeName() = %q; want %q", testCase.value, result, testCase.expected)
			}
		})
	}
}

func TestEqualsWithValuesOfDifferentTypes(t *testing.T) {
	testCases := []valueEqualityTestCase{
		{"integer and string", integer(1), stringValue("1"), false},
		{"string and integer", stringValue("1"), integer(1), false},
		{"boolean and integer", booleanValue(true), integer(1), false},
		{"integer zero and false", integer(0), booleanValue(false), false},
		{"string and boolean", stringValue("True"), booleanValue(true), false},
		{"empty string and empty struct", stringValue(""), structValue("Vacio"), false},
		{"struct and integer", punto(1, 2), integer(1), false},
		{"integer and struct", integer(1), punto(1, 2), false},
		{"fields of different types", structValue("Dato", field("x", integer(1))), structValue("Dato", field("x", stringValue("1"))), false},
		{"integer and float are both numbers", integer(1), float(1), true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left.Equals(testCase.right)

			if result != testCase.expected {
				t.Errorf("(%v).Equals(%v) = %v; want %v", testCase.left, testCase.right, result, testCase.expected)
			}
		})
	}
}
