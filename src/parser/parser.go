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
	return p.parseNextExpressionStatement()
}

func (p *Parser) parseNextExpressionStatement() (common.ExpressionStatement, error) {
	expression, err := p.parseNextExpression()
	if err != nil {
		return common.ExpressionStatement{}, err
	}
	currentToken := p.tokens[p.currentPos]
	if currentToken.TokenType != common.SEMICOLON {
		return common.ExpressionStatement{}, fmt.Errorf("Expected ';' after expression")
	}
	p.currentPos++
	return common.ExpressionStatement{Expression: expression}, nil
}

func (p *Parser) parseNextExpression() (common.Expression, error) {
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
		currentExpression = common.BinaryExpression{
			LeftExpression:  currentExpression,
			Operator:        operator,
			RightExpression: rightExpression,
		}
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
		currentExpression = common.BinaryExpression{
			LeftExpression:  currentExpression,
			Operator:        operator,
			RightExpression: rightExpression,
		}
	}
	return currentExpression, nil
}

func (p *Parser) parseNextFactor() (common.Expression, error) {
	if p.tokens[p.currentPos].TokenType != common.MINUS {
		return p.parseNextPower()
	}
	operator := p.tokens[p.currentPos]
	p.currentPos++
	expression, err := p.parseNextFactor()
	if err != nil {
		return nil, err
	}
	return common.UnaryExpression{
		Operator:   operator,
		Expression: expression,
	}, nil
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
	return common.BinaryExpression{
		LeftExpression:  currentPrimary,
		Operator:        operator,
		RightExpression: rightExpression,
	}, nil
}

func (p *Parser) parseNextPrimary() (common.Expression, error) {
	if p.tokens[p.currentPos].TokenType == common.OPEN_PAR {
		return p.parseNextGroupingExpression()
	}
	validTokenTypes := []common.TokenType{common.INTEGER, common.FLOAT}
	if p.isAtTheEnd() || !slices.Contains(validTokenTypes, p.tokens[p.currentPos].TokenType) {
		return nil, fmt.Errorf("Invalid primary expression")
	}
	token := p.tokens[p.currentPos]
	value, err := parseLiteralValue(token)
	if err != nil {
		return nil, err
	}
	p.currentPos++
	return common.LiteralExpression{
		Token: token,
		Value: value,
	}, nil
}

func parseLiteralValue(token common.Token) (types.Number, error) {
	switch token.TokenType {
	case common.INTEGER:
		value, err := strconv.ParseInt(token.Lexeme, 10, 64)
		if err != nil {
			return types.Number{}, fmt.Errorf("Invalid integer: %s", token.Lexeme)
		}
		return types.NewInteger(value), nil
	case common.FLOAT:
		value, err := strconv.ParseFloat(token.Lexeme, 64)
		if err != nil {
			return types.Number{}, fmt.Errorf("Invalid float: %s", token.Lexeme)
		}
		return types.NewFloat(value), nil
	default:
		return types.Number{}, fmt.Errorf("Invalid literal: %s", token.Lexeme)
	}
}

func (p *Parser) parseNextGroupingExpression() (common.GroupingExpression, error) {
	p.currentPos++
	expression, err := p.parseNextExpression()
	if err != nil {
		return common.GroupingExpression{}, err
	}
	if p.isAtTheEnd() || p.tokens[p.currentPos].TokenType != common.CLOSED_PAR {
		return common.GroupingExpression{}, fmt.Errorf("Grouping expression without close")
	}
	p.currentPos++
	return common.GroupingExpression{
		Expression: expression,
	}, nil
}

func (p *Parser) isAtTheEnd() bool {
	return p.tokens[p.currentPos].TokenType == common.EOF
}
