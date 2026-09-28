package interpreter

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type environment struct {
	values    map[string]types.Value
	enclosing *environment
}

func newEnvironment(enclosing *environment) *environment {
	return &environment{
		values:    map[string]types.Value{},
		enclosing: enclosing,
	}
}

func (e *environment) define(name string, value types.Value) {
	e.values[name] = value
}

func (e *environment) get(name string, distance int) types.Value {
	scope := e.getAncestorScope(distance)
	value, ok := scope.values[name]
	if ok {
		return value
	}
	panic(fmt.Sprintf("Undefined variable '%s'", name))
}

func (e *environment) assign(name string, value types.Value, distance int) types.Value {
	scope := e.getAncestorScope(distance)
	if _, ok := scope.values[name]; ok {
		scope.values[name] = value
		return value
	}
	panic(fmt.Sprintf("Cannot assign value to undefined variable '%s'", name))
}

func (e *environment) getAncestorScope(distance int) *environment {
	currentEnvironment := e
	for range distance {
		if currentEnvironment.enclosing == nil {
			panic(fmt.Sprintf("No enclosing environment was found for environment at distance '%d'", distance))
		}
		currentEnvironment = currentEnvironment.enclosing
	}
	return currentEnvironment
}
