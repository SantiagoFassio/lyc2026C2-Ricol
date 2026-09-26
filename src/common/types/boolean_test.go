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

func TestBooleanEquals(t *testing.T) {
	testCases := []struct {
		name     string
		left     types.Boolean
		right    types.Boolean
		expected bool
	}{
		{"both true", booleanValue(true), booleanValue(true), true},
		{"both false", booleanValue(false), booleanValue(false), true},
		{"true and false", booleanValue(true), booleanValue(false), false},
		{"false and true", booleanValue(false), booleanValue(true), false},
		{"zero value equals false", types.Boolean{}, booleanValue(false), true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left.Equals(testCase.right)
			expected := types.NewBoolean(testCase.expected)

			if result != expected {
				t.Errorf("(%v).Equals(%v) = %v; want %v", testCase.left, testCase.right, result, expected)
			}
		})
	}
}

func TestBooleanNot(t *testing.T) {
	testCases := []struct {
		name     string
		value    types.Boolean
		expected types.Boolean
	}{
		{"not true", booleanValue(true), booleanValue(false)},
		{"not false", booleanValue(false), booleanValue(true)},
		{"double negation", booleanValue(true).Not(), booleanValue(true)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.Not()

			if result != testCase.expected {
				t.Errorf("(%v).Not() = %v; want %v", testCase.value, result, testCase.expected)
			}
		})
	}
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
