package typechecker

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type TypeChecker struct {
	statements      []common.Statement
	errors          common.RicolErrorList
	nestedLoops     int
	currentFunction *common.FuncDeclarationStatement
	scopes          *scopeStack
	distances       map[common.Expression]int
}

func NewTypeChecker(statements []common.Statement) *TypeChecker {
	return &TypeChecker{
		statements: statements,
		scopes:     newScopeStack(),
		distances:  map[common.Expression]int{},
	}
}

func (t *TypeChecker) Check() (map[common.Expression]int, common.RicolErrorList) {
	t.checkStatements(t.statements)
	return t.distances, t.errors
}

func (t *TypeChecker) checkStatements(statements []common.Statement) {
	for i, statement := range statements {
		if _, ok := statement.(*common.FuncDeclarationStatement); ok && !isFuncDeclaration(statements, i-1) {
			t.declareFunctionGroup(statements[i:])
		}
		t.checkStatement(statement)
	}
}

func (t *TypeChecker) declareFunctionGroup(statements []common.Statement) {
	for _, statement := range statements {
		funcDeclaration, ok := statement.(*common.FuncDeclarationStatement)
		if !ok {
			return
		}
		nameToken := funcDeclaration.NameToken
		if !t.scopes.declareFunction(nameToken.Lexeme, funcDeclaration.Signature()) {
			t.reportError(nameToken.Position, fmt.Sprintf("Function '%s' already declared in this scope", nameToken.Lexeme))
		}
	}
}

func isFuncDeclaration(statements []common.Statement, index int) bool {
	if index < 0 {
		return false
	}
	_, ok := statements[index].(*common.FuncDeclarationStatement)
	return ok
}

func (t *TypeChecker) checkStatement(statement common.Statement) {
	switch typedStatement := statement.(type) {
	case *common.PrintStatement:
		t.checkPrintStatement(typedStatement)
	case *common.BlockStatement:
		t.scopes.push()
		t.checkStatements(typedStatement.Statements)
		t.scopes.pop()
	case *common.IfStatement:
		t.checkIfStatement(typedStatement)
	case *common.WhileStatement:
		t.checkWhileStatement(typedStatement)
	case *common.ContinueStatement:
		t.checkContinueStatement(typedStatement)
	case *common.BreakStatement:
		t.checkBreakStatement(typedStatement)
	case *common.VarDeclarationStatement:
		t.checkVarDeclarationStatement(typedStatement)
	case *common.FuncDeclarationStatement:
		t.checkFuncDeclarationStatement(typedStatement)
	case *common.ReturnStatement:
		t.checkReturnStatement(typedStatement)
	case *common.ExpressionStatement:
		t.checkExpression(typedStatement.Expression)
	default:
		panic(fmt.Sprintf("Unknown statement node: %T", statement))
	}
}

func (t *TypeChecker) checkPrintStatement(printStatement *common.PrintStatement) {
	if t.checkExpression(printStatement.Expression) == types.Void {
		t.reportError(voidExpressionPosition(printStatement.Expression), "Cannot print a Void value")
	}
}

func (t *TypeChecker) checkIfStatement(ifStatement *common.IfStatement) {
	conditionType := t.checkExpression(ifStatement.Condition)
	if conditionType != types.Invalid && !isBool(conditionType) {
		t.reportError(ifStatement.IfToken.Position,
			fmt.Sprintf("Non boolean expression in if condition: %s", conditionType))
	}
	t.checkStatement(ifStatement.IfBranch)
	if ifStatement.ElseBranch != nil {
		t.checkStatement(ifStatement.ElseBranch)
	}
}

func (t *TypeChecker) checkWhileStatement(whileStatement *common.WhileStatement) {
	conditionType := t.checkExpression(whileStatement.Condition)
	if conditionType != types.Invalid && !isBool(conditionType) {
		t.reportError(whileStatement.WhileToken.Position,
			fmt.Sprintf("Non boolean expression in while condition: %s", conditionType))
	}
	t.nestedLoops++
	t.checkStatement(whileStatement.Body)
	t.nestedLoops--
}

func (t *TypeChecker) checkContinueStatement(continueStatement *common.ContinueStatement) {
	if t.nestedLoops == 0 {
		t.reportError(continueStatement.ContinueToken.Position, "'continue' outside loop")
	}
}

