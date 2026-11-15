package optimizer

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/interpreter"
)

func Optimize(statements []common.Statement) []common.Statement {
	return optimizeStatements(statements)
}

func optimizeStatements(statements []common.Statement) []common.Statement {
	optimized := make([]common.Statement, 0, len(statements))
	for _, statement := range statements {
		optimizedStatement := optimizeStatement(statement)
		if optimizedStatement == nil {
			continue
		}
		optimized = append(optimized, optimizedStatement)
		if endsControlFlow(optimizedStatement) {
			break
		}
	}
	return optimized
}

func optimizeBranch(statement common.Statement) common.Statement {
	optimized := optimizeStatement(statement)
	if optimized == nil {
		return common.NewBlockStatement([]common.Statement{})
	}
	return optimized
}

func endsControlFlow(statement common.Statement) bool {
	switch statement.(type) {
	case *common.ReturnStatement, *common.BreakStatement, *common.ContinueStatement:
		return true
	default:
		return false
	}
}

func optimizeStatement(statement common.Statement) common.Statement {
	switch typedStatement := statement.(type) {
	case *common.ExpressionStatement:
		typedStatement.Expression = foldExpression(typedStatement.Expression)
	case *common.PrintStatement:
		typedStatement.Expression = foldExpression(typedStatement.Expression)
	case *common.BlockStatement:
		typedStatement.Statements = optimizeStatements(typedStatement.Statements)
	case *common.IfStatement:
		return optimizeIfStatement(typedStatement)
	case *common.WhileStatement:
		return optimizeWhileStatement(typedStatement)
	case *common.VarDeclarationStatement:
		typedStatement.ValueExpression = foldExpression(typedStatement.ValueExpression)
	case *common.FuncDeclarationStatement:
		typedStatement.Body = optimizeStatements(typedStatement.Body)
	case *common.ReturnStatement:
		if typedStatement.ValueExpression != nil {
			typedStatement.ValueExpression = foldExpression(typedStatement.ValueExpression)
		}
	case *common.MatchStatement:
		typedStatement.Expression = foldExpression(typedStatement.Expression)
		for index := range typedStatement.Arms {
			typedStatement.Arms[index].Body = optimizeBranch(typedStatement.Arms[index].Body)
		}
	case *common.ServerDeclarationStatement:
		typedStatement.Body = optimizeStatements(typedStatement.Body)
	case *common.SessionStatement:
		typedStatement.Connection = foldExpression(typedStatement.Connection)
		if typedStatement.Timeout != nil {
			typedStatement.Timeout = foldExpression(typedStatement.Timeout)
		}
		typedStatement.Body = optimizeStatements(typedStatement.Body)
		typedStatement.RescueBody = optimizeStatements(typedStatement.RescueBody)
	case *common.ContinueStatement, *common.BreakStatement, *common.StructDeclarationStatement,
		*common.EnumDeclarationStatement, *common.ImportStatement, *common.ServiceDeclarationStatement:
	default:
		panic(fmt.Sprintf("Unknown statement node: %T", statement))
	}
	return statement
}

func optimizeIfStatement(statement *common.IfStatement) common.Statement {
	statement.Condition = foldExpression(statement.Condition)
	if condition, isLiteral := literalBoolean(statement.Condition); isLiteral {
		if condition {
			return optimizeStatement(statement.IfBranch)
		}
		if statement.ElseBranch == nil {
			return nil
		}
		return optimizeStatement(statement.ElseBranch)
	}
	statement.IfBranch = optimizeBranch(statement.IfBranch)
	if statement.ElseBranch != nil {
		statement.ElseBranch = optimizeStatement(statement.ElseBranch)
	}
	return statement
}

func optimizeWhileStatement(statement *common.WhileStatement) common.Statement {
	statement.Condition = foldExpression(statement.Condition)
	if condition, isLiteral := literalBoolean(statement.Condition); isLiteral && !condition {
		return nil
	}
	statement.Body = optimizeBranch(statement.Body)
	return statement
}

