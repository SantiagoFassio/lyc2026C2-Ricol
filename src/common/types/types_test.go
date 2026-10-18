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

type enumTypeVariantTestCase struct {
	name     string
	enumType *types.EnumType
	variant  string
	expected types.EnumVariant
}

type enumTypeMissingVariantTestCase struct {
	name     string
	enumType *types.EnumType
	variant  string
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

func listType() *types.EnumType {
	list := types.NewEnumType("List")
	list.Variants = []types.EnumVariant{
		{Name: "Empty"},
		{Name: "Node", Fields: []types.StructField{{Name: "value", Type: types.Int}, {Name: "rest", Type: list}}},
	}
	return list
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
		{"enum type", listType(), "List"},
		{"enum type without variants", types.NewEnumType("Nothing"), "Nothing"},
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
		{"enum types", types.Signature{Params: []types.Type{listType()}, Return: types.NewEnumType("Result")},
			"(List) -> Result"},
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

func TestEnumTypeVariant(t *testing.T) {
	list := listType()
	testCases := []enumTypeVariantTestCase{
		{"variant without fields", list, "Empty", list.Variants[0]},
		{"variant with fields", list, "Node", list.Variants[1]},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := testCase.enumType.Variant(testCase.variant)

			if !ok {
				t.Fatalf("(%v).Variant(%q) not found; want %v", testCase.enumType, testCase.variant, testCase.expected)
			}
			if result.Name != testCase.expected.Name || len(result.Fields) != len(testCase.expected.Fields) {
				t.Fatalf("(%v).Variant(%q) = %v; want %v", testCase.enumType, testCase.variant, result, testCase.expected)
			}
			for i, resultField := range result.Fields {
				if resultField != testCase.expected.Fields[i] {
					t.Errorf("(%v).Variant(%q).Fields[%d] = %v; want %v",
						testCase.enumType, testCase.variant, i, resultField, testCase.expected.Fields[i])
				}
			}
		})
	}
}

func TestRecursiveEnumTypeFieldIsTheSameType(t *testing.T) {
	list := listType()

	node, _ := list.Variant("Node")

	if node.Fields[1].Type != list {
		t.Errorf("(%v).Variant(%q).Fields[1].Type = %p; want the same *EnumType %p", list, "Node", node.Fields[1].Type, list)
	}
}

func TestEnumTypeMissingVariant(t *testing.T) {
	testCases := []enumTypeMissingVariantTestCase{
		{"unknown variant", listType(), "Other"},
		{"variant names are case sensitive", listType(), "empty"},
		{"field name is not a variant", listType(), "value"},
		{"enum without variants", types.NewEnumType("Nothing"), "Empty"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := testCase.enumType.Variant(testCase.variant)

			if ok {
				t.Errorf("(%v).Variant(%q) = %v; want not found", testCase.enumType, testCase.variant, result)
			}
		})
	}
}

func TestTypeIdentity(t *testing.T) {
	punto := puntoType()
	list := listType()
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
		{"same enum declaration", list, list, true},
		{"enum declarations with the same name and variants", list, listType(), false},
		{"named type and its enum type", types.NamedType{Name: "List"}, list, false},
		{"struct and enum with the same name", types.NewStructType("List", nil), types.NewEnumType("List"), false},
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
