package types_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type structRepresentationTestCase struct {
	name     string
	value    types.Struct
	expected string
}

type structFieldTestCase struct {
	name      string
	value     types.Struct
	fieldName string
	expected  types.Value
}

type structMissingFieldTestCase struct {
	name      string
	value     types.Struct
	fieldName string
}

type structEqualityTestCase struct {
	name     string
	left     types.Struct
	right    types.Struct
	expected bool
}

func field(name string, value types.Value) types.FieldValue {
	return types.FieldValue{Name: name, Value: value}
}

func structValue(name string, fields ...types.FieldValue) types.Struct {
	return types.NewStruct(name, fields)
}

func punto(x int64, y int64) types.Struct {
	return structValue("Punto", field("x", integer(x)), field("y", integer(y)))
}

func persona(nombre string) types.Struct {
	return structValue("Persona", field("nombre", stringValue(nombre)))
}

func linea(desde types.Struct, hasta types.Struct) types.Struct {
	return structValue("Linea", field("desde", desde), field("hasta", hasta))
}

func TestStructRepresentation(t *testing.T) {
	testCases := []structRepresentationTestCase{
		{"one field", structValue("Edad", field("valor", integer(30))), "Edad { valor: 30 }"},
		{"several fields", punto(1, 2), "Punto { x: 1, y: 2 }"},
		{"negative and float fields", structValue("Medida", field("delta", integer(-3)), field("valor", float(2.5))),
			"Medida { delta: -3, valor: 2.5 }"},
		{"string field is quoted", persona("Ana"), `Persona { nombre: "Ana" }`},
		{"string field with escape sequences", persona("a\n\"b\""), `Persona { nombre: "a\n\"b\"" }`},
		{"boolean field", structValue("Flag", field("activo", booleanValue(true))), "Flag { activo: True }"},
		{"nested structs", linea(punto(0, 0), punto(1, 2)), "Linea { desde: Punto { x: 0, y: 0 }, hasta: Punto { x: 1, y: 2 } }"},
		{"empty struct", structValue("Vacio"), "Vacio {}"},
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

func TestStructDisplay(t *testing.T) {
	testCases := []structRepresentationTestCase{
		{"several fields", punto(1, 2), "Punto { x: 1, y: 2 }"},
		{"string field stays quoted", persona("Ana"), `Persona { nombre: "Ana" }`},
		{"nested structs", linea(punto(0, 0), punto(1, 2)), "Linea { desde: Punto { x: 0, y: 0 }, hasta: Punto { x: 1, y: 2 } }"},
		{"empty struct", structValue("Vacio"), "Vacio {}"},
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

func TestStructField(t *testing.T) {
	testCases := []structFieldTestCase{
		{"first field", punto(1, 2), "x", integer(1)},
		{"last field", punto(1, 2), "y", integer(2)},
		{"string field", persona("Ana"), "nombre", stringValue("Ana")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := testCase.value.Field(testCase.fieldName)

			if !ok {
				t.Fatalf("(%v).Field(%q) not found; want %v", testCase.value, testCase.fieldName, testCase.expected)
			}
			if result != testCase.expected {
				t.Errorf("(%v).Field(%q) = %v; want %v", testCase.value, testCase.fieldName, result, testCase.expected)
			}
		})
	}
}

func TestStructNestedField(t *testing.T) {
	desde := punto(0, 0)
	value := linea(desde, punto(1, 2))

	result, ok := value.Field("desde")

	if !ok {
		t.Fatalf("(%v).Field(%q) not found; want %v", value, "desde", desde)
	}
	nested, isStruct := result.(types.Struct)
	if !isStruct || !nested.Equals(desde) {
		t.Errorf("(%v).Field(%q) = %v; want %v", value, "desde", result, desde)
	}
}

func TestStructMissingField(t *testing.T) {
	testCases := []structMissingFieldTestCase{
		{"unknown field", punto(1, 2), "z"},
		{"field names are case sensitive", punto(1, 2), "X"},
		{"empty struct", structValue("Vacio"), "x"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := testCase.value.Field(testCase.fieldName)

			if ok {
				t.Errorf("(%v).Field(%q) = %v; want not found", testCase.value, testCase.fieldName, result)
			}
		})
	}
}

func TestStructEquals(t *testing.T) {
	sameValue := punto(1, 2)
	testCases := []structEqualityTestCase{
		{"same value", sameValue, sameValue, true},
		{"equal fields", punto(1, 2), punto(1, 2), true},
		{"different first field", punto(1, 2), punto(3, 2), false},
		{"different last field", punto(1, 2), punto(1, 3), false},
		{"swapped values", punto(1, 2), punto(2, 1), false},
		{"equal strings", persona("Ana"), persona("Ana"), true},
		{"different strings", persona("Ana"), persona("ana"), false},
		{"different booleans", structValue("Flag", field("activo", booleanValue(true))),
			structValue("Flag", field("activo", booleanValue(false))), false},
		{"equal floats", structValue("Medida", field("valor", float(2.5))), structValue("Medida", field("valor", float(2.5))), true},
		{"nested structs with equal fields", linea(punto(0, 0), punto(1, 2)), linea(punto(0, 0), punto(1, 2)), true},
		{"nested structs with different fields", linea(punto(0, 0), punto(1, 2)), linea(punto(0, 0), punto(1, 3)), false},
		{"empty structs", structValue("Vacio"), structValue("Vacio"), true},
		{"different struct names", structValue("A", field("x", integer(1))), structValue("B", field("x", integer(1))), false},
		{"fields of different types", structValue("A", field("x", booleanValue(true))), structValue("A", field("x", stringValue("True"))), false},
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

func TestStructKeepsACopyOfItsFields(t *testing.T) {
	fields := []types.FieldValue{field("x", integer(1))}
	value := types.NewStruct("Punto", fields)

	fields[0] = field("x", integer(2))

	result, _ := value.Field("x")
	if result != integer(1) {
		t.Errorf("modifying the fields passed to NewStruct changed the struct: %v; want x = 1", value)
	}
}
