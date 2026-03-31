package parser

import (
	"fmt"
	"strconv"

	"github.com/TRC-Loop/ccolon/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peek() lexer.Token {
	if p.pos+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) advance() lexer.Token {
	tok := p.current()
	p.pos++
	return tok
}

func (p *Parser) expect(t lexer.TokenType) (lexer.Token, error) {
	tok := p.current()
	if tok.Type != t {
		return tok, fmt.Errorf("line %d:%d: expected '%s', got '%s'",
			tok.Line, tok.Col, t.String(), tok.Literal)
	}
	p.advance()
	return tok, nil
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.current().Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) Parse() (*Program, error) {
	prog := &Program{}
	for p.current().Type != lexer.TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Stmts = append(prog.Stmts, stmt)
		}
	}
	return prog, nil
}

func (p *Parser) parseStatement() (Stmt, error) {
	// skip stray semicolons
	for p.current().Type == lexer.TOKEN_SEMICOLON {
		p.advance()
	}
	if p.current().Type == lexer.TOKEN_EOF {
		return nil, nil
	}

	switch p.current().Type {
	case lexer.TOKEN_IMPORT:
		return p.parseImport()
	case lexer.TOKEN_VAR:
		return p.parseVarDecl()
	case lexer.TOKEN_FUNCTION:
		return p.parseFuncDecl()
	case lexer.TOKEN_IF:
		return p.parseIfStmt()
	case lexer.TOKEN_WHILE:
		return p.parseWhileStmt()
	case lexer.TOKEN_FOR:
		return p.parseForInStmt()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStmt()
	case lexer.TOKEN_BREAK:
		tok := p.advance()
		return &BreakStmt{P: Position{tok.Line, tok.Col}}, nil
	case lexer.TOKEN_CONTINUE:
		tok := p.advance()
		return &ContinueStmt{P: Position{tok.Line, tok.Col}}, nil
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseImport() (*ImportStmt, error) {
	tok := p.advance() // consume 'import'
	name, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	return &ImportStmt{Module: name.Literal, P: Position{tok.Line, tok.Col}}, nil
}

func (p *Parser) parseVarDecl() (*VarDecl, error) {
	tok := p.advance() // consume 'var'

	// type name
	typeTok := p.current()
	if !lexer.IsTypeKeyword(typeTok.Type) {
		return nil, fmt.Errorf("line %d:%d: expected type, got '%s'", typeTok.Line, typeTok.Col, typeTok.Literal)
	}
	p.advance()

	name, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_ASSIGN); err != nil {
		return nil, err
	}

	value, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	return &VarDecl{
		TypeName: typeTok.Literal,
		Name:     name.Literal,
		Value:    value,
		P:        Position{tok.Line, tok.Col},
	}, nil
}

func (p *Parser) parseFuncDecl() (*FuncDecl, error) {
	tok := p.advance() // consume 'function'

	name, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}

	var params []Param
	for p.current().Type != lexer.TOKEN_RPAREN {
		if len(params) > 0 {
			if _, err := p.expect(lexer.TOKEN_COMMA); err != nil {
				return nil, err
			}
		}
		paramType := p.current()
		if !lexer.IsTypeKeyword(paramType.Type) {
			return nil, fmt.Errorf("line %d:%d: expected parameter type, got '%s'",
				paramType.Line, paramType.Col, paramType.Literal)
		}
		p.advance()
		paramName, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		params = append(params, Param{TypeName: paramType.Literal, Name: paramName.Literal})
	}
	p.advance() // consume ')'

	// optional return type
	retType := ""
	if lexer.IsTypeKeyword(p.current().Type) {
		retType = p.current().Literal
		p.advance()
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &FuncDecl{
		Name:       name.Literal,
		Params:     params,
		ReturnType: retType,
		Body:       body,
		P:          Position{tok.Line, tok.Col},
	}, nil
}

func (p *Parser) parseIfStmt() (*IfStmt, error) {
	tok := p.advance() // consume 'if'

	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	cond, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseBody []Stmt
	if p.current().Type == lexer.TOKEN_ELSE {
		p.advance()
		if p.current().Type == lexer.TOKEN_IF {
			// else if
			elseIf, err := p.parseIfStmt()
			if err != nil {
				return nil, err
			}
			elseBody = []Stmt{elseIf}
		} else {
			elseBody, err = p.parseBlock()
			if err != nil {
				return nil, err
			}
		}
	}

	return &IfStmt{Cond: cond, Body: body, ElseBody: elseBody, P: Position{tok.Line, tok.Col}}, nil
}

func (p *Parser) parseWhileStmt() (*WhileStmt, error) {
	tok := p.advance() // consume 'while'

	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	cond, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &WhileStmt{Cond: cond, Body: body, P: Position{tok.Line, tok.Col}}, nil
}

func (p *Parser) parseForInStmt() (*ForInStmt, error) {
	tok := p.advance() // consume 'for'

	varName, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_IN); err != nil {
		return nil, err
	}

	iterable, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ForInStmt{
		VarName:  varName.Literal,
		Iterable: iterable,
		Body:     body,
		P:        Position{tok.Line, tok.Col},
	}, nil
}

