package types

// Value es cualquier resultado que puede producir una expresión.
type Value interface {
	isValue()
	TypeName() string
	String() string
	Display() string
}