func literalBoolean(expression common.Expression) (bool, bool) {
	literal, ok := expression.(*common.LiteralExpression)
	if !ok {
		return false, false
	}
	boolean, ok := literal.Value.(types.Boolean)
	if !ok {
		return false, false
	}
	return boolean.IsTrue(), true
}

func foldExpression(expression common.Expression) common.Expression {
	switch typedExpression := expression.(type) {
	case *common.BinaryExpression:
		return foldBinaryExpression(typedExpression)
	case *common.GroupingExpression:
		typedExpression.Expression = foldExpression(typedExpression.Expression)
		if literal, ok := typedExpression.Expression.(*common.LiteralExpression); ok {
			return literal
		}
	case *common.UnaryExpression:
		typedExpression.Expression = foldExpression(typedExpression.Expression)
		if _, ok := typedExpression.Expression.(*common.LiteralExpression); ok {
			return evaluateOrKeep(typedExpression, typedExpression.Operator.Position)
		}
	case *common.VarAssignmentExpression:
		typedExpression.ValueExpression = foldExpression(typedExpression.ValueExpression)
	case *common.CallExpression:
		foldExpressions(typedExpression.Arguments)
	case *common.StructLiteralExpression:
		for index := range typedExpression.Fields {
			typedExpression.Fields[index].ValueExpression = foldExpression(typedExpression.Fields[index].ValueExpression)
		}
	case *common.FieldAccessExpression:
		typedExpression.Expression = foldExpression(typedExpression.Expression)
	case *common.MemberCallExpression:
		typedExpression.Expression = foldExpression(typedExpression.Expression)
		foldExpressions(typedExpression.Arguments)
	case *common.LiteralExpression, *common.VariableExpression, *common.ConnectExpression, *common.InputExpression:
	default:
		panic(fmt.Sprintf("Unknown expression node: %T", expression))
	}
	return expression
}

func foldExpressions(expressions []common.Expression) {
	for index, expression := range expressions {
		expressions[index] = foldExpression(expression)
	}
}

func foldBinaryExpression(expression *common.BinaryExpression) common.Expression {
	expression.LeftExpression = foldExpression(expression.LeftExpression)
	expression.RightExpression = foldExpression(expression.RightExpression)
	left, leftIsLiteral := expression.LeftExpression.(*common.LiteralExpression)
	if !leftIsLiteral {
		return expression
	}
	if isLogicalOperator(expression.Operator) {
		return foldLogicalExpression(expression, left)
	}
	if _, rightIsLiteral := expression.RightExpression.(*common.LiteralExpression); rightIsLiteral {
		return evaluateOrKeep(expression, left.Token.Position)
	}
	return expression
}

func foldLogicalExpression(expression *common.BinaryExpression, left *common.LiteralExpression) common.Expression {
	leftValue, ok := left.Value.(types.Boolean)
	if !ok {
		return expression
	}
	shortCircuits := leftValue.IsTrue() == (expression.Operator.TokenType == common.OR)
	if shortCircuits {
		return left
	}
	return expression.RightExpression
}

func evaluateOrKeep(expression common.Expression, position common.Position) common.Expression {
	value, err := interpreter.EvaluateConstant(expression)
	if err != nil {
		return expression
	}
	return common.NewLiteralExpression(common.NewToken(literalTokenType(value), value.String(), position), value)
}

func literalTokenType(value types.Value) common.TokenType {
	switch typedValue := value.(type) {
	case types.Number:
		if _, isInteger := typedValue.IntegerValue(); isInteger {
			return common.INTEGER
		}
		return common.FLOAT
	case types.String:
		return common.STRING
	case types.Boolean:
		if typedValue.IsTrue() {
			return common.TRUE
		}
		return common.FALSE
	default:
		panic(fmt.Sprintf("Unexpected constant value: %T", value))
	}
}

func isLogicalOperator(operator common.Token) bool {
	return operator.TokenType == common.AND || operator.TokenType == common.OR
}
