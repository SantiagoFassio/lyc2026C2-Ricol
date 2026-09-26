package types

type Boolean struct {
	value bool
}

func NewBoolean(value bool) Boolean {
	return Boolean{value: value}
}

func (b Boolean) Equals(other Boolean) bool {
	return b.value == other.value
}

func (b Boolean) TypeName() string {
	return "boolean"
}

// Se muestra igual que se escribe en el código fuente.
func (b Boolean) String() string {
	if b.value {
		return "True"
	}
	return "False"
}

func (Boolean) isValue() {}
