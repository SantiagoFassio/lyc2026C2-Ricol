package types

type Boolean struct {
	value bool
}

func NewBoolean(value bool) Boolean {
	return Boolean{value: value}
}

func (b Boolean) Equals(other Boolean) Boolean {
	return NewBoolean(b.value == other.value)
}

func (b Boolean) Not() Boolean {
	return NewBoolean(!b.value)
}

func (b Boolean) TypeName() string {
	return "Bool"
}

func (b Boolean) String() string {
	if b.value {
		return "True"
	}
	return "False"
}

func (b Boolean) Display() string {
	return b.String()
}

func (Boolean) isValue() {}
