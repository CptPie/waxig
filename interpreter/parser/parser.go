package parser

import (
	"strconv"
	"waixg/interpreter/ast"
	"waixg/interpreter/errors"
	"waixg/interpreter/lexer"
	"waixg/interpreter/token"
)

const enableTraces = false

// Precedence is the precedence of operators
//
// The higher the value, the higher the precedence (the more important the operator).
// The highest precedence is `CALL` with a value of 7 and represents function calls
// The lowest precedence is `LOWEST` with a value of 1 and is the default value if no operator is found
type Precedence int

// Define the precedence of operators
const (
	_ Precedence = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	EXPONENT    // ^
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

// Precedence table associating token types with their precedence
var precedences = map[token.TokenType]Precedence{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.GTEQ:     LESSGREATER,
	token.LTEQ:     LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.HAT:      EXPONENT,
	token.LPAREN:   CALL,
}

type (
	prefixParseFn func() ast.Expression
	// for infixParseFn, the first parameter is the left-hand side of the expression
	infixParseFn func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	errors []error

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []error{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.HAT, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GTEQ, p.parseInfixExpression)
	p.registerInfix(token.LTEQ, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// Read two tokens, so curToken and peekToken are both set
	// curToken will be the first token in the input
	// peekToken will be the second token in the input
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) Errors() []error {
	return p.errors
}

func (p *Parser) addError(err error) {
	p.errors = append(p.errors, err)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()

	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() ast.Statement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the next token is of the expected type
// if it is, it advances the parser and returns true
// if it is not, it adds an error and returns false
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.addError(&errors.PeekTypeMismatch{
			Expected: t,
			Actual:   p.peekToken.Type,
		})
		return false
	}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	if enableTraces {
		defer untrace(trace("parseExpressionStatement"))
	}

	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	if enableTraces {
		defer untrace(trace("parseExpression"))
	}

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.addError(&errors.NoPrefixParseFnError{
			TokenType: p.curToken.Type,
		})
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	if enableTraces {
		defer untrace(trace("parseIntegerLiteral"))
	}
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.addError(&errors.InvalidIntegerLiteral{
			Literal: p.curToken.Literal,
		})
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	if enableTraces {
		defer untrace(trace("parsePrefixExpression"))
	}
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) peekPrecedence() Precedence {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() Precedence {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	if enableTraces {
		defer untrace(trace("parseIfExpression"))
	}

	expression := &ast.IfExpression{Token: p.curToken}

	//we expect a `(` after the `if` keyword
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// advance the tokens
	p.nextToken()

	// parse the condition (stuff inside parenthesis)
	expression.Condition = p.parseExpression(LOWEST)

	// we expect a `)` after the condition
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// after the closing parenthesis, we expect a `{`
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// parse the consequence
	// parseBlockStatement() will consume the trailing `}`
	expression.Consequence = p.parseBlockStatement()

	// if we have an `else` keyword, we parse the alternative
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		// we expect a `{` after the `else` keyword
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		// parse the alternative
		// parseBlockStatement() will consume the trailing `}`
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	if enableTraces {
		defer untrace(trace("parseBlockStatement"))
	}

	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	// everything until the next `}` is part of the block
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	if enableTraces {
		defer untrace(trace("parseFunctionLiteral"))
	}

	lit := &ast.FunctionLiteral{Token: p.curToken}

	// we expect a `(` after the `fn` keyword
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// parse the function parameters (list of identifiers/expressions)
	// parseFunctionParameters() will consume the trailing `)`
	lit.Parameters = p.parseFunctionParameters()

	// we expect a `{` after the parameters
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// parse the function body (block statement)
	// parseBlockStatement() will consume the trailing `}`
	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// if we have a `)` after the `(`, we have no parameters
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	// advance the tokens
	p.nextToken()

	// parse the first parameter
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// if we have a `,` after the first parameter, we have more parameters
	for p.peekTokenIs(token.COMMA) {
		// consume the `,`
		p.nextToken()
		// advance the token to the next parameter
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// we expect a `)` after the last parameter
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	if enableTraces {
		defer untrace(trace("parseCallExpression"))
	}

	exp := &ast.CallExpression{Token: p.curToken, Function: function}

	// parse the call arguments (list of expressions)
	// parseCallArguments() will consume the trailing `)`
	exp.Arguments = p.parseCallArguments()

	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	// if we have a `)` after the `(`, we have no arguments
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	// advance the tokens
	p.nextToken()

	// parse the first argument
	args = append(args, p.parseExpression(LOWEST))

	// if we have a `,` after the first argument, we have more arguments
	for p.peekTokenIs(token.COMMA) {
		// consume the `,`
		p.nextToken()
		// advance the token to the next argument
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	// we expect a `)` after the last argument
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}
