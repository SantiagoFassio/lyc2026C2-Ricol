package types

// Value es cualquier resultado que puede producir una expresión.
type Value interface {
	isValue()
	equalsValue(other Value) bool
	TypeName() string
	String() string
	Display() string
}
