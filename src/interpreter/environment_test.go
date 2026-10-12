package interpreter

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type environmentPanicTestCase struct {
	name     string
	function func(global *environment, inner *environment)
}

func assertGet(t *testing.T, env *environment, name string, distance int, expected types.Value) {
	t.Helper()

	if value := env.get(name, distance); value != expected {
		t.Errorf("get(%q, %d) = %#v; want %#v", name, distance, value, expected)
	}
}

func assertPanics(t *testing.T, description string, function func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Errorf("%s did not panic", description)
		}
	}()

	function()
}

func TestEnvironmentDefineAndGet(t *testing.T) {
	env := newEnvironment(nil)

	env.define("x", types.NewInteger(1))
	env.define("y", types.NewString("a"))

	assertGet(t, env, "x", 0, types.NewInteger(1))
	assertGet(t, env, "y", 0, types.NewString("a"))
}

func TestEnvironmentDefineOverwritesInTheSameScope(t *testing.T) {
	env := newEnvironment(nil)

	env.define("x", types.NewInteger(1))
	env.define("x", types.NewBoolean(true))

	assertGet(t, env, "x", 0, types.NewBoolean(true))
}

func TestEnvironmentGetFromEnclosingScopes(t *testing.T) {
	global := newEnvironment(nil)
	global.define("x", types.NewInteger(1))
	middle := newEnvironment(global)
	middle.define("y", types.NewInteger(2))
	inner := newEnvironment(middle)
	inner.define("z", types.NewInteger(3))

	assertGet(t, inner, "x", 2, types.NewInteger(1))
	assertGet(t, inner, "y", 1, types.NewInteger(2))
	assertGet(t, inner, "z", 0, types.NewInteger(3))
	assertGet(t, middle, "x", 1, types.NewInteger(1))
}

func TestEnvironmentShadowing(t *testing.T) {
	global := newEnvironment(nil)
	global.define("x", types.NewString("outer"))
	inner := newEnvironment(global)
	inner.define("x", types.NewString("inner"))

	assertGet(t, inner, "x", 0, types.NewString("inner"))
	assertGet(t, inner, "x", 1, types.NewString("outer"))
	assertGet(t, global, "x", 0, types.NewString("outer"))
}

func TestEnvironmentAssign(t *testing.T) {
	env := newEnvironment(nil)
	env.define("x", types.NewInteger(1))

	if value := env.assign("x", types.NewInteger(2), 0); value != types.NewInteger(2) {
		t.Errorf("assign(%q, 2, 0) = %#v; want %#v", "x", value, types.NewInteger(2))
	}
	assertGet(t, env, "x", 0, types.NewInteger(2))
}

func TestEnvironmentAssignToEnclosingScope(t *testing.T) {
	global := newEnvironment(nil)
	global.define("x", types.NewInteger(1))
	inner := newEnvironment(global)

	inner.assign("x", types.NewInteger(2), 1)

	assertGet(t, global, "x", 0, types.NewInteger(2))
	assertGet(t, inner, "x", 1, types.NewInteger(2))
}

func TestEnvironmentAssignOnlyChangesTheScopeAtTheDistance(t *testing.T) {
	global := newEnvironment(nil)
	global.define("x", types.NewString("outer"))
	inner := newEnvironment(global)
	inner.define("x", types.NewString("inner"))

	inner.assign("x", types.NewString("new inner"), 0)
	assertGet(t, inner, "x", 0, types.NewString("new inner"))
	assertGet(t, global, "x", 0, types.NewString("outer"))

	inner.assign("x", types.NewString("new outer"), 1)
	assertGet(t, inner, "x", 0, types.NewString("new inner"))
	assertGet(t, global, "x", 0, types.NewString("new outer"))
}

func TestEnvironmentPanics(t *testing.T) {
	testCases := []struct {
		name     string
		function func(global *environment, inner *environment)
	}{
		{"get an undefined variable", func(global *environment, inner *environment) { global.get("y", 0) }},
		{"get a variable from the wrong scope", func(global *environment, inner *environment) { inner.get("x", 0) }},
		{"get beyond the global scope", func(global *environment, inner *environment) { inner.get("x", 2) }},
		{"assign an undefined variable", func(global *environment, inner *environment) { global.assign("y", types.NewInteger(2), 0) }},
		{
			"assign a variable in the wrong scope",
			func(global *environment, inner *environment) { inner.assign("x", types.NewInteger(2), 0) },
		},
		{
			"assign beyond the global scope",
			func(global *environment, inner *environment) { inner.assign("x", types.NewInteger(2), 2) },
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			global := newEnvironment(nil)
			global.define("x", types.NewInteger(1))
			inner := newEnvironment(global)

			assertPanics(t, testCase.name, func() { testCase.function(global, inner) })
		})
	}
}

func funcDeclaration(name string) *common.FuncDeclarationStatement {
	return common.NewFuncDeclarationStatement(
		common.NewToken(common.FUNC, "func", common.Position{}),
		common.NewToken(common.IDENTIFIER, name, common.Position{}),
		[]common.Parameter{},
		types.Void,
		[]common.Statement{},
	)
}

func assertGetFunction(
	t *testing.T,
	env *environment,
	name string,
	distance int,
	expectedDeclaration *common.FuncDeclarationStatement,
	expectedScope *environment,
) {
	t.Helper()

	declaration, scope := env.getFunction(name, distance)
	if declaration != expectedDeclaration {
		t.Errorf("getFunction(%q, %d) declaration = %v; want %v", name, distance, declaration, expectedDeclaration)
	}
	if scope != expectedScope {
		t.Errorf("getFunction(%q, %d) returned the wrong scope", name, distance)
	}
}

