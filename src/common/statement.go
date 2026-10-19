package common

import (
	"fmt"
	"strings"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type Statement interface {
	isStatement()
	String() string
}

type ExpressionStatement struct {
	Expression Expression
}

type PrintStatement struct {
	Expression Expression
}

type BlockStatement struct {
	Statements []Statement
}

type IfStatement struct {
	IfToken    Token
	Condition  Expression
	IfBranch   Statement
	ElseBranch Statement
}

type WhileStatement struct {
	WhileToken Token
	Condition  Expression
	Body       Statement
}

type ContinueStatement struct {
	ContinueToken Token
}

type BreakStatement struct {
	BreakToken Token
}

type VarDeclarationStatement struct {
	LetToken        Token
	NameToken       Token
	VarType         types.Type
	ValueExpression Expression
}

type Parameter struct {
	NameToken Token
	ParamType types.Type
}

type FuncDeclarationStatement struct {
	FuncToken  Token
	NameToken  Token
	Parameters []Parameter
	ReturnType types.Type
	Body       []Statement
}

type ReturnStatement struct {
	ReturnToken     Token
	ValueExpression Expression
}

type FieldDeclaration struct {
	NameToken Token
	FieldType types.Type
}

type StructDeclarationStatement struct {
	StructToken Token
	NameToken   Token
	Fields      []FieldDeclaration
}

type VariantDeclaration struct {
	NameToken Token
	Fields    []FieldDeclaration
}

type EnumDeclarationStatement struct {
	EnumToken Token
	NameToken Token
	Variants  []VariantDeclaration
}

type MatchPattern struct {
	NameToken Token
	Bindings  []Token
}

type MatchArm struct {
	Pattern MatchPattern
	Body    Statement
}

type MatchStatement struct {
	MatchToken Token
	Expression Expression
	Arms       []MatchArm
}

func NewExpressionStatement(expression Expression) *ExpressionStatement {
	return &ExpressionStatement{
		Expression: expression,
	}
}

func NewPrintStatement(expression Expression) *PrintStatement {
	return &PrintStatement{
		Expression: expression,
	}
}

func NewBlockStatement(statements []Statement) *BlockStatement {
	return &BlockStatement{
		Statements: statements,
	}
}

func NewIfStatement(ifToken Token, condition Expression, ifBranch Statement, elseBranch Statement) *IfStatement {
	return &IfStatement{
		IfToken:    ifToken,
		Condition:  condition,
		IfBranch:   ifBranch,
		ElseBranch: elseBranch,
	}
}

func NewWhileStatement(whileToken Token, condition Expression, body Statement) *WhileStatement {
	return &WhileStatement{
		WhileToken: whileToken,
		Condition:  condition,
		Body:       body,
	}
}

func NewContinueStatement(continueToken Token) *ContinueStatement {
	return &ContinueStatement{
		ContinueToken: continueToken,
	}
}

func NewBreakStatement(breakToken Token) *BreakStatement {
	return &BreakStatement{
		BreakToken: breakToken,
	}
}

func NewVarDeclarationStatement(
	letToken Token,
	nameToken Token,
	varType types.Type,
	valueExpression Expression,
) *VarDeclarationStatement {
	return &VarDeclarationStatement{
		LetToken:        letToken,
		NameToken:       nameToken,
		VarType:         varType,
		ValueExpression: valueExpression,
	}
}

func NewParameter(nameToken Token, paramType types.Type) Parameter {
	return Parameter{
		NameToken: nameToken,
		ParamType: paramType,
	}
}

func NewFuncDeclarationStatement(
	funcToken Token,
	nameToken Token,
	parameters []Parameter,
	returnType types.Type,
	body []Statement,
) *FuncDeclarationStatement {
	return &FuncDeclarationStatement{
		FuncToken:  funcToken,
		NameToken:  nameToken,
		Parameters: parameters,
		ReturnType: returnType,
		Body:       body,
	}
}

func NewReturnStatement(returnToken Token, valueExpression Expression) *ReturnStatement {
	return &ReturnStatement{
		ReturnToken:     returnToken,
		ValueExpression: valueExpression,
	}
}

func NewFieldDeclaration(nameToken Token, fieldType types.Type) FieldDeclaration {
	return FieldDeclaration{
		NameToken: nameToken,
		FieldType: fieldType,
	}
}

func NewStructDeclarationStatement(structToken Token, nameToken Token, fields []FieldDeclaration) *StructDeclarationStatement {
	return &StructDeclarationStatement{
		StructToken: structToken,
		NameToken:   nameToken,
		Fields:      fields,
	}
}

func NewVariantDeclaration(nameToken Token, fields []FieldDeclaration) VariantDeclaration {
	return VariantDeclaration{
		NameToken: nameToken,
		Fields:    fields,
	}
}

func NewEnumDeclarationStatement(
	enumToken Token,
	nameToken Token,
	variants []VariantDeclaration,
) *EnumDeclarationStatement {
	return &EnumDeclarationStatement{
		EnumToken: enumToken,
		NameToken: nameToken,
		Variants:  variants,
	}
}

func NewMatchPattern(nameToken Token, bindings []Token) MatchPattern {
	return MatchPattern{
		NameToken: nameToken,
		Bindings:  bindings,
	}
}

func NewMatchArm(pattern MatchPattern, body Statement) MatchArm {
	return MatchArm{
		Pattern: pattern,
		Body:    body,
	}
}

func (a MatchArm) BodyStatements() []Statement {
	if block, ok := a.Body.(*BlockStatement); ok {
		return block.Statements
	}
	return []Statement{a.Body}
}

func NewMatchStatement(matchToken Token, expression Expression, arms []MatchArm) *MatchStatement {
	return &MatchStatement{
		MatchToken: matchToken,
		Expression: expression,
		Arms:       arms,
	}
}

func (f *FuncDeclarationStatement) Signature() types.Signature {
	params := make([]types.Type, len(f.Parameters))
	for i, parameter := range f.Parameters {
		params[i] = parameter.ParamType
	}
	return types.Signature{Params: params, Return: f.ReturnType}
}

func (e *ExpressionStatement) String() string {
	return fmt.Sprintf("%s;\n", e.Expression.String())
}

func (p *PrintStatement) String() string {
	return fmt.Sprintf("print %v;\n", p.Expression)
}

func (b *BlockStatement) String() string {
	return blockString(b.Statements)
}

func blockString(statements []Statement) string {
	var stringBuilder strings.Builder
	stringBuilder.WriteString("{\n")
	for _, statement := range statements {
		writeIndented(&stringBuilder, statement.String())
	}
	stringBuilder.WriteString("}\n")
	return stringBuilder.String()
}

func writeIndented(stringBuilder *strings.Builder, text string) {
	for _, line := range strings.SplitAfter(text, "\n") {
		if line != "" {
			stringBuilder.WriteString("  " + line)
		}
	}
}

func (i *IfStatement) String() string {
	ifString := fmt.Sprintf("if (%v) %v", i.Condition, i.IfBranch)
	if i.ElseBranch == nil {
		return ifString
	}
	return fmt.Sprintf("%s else %v", strings.TrimSuffix(ifString, "\n"), i.ElseBranch)
}

func (w *WhileStatement) String() string {
	return fmt.Sprintf("while (%v) %v", w.Condition, w.Body)
}

func (c *ContinueStatement) String() string {
	return "continue;\n"
}

func (b *BreakStatement) String() string {
	return "break;\n"
}

func (v *VarDeclarationStatement) String() string {
	return fmt.Sprintf("let %s : %v = %v;\n", v.NameToken.Lexeme, v.VarType, v.ValueExpression)
}

func (f *FuncDeclarationStatement) String() string {
	params := make([]string, len(f.Parameters))
	for i, parameter := range f.Parameters {
		params[i] = fmt.Sprintf("%s : %v", parameter.NameToken.Lexeme, parameter.ParamType)
	}
	return fmt.Sprintf("func %s(%s) -> %v %s", f.NameToken.Lexeme, strings.Join(params, ", "), f.ReturnType, blockString(f.Body))
}

func (r *ReturnStatement) String() string {
	if r.ValueExpression == nil {
		return "return;\n"
	}
	return fmt.Sprintf("return %v;\n", r.ValueExpression)
}

func (s *StructDeclarationStatement) String() string {
	var stringBuilder strings.Builder
	fmt.Fprintf(&stringBuilder, "struct %s {\n", s.NameToken.Lexeme)
	for _, field := range s.Fields {
		fmt.Fprintf(&stringBuilder, "  %v;\n", field)
	}
	stringBuilder.WriteString("}\n")
	return stringBuilder.String()
}

func (e *EnumDeclarationStatement) String() string {
	var stringBuilder strings.Builder
	fmt.Fprintf(&stringBuilder, "enum %s {\n", e.NameToken.Lexeme)
	for _, variant := range e.Variants {
		fmt.Fprintf(&stringBuilder, "  %v;\n", variant)
	}
	stringBuilder.WriteString("}\n")
	return stringBuilder.String()
}

func (f FieldDeclaration) String() string {
	return fmt.Sprintf("%s : %v", f.NameToken.Lexeme, f.FieldType)
}

func (v VariantDeclaration) String() string {
	if len(v.Fields) == 0 {
		return v.NameToken.Lexeme
	}
	fields := make([]string, len(v.Fields))
	for i, field := range v.Fields {
		fields[i] = field.String()
	}
	return fmt.Sprintf("%s(%s)", v.NameToken.Lexeme, strings.Join(fields, ", "))
}

func (p MatchPattern) String() string {
	if len(p.Bindings) == 0 {
		return p.NameToken.Lexeme
	}
	bindings := make([]string, len(p.Bindings))
	for i, binding := range p.Bindings {
		bindings[i] = binding.Lexeme
	}
	return fmt.Sprintf("%s(%s)", p.NameToken.Lexeme, strings.Join(bindings, ", "))
}

func (m *MatchStatement) String() string {
	var stringBuilder strings.Builder
	fmt.Fprintf(&stringBuilder, "match (%v) {\n", m.Expression)
	for _, arm := range m.Arms {
		writeIndented(&stringBuilder, fmt.Sprintf("%v => %v", arm.Pattern, arm.Body))
	}
	stringBuilder.WriteString("}\n")
	return stringBuilder.String()
}

func (e *ExpressionStatement) isStatement()        {}
func (p *PrintStatement) isStatement()             {}
func (b *BlockStatement) isStatement()             {}
func (i *IfStatement) isStatement()                {}
func (w *WhileStatement) isStatement()             {}
func (c *ContinueStatement) isStatement()          {}
func (b *BreakStatement) isStatement()             {}
func (v *VarDeclarationStatement) isStatement()    {}
func (f *FuncDeclarationStatement) isStatement()   {}
func (r *ReturnStatement) isStatement()            {}
func (s *StructDeclarationStatement) isStatement() {}
func (e *EnumDeclarationStatement) isStatement()   {}
func (m *MatchStatement) isStatement()             {}
