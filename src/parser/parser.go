package parser

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type Parser struct {
	tokens     []common.Token
	currentPos int
}

func NewParser(tokens []common.Token) *Parser {
	return &Parser{
		tokens:     tokens,
		currentPos: 0,
	}
}

func (p *Parser) Parse() ([]common.Statement, error) {
	statements := []common.Statement{}
	for !p.isAtTheEnd() {
		nextStatement, err := p.parseNextStatement()
		if err != nil {
			return []common.Statement{}, err
		}
		statements = append(statements, nextStatement)
	}
	return statements, nil
}

func (p *Parser) parseNextStatement() (common.Statement, error) {
	switch p.tokens[p.currentPos].TokenType {
	case common.PRINT:
		return p.parseNextPrintStatement()
	case common.OPEN_BRACE:
		return p.parseNextBlockStatement()
	case common.IF:
		return p.parseNextIfStatement()
	case common.WHILE:
		return p.parseNextWhileStatement()
	case common.CONTINUE:
		return p.parseNextContinueStatement()
	case common.BREAK:
		return p.parseNextBreakStatement()
	case common.LET:
		return p.parseNextVarDeclarationStatement()
	case common.FUNC:
		return p.parseNextFuncDeclarationStatement()
	case common.RETURN:
		return p.parseNextReturnStatement()
	default:
		return p.parseNextExpressionStatement()
	}
}

func (p *Parser) parseNextPrintStatement() (common.Statement, error) {
	p.currentPos++
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType == common.SEMICOLON {
		return nil, common.NewRicolError(p.currentPosition(), "Expected expression after 'PRINT'")
	}
	expression, err := p.parseNextExpression()
	if err != nil {
		return nil, err
	}
	if err := p.consumeSemicolon("expression"); err != nil {
		return nil, err
	}
	return common.NewPrintStatement(expression), nil
}

func (p *Parser) parseNextBlockStatement() (common.Statement, error) {
	statements, err := p.parseNextBlockStatements()
	if err != nil {
		return nil, err
	}
	return common.NewBlockStatement(statements), nil
}

func (p *Parser) parseNextBlockStatements() ([]common.Statement, error) {
	p.currentPos++
	statements := []common.Statement{}
	for !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType != common.CLOSED_BRACE {
		nextStatement, err := p.parseNextStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, nextStatement)
	}
	if p.isAtTheEnd() {
		return nil, common.NewRicolError(p.currentPosition(), "Expected '}' after block")
	}
	p.currentPos++
	return statements, nil
}

func (p *Parser) parseNextIfStatement() (common.Statement, error) {
	ifToken := p.tokens[p.currentPos]
	p.currentPos++
	condition, err := p.parseNextCondition(ifToken)
	if err != nil {
		return nil, err
	}
	ifBranch, err := p.parseNextBody()
	if err != nil {
		return nil, err
	}
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.ELSE {
		return common.NewIfStatement(ifToken, condition, ifBranch, nil), nil
	}
	p.currentPos++
	var elseBranch common.Statement
	switch {
	case !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.OPEN_BRACE:
		elseBranch, err = p.parseNextBlockStatement()
	case !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.IF:
		elseBranch, err = p.parseNextIfStatement()
	default:
		return nil, common.NewRicolError(p.currentPosition(), "Expected '{' or 'if' after 'else'")
	}
	if err != nil {
		return nil, err
	}
	return common.NewIfStatement(ifToken, condition, ifBranch, elseBranch), nil
}

func (p *Parser) parseNextWhileStatement() (common.Statement, error) {
	whileToken := p.tokens[p.currentPos]
	p.currentPos++
	condition, err := p.parseNextCondition(whileToken)
	if err != nil {
		return nil, err
	}
	body, err := p.parseNextBody()
	if err != nil {
		return nil, err
	}
	return common.NewWhileStatement(whileToken, condition, body), nil
}

func (p *Parser) parseNextCondition(keyword common.Token) (common.Expression, error) {
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.OPEN_PAR {
		return nil, common.NewRicolError(p.currentPosition(), fmt.Sprintf("Expected '(' after '%s'", keyword.Lexeme))
	}
	p.currentPos++
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType == common.SEMICOLON {
		return nil, common.NewRicolError(p.currentPosition(),
			fmt.Sprintf("Expected expression after '%s ('", keyword.Lexeme))
	}
	condition, err := p.parseNextExpression()
	if err != nil {
		return nil, err
	}
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.CLOSED_PAR {
		return nil, common.NewRicolError(p.currentPosition(), "Expected ')' after expression")
	}
	p.currentPos++
	return condition, nil
}

