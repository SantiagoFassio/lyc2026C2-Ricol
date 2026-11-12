package wire

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

const canonicalVersion = "ricol-interface-v1"

func canonicalForm(service *types.ServiceType) string {
	builder := canonicalBuilder{declared: map[string]types.Type{}}
	lines := []string{canonicalVersion, "service " + service.Name}
	for _, method := range service.Methods {
		lines = append(lines, "method "+method.Name+canonicalSignature(method.Signature))
		for _, param := range method.Signature.Params {
			builder.visit(param)
		}
		builder.visit(method.Signature.Return)
	}
	return strings.Join(append(lines, builder.declarations...), "\n")
}

func hashOf(canonical string) [32]byte {
	return sha256.Sum256([]byte(canonical))
}

type canonicalBuilder struct {
	declared     map[string]types.Type
	declarations []string
}

func (b *canonicalBuilder) visit(valueType types.Type) {
	switch valueType := valueType.(type) {
	case *types.StructType:
		if b.declare(valueType.Name, valueType) {
			b.declarations = append(b.declarations,
				"struct "+valueType.Name+" { "+joinTerminated(canonicalFields(valueType.Fields), "; ")+"}")
			b.visitFields(valueType.Fields)
		}
	case *types.EnumType:
		if b.declare(valueType.Name, valueType) {
			variants := make([]string, len(valueType.Variants))
			for i, variant := range valueType.Variants {
				variants[i] = canonicalVariant(variant)
			}
			b.declarations = append(b.declarations,
				"enum "+valueType.Name+" { "+joinTerminated(variants, "; ")+"}")
			for _, variant := range valueType.Variants {
				b.visitFields(variant.Fields)
			}
		}
	}
}

func (b *canonicalBuilder) visitFields(fields []types.StructField) {
	for _, field := range fields {
		b.visit(field.Type)
	}
}

func (b *canonicalBuilder) declare(name string, valueType types.Type) bool {
	existing, ok := b.declared[name]
	if !ok {
		b.declared[name] = valueType
		return true
	}
	if existing != valueType {
		panic(fmt.Sprintf("Two different types are named '%s' in the same service", name))
	}
	return false
}

func canonicalSignature(signature types.Signature) string {
	params := make([]string, len(signature.Params))
	for i, param := range signature.Params {
		params[i] = canonicalTypeName(param)
	}
	return "(" + strings.Join(params, ", ") + ") -> " + canonicalTypeName(signature.Return)
}

func canonicalVariant(variant types.EnumVariant) string {
	if len(variant.Fields) == 0 {
		return variant.Name
	}
	return variant.Name + "(" + strings.Join(canonicalFields(variant.Fields), ", ") + ")"
}

func canonicalFields(fields []types.StructField) []string {
	result := make([]string, len(fields))
	for i, field := range fields {
		result[i] = field.Name + ": " + canonicalTypeName(field.Type)
	}
	return result
}

func joinTerminated(items []string, terminator string) string {
	var result strings.Builder
	for _, item := range items {
		result.WriteString(item + terminator)
	}
	return result.String()
}

func canonicalTypeName(valueType types.Type) string {
	switch valueType := valueType.(type) {
	case *types.StructType:
		return valueType.Name
	case *types.EnumType:
		return valueType.Name
	}
	switch valueType {
	case types.Int, types.Float, types.Bool, types.Str, types.Void:
		return valueType.String()
	}
	panic(fmt.Sprintf("Cannot write type %v (%T) in an interface", valueType, valueType))
}