func (p *Parser) parseReturnStmt() (*ReturnStmt, error) {
	tok := p.advance() // consume 'return'

	// return with no value
	if p.current().Type == lexer.TOKEN_RBRACE || p.current().Type == lexer.TOKEN_EOF ||
		p.current().Type == lexer.TOKEN_SEMICOLON {
		return &ReturnStmt{P: Position{tok.Line, tok.Col}}, nil
	}

	value, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	return &ReturnStmt{Value: value, P: Position{tok.Line, tok.Col}}, nil
}

func (p *Parser) parseExpressionStatement() (Stmt, error) {
	pos := Position{p.current().Line, p.current().Col}
	expr, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	// check for assignment: ident = expr or expr[index] = expr
	if p.current().Type == lexer.TOKEN_ASSIGN {
		p.advance()
		value, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		return &AssignStmt{Target: expr, Value: value, P: pos}, nil
	}

	return &ExprStmt{Expression: expr, P: pos}, nil
}

func (p *Parser) parseBlock() ([]Stmt, error) {
	if _, err := p.expect(lexer.TOKEN_LBRACE); err != nil {
		return nil, err
	}

	var stmts []Stmt
	for p.current().Type != lexer.TOKEN_RBRACE && p.current().Type != lexer.TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	if _, err := p.expect(lexer.TOKEN_RBRACE); err != nil {
		return nil, err
	}

	return stmts, nil
}

// Pratt parser for expressions

const (
	precNone       = 0
	precOr         = 1
	precAnd        = 2
	precEquality   = 3
	precComparison = 4
	precTerm       = 5
	precFactor     = 6
	precUnary      = 7
	precCall       = 8
)

func (p *Parser) parseExpression(minPrec int) (Expr, error) {
	left, err := p.parsePrefix()
	if err != nil {
		return nil, err
	}

	for {
		prec := p.infixPrecedence()
		if prec <= minPrec {
			break
		}

		left, err = p.parseInfix(left, prec)
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func (p *Parser) infixPrecedence() int {
	switch p.current().Type {
	case lexer.TOKEN_OR:
		return precOr
	case lexer.TOKEN_AND:
		return precAnd
	case lexer.TOKEN_EQ, lexer.TOKEN_NEQ:
		return precEquality
	case lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_GTE:
		return precComparison
	case lexer.TOKEN_PLUS, lexer.TOKEN_MINUS:
		return precTerm
	case lexer.TOKEN_STAR, lexer.TOKEN_SLASH, lexer.TOKEN_PERCENT:
		return precFactor
	case lexer.TOKEN_DOT, lexer.TOKEN_LPAREN, lexer.TOKEN_LBRACKET:
		return precCall
	default:
		return precNone
	}
}

func (p *Parser) parsePrefix() (Expr, error) {
	tok := p.current()

	switch tok.Type {
	case lexer.TOKEN_INT_LIT:
		p.advance()
		val, err := strconv.ParseInt(tok.Literal, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d:%d: invalid integer '%s'", tok.Line, tok.Col, tok.Literal)
		}
		return &IntLiteral{Value: val, P: Position{tok.Line, tok.Col}}, nil

	case lexer.TOKEN_FLOAT_LIT:
		p.advance()
		val, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d:%d: invalid float '%s'", tok.Line, tok.Col, tok.Literal)
		}
		return &FloatLiteral{Value: val, P: Position{tok.Line, tok.Col}}, nil

	case lexer.TOKEN_STRING_LIT:
		p.advance()
		return &StringLiteral{Value: tok.Literal, P: Position{tok.Line, tok.Col}}, nil

	case lexer.TOKEN_TRUE:
		p.advance()
		return &BoolLiteral{Value: true, P: Position{tok.Line, tok.Col}}, nil

	case lexer.TOKEN_FALSE:
		p.advance()
		return &BoolLiteral{Value: false, P: Position{tok.Line, tok.Col}}, nil

	case lexer.TOKEN_IDENT:
		p.advance()
		return &Identifier{Name: tok.Literal, P: Position{tok.Line, tok.Col}}, nil

	case lexer.TOKEN_MINUS, lexer.TOKEN_NOT:
		p.advance()
		operand, err := p.parseExpression(precUnary)
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: tok.Type, Operand: operand, P: Position{tok.Line, tok.Col}}, nil

	case lexer.TOKEN_LPAREN:
		p.advance()
		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
			return nil, err
		}
		return expr, nil

	case lexer.TOKEN_LBRACKET:
		return p.parseListLiteral()

	case lexer.TOKEN_RANGE:
		return p.parseRange()

	case lexer.TOKEN_FIXED:
		return p.parseFixed()

	default:
		return nil, fmt.Errorf("line %d:%d: unexpected token '%s'", tok.Line, tok.Col, tok.Literal)
	}
}