func (p *Parser) parseNextBody() (common.Statement, error) {
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.OPEN_BRACE {
		return nil, common.NewRicolError(p.currentPosition(), "Expected block statement")
	}
	return p.parseNextBlockStatement()
}

func (p *Parser) parseNextContinueStatement() (common.Statement, error) {
	continueToken := p.tokens[p.currentPos]
	p.currentPos++
	if err := p.consumeSemicolon("'continue'"); err != nil {
		return nil, err
	}
	return common.NewContinueStatement(continueToken), nil
}

func (p *Parser) parseNextBreakStatement() (common.Statement, error) {
	breakToken := p.tokens[p.currentPos]
	p.currentPos++
	if err := p.consumeSemicolon("'break'"); err != nil {
		return nil, err
	}
	return common.NewBreakStatement(breakToken), nil
}

func (p *Parser) parseNextVarDeclarationStatement() (common.Statement, error) {
	letToken := p.tokens[p.currentPos]
	p.currentPos++
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.IDENTIFIER {
		return nil, common.NewRicolError(p.currentPosition(), "Expected identifier after let")
	}
	varNameToken := p.tokens[p.currentPos]
	p.currentPos++
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.COLON {
		return nil, common.NewRicolError(p.currentPosition(), fmt.Sprintf("Expected ':' after let %s", varNameToken.Lexeme))
	}
	p.currentPos++
	if p.isAtTheEnd() {
		return nil, common.NewRicolError(p.currentPosition(),
			fmt.Sprintf("Expected type after let %s :", varNameToken.Lexeme))
	}
	varTypeToken := p.tokens[p.currentPos]
	varType, ok := common.TypeTokenTypes[varTypeToken.TokenType]
	if !ok {
		return nil, common.NewRicolError(p.currentPosition(),
			fmt.Sprintf("Expected type after let %s :", varNameToken.Lexeme))
	}
	p.currentPos++
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.EQUAL {
		return nil, common.NewRicolError(p.currentPosition(),
			fmt.Sprintf("Expected '=' after let %s : %s", varNameToken.Lexeme, varTypeToken.Lexeme))
	}
	p.currentPos++
	valueExpression, err := p.parseNextExpression()
	if err != nil {
		return nil, err
	}
	if err := p.consumeSemicolon("variable declaration"); err != nil {
		return nil, err
	}
	return common.NewVarDeclarationStatement(letToken, varNameToken, varType, valueExpression), nil
}

func (p *Parser) parseNextFuncDeclarationStatement() (common.Statement, error) {
	funcToken := p.tokens[p.currentPos]
	p.currentPos++
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.IDENTIFIER {
		return nil, common.NewRicolError(p.currentPosition(), "Expected identifier after func")
	}
	funcNameToken := p.tokens[p.currentPos]
	p.currentPos++
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.OPEN_PAR {
		return nil, common.NewRicolError(p.currentPosition(), fmt.Sprintf("Expected '(' after func %s", funcNameToken.Lexeme))
	}
	p.currentPos++
	parameters, err := p.parseNextParameters(funcNameToken)
	if err != nil {
		return nil, err
	}
	returnType := types.Void
	if !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.ARROW {
		p.currentPos++
		returnType, err = p.parseNextType("Expected return type after '->'")
		if err != nil {
			return nil, err
		}
	}
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.OPEN_BRACE {
		return nil, common.NewRicolError(p.currentPosition(), "Expected block statement")
	}
	body, err := p.parseNextBlockStatements()
	if err != nil {
		return nil, err
	}
	return common.NewFuncDeclarationStatement(funcToken, funcNameToken, parameters, returnType, body), nil
}

