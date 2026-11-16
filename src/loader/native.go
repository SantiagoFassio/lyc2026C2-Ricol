package loader

import (
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

var nativeModules = map[string]func() []common.Statement{
	"random": randomModule,
}

func randomModule() []common.Statement {
	return []common.Statement{
		nativeFunction("random", "randomInt", []nativeParameter{{"bound", types.Int}}, types.Int),
	}
}

type nativeParameter struct {
	name      string
	paramType types.Type
}

func nativeFunction(module string, name string, parameters []nativeParameter, returnType types.Type) *common.FuncDeclarationStatement {
	position := common.Position{File: module, Line: 1, Column: 1}
	declaredParameters := make([]common.Parameter, len(parameters))
	for i, parameter := range parameters {
		declaredParameters[i] = common.Parameter{
			NameToken: common.NewToken(common.IDENTIFIER, parameter.name, position),
			ParamType: parameter.paramType,
		}
	}
	return common.NewNativeFuncDeclarationStatement(
		common.NewToken(common.FUNC, "func", position),
		common.NewToken(common.IDENTIFIER, name, position),
		declaredParameters,
		returnType,
		module+"."+name,
	)
}
