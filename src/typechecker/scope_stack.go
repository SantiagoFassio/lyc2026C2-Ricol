package typechecker

import "github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"

type symbol struct {
	varType   types.Type
	signature *types.Signature
}

type scopeStack struct {
	scopes []map[string]symbol
}

func newScopeStack() *scopeStack {
	return &scopeStack{
		scopes: []map[string]symbol{{}},
	}
}

func (s *scopeStack) push() {
	s.scopes = append(s.scopes, map[string]symbol{})
}

func (s *scopeStack) pop() {
	s.scopes = s.scopes[:len(s.scopes)-1]
}

func (s *scopeStack) declareVariable(name string, varType types.Type) bool {
	return s.declare(name, symbol{varType: varType})
}

func (s *scopeStack) declareFunction(name string, signature types.Signature) bool {
	return s.declare(name, symbol{signature: &signature})
}

func (s *scopeStack) declare(name string, declared symbol) bool {
	currentScope := s.scopes[len(s.scopes)-1]
	if _, ok := currentScope[name]; ok {
		return false
	}
	currentScope[name] = declared
	return true
}

func (s *scopeStack) lookup(name string) (symbol, int, bool) {
	for i := len(s.scopes) - 1; i >= 0; i-- {
		if declared, ok := s.scopes[i][name]; ok {
			return declared, len(s.scopes) - 1 - i, true
		}
	}
	return symbol{}, 0, false
}