func (p *Parser) parseNextParameters(funcNameToken common.Token) ([]common.Parameter, error) {
	parameters := []common.Parameter{}
	if !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.CLOSED_PAR {
		p.currentPos++
		return parameters, nil
	}
	for {
		if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.IDENTIFIER {
			return nil, common.NewRicolError(p.currentPosition(),
				fmt.Sprintf("Expected parameter name in func %s", funcNameToken.Lexeme))
		}
		paramNameToken := p.tokens[p.currentPos]
		p.currentPos++
		if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.COLON {
			return nil, common.NewRicolError(p.currentPosition(),
				fmt.Sprintf("Expected ':' after parameter %s", paramNameToken.Lexeme))
		}
		p.currentPos++
		paramType, err := p.parseNextType(fmt.Sprintf("Expected type after parameter %s :", paramNameToken.Lexeme))
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, common.NewParameter(paramNameToken, paramType))
		if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.COMMA {
			break
		}
		p.currentPos++
	}
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.CLOSED_PAR {
		return nil, common.NewRicolError(p.currentPosition(), "Expected ',' or ')' after parameter")
	}
	p.currentPos++
	return parameters, nil
}

func (p *Parser) parseNextType(errorMessage string) (types.Type, error) {
	if p.isAtTheEnd() {
		return nil, common.NewRicolError(p.currentPosition(), errorMessage)
	}
	parsedType, ok := common.TypeTokenTypes[p.tokens[p.currentPos].TokenType]
	if !ok {
		return nil, common.NewRicolError(p.currentPosition(), errorMessage)
	}
	p.currentPos++
	return parsedType, nil
}

func (p *Parser) parseNextReturnStatement() (common.Statement, error) {
	returnToken := p.tokens[p.currentPos]
	p.currentPos++
	var valueExpression common.Expression
	if !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType != common.SEMICOLON {
		var err error
		valueExpression, err = p.parseNextExpression()
		if err != nil {
			return nil, err
		}
	}
	if err := p.consumeSemicolon("'return'"); err != nil {
		return nil, err
	}
	return common.NewReturnStatement(returnToken, valueExpression), nil
}

func (p *Parser) parseNextExpressionStatement() (common.Statement, error) {
	expression, err := p.parseNextExpression()
	if err != nil {
		return nil, err
	}
	if err := p.consumeSemicolon("expression"); err != nil {
		return nil, err
	}
	return common.NewExpressionStatement(expression), nil
}

func (p *Parser) consumeSemicolon(after string) error {
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.SEMICOLON {
		return common.NewRicolError(p.currentPosition(), fmt.Sprintf("Expected ';' after %s", after))
	}
	p.currentPos++
	return nil
}

func (p *Parser) parseNextExpression() (common.Expression, error) {
	return p.parseNextVarAssignmentExpression()
}

func (p *Parser) parseNextVarAssignmentExpression() (common.Expression, error) {
	nextPosAtTheEnd := p.currentPos+1 >= len(p.tokens) || p.tokens[p.currentPos+1].TokenType == common.EOF
	if !nextPosAtTheEnd && p.tokens[p.currentPos].TokenType == common.IDENTIFIER &&
		p.tokens[p.currentPos+1].TokenType == common.EQUAL {
		nameToken := p.tokens[p.currentPos]
		p.currentPos += 2
		valueExpression, err := p.parseNextExpression()
		if err != nil {
			return nil, err
		}
		return common.NewVarAssignmentExpression(nameToken, valueExpression), nil
	}
	return p.parseNextLogicOr()
}

func (p *Parser) parseNextLogicOr() (common.Expression, error) {
	currentExpression, err := p.parseNextLogicAnd()
	if err != nil {
		return nil, err
	}
	for !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.OR {
		operator := p.tokens[p.currentPos]
		p.currentPos++
		rightExpression, err := p.parseNextLogicAnd()
		if err != nil {
			return nil, err
		}
		currentExpression = common.NewBinaryExpression(currentExpression, operator, rightExpression)
	}
	return currentExpression, nil
}

func (p *Parser) parseNextLogicAnd() (common.Expression, error) {
	currentExpression, err := p.parseNextLogicNot()
	if err != nil {
		return nil, err
	}
	for !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.AND {
		operator := p.tokens[p.currentPos]
		p.currentPos++
		rightExpression, err := p.parseNextLogicNot()
		if err != nil {
			return nil, err
		}
		currentExpression = common.NewBinaryExpression(currentExpression, operator, rightExpression)
	}
	return currentExpression, nil
}

func (p *Parser) parseNextLogicNot() (common.Expression, error) {
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.NOT {
		return p.parseNextEquality()
	}
	operator := p.tokens[p.currentPos]
	p.currentPos++
	expression, err := p.parseNextLogicNot()
	if err != nil {
		return nil, err
	}
	return common.NewUnaryExpression(operator, expression), nil
}

