package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []string

	// to check if curToken.Type has associated prefix or infix
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

// measure precedence of an exrpession -> * should have higher precendence than + or -
// ,higher the num 0 -> 7, stronger the precedence
const (
	_           int = iota
	LOWEST          // wear precedence
	EQUALS          // ==
	LESSGREATER     // > or <
	SUM             // + or -
	PRODUCT         // * or /
	PREFIX          // -X or !X
	CALL            // foo(x)
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)

	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	p.registerPrefix(token.IF, p.parseIfExpression)

	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.registerPrefix(token.STRING, p.parseStringLiteral)

	// move 2 to get curr and peak
	p.nextToken() // curr 0, next 1, we don't want 0
	p.nextToken() // curr 1, next 2
	return p
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectedPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function} // ( and function name e.g., foo
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) { // while next token is " ,"
		p.nextToken() // ","
		p.nextToken() // new token
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectedPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectedPeek(token.LPAREN) {
		return nil
	}
	// move from (x, y) <- to this RPAREN
	lit.Parameters = p.parseFunctionParameters()
	// next one should be {
	if !p.expectedPeek(token.LBRACE) {
		return nil
	}
	// inside of braces
	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectedPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectedPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	// remember that after each peek we move to the next token in the exptectedPeek function
	if !p.expectedPeek(token.RPAREN) {
		return nil
	}

	if !p.expectedPeek(token.LBRACE) {
		return nil
	}

	// moves parser to the end of the if block
	expression.Consequence = p.parseBlockStatement()

	// if else exists, we are there now
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectedPeek(token.LBRACE) {
			return nil
		}
		// same logic as for if block
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) { // untill { x + y; } <- here or till input ends
		stmt := p.parseStatement() // makes a statement node with the contents of each line e.g., x; or x + y;
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectedPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	// in middle of an expression
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left, // int literal from parseExpression
	}
	// how strong are we
	precedence := p.curPrecedence()
	p.nextToken()

	//
	expression.Right = p.parseExpression(precedence)

	return expression
}

// is something going to have higher precedence than me?
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// what is my current precedence?
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	// INIT root node
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
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

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	// associated prefix of current
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	// curr is left node expression like number 1
	leftExp := prefix()
	// not end of line and next has higher precedence
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// associated infix of next
		infix := p.infixParseFns[p.peekToken.Type]

		// no next infix, int is the left one
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		// if we are e.g., + we go after 1 -> 1 +,  becomes the left node and 1 below it + -> 1
		leftExp = infix(leftExp)
	}
	// returns left node
	return leftExp
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("I could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectedPeek(token.IDENT) {
		return nil
	}

	// we know a var is named
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// var not assigned to value
	if !p.expectedPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	// go till end of expression thing let foo = 'bar';
	for !p.curTokenIs(token.SEMICOLON) {
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

func (p *Parser) expectedPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken() // we want to take out the peek token
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