func TestEnvironmentDefineAndGetFunction(t *testing.T) {
	env := newEnvironment(nil)
	f := funcDeclaration("f")
	env.defineFunction("f", f)

	assertGetFunction(t, env, "f", 0, f, env)
}

func TestEnvironmentGetFunctionFromEnclosingScopes(t *testing.T) {
	global := newEnvironment(nil)
	f := funcDeclaration("f")
	global.defineFunction("f", f)
	middle := newEnvironment(global)
	g := funcDeclaration("g")
	middle.defineFunction("g", g)
	inner := newEnvironment(middle)

	assertGetFunction(t, inner, "f", 2, f, global)
	assertGetFunction(t, inner, "g", 1, g, middle)
}

func TestEnvironmentFunctionShadowing(t *testing.T) {
	global := newEnvironment(nil)
	outer := funcDeclaration("f")
	global.defineFunction("f", outer)
	inner := newEnvironment(global)
	shadowing := funcDeclaration("f")
	inner.defineFunction("f", shadowing)

	assertGetFunction(t, inner, "f", 0, shadowing, inner)
	assertGetFunction(t, inner, "f", 1, outer, global)
}

func TestEnvironmentFunctionsAndVariablesAreSeparate(t *testing.T) {
	global := newEnvironment(nil)
	global.define("x", types.NewInteger(1))
	f := funcDeclaration("f")
	global.defineFunction("f", f)

	assertGet(t, global, "x", 0, types.NewInteger(1))
	assertGetFunction(t, global, "f", 0, f, global)
	assertPanics(t, "get a function as a variable", func() { global.get("f", 0) })
	assertPanics(t, "get a variable as a function", func() { global.getFunction("x", 0) })
}

func TestEnvironmentFunctionPanics(t *testing.T) {
	testCases := []struct {
		name     string
		function func(global *environment, inner *environment)
	}{
		{"get an undefined function", func(global *environment, inner *environment) { global.getFunction("g", 0) }},
		{"get a function from the wrong scope", func(global *environment, inner *environment) { inner.getFunction("f", 0) }},
		{"get a function beyond the global scope", func(global *environment, inner *environment) { inner.getFunction("f", 2) }},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			global := newEnvironment(nil)
			global.defineFunction("f", funcDeclaration("f"))
			inner := newEnvironment(global)

			assertPanics(t, testCase.name, func() { testCase.function(global, inner) })
		})
	}
}

func structDeclaration(name string) *common.StructDeclarationStatement {
	return common.NewStructDeclarationStatement(
		common.NewToken(common.STRUCT, "struct", common.Position{}),
		common.NewToken(common.IDENTIFIER, name, common.Position{}),
		[]common.FieldDeclaration{},
	)
}

func assertGetStruct(t *testing.T, env *environment, name string, distance int, expected *common.StructDeclarationStatement) {
	t.Helper()

	if declaration := env.getStruct(name, distance); declaration != expected {
		t.Errorf("getStruct(%q, %d) = %v; want %v", name, distance, declaration, expected)
	}
}

func TestEnvironmentDefineAndGetStruct(t *testing.T) {
	env := newEnvironment(nil)
	p := structDeclaration("P")
	env.defineStruct("P", p)

	assertGetStruct(t, env, "P", 0, p)
}

func TestEnvironmentGetStructFromEnclosingScopes(t *testing.T) {
	global := newEnvironment(nil)
	p := structDeclaration("P")
	global.defineStruct("P", p)
	middle := newEnvironment(global)
	q := structDeclaration("Q")
	middle.defineStruct("Q", q)
	inner := newEnvironment(middle)

	assertGetStruct(t, inner, "P", 2, p)
	assertGetStruct(t, inner, "Q", 1, q)
}

func TestEnvironmentStructShadowing(t *testing.T) {
	global := newEnvironment(nil)
	outer := structDeclaration("P")
	global.defineStruct("P", outer)
	inner := newEnvironment(global)
	shadowing := structDeclaration("P")
	inner.defineStruct("P", shadowing)

	assertGetStruct(t, inner, "P", 0, shadowing)
	assertGetStruct(t, inner, "P", 1, outer)
}

func TestEnvironmentStructsFunctionsAndVariablesAreSeparate(t *testing.T) {
	global := newEnvironment(nil)
	global.define("x", types.NewInteger(1))
	f := funcDeclaration("f")
	global.defineFunction("f", f)
	p := structDeclaration("P")
	global.defineStruct("P", p)

	assertGet(t, global, "x", 0, types.NewInteger(1))
	assertGetFunction(t, global, "f", 0, f, global)
	assertGetStruct(t, global, "P", 0, p)
	assertPanics(t, "get a struct as a variable", func() { global.get("P", 0) })
	assertPanics(t, "get a struct as a function", func() { global.getFunction("P", 0) })
	assertPanics(t, "get a variable as a struct", func() { global.getStruct("x", 0) })
	assertPanics(t, "get a function as a struct", func() { global.getStruct("f", 0) })
}

func TestEnvironmentStructPanics(t *testing.T) {
	testCases := []environmentPanicTestCase{
		{"get an undefined struct", func(global *environment, inner *environment) { global.getStruct("Q", 0) }},
		{"get a struct from the wrong scope", func(global *environment, inner *environment) { inner.getStruct("P", 0) }},
		{"get a struct beyond the global scope", func(global *environment, inner *environment) { inner.getStruct("P", 2) }},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			global := newEnvironment(nil)
			global.defineStruct("P", structDeclaration("P"))
			inner := newEnvironment(global)

			assertPanics(t, testCase.name, func() { testCase.function(global, inner) })
		})
	}
}