func (p *Parser) parseNextEquality() (common.Expression, error) {
	currentExpression, err := p.parseNextComparison()
	if err != nil {
		return nil, err
	}
	validTokenTypes := []common.TokenType{common.DOUBLE_EQUAL, common.NOT_EQUAL}
	for !p.isAtTheEnd() && slices.Contains(validTokenTypes, p.tokens[p.currentPos].TokenType) {
		operator := p.tokens[p.currentPos]
		p.currentPos++
		rightExpression, err := p.parseNextComparison()
		if err != nil {
			return nil, err
		}
		currentExpression = common.NewBinaryExpression(currentExpression, operator, rightExpression)
	}
	return currentExpression, nil
}

func (p *Parser) parseNextComparison() (common.Expression, error) {
	currentExpression, err := p.parseNextAddition()
	if err != nil {
		return nil, err
	}
	validTokenTypes := []common.TokenType{common.LESS, common.LESS_EQUAL, common.GREATER, common.GREATER_EQUAL}
	for !p.isAtTheEnd() && slices.Contains(validTokenTypes, p.tokens[p.currentPos].TokenType) {
		operator := p.tokens[p.currentPos]
		p.currentPos++
		rightExpression, err := p.parseNextAddition()
		if err != nil {
			return nil, err
		}
		currentExpression = common.NewBinaryExpression(currentExpression, operator, rightExpression)
	}
	return currentExpression, nil
}

func (p *Parser) parseNextAddition() (common.Expression, error) {
	currentExpression, err := p.parseNextTerm()
	if err != nil {
		return nil, err
	}
	validTokenTypes := []common.TokenType{common.PLUS, common.MINUS}
	for !p.isAtTheEnd() && slices.Contains(validTokenTypes, p.tokens[p.currentPos].TokenType) {
		operator := p.tokens[p.currentPos]
		p.currentPos++
		rightExpression, err := p.parseNextTerm()
		if err != nil {
			return nil, err
		}
		currentExpression = common.NewBinaryExpression(currentExpression, operator, rightExpression)
	}
	return currentExpression, nil
}

func (p *Parser) parseNextTerm() (common.Expression, error) {
	currentExpression, err := p.parseNextFactor()
	if err != nil {
		return nil, err
	}
	validTokenTypes := []common.TokenType{common.STAR, common.SLASH, common.DOUBLE_SLASH, common.PERCENTAGE}
	for !p.isAtTheEnd() && slices.Contains(validTokenTypes, p.tokens[p.currentPos].TokenType) {
		operator := p.tokens[p.currentPos]
		p.currentPos++
		rightExpression, err := p.parseNextFactor()
		if err != nil {
			return nil, err
		}
		currentExpression = common.NewBinaryExpression(currentExpression, operator, rightExpression)
	}
	return currentExpression, nil
}

func (p *Parser) parseNextFactor() (common.Expression, error) {
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.MINUS {
		return p.parseNextPower()
	}
	operator := p.tokens[p.currentPos]
	p.currentPos++
	expression, err := p.parseNextFactor()
	if err != nil {
		return nil, err
	}
	return common.NewUnaryExpression(operator, expression), nil
}

func (p *Parser) parseNextPower() (common.Expression, error) {
	currentPrimary, err := p.parseNextPrimary()
	if err != nil {
		return nil, err
	}
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.DOUBLE_STAR {
		return currentPrimary, nil
	}
	operator := p.tokens[p.currentPos]
	p.currentPos++
	rightExpression, err := p.parseNextFactor()
	if err != nil {
		return nil, err
	}
	return common.NewBinaryExpression(currentPrimary, operator, rightExpression), nil
}

func (p *Parser) parseNextPrimary() (common.Expression, error) {
	if !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.OPEN_PAR {
		expression, err := p.parseNextGroupingExpression()
		if err != nil {
			return nil, err
		}
		return expression, nil
	}
	if !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.IDENTIFIER {
		nameToken := p.tokens[p.currentPos]
		p.currentPos++
		if !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.OPEN_PAR {
			return p.parseNextCallExpression(nameToken)
		}
		return common.NewVariableExpression(nameToken), nil
	}
	validTokenTypes := []common.TokenType{common.INTEGER, common.FLOAT, common.STRING, common.TRUE, common.FALSE}
	if p.isAtTheEnd() || !slices.Contains(validTokenTypes, p.tokens[p.currentPos].TokenType) {
		return nil, common.NewRicolError(p.currentPosition(), "Invalid primary expression")
	}
	token := p.tokens[p.currentPos]
	value, err := parseLiteralValue(token)
	if err != nil {
		return nil, err
	}
	p.currentPos++
	return common.NewLiteralExpression(token, value), nil
}

