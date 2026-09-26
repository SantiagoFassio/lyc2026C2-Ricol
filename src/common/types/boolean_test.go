package types_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type booleanRepresentationTestCase struct {
	name     string
	value    types.Boolean
	expected string
}

func booleanValue(value bool) types.Boolean {
	return types.NewBoolean(value)
}

func TestBooleanRepresentation(t *testing.T) {
	testCases := []booleanRepresentationTestCase{
		{"true", booleanValue(true), "True"},
		{"false", booleanValue(false), "False"},
		{"zero value is false", types.Boolean{}, "False"},
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
