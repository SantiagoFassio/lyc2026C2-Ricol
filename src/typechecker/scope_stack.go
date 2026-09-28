package typechecker

import "github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"

type scopeStack struct {
	scopes []map[string]types.Type
}

func newScopeStack() *scopeStack {
	return &scopeStack{
		scopes: []map[string]types.Type{{}},
	}
}

func (s *scopeStack) push() {
	s.scopes = append(s.scopes, map[string]types.Type{})
}

func (s *scopeStack) pop() {
	s.scopes = s.scopes[:len(s.scopes)-1]
}

func (s *scopeStack) declare(name string, varType types.Type) bool {
	currentScope := s.scopes[len(s.scopes)-1]
	if _, ok := currentScope[name]; ok {
		return false
	}
	currentScope[name] = varType
	return true
}

func (s *scopeStack) lookup(name string) (types.Type, int, bool) {
	for i := len(s.scopes) - 1; i >= 0; i-- {
		if varType, ok := s.scopes[i][name]; ok {
			return varType, len(s.scopes) - 1 - i, true
		}
	}
	return nil, 0, false
}
