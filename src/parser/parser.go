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
	statement, err := p.parseNextExpressionStatement()
	if err != nil {
		return nil, err
	}
	return statement, nil
}

func (p *Parser) parseNextExpressionStatement() (*common.ExpressionStatement, error) {
	expression, err := p.parseNextExpression()
	if err != nil {
		return nil, err
	}
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.SEMICOLON {
		return nil, common.NewRicolError(p.currentPosition(), "Expected ';' after expression")
	}
	p.currentPos++
	return common.NewExpressionStatement(expression), nil
}

func (p *Parser) parseNextExpression() (common.Expression, error) {
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
