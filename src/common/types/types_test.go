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
		{"void", types.Void, "Void"},
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

type signatureRepresentationTestCase struct {
	name      string
	signature types.Signature
	expected  string
}

func TestSignatureRepresentation(t *testing.T) {
	testCases := []signatureRepresentationTestCase{
		{"no parameters", types.Signature{Return: types.Int}, "() -> Int"},
		{"one parameter", types.Signature{Params: []types.Type{types.Str}, Return: types.Bool}, "(String) -> Bool"},
		{"several parameters", types.Signature{Params: []types.Type{types.Int, types.Float, types.Bool}, Return: types.Float},
			"(Int, Float, Bool) -> Float"},
		{"void return", types.Signature{Params: []types.Type{types.Int}, Return: types.Void}, "(Int) -> Void"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.signature.String()

			if result != testCase.expected {
				t.Errorf("(%#v).String() = %q; want %q", testCase.signature, result, testCase.expected)
			}
		})
	}
}