func (t *TypeChecker) checkBreakStatement(breakStatement *common.BreakStatement) {
	if t.nestedLoops == 0 {
		t.reportError(breakStatement.BreakToken.Position, "'break' outside loop")
	}
}

func (t *TypeChecker) checkVarDeclarationStatement(varDeclarationStatement *common.VarDeclarationStatement) {
	valueExpressionType := t.checkExpression(varDeclarationStatement.ValueExpression)
	if valueExpressionType != types.Invalid && valueExpressionType != varDeclarationStatement.VarType {
		t.reportError(varDeclarationStatement.LetToken.Position,
			fmt.Sprintf("Cannot assign %v to variable of type %v", valueExpressionType, varDeclarationStatement.VarType))
	}
	nameToken := varDeclarationStatement.NameToken
	if !t.scopes.declareVariable(nameToken.Lexeme, varDeclarationStatement.VarType) {
		t.reportError(nameToken.Position, fmt.Sprintf("Variable '%s' already declared in this scope", nameToken.Lexeme))
	}
}

func (t *TypeChecker) checkFuncDeclarationStatement(funcDeclaration *common.FuncDeclarationStatement) {
	previousFunction, previousNestedLoops := t.currentFunction, t.nestedLoops
	t.currentFunction, t.nestedLoops = funcDeclaration, 0
	t.scopes.push()
	defer func() {
		t.scopes.pop()
		t.currentFunction, t.nestedLoops = previousFunction, previousNestedLoops
	}()

	for _, parameter := range funcDeclaration.Parameters {
		if !t.scopes.declareVariable(parameter.NameToken.Lexeme, parameter.ParamType) {
			t.reportError(parameter.NameToken.Position, fmt.Sprintf("Duplicate parameter '%s' in function '%s'",
				parameter.NameToken.Lexeme, funcDeclaration.NameToken.Lexeme))
		}
	}
	t.checkStatements(funcDeclaration.Body)
	if funcDeclaration.ReturnType != types.Void && !alwaysReturns(funcDeclaration.Body) {
		t.reportError(funcDeclaration.NameToken.Position,
			fmt.Sprintf("Function '%s' does not return a value on every path", funcDeclaration.NameToken.Lexeme))
	}
}

func alwaysReturns(statements []common.Statement) bool {
	for _, statement := range statements {
		switch typedStatement := statement.(type) {
		case *common.ReturnStatement:
			return true
		case *common.BlockStatement:
			if alwaysReturns(typedStatement.Statements) {
				return true
			}
		case *common.IfStatement:
			if typedStatement.ElseBranch != nil &&
				alwaysReturns([]common.Statement{typedStatement.IfBranch}) &&
				alwaysReturns([]common.Statement{typedStatement.ElseBranch}) {
				return true
			}
		}
	}
	return false
}

func (t *TypeChecker) checkReturnStatement(returnStatement *common.ReturnStatement) {
	valueType := types.Void
	if returnStatement.ValueExpression != nil {
		valueType = t.checkExpression(returnStatement.ValueExpression)
	}
	position := returnStatement.ReturnToken.Position
	if t.currentFunction == nil {
		t.reportError(position, "'return' outside function")
		return
	}
	functionName := t.currentFunction.NameToken.Lexeme
	returnType := t.currentFunction.ReturnType
	switch {
	case valueType == types.Invalid:
	case returnType == types.Void && returnStatement.ValueExpression != nil:
		t.reportError(position, fmt.Sprintf("Function '%s' cannot return a value", functionName))
	case returnType != types.Void && returnStatement.ValueExpression == nil:
		t.reportError(position, fmt.Sprintf("Function '%s' must return a value of type %v", functionName, returnType))
	case valueType != returnType:
		t.reportError(position, fmt.Sprintf("Cannot return %v from function '%s' of type %v", valueType, functionName, returnType))
	}
}

func (t *TypeChecker) checkExpression(expression common.Expression) types.Type {
	switch typedExpression := expression.(type) {
	case *common.BinaryExpression:
		return t.checkBinaryExpression(typedExpression)
	case *common.GroupingExpression:
		return t.checkExpression(typedExpression.Expression)
	case *common.LiteralExpression:
		return t.checkLiteralExpression(typedExpression)
	case *common.UnaryExpression:
		return t.checkUnaryExpression(typedExpression)
	case *common.VariableExpression:
		return t.checkVariableExpression(typedExpression)
	case *common.VarAssignmentExpression:
		return t.checkVarAssignmentExpression(typedExpression)
	case *common.CallExpression:
		return t.checkCallExpression(typedExpression)
	default:
		panic(fmt.Sprintf("Unknown expression node: %T", expression))
	}
}

