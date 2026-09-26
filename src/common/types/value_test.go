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

func TestTypeName(t *testing.T) {
	testCases := []typeNameTestCase{
		{"integer", integer(3), "integer"},
		{"float", float(2.5), "float"},
		{"float without decimals", float(2), "float"},
		{"string", stringValue("hola"), "string"},
		{"true", booleanValue(true), "boolean"},
		{"false", booleanValue(false), "boolean"},
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
