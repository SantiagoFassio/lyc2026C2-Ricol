package types_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type enumRepresentationTestCase struct {
	name     string
	value    types.Enum
	expected string
}

type enumFieldsTestCase struct {
	name     string
	value    types.Enum
	expected []types.Value
}

type enumEqualityTestCase struct {
	name     string
	left     types.Enum
	right    types.Enum
	expected bool
}

func enumValue(name string, variant string, fields ...types.Value) types.Enum {
	return types.NewEnum(name, variant, fields)
}

func emptyList() types.Enum {
	return enumValue("List", "Empty")
}

func node(value int64, rest types.Enum) types.Enum {
	return enumValue("List", "Node", integer(value), rest)
}

func TestEnumRepresentation(t *testing.T) {
	testCases := []enumRepresentationTestCase{
		{"variant without fields", enumValue("Color", "Red"), "Color.Red"},
		{"variant with one field", enumValue("R", "InsufficientFunds", integer(10)), "R.InsufficientFunds(10)"},
		{"variant with several fields", enumValue("D", "All", integer(-3), float(2.5), booleanValue(true)), "D.All(-3, 2.5, True)"},
		{"string field is quoted", enumValue("R", "Error", stringValue("a\n\"b\"")), `R.Error("a\n\"b\"")`},
		{"struct field", enumValue("R", "Ok", punto(1, 2)), "R.Ok(Punto { x: 1, y: 2 })"},
		{"recursive enum", node(1, node(2, emptyList())), "List.Node(1, List.Node(2, List.Empty))"},
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

func TestEnumDisplay(t *testing.T) {
	testCases := []enumRepresentationTestCase{
		{"variant without fields", enumValue("Color", "Red"), "Color.Red"},
		{"string field stays quoted", enumValue("R", "Error", stringValue("Ann")), `R.Error("Ann")`},
		{"recursive enum", node(1, emptyList()), "List.Node(1, List.Empty)"},
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

func TestEnumVariant(t *testing.T) {
	testCases := []enumRepresentationTestCase{
		{"variant without fields", emptyList(), "Empty"},
		{"variant with fields", node(1, emptyList()), "Node"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.Variant()

			if result != testCase.expected {
				t.Errorf("(%v).Variant() = %q; want %q", testCase.value, result, testCase.expected)
			}
		})
	}
}

func TestEnumFields(t *testing.T) {
	testCases := []enumFieldsTestCase{
		{"variant without fields", emptyList(), []types.Value{}},
		{"variant with one field", enumValue("R", "Error", stringValue("x")), []types.Value{stringValue("x")}},
		{"fields keep their order", enumValue("D", "Pair", integer(2), integer(1)), []types.Value{integer(2), integer(1)}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.Fields()

			if len(result) != len(testCase.expected) {
				t.Fatalf("(%v).Fields() = %v; want %v", testCase.value, result, testCase.expected)
			}
			for i := range result {
				if result[i] != testCase.expected[i] {
					t.Errorf("(%v).Fields()[%d] = %v; want %v", testCase.value, i, result[i], testCase.expected[i])
				}
			}
		})
	}
}

func TestEnumNestedFields(t *testing.T) {
	rest := node(2, emptyList())
	value := node(1, rest)

	result := value.Fields()

	nested, isEnum := result[1].(types.Enum)
	if !isEnum || !nested.Equals(rest) {
		t.Errorf("(%v).Fields()[1] = %v; want %v", value, result[1], rest)
	}
}

func TestEnumEquals(t *testing.T) {
	sameValue := node(1, emptyList())
	testCases := []enumEqualityTestCase{
		{"same value", sameValue, sameValue, true},
		{"same variant without fields", emptyList(), emptyList(), true},
		{"equal fields", enumValue("R", "Ok", integer(1)), enumValue("R", "Ok", integer(1)), true},
		{"different fields", enumValue("R", "Ok", integer(1)), enumValue("R", "Ok", integer(2)), false},
		{"different variants with equal fields", enumValue("R", "A", integer(1)), enumValue("R", "B", integer(1)), false},
		{"variant with and without fields", emptyList(), node(1, emptyList()), false},
		{"swapped values", enumValue("D", "Pair", integer(1), integer(2)), enumValue("D", "Pair", integer(2), integer(1)), false},
		{"equal strings", enumValue("R", "Error", stringValue("x")), enumValue("R", "Error", stringValue("x")), true},
		{"different strings", enumValue("R", "Error", stringValue("x")), enumValue("R", "Error", stringValue("X")), false},
		{"different booleans", enumValue("R", "Flag", booleanValue(true)), enumValue("R", "Flag", booleanValue(false)), false},
		{"equal floats", enumValue("R", "Measure", float(2.5)), enumValue("R", "Measure", float(2.5)), true},
		{"equal struct fields", enumValue("R", "Ok", punto(1, 2)), enumValue("R", "Ok", punto(1, 2)), true},
		{"different struct fields", enumValue("R", "Ok", punto(1, 2)), enumValue("R", "Ok", punto(1, 3)), false},
		{"equal recursive enums", node(1, node(2, emptyList())), node(1, node(2, emptyList())), true},
		{"recursive enums with a different deep field", node(1, node(2, emptyList())), node(1, node(3, emptyList())), false},
		{"recursive enums of different length", node(1, emptyList()), node(1, node(2, emptyList())), false},
		{"different enum names", enumValue("A", "X", integer(1)), enumValue("B", "X", integer(1)), false},
		{"fields of different types", enumValue("R", "Ok", integer(1)), enumValue("R", "Ok", stringValue("1")), false},
		{"enum and struct fields", enumValue("R", "Ok", emptyList()), enumValue("R", "Ok", structValue("List")), false},
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

func TestStructWithEnumFieldEquals(t *testing.T) {
	testCases := []structEqualityTestCase{
		{"equal enum fields", structValue("Box", field("list", node(1, emptyList()))), structValue("Box", field("list", node(1, emptyList()))), true},
		{"different enum fields", structValue("Box", field("list", node(1, emptyList()))), structValue("Box", field("list", emptyList())), false},
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

func TestEnumKeepsACopyOfItsFields(t *testing.T) {
	fields := []types.Value{integer(1)}
	value := types.NewEnum("R", "Ok", fields)

	fields[0] = integer(2)

	if result := value.Fields()[0]; result != integer(1) {
		t.Errorf("modifying the fields passed to NewEnum changed the enum: %v; want R.Ok(1)", value)
	}
}

func TestEnumFieldsReturnsACopy(t *testing.T) {
	value := types.NewEnum("R", "Ok", []types.Value{integer(1)})

	value.Fields()[0] = integer(2)

	if result := value.Fields()[0]; result != integer(1) {
		t.Errorf("modifying the fields returned by Fields changed the enum: %v; want R.Ok(1)", value)
	}
}
