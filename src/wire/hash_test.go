package wire

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type canonicalFormTestCase struct {
	name     string
	service  *types.ServiceType
	expected []string
}

type serviceChange func(service *types.ServiceType)

type hashChangeTestCase struct {
	name        string
	before      serviceChange
	after       serviceChange
	changesHash bool
}

type invalidServiceTestCase struct {
	name    string
	service *types.ServiceType
}

const accountHash = "c586622ea17b6f2c4b7d9f56c6b5cc9b94507cf19000c78065ce408839bd8a46"

func method(name string, returnType types.Type, params ...types.Type) types.ServiceMethod {
	return types.ServiceMethod{Name: name, Signature: types.Signature{Params: params, Return: returnType}}
}

func service(name string, methods ...types.ServiceMethod) *types.ServiceType {
	return types.NewServiceType(name, methods)
}

func accountService() *types.ServiceType {
	movementType := types.NewStructType("Movement", []types.StructField{
		{Name: "amount", Type: types.Int},
		{Name: "finalBalance", Type: types.Int},
	})
	withdrawResultType := enumType("WithdrawResult",
		types.EnumVariant{Name: "Ok", Fields: []types.StructField{{Name: "movement", Type: movementType}}},
		types.EnumVariant{Name: "InsufficientFunds", Fields: []types.StructField{{Name: "available", Type: types.Int}}},
	)
	return service("Account",
		method("balance", types.Int),
		method("deposit", movementType, types.Int),
		method("withdraw", withdrawResultType, types.Int),
		method("reset", types.Void),
	)
}

func movementOf(account *types.ServiceType) *types.StructType {
	return account.Methods[1].Signature.Return.(*types.StructType)
}

func withdrawResultOf(account *types.ServiceType) *types.EnumType {
	return account.Methods[2].Signature.Return.(*types.EnumType)
}

func accountHashWith(change serviceChange) [32]byte {
	account := accountService()
	if change != nil {
		change(account)
	}
	return hashOf(canonicalForm(account))
}

func TestCanonicalForm(t *testing.T) {
	testCases := []canonicalFormTestCase{
		{"account", accountService(), []string{
			"ricol-interface-v1",
			"service Account",
			"method balance() -> Int",
			"method deposit(Int) -> Movement",
			"method withdraw(Int) -> WithdrawResult",
			"method reset() -> Void",
			"struct Movement { amount: Int; finalBalance: Int; }",
			"enum WithdrawResult { Ok(movement: Movement); InsufficientFunds(available: Int); }",
		}},
		{"recursive enum", service("Stack",
			method("push", listType, types.Int),
			method("sum", types.Int, listType),
		), []string{
			"ricol-interface-v1",
			"service Stack",
			"method push(Int) -> List",
			"method sum(List) -> Int",
			"enum List { Empty; Node(value: Int, rest: List); }",
		}},
		{"nested types in depth-first order", service("Canvas",
			method("draw", types.Bool, drawingType, segmentType),
			method("clear", types.Void),
		), []string{
			"ricol-interface-v1",
			"service Canvas",
			"method draw(Drawing, Segment) -> Bool",
			"method clear() -> Void",
			"struct Drawing { shape: Shape; layer: Int; }",
			"enum Shape { Blank; Circle(center: Point, radius: Float); Text(content: String, bold: Bool); }",
			"struct Point { x: Int; y: Int; }",
			"struct Segment { start: Point; end: Point; label: String; }",
		}},
		{"parameters before the return type", service("Geometry",
			method("move", pointType, segmentType, types.Float, types.Str),
		), []string{
			"ricol-interface-v1",
			"service Geometry",
			"method move(Segment, Float, String) -> Point",
			"struct Segment { start: Point; end: Point; label: String; }",
			"struct Point { x: Int; y: Int; }",
		}},
		{"type used by several methods", service("Map",
			method("origin", pointType),
			method("distance", types.Float, pointType, pointType),
		), []string{
			"ricol-interface-v1",
			"service Map",
			"method origin() -> Point",
			"method distance(Point, Point) -> Float",
			"struct Point { x: Int; y: Int; }",
		}},
		{"struct without fields", service("Ping", method("ping", nothingType, nothingType)), []string{
			"ricol-interface-v1",
			"service Ping",
			"method ping(Nothing) -> Nothing",
			"struct Nothing { }",
		}},
		{"service without methods", service("Idle"), []string{
			"ricol-interface-v1",
			"service Idle",
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := canonicalForm(testCase.service)

			if expected := strings.Join(testCase.expected, "\n"); result != expected {
				t.Errorf("canonicalForm() =\n%s\nwant\n%s", result, expected)
			}
		})
	}
}

func TestAccountHash(t *testing.T) {
	result := hashOf(canonicalForm(accountService()))

	if encoded := hex.EncodeToString(result[:]); encoded != accountHash {
		t.Errorf("hashOf(canonicalForm(Account)) = %s; want %s", encoded, accountHash)
	}
}

