package types

import "strings"

type Type interface {
	isType()
	String() string
}

type primitiveType struct {
	name string
}

func (primitiveType) isType() {}

func (p primitiveType) String() string { return p.name }

var (
	Int     Type = primitiveType{name: "Int"}
	Float   Type = primitiveType{name: "Float"}
	Str     Type = primitiveType{name: "String"}
	Bool    Type = primitiveType{name: "Bool"}
	Void    Type = primitiveType{name: "Void"}
	Invalid Type = primitiveType{name: "<Invalid>"}
)

type Signature struct {
	Params []Type
	Return Type
}

func (s Signature) String() string {
	params := make([]string, len(s.Params))
	for i, param := range s.Params {
		params[i] = param.String()
	}
	return "(" + strings.Join(params, ", ") + ") -> " + s.Return.String()
}