func (t *TypeChecker) checkVariableExpression(expression *common.VariableExpression) types.Type {
	declared, ok := t.resolve(expression, expression.NameToken, "variable")
	if !ok {
		return types.Invalid
	}
	if declared.signature != nil {
		return t.reportError(expression.NameToken.Position,
			fmt.Sprintf("Cannot use function '%s' as a value", expression.NameToken.Lexeme))
	}
	return declared.varType
}

func (t *TypeChecker) checkVarAssignmentExpression(expression *common.VarAssignmentExpression) types.Type {
	valueType := t.checkExpression(expression.ValueExpression)
	declared, ok := t.resolve(expression, expression.NameToken, "variable")
	if !ok {
		return types.Invalid
	}
	if declared.signature != nil {
		return t.reportError(expression.NameToken.Position,
			fmt.Sprintf("Cannot assign to function '%s'", expression.NameToken.Lexeme))
	}
	varType := declared.varType
	if valueType != types.Invalid && valueType != varType {
		t.reportError(expression.NameToken.Position,
			fmt.Sprintf("Cannot assign %v to variable of type %v", valueType, varType))
	}
	return varType
}

func (t *TypeChecker) resolve(expression common.Expression, nameToken common.Token, kind string) (symbol, bool) {
	declared, distance, ok := t.scopes.lookup(nameToken.Lexeme)
	if !ok {
		t.reportError(nameToken.Position, fmt.Sprintf("Undefined %s '%s'", kind, nameToken.Lexeme))
		return symbol{}, false
	}
	t.distances[expression] = distance
	return declared, true
}

func (t *TypeChecker) checkCallExpression(expression *common.CallExpression) types.Type {
	argumentTypes := make([]types.Type, len(expression.Arguments))
	for i, argument := range expression.Arguments {
		argumentTypes[i] = t.checkExpression(argument)
	}
	nameToken := expression.NameToken
	declared, ok := t.resolve(expression, nameToken, "function")
	if !ok {
		return types.Invalid
	}
	if declared.signature == nil {
		return t.reportError(nameToken.Position, fmt.Sprintf("'%s' is not a function", nameToken.Lexeme))
	}
	signature := declared.signature
	if len(argumentTypes) != len(signature.Params) {
		t.reportError(nameToken.Position, fmt.Sprintf("Function '%s' expects %s, got %d",
			nameToken.Lexeme, pluralize(len(signature.Params), "argument"), len(argumentTypes)))
		return signature.Return
	}
	for i, argumentType := range argumentTypes {
		paramType := signature.Params[i]
		if argumentType != types.Invalid && argumentType != paramType {
			t.reportError(nameToken.Position, fmt.Sprintf("Argument %d of function '%s' must be %v, got %v",
				i+1, nameToken.Lexeme, paramType, argumentType))
		}
	}
	return signature.Return
}

func (t *TypeChecker) checkLiteralExpression(expression *common.LiteralExpression) types.Type {
	switch expression.Token.TokenType {
	case common.INTEGER:
		return types.Int
	case common.FLOAT:
		return types.Float
	case common.STRING:
		return types.Str
	case common.TRUE, common.FALSE:
		return types.Bool
	default:
		panic(fmt.Sprintf("Unknown literal token: %v", expression.Token))
	}
}

func (t *TypeChecker) checkUnaryExpression(expression *common.UnaryExpression) types.Type {
	operandType := t.checkExpression(expression.Expression)
	if operandType == types.Invalid {
		return types.Invalid
	}
	switch expression.Operator.TokenType {
	case common.MINUS:
		if !isNumeric(operandType) {
			return t.unsupportedOperandError(expression.Operator, operandType)
		}
		return operandType
	case common.NOT:
		if !isBool(operandType) {
			return t.unsupportedOperandError(expression.Operator, operandType)
		}
		return types.Bool
	default:
		return t.reportError(expression.Operator.Position,
			fmt.Sprintf("Invalid unary operator: %v", expression.Operator))
	}
}