func TestWhatChangesTheHash(t *testing.T) {
	testCases := []hashChangeTestCase{
		{"rebuild the same service", nil, nil, false},
		{"add a method", nil, func(account *types.ServiceType) {
			account.Methods = append(account.Methods, method("close", types.Void))
		}, true},
		{"remove a method", nil, func(account *types.ServiceType) {
			account.Methods = account.Methods[:3]
		}, true},
		{"reorder methods", nil, func(account *types.ServiceType) {
			account.Methods[0], account.Methods[3] = account.Methods[3], account.Methods[0]
		}, true},
		{"add a parameter", nil, func(account *types.ServiceType) {
			account.Methods[1].Signature.Params = append(account.Methods[1].Signature.Params, types.Str)
		}, true},
		{"remove a parameter", nil, func(account *types.ServiceType) {
			account.Methods[1].Signature.Params = nil
		}, true},
		{"reorder parameters", func(account *types.ServiceType) {
			account.Methods[1].Signature.Params = []types.Type{types.Int, types.Str}
		}, func(account *types.ServiceType) {
			account.Methods[1].Signature.Params = []types.Type{types.Str, types.Int}
		}, true},
		{"change a parameter type", nil, func(account *types.ServiceType) {
			account.Methods[1].Signature.Params[0] = types.Float
		}, true},
		{"change a return type", nil, func(account *types.ServiceType) {
			account.Methods[0].Signature.Return = types.Float
		}, true},
		{"add a field", nil, func(account *types.ServiceType) {
			movement := movementOf(account)
			movement.Fields = append(movement.Fields, types.StructField{Name: "note", Type: types.Str})
		}, true},
		{"remove a field", nil, func(account *types.ServiceType) {
			movement := movementOf(account)
			movement.Fields = movement.Fields[:1]
		}, true},
		{"reorder fields", nil, func(account *types.ServiceType) {
			movement := movementOf(account)
			movement.Fields[0], movement.Fields[1] = movement.Fields[1], movement.Fields[0]
		}, true},
		{"change a field type", nil, func(account *types.ServiceType) {
			movementOf(account).Fields[0].Type = types.Float
		}, true},
		{"add a variant", nil, func(account *types.ServiceType) {
			withdrawResult := withdrawResultOf(account)
			withdrawResult.Variants = append(withdrawResult.Variants, types.EnumVariant{Name: "Locked"})
		}, true},
		{"remove a variant", nil, func(account *types.ServiceType) {
			withdrawResult := withdrawResultOf(account)
			withdrawResult.Variants = withdrawResult.Variants[:1]
		}, true},
		{"reorder variants", nil, func(account *types.ServiceType) {
			variants := withdrawResultOf(account).Variants
			variants[0], variants[1] = variants[1], variants[0]
		}, true},
		{"add a field to a variant", nil, func(account *types.ServiceType) {
			variant := &withdrawResultOf(account).Variants[1]
			variant.Fields = append(variant.Fields, types.StructField{Name: "missing", Type: types.Int})
		}, true},
		{"change the type of a variant field", nil, func(account *types.ServiceType) {
			withdrawResultOf(account).Variants[1].Fields[0].Type = types.Float
		}, true},
		{"rename the service", nil, func(account *types.ServiceType) {
			account.Name = "Wallet"
		}, true},
		{"rename a method", nil, func(account *types.ServiceType) {
			account.Methods[1].Name = "put"
		}, true},
		{"rename a struct", nil, func(account *types.ServiceType) {
			movementOf(account).Name = "Transfer"
		}, true},
		{"rename a struct field", nil, func(account *types.ServiceType) {
			movementOf(account).Fields[0].Name = "value"
		}, true},
		{"rename an enum", nil, func(account *types.ServiceType) {
			withdrawResultOf(account).Name = "Withdrawal"
		}, true},
		{"rename a variant", nil, func(account *types.ServiceType) {
			withdrawResultOf(account).Variants[0].Name = "Done"
		}, true},
		{"rename a variant field", nil, func(account *types.ServiceType) {
			withdrawResultOf(account).Variants[1].Fields[0].Name = "balance"
		}, true},
		{"turn a struct into an enum with the same name", nil, func(account *types.ServiceType) {
			account.Methods[1].Signature.Return = enumType("Movement")
			withdrawResultOf(account).Variants[0].Fields[0].Type = account.Methods[1].Signature.Return
		}, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			before := accountHashWith(testCase.before)
			after := accountHashWith(testCase.after)

			if changed := before != after; changed != testCase.changesHash {
				t.Errorf("hash changed = %v; want %v", changed, testCase.changesHash)
			}
		})
	}
}

func TestCanonicalFormWithInvalidServicePanics(t *testing.T) {
	otherPointType := types.NewStructType("Point", []types.StructField{
		{Name: "x", Type: types.Int},
		{Name: "y", Type: types.Int},
	})
	testCases := []invalidServiceTestCase{
		{"two structs with the same name", service("Map",
			method("move", otherPointType, pointType),
		)},
		{"two types with the same name in nested fields", service("Map",
			method("draw", types.Void, segmentType, types.NewStructType("Wrapper", []types.StructField{
				{Name: "point", Type: otherPointType},
			})),
		)},
		{"a struct and an enum with the same name", service("Map",
			method("draw", enumType("Point"), pointType),
		)},
		{"invalid parameter type", service("Map", method("draw", types.Void, types.Invalid))},
		{"unresolved return type", service("Map", method("draw", types.NamedType{Name: "Point"}))},
		{"missing field type", service("Map", method("draw", types.NewStructType("Broken", []types.StructField{
			{Name: "field", Type: nil},
		})))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertPanics(t, "canonicalForm", func() { canonicalForm(testCase.service) })
		})
	}
}