func parseLiteralValue(token common.Token) (types.Value, error) {
	switch token.TokenType {
	case common.INTEGER:
		value, err := strconv.ParseInt(token.Lexeme, 10, 64)
		if err != nil {
			return nil, common.NewRicolError(token.Position, fmt.Sprintf("Invalid integer: %s", token.Lexeme))
		}
		return types.NewInteger(value), nil
	case common.FLOAT:
		value, err := strconv.ParseFloat(token.Lexeme, 64)
		if err != nil {
			return nil, common.NewRicolError(token.Position, fmt.Sprintf("Invalid float: %s", token.Lexeme))
		}
		return types.NewFloat(value), nil
	case common.STRING:
		return parseStringValue(token)
	case common.TRUE:
		return types.NewBoolean(true), nil
	case common.FALSE:
		return types.NewBoolean(false), nil
	default:
		return nil, common.NewRicolError(token.Position, fmt.Sprintf("Invalid literal: %s", token.Lexeme))
	}
}

// Quita las comillas del lexema y reemplaza cada secuencia de escape por el carácter que representa.
func parseStringValue(token common.Token) (types.Value, error) {
	lexeme := token.Lexeme
	lexemeChars := []rune(lexeme)
	if len(lexemeChars) < 2 || lexemeChars[0] != '"' || lexemeChars[len(lexemeChars)-1] != '"' {
		return nil, common.NewRicolError(token.Position, fmt.Sprintf("Invalid string: %s", lexeme))
	}
	valueChars := []rune{}
	escaping := false
	for _, char := range lexemeChars[1 : len(lexemeChars)-1] {
		switch {
		case escaping:
			escapedChar, ok := common.EscapeSequences[char]
			if !ok {
				return nil, common.NewRicolError(token.Position, fmt.Sprintf("Invalid string: %s", lexeme))
			}
			valueChars = append(valueChars, escapedChar)
			escaping = false
		case char == '\\':
			escaping = true
		case char == '"':
			return nil, common.NewRicolError(token.Position, fmt.Sprintf("Invalid string: %s", lexeme))
		default:
			valueChars = append(valueChars, char)
		}
	}
	if escaping {
		return nil, common.NewRicolError(token.Position, fmt.Sprintf("Invalid string: %s", lexeme))
	}
	return types.NewString(string(valueChars)), nil
}

func (p *Parser) parseNextCallExpression(nameToken common.Token) (common.Expression, error) {
	p.currentPos++
	arguments := []common.Expression{}
	if !p.isAtTheEnd() && p.tokens[p.currentPos].TokenType == common.CLOSED_PAR {
		p.currentPos++
		return common.NewCallExpression(nameToken, arguments), nil
	}
	for {
		argument, err := p.parseNextExpression()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, argument)
		if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.COMMA {
			break
		}
		p.currentPos++
	}
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.CLOSED_PAR {
		return nil, common.NewRicolError(p.currentPosition(), "Expected ',' or ')' after argument")
	}
	p.currentPos++
	return common.NewCallExpression(nameToken, arguments), nil
}

func (p *Parser) parseNextGroupingExpression() (*common.GroupingExpression, error) {
	openPar := p.tokens[p.currentPos]
	p.currentPos++
	expression, err := p.parseNextExpression()
	if err != nil {
		return nil, err
	}
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.CLOSED_PAR {
		return nil, common.NewRicolError(p.currentPosition(), "Grouping expression without close")
	}
	p.currentPos++
	return common.NewGroupingExpression(openPar, expression), nil
}

func (p *Parser) currentPosition() common.Position {
	if p.currentPos < len(p.tokens) {
		return p.tokens[p.currentPos].Position
	}
	if len(p.tokens) == 0 {
		return common.Position{Line: 1, Column: 1}
	}
	return p.tokens[len(p.tokens)-1].Position
}

func (p *Parser) isAtTheEnd() bool {
	return p.currentPos == len(p.tokens) || p.tokens[p.currentPos].TokenType == common.EOF
}
