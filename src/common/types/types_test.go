package types_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type typeRepresentationTestCase struct {
	name     string
	dataType types.Type
	expected string
}

func TestTypeRepresentation(t *testing.T) {
	testCases := []typeRepresentationTestCase{
		{"integer", types.Int, "Int"},
		{"float", types.Float, "Float"},
		{"string", types.Str, "String"},
		{"invalid", types.Invalid, "<Invalid>"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.dataType.String()

			if result != testCase.expected {
				t.Errorf("(%#v).String() = %q; want %q", testCase.dataType, result, testCase.expected)
			}
		})
	}
}
