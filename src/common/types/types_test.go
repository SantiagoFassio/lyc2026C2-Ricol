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

type signatureRepresentationTestCase struct {
	name      string
	signature types.Signature
	expected  string
}

type structTypeFieldTestCase struct {
	name       string
	structType *types.StructType
	fieldName  string
	expected   types.Type
}

type structTypeMissingFieldTestCase struct {
	name       string
	structType *types.StructType
	fieldName  string
}

type typeIdentityTestCase struct {
	name     string
	left     types.Type
	right    types.Type
	expected bool
}

func puntoType() *types.StructType {
	return types.NewStructType("Punto", []types.StructField{
		{Name: "x", Type: types.Int},
		{Name: "nombre", Type: types.Str},
	})
}

func TestTypeRepresentation(t *testing.T) {
	testCases := []typeRepresentationTestCase{
		{"integer", types.Int, "Int"},
		{"float", types.Float, "Float"},
		{"string", types.Str, "String"},
		{"void", types.Void, "Void"},
		{"invalid", types.Invalid, "<Invalid>"},
		{"named type", types.NamedType{Name: "Punto"}, "Punto"},
		{"struct type", puntoType(), "Punto"},
		{"empty struct type", types.NewStructType("Vacio", nil), "Vacio"},
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

func TestSignatureRepresentation(t *testing.T) {
	testCases := []signatureRepresentationTestCase{
		{"no parameters", types.Signature{Return: types.Int}, "() -> Int"},
		{"one parameter", types.Signature{Params: []types.Type{types.Str}, Return: types.Bool}, "(String) -> Bool"},
		{"several parameters", types.Signature{Params: []types.Type{types.Int, types.Float, types.Bool}, Return: types.Float},
			"(Int, Float, Bool) -> Float"},
		{"void return", types.Signature{Params: []types.Type{types.Int}, Return: types.Void}, "(Int) -> Void"},
		{"struct types", types.Signature{Params: []types.Type{puntoType(), types.Int}, Return: types.NewStructType("Linea", nil)},
			"(Punto, Int) -> Linea"},
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

func TestStructTypeField(t *testing.T) {
	punto := puntoType()
	linea := types.NewStructType("Linea", []types.StructField{{Name: "desde", Type: punto}})
	testCases := []structTypeFieldTestCase{
		{"first field", punto, "x", types.Int},
		{"last field", punto, "nombre", types.Str},
		{"struct field", linea, "desde", punto},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := testCase.structType.Field(testCase.fieldName)

			if !ok {
				t.Fatalf("(%v).Field(%q) not found; want %v", testCase.structType, testCase.fieldName, testCase.expected)
			}
			if result != testCase.expected {
				t.Errorf("(%v).Field(%q) = %v; want %v", testCase.structType, testCase.fieldName, result, testCase.expected)
			}
		})
	}
}

func TestStructTypeMissingField(t *testing.T) {
	testCases := []structTypeMissingFieldTestCase{
		{"unknown field", puntoType(), "z"},
		{"field names are case sensitive", puntoType(), "X"},
		{"empty struct", types.NewStructType("Vacio", nil), "x"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := testCase.structType.Field(testCase.fieldName)

			if ok {
				t.Errorf("(%v).Field(%q) = %v; want not found", testCase.structType, testCase.fieldName, result)
			}
		})
	}
}

func TestTypeIdentity(t *testing.T) {
	punto := puntoType()
	testCases := []typeIdentityTestCase{
		{"same struct declaration", punto, punto, true},
		{"struct declarations with the same name and fields", punto, puntoType(), false},
		{"named types with the same name and position", types.NamedType{Name: "Punto", Line: 1, Column: 8},
			types.NamedType{Name: "Punto", Line: 1, Column: 8}, true},
		{"named types with the same name in different positions", types.NamedType{Name: "Punto", Line: 1, Column: 8},
			types.NamedType{Name: "Punto", Line: 2, Column: 8}, false},
		{"named types with different names", types.NamedType{Name: "Punto"}, types.NamedType{Name: "Linea"}, false},
		{"named type and its struct type", types.NamedType{Name: "Punto"}, punto, false},
		{"primitive type and named type with its name", types.Int, types.NamedType{Name: "Int"}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left == testCase.right

			if result != testCase.expected {
				t.Errorf("%v == %v is %v; want %v", testCase.left, testCase.right, result, testCase.expected)
			}
		})
	}
}
