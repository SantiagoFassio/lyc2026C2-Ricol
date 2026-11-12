package types_test

import (
	"slices"
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

type serviceTypeMethodTestCase struct {
	name        string
	serviceType *types.ServiceType
	method      string
	expected    types.Signature
}

type serviceTypeMissingMethodTestCase struct {
	name        string
	serviceType *types.ServiceType
	method      string
}

type signatureEqualityTestCase struct {
	name     string
	left     types.Signature
	right    types.Signature
	expected bool
}

func accountServiceType() *types.ServiceType {
	return types.NewServiceType("Account", []types.ServiceMethod{
		{Name: "balance", Signature: types.Signature{Params: []types.Type{}, Return: types.Int}},
		{Name: "deposit", Signature: types.Signature{Params: []types.Type{types.Int, types.Str}, Return: types.Bool}},
		{Name: "reset", Signature: types.Signature{Params: []types.Type{}, Return: types.Void}},
	})
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

func namedConnectionType(service string, line int, column int) types.NamedConnectionType {
	serviceName := types.NamedType{Name: service, Line: line, Column: column + 11}
	return types.NamedConnectionType{Service: serviceName, Line: line, Column: column}
}

func namedSessionType(service string, line int, column int) types.NamedSessionType {
	serviceName := types.NamedType{Name: service, Line: line, Column: column + 8}
	return types.NamedSessionType{Service: serviceName, Line: line, Column: column}
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
		{"named connection type", types.NamedConnectionType{Service: types.NamedType{Name: "Account"}}, "connection<Account>"},
		{"named session type", types.NamedSessionType{Service: types.NamedType{Name: "Account"}}, "session<Account>"},
		{"connection type", types.ConnectionType{Service: accountServiceType()}, "connection<Account>"},
		{"session type", types.SessionType{Service: accountServiceType()}, "session<Account>"},
		{"session error type", types.SessionError, "SessionError"},
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
	account := accountServiceType()
	testCases := []typeIdentityTestCase{
		{"same struct declaration", punto, punto, true},
		{"struct declarations with the same name and fields", punto, puntoType(), false},
		{"named types with the same name and position", types.NamedType{Name: "Punto", Line: 1, Column: 8},
			types.NamedType{Name: "Punto", Line: 1, Column: 8}, true},
		{"named types with the same name in different positions", types.NamedType{Name: "Punto", Line: 1, Column: 8},
			types.NamedType{Name: "Punto", Line: 2, Column: 8}, false},
		{"named types with the same name and position in different files", types.NamedType{Name: "Punto", File: "a.ric", Line: 1, Column: 8},
			types.NamedType{Name: "Punto", File: "b.ric", Line: 1, Column: 8}, false},
		{"named types with different names", types.NamedType{Name: "Punto"}, types.NamedType{Name: "Linea"}, false},
		{"named type and its struct type", types.NamedType{Name: "Punto"}, punto, false},
		{"primitive type and named type with its name", types.Int, types.NamedType{Name: "Int"}, false},
		{"same enum declaration", list, list, true},
		{"enum declarations with the same name and variants", list, listType(), false},
		{"named type and its enum type", types.NamedType{Name: "List"}, list, false},
		{"struct and enum with the same name", types.NewStructType("List", nil), types.NewEnumType("List"), false},
		{"connection types of the same service", types.ConnectionType{Service: account}, types.ConnectionType{Service: account}, true},
		{"connection types of services with the same name and methods", types.ConnectionType{Service: account},
			types.ConnectionType{Service: accountServiceType()}, false},
		{"session types of the same service", types.SessionType{Service: account}, types.SessionType{Service: account}, true},
		{"session types of services with the same name and methods", types.SessionType{Service: account},
			types.SessionType{Service: accountServiceType()}, false},
		{"connection and session types of the same service", types.ConnectionType{Service: account}, types.SessionType{Service: account}, false},
		{"named connection types with the same service and position", namedConnectionType("Account", 1, 8),
			namedConnectionType("Account", 1, 8), true},
		{"named connection types in different positions", namedConnectionType("Account", 1, 8),
			namedConnectionType("Account", 2, 8), false},
		{"named connection and session types with the same service and position", namedConnectionType("Account", 1, 8),
			namedSessionType("Account", 1, 8), false},
		{"named connection type and its connection type", namedConnectionType("Account", 1, 8), types.ConnectionType{Service: account}, false},
		{"named type and connection type of the same service", types.NamedType{Name: "Account"}, types.ConnectionType{Service: account}, false},
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

func TestServiceTypeMethod(t *testing.T) {
	account := accountServiceType()
	testCases := []serviceTypeMethodTestCase{
		{"method without parameters", account, "balance", account.Methods[0].Signature},
		{"method with parameters", account, "deposit", account.Methods[1].Signature},
		{"method without return value", account, "reset", account.Methods[2].Signature},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := testCase.serviceType.Method(testCase.method)

			if !ok {
				t.Fatalf("(%v).Method(%q) not found; want %v", testCase.serviceType.Name, testCase.method, testCase.expected)
			}
			if !result.Equals(testCase.expected) {
				t.Errorf("(%v).Method(%q) = %v; want %v", testCase.serviceType.Name, testCase.method, result, testCase.expected)
			}
		})
	}
}

func TestServiceTypeMissingMethod(t *testing.T) {
	testCases := []serviceTypeMissingMethodTestCase{
		{"unknown method", accountServiceType(), "withdraw"},
		{"method names are case sensitive", accountServiceType(), "Balance"},
		{"service name is not a method", accountServiceType(), "Account"},
		{"service without methods", types.NewServiceType("Empty", nil), "balance"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := testCase.serviceType.Method(testCase.method)

			if ok {
				t.Errorf("(%v).Method(%q) = %v; want not found", testCase.serviceType.Name, testCase.method, result)
			}
		})
	}
}

func TestNewServiceTypeIdentity(t *testing.T) {
	account := accountServiceType()

	if account == accountServiceType() {
		t.Errorf("two NewServiceType calls with the same name and methods returned the same *ServiceType")
	}
	if account.Name != "Account" || len(account.Methods) != 3 {
		t.Errorf("NewServiceType() = %v with %d methods; want Account with 3 methods", account.Name, len(account.Methods))
	}
}

func TestSignatureEquality(t *testing.T) {
	point := puntoType()
	list := listType()
	account := accountServiceType()
	testCases := []signatureEqualityTestCase{
		{"same primitive types", types.Signature{Params: []types.Type{types.Int, types.Str}, Return: types.Bool},
			types.Signature{Params: []types.Type{types.Int, types.Str}, Return: types.Bool}, true},
		{"no parameters", types.Signature{Params: []types.Type{}, Return: types.Void},
			types.Signature{Params: []types.Type{}, Return: types.Void}, true},
		{"nil and empty parameters", types.Signature{Params: nil, Return: types.Int},
			types.Signature{Params: []types.Type{}, Return: types.Int}, true},
		{"same struct and enum types", types.Signature{Params: []types.Type{point}, Return: list},
			types.Signature{Params: []types.Type{point}, Return: list}, true},
		{"different number of parameters", types.Signature{Params: []types.Type{types.Int}, Return: types.Int},
			types.Signature{Params: []types.Type{types.Int, types.Int}, Return: types.Int}, false},
		{"different parameter type", types.Signature{Params: []types.Type{types.Int}, Return: types.Int},
			types.Signature{Params: []types.Type{types.Float}, Return: types.Int}, false},
		{"same parameter types in different order", types.Signature{Params: []types.Type{types.Int, types.Str}, Return: types.Int},
			types.Signature{Params: []types.Type{types.Str, types.Int}, Return: types.Int}, false},
		{"different return type", types.Signature{Params: []types.Type{}, Return: types.Int},
			types.Signature{Params: []types.Type{}, Return: types.Float}, false},
		{"void and non void return", types.Signature{Params: []types.Type{}, Return: types.Void},
			types.Signature{Params: []types.Type{}, Return: types.Int}, false},
		{"struct declarations with the same name and fields", types.Signature{Params: []types.Type{point}, Return: types.Void},
			types.Signature{Params: []types.Type{puntoType()}, Return: types.Void}, false},
		{"enum declarations with the same name and variants", types.Signature{Params: []types.Type{}, Return: list},
			types.Signature{Params: []types.Type{}, Return: listType()}, false},
		{"named type and its struct type", types.Signature{Params: []types.Type{types.NamedType{Name: "Punto"}}, Return: types.Void},
			types.Signature{Params: []types.Type{point}, Return: types.Void}, false},
		{"connection types of the same service", types.Signature{Params: []types.Type{types.ConnectionType{Service: account}}, Return: types.Void},
			types.Signature{Params: []types.Type{types.ConnectionType{Service: account}}, Return: types.Void}, true},
		{"connection types of different services", types.Signature{Params: []types.Type{}, Return: types.ConnectionType{Service: account}},
			types.Signature{Params: []types.Type{}, Return: types.ConnectionType{Service: accountServiceType()}}, false},
		{"connection and session types", types.Signature{Params: []types.Type{types.SessionType{Service: account}}, Return: types.Void},
			types.Signature{Params: []types.Type{types.ConnectionType{Service: account}}, Return: types.Void}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left.Equals(testCase.right)
			reversed := testCase.right.Equals(testCase.left)

			if result != testCase.expected || reversed != testCase.expected {
				t.Errorf("(%v).Equals(%v) = %v, reversed = %v; want %v", testCase.left, testCase.right, result, reversed, testCase.expected)
			}
		})
	}
}

func TestSessionErrorType(t *testing.T) {
	message := []types.StructField{{Name: "message", Type: types.Str}}
	expected := []types.EnumVariant{
		{Name: "ConnectionFailed", Fields: message},
		{Name: "ConnectionLost", Fields: message},
		{Name: "Timeout"},
		{Name: "ProtocolError", Fields: message},
		{Name: "ServerError", Fields: message},
	}

	variants := types.SessionError.Variants
	if len(variants) != len(expected) {
		t.Fatalf("SessionError has %d variants; want %d", len(variants), len(expected))
	}
	for i, variant := range variants {
		if variant.Name != expected[i].Name || !slices.Equal(variant.Fields, expected[i].Fields) {
			t.Errorf("SessionError.Variants[%d] = %v; want %v", i, variant, expected[i])
		}
	}
}
