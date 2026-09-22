package types

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
	Invalid Type = primitiveType{name: "<Invalid>"}
)