func (t *TypeChecker) unsupportedOperandError(operator common.Token, operandType types.Type) types.Type {
	return t.reportError(operator.Position,
		fmt.Sprintf("Unsupported operand type for %s: %s", operator.Lexeme, operandType))
}

func (t *TypeChecker) checkBinaryExpression(expression *common.BinaryExpression) types.Type {
	leftType := t.checkExpression(expression.LeftExpression)
	rightType := t.checkExpression(expression.RightExpression)
	if leftType == types.Invalid || rightType == types.Invalid {
		return types.Invalid
	}
	if isLogicalOperator(expression.Operator) {
		return t.checkLogicalOperation(expression.Operator, leftType, rightType)
	}
	if isComparisonOperator(expression.Operator) {
		return t.checkComparison(expression.Operator, leftType, rightType)
	}
	if leftType == types.Str && rightType == types.Str && expression.Operator.TokenType == common.PLUS {
		return types.Str
	}
	if isNumeric(leftType) && isNumeric(rightType) {
		return t.checkNumericOperation(expression.Operator, leftType, rightType)
	}
	return t.reportError(expression.Operator.Position,
		fmt.Sprintf("Unsupported operand types for %s: %s and %s", expression.Operator.Lexeme, leftType, rightType))
}

func (t *TypeChecker) checkNumericOperation(operator common.Token, leftType types.Type, rightType types.Type) types.Type {
	switch operator.TokenType {
	case common.PLUS, common.MINUS, common.STAR, common.DOUBLE_SLASH, common.PERCENTAGE, common.DOUBLE_STAR:
		return numericResult(leftType, rightType)
	case common.SLASH:
		return types.Float
	default:
		return t.reportError(operator.Position, fmt.Sprintf("Invalid binary operator: %v", operator))
	}
}

// Se comparan numeros con numeros, strings con strings y booleanos con booleanos.
// Los booleanos solo se comparan por igualdad, porque no tienen orden.
func (t *TypeChecker) checkComparison(operator common.Token, leftType types.Type, rightType types.Type) types.Type {
	sameKind := (isNumeric(leftType) && isNumeric(rightType)) || (leftType == rightType && leftType != types.Void)
	isOrdering := operator.TokenType != common.DOUBLE_EQUAL && operator.TokenType != common.NOT_EQUAL
	if !sameKind || (isOrdering && isBool(leftType)) {
		return t.reportError(operator.Position,
			fmt.Sprintf("Unsupported operand types for %s: %s and %s", operator.Lexeme, leftType, rightType))
	}
	return types.Bool
}

func (t *TypeChecker) checkLogicalOperation(operator common.Token, leftType types.Type, rightType types.Type) types.Type {
	if !isBool(leftType) || !isBool(rightType) {
		return t.reportError(operator.Position,
			fmt.Sprintf("Unsupported operand types for %s: %s and %s", operator.Lexeme, leftType, rightType))
	}
	return types.Bool
}

func (t *TypeChecker) reportError(position common.Position, message string) types.Type {
	t.errors = append(t.errors, common.NewRicolError(position, message))
	return types.Invalid
}

func voidExpressionPosition(expression common.Expression) common.Position {
	switch typedExpression := expression.(type) {
	case *common.CallExpression:
		return typedExpression.NameToken.Position
	case *common.GroupingExpression:
		return voidExpressionPosition(typedExpression.Expression)
	default:
		panic(fmt.Sprintf("Void value from a non-call expression: %T", expression))
	}
}

func pluralize(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, noun)
	}
	return fmt.Sprintf("%d %ss", count, noun)
}

func numericResult(leftType types.Type, rightType types.Type) types.Type {
	if leftType == types.Int && rightType == types.Int {
		return types.Int
	}
	return types.Float
}

func isLogicalOperator(operator common.Token) bool {
	return operator.TokenType == common.AND || operator.TokenType == common.OR
}

func isComparisonOperator(operator common.Token) bool {
	switch operator.TokenType {
	case common.DOUBLE_EQUAL, common.NOT_EQUAL, common.LESS, common.LESS_EQUAL, common.GREATER, common.GREATER_EQUAL:
		return true
	default:
		return false
	}
}

func isNumeric(expressionType types.Type) bool {
	return expressionType == types.Int || expressionType == types.Float
}

func isBool(expressionType types.Type) bool {
	return expressionType == types.Bool
}