func (p *Parser) parseInfix(left Expr, prec int) (Expr, error) {
	tok := p.current()

	switch tok.Type {
	case lexer.TOKEN_DOT:
		p.advance()
		method, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		if p.current().Type == lexer.TOKEN_LPAREN {
			p.advance()
			args, err := p.parseArgList()
			if err != nil {
				return nil, err
			}
			return &MethodCallExpr{
				Object: left,
				Method: method.Literal,
				Args:   args,
				P:      Position{tok.Line, tok.Col},
			}, nil
		}
		// property access (no parens) - treat as method call with no args for now
		return &MethodCallExpr{
			Object: left,
			Method: method.Literal,
			Args:   nil,
			P:      Position{tok.Line, tok.Col},
		}, nil

	case lexer.TOKEN_LPAREN:
		p.advance()
		args, err := p.parseArgList()
		if err != nil {
			return nil, err
		}
		return &CallExpr{Callee: left, Args: args, P: Position{tok.Line, tok.Col}}, nil

	case lexer.TOKEN_LBRACKET:
		p.advance()
		index, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.TOKEN_RBRACKET); err != nil {
			return nil, err
		}
		return &IndexExpr{Object: left, Index: index, P: Position{tok.Line, tok.Col}}, nil

	default:
		// binary operator
		p.advance()
		right, err := p.parseExpression(prec)
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Left: left, Op: tok.Type, Right: right, P: Position{tok.Line, tok.Col}}, nil
	}
}

func (p *Parser) parseArgList() ([]Expr, error) {
	var args []Expr
	if p.current().Type == lexer.TOKEN_RPAREN {
		p.advance()
		return args, nil
	}
	for {
		arg, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.current().Type == lexer.TOKEN_COMMA {
			p.advance()
		} else {
			break
		}
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}
	return args, nil
}

func (p *Parser) parseListLiteral() (Expr, error) {
	tok := p.advance() // consume '['
	var elements []Expr
	if p.current().Type == lexer.TOKEN_RBRACKET {
		p.advance()
		return &ListLiteral{Elements: elements, P: Position{tok.Line, tok.Col}}, nil
	}
	for {
		elem, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
		if p.current().Type == lexer.TOKEN_COMMA {
			p.advance()
		} else {
			break
		}
	}
	if _, err := p.expect(lexer.TOKEN_RBRACKET); err != nil {
		return nil, err
	}
	return &ListLiteral{Elements: elements, P: Position{tok.Line, tok.Col}}, nil
}

func (p *Parser) parseRange() (Expr, error) {
	tok := p.advance() // consume 'range'
	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	first, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	var start, end Expr
	if p.current().Type == lexer.TOKEN_COMMA {
		p.advance()
		end, err = p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		start = first
	} else {
		start = &IntLiteral{Value: 0, P: Position{tok.Line, tok.Col}}
		end = first
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}
	return &RangeExpr{Start: start, End: end, P: Position{tok.Line, tok.Col}}, nil
}

func (p *Parser) parseFixed() (Expr, error) {
	tok := p.advance() // consume 'fixed'
	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_LBRACKET); err != nil {
		return nil, err
	}
	var elements []Expr
	if p.current().Type != lexer.TOKEN_RBRACKET {
		for {
			elem, err := p.parseExpression(0)
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)
			if p.current().Type == lexer.TOKEN_COMMA {
				p.advance()
			} else {
				break
			}
		}
	}
	if _, err := p.expect(lexer.TOKEN_RBRACKET); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}
	return &FixedArrayLiteral{Elements: elements, P: Position{tok.Line, tok.Col}}, nil
}
