package parser

import (
	"fmt"
	"math/big"
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
		return p.parseVarDecl(false)
	case lexer.TOKEN_CONST:
		return p.parseVarDecl(true)
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
	case lexer.TOKEN_CLASS:
		return p.parseClassDecl()
	case lexer.TOKEN_TRY:
		return p.parseTryCatch()
	case lexer.TOKEN_THROW:
		return p.parseThrow()
	case lexer.TOKEN_WITH:
		return p.parseWithStmt()
	default:
		// better error for mistyped keywords: if an identifier is followed by something
		// that looks like a function or block declaration, the user likely mistyped a keyword
		if p.current().Type == lexer.TOKEN_IDENT {
			next := p.peek()
			if next.Type == lexer.TOKEN_IDENT || next.Type == lexer.TOKEN_LBRACE || next.Type == lexer.TOKEN_LPAREN {
				// check if it looks like a mistyped keyword
				lit := p.current().Literal
				if lit != "true" && lit != "false" && lit != "nil" && lit != "None" {
					suggestions := map[string]string{
						"func": "function", "fn": "function", "fnction": "function",
						"funciton": "function", "fucntion": "function",
						"clas": "class", "classs": "class",
						"imprt": "import", "impor": "import",
						"retrun": "return", "reutrn": "return",
						"whlie": "while", "whlile": "while",
					}
					if suggestion, ok := suggestions[lit]; ok {
						return nil, fmt.Errorf("line %d:%d: unknown keyword '%s', did you mean '%s'?",
							p.current().Line, p.current().Col, lit, suggestion)
					}
				}
			}
		}
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseImport() (*ImportStmt, error) {
	tok := p.advance() // consume 'import'

	// import "file.ccl" or import modulename
	if p.current().Type == lexer.TOKEN_STRING_LIT {
		path := p.advance()
		return &ImportStmt{Module: path.Literal, IsFile: true, P: Position{tok.Line, tok.Col}}, nil
	}

	name, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	return &ImportStmt{Module: name.Literal, P: Position{tok.Line, tok.Col}}, nil
}

func (p *Parser) parseVarDecl(isConst bool) (*VarDecl, error) {
	tok := p.advance() // consume 'var' or 'const'

	// type name (built-in type or class name)
	typeTok := p.current()
	if !lexer.IsClassTypeKeyword(typeTok.Type) {
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
		IsConst:  isConst,
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

	params, err := p.parseParamList()
	if err != nil {
		return nil, err
	}
	p.advance() // consume ')'

	// optional return type
	retType := ""
	if lexer.IsClassTypeKeyword(p.current().Type) {
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
			// too large for int64, use sint (arbitrary precision)
			bigVal := new(big.Int)
			if _, ok := bigVal.SetString(tok.Literal, 10); !ok {
				return nil, fmt.Errorf("line %d:%d: invalid integer '%s'", tok.Line, tok.Col, tok.Literal)
			}
			return &SintLiteral{Value: bigVal, P: Position{tok.Line, tok.Col}}, nil
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

	case lexer.TOKEN_FSTRING_LIT:
		p.advance()
		return p.parseFString(tok)

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

	case lexer.TOKEN_SELF:
		p.advance()
		return &SelfExpr{P: Position{tok.Line, tok.Col}}, nil

	case lexer.TOKEN_SUPER:
		return p.parseSuperCall()

	case lexer.TOKEN_LBRACE:
		return p.parseDictLiteral()

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
		// property access (no parens) - field access
		return &FieldAccessExpr{
			Object: left,
			Field:  method.Literal,
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

func (p *Parser) parseClassDecl() (*ClassDecl, error) {
	tok := p.advance() // consume 'class'

	name, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	superName := ""
	if p.current().Type == lexer.TOKEN_EXTENDS {
		p.advance()
		super, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		superName = super.Literal
	}

	if _, err := p.expect(lexer.TOKEN_LBRACE); err != nil {
		return nil, err
	}

	var fields []FieldDecl
	var methods []MethodDecl

	for p.current().Type != lexer.TOKEN_RBRACE && p.current().Type != lexer.TOKEN_EOF {
		// skip stray semicolons
		for p.current().Type == lexer.TOKEN_SEMICOLON {
			p.advance()
		}
		if p.current().Type == lexer.TOKEN_RBRACE {
			break
		}

		if p.current().Type == lexer.TOKEN_VAR {
			field, err := p.parseFieldDecl()
			if err != nil {
				return nil, err
			}
			fields = append(fields, *field)
		} else if p.current().Type == lexer.TOKEN_PUBLIC || p.current().Type == lexer.TOKEN_PRIVATE {
			visibility := p.current().Literal
			p.advance()

			if p.current().Type == lexer.TOKEN_FUNCTION {
				method, err := p.parseMethodDecl(visibility)
				if err != nil {
					return nil, err
				}
				methods = append(methods, *method)
			} else {
				return nil, fmt.Errorf("line %d:%d: expected 'function' after '%s', got '%s'",
					p.current().Line, p.current().Col, visibility, p.current().Literal)
			}
		} else {
			return nil, fmt.Errorf("line %d:%d: expected 'var', 'public', or 'private' in class body, got '%s'",
				p.current().Line, p.current().Col, p.current().Literal)
		}
	}

	if _, err := p.expect(lexer.TOKEN_RBRACE); err != nil {
		return nil, err
	}

	return &ClassDecl{
		Name:      name.Literal,
		SuperName: superName,
		Fields:    fields,
		Methods:   methods,
		P:         Position{tok.Line, tok.Col},
	}, nil
}

func (p *Parser) parseFieldDecl() (*FieldDecl, error) {
	tok := p.advance() // consume 'var'

	// expect public or private
	if p.current().Type != lexer.TOKEN_PUBLIC && p.current().Type != lexer.TOKEN_PRIVATE {
		return nil, fmt.Errorf("line %d:%d: expected 'public' or 'private' after 'var' in class, got '%s'",
			p.current().Line, p.current().Col, p.current().Literal)
	}
	visibility := p.current().Literal
	p.advance()

	// type
	typeTok := p.current()
	if !lexer.IsClassTypeKeyword(typeTok.Type) {
		return nil, fmt.Errorf("line %d:%d: expected type, got '%s'", typeTok.Line, typeTok.Col, typeTok.Literal)
	}
	p.advance()

	// name
	nameTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	// optional default value
	var defaultVal Expr
	if p.current().Type == lexer.TOKEN_ASSIGN {
		p.advance()
		defaultVal, err = p.parseExpression(0)
		if err != nil {
			return nil, err
		}
	}

	return &FieldDecl{
		Visibility: visibility,
		TypeName:   typeTok.Literal,
		Name:       nameTok.Literal,
		Default:    defaultVal,
		P:          Position{tok.Line, tok.Col},
	}, nil
}

func (p *Parser) parseMethodDecl(visibility string) (*MethodDecl, error) {
	tok := p.advance() // consume 'function'

	nameTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	// init is always private
	methodName := nameTok.Literal
	if methodName == "init" {
		visibility = "private"
	}

	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}

	params, err := p.parseParamList()
	if err != nil {
		return nil, err
	}

	p.advance() // consume ')'

	// optional return type
	retType := ""
	if lexer.IsClassTypeKeyword(p.current().Type) {
		retType = p.current().Literal
		p.advance()
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &MethodDecl{
		Visibility: visibility,
		Name:       methodName,
		Params:     params,
		ReturnType: retType,
		Body:       body,
		P:          Position{tok.Line, tok.Col},
	}, nil
}

func (p *Parser) parseParamList() ([]Param, error) {
	var params []Param
	hasOptional := false
	for p.current().Type != lexer.TOKEN_RPAREN {
		if len(params) > 0 {
			if _, err := p.expect(lexer.TOKEN_COMMA); err != nil {
				return nil, err
			}
		}
		paramType := p.current()
		if !lexer.IsClassTypeKeyword(paramType.Type) {
			return nil, fmt.Errorf("line %d:%d: expected parameter type, got '%s'",
				paramType.Line, paramType.Col, paramType.Literal)
		}
		p.advance()
		paramName, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		param := Param{TypeName: paramType.Literal, Name: paramName.Literal}

		// optional default value
		if p.current().Type == lexer.TOKEN_ASSIGN {
			p.advance()
			def, err := p.parseExpression(0)
			if err != nil {
				return nil, err
			}
			param.Default = def
			hasOptional = true
		} else if hasOptional {
			return nil, fmt.Errorf("line %d:%d: required parameter '%s' cannot follow optional parameters",
				paramName.Line, paramName.Col, paramName.Literal)
		}

		params = append(params, param)
	}
	return params, nil
}

func (p *Parser) parseSuperCall() (Expr, error) {
	tok := p.advance() // consume 'super'
	if _, err := p.expect(lexer.TOKEN_DOT); err != nil {
		return nil, err
	}
	method, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	args, err := p.parseArgList()
	if err != nil {
		return nil, err
	}
	return &SuperCallExpr{
		Method: method.Literal,
		Args:   args,
		P:      Position{tok.Line, tok.Col},
	}, nil
}

func (p *Parser) parseDictLiteral() (Expr, error) {
	tok := p.advance() // consume '{'
	var keys, values []Expr

	if p.current().Type == lexer.TOKEN_RBRACE {
		p.advance()
		return &DictLiteral{Keys: keys, Values: values, P: Position{tok.Line, tok.Col}}, nil
	}

	for {
		key, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)

		if _, err := p.expect(lexer.TOKEN_COLON); err != nil {
			return nil, err
		}

		value, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		values = append(values, value)

		if p.current().Type == lexer.TOKEN_COMMA {
			p.advance()
			if p.current().Type == lexer.TOKEN_RBRACE {
				break // trailing comma
			}
		} else {
			break
		}
	}

	if _, err := p.expect(lexer.TOKEN_RBRACE); err != nil {
		return nil, err
	}
	return &DictLiteral{Keys: keys, Values: values, P: Position{tok.Line, tok.Col}}, nil
}

func (p *Parser) parseTryCatch() (*TryCatchStmt, error) {
	tok := p.advance() // consume 'try'

	tryBody, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_CATCH); err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}

	catchType, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	catchName, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}

	catchBody, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &TryCatchStmt{
		TryBody:   tryBody,
		CatchType: catchType.Literal,
		CatchName: catchName.Literal,
		CatchBody: catchBody,
		P:         Position{tok.Line, tok.Col},
	}, nil
}

func (p *Parser) parseThrow() (*ThrowStmt, error) {
	tok := p.advance() // consume 'throw'
	value, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	return &ThrowStmt{Value: value, P: Position{tok.Line, tok.Col}}, nil
}

func (p *Parser) parseWithStmt() (*WithStmt, error) {
	tok := p.advance() // consume 'with'

	expr, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_AS); err != nil {
		return nil, err
	}

	varName, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &WithStmt{
		Expr:    expr,
		VarName: varName.Literal,
		Body:    body,
		P:       Position{tok.Line, tok.Col},
	}, nil
}

func (p *Parser) parseFString(tok lexer.Token) (Expr, error) {
	raw := []rune(tok.Literal)
	pos := Position{tok.Line, tok.Col}
	var parts []Expr
	var buf []rune

	i := 0
	for i < len(raw) {
		if raw[i] == '{' {
			// flush text before this
			if len(buf) > 0 {
				parts = append(parts, &StringLiteral{Value: string(buf), P: pos})
				buf = nil
			}
			// find matching closing brace
			i++
			depth := 1
			var exprRunes []rune
			for i < len(raw) && depth > 0 {
				if raw[i] == '{' {
					depth++
				} else if raw[i] == '}' {
					depth--
					if depth == 0 {
						break
					}
				}
				exprRunes = append(exprRunes, raw[i])
				i++
			}
			if depth != 0 {
				return nil, fmt.Errorf("line %d:%d: unclosed '{' in f-string", tok.Line, tok.Col)
			}
			i++ // skip closing }

			// sub-parse the expression
			subLexer := lexer.New(string(exprRunes))
			subTokens, err := subLexer.Tokenize()
			if err != nil {
				return nil, fmt.Errorf("line %d:%d: f-string expression: %s", tok.Line, tok.Col, err)
			}
			subParser := New(subTokens)
			expr, err := subParser.parseExpression(0)
			if err != nil {
				return nil, fmt.Errorf("line %d:%d: f-string expression: %s", tok.Line, tok.Col, err)
			}
			parts = append(parts, expr)
		} else {
			buf = append(buf, raw[i])
			i++
		}
	}
	// flush remaining text
	if len(buf) > 0 {
		parts = append(parts, &StringLiteral{Value: string(buf), P: pos})
	}

	if len(parts) == 0 {
		return &StringLiteral{Value: "", P: pos}, nil
	}
	if len(parts) == 1 {
		if sl, ok := parts[0].(*StringLiteral); ok {
			return sl, nil
		}
	}

	return &FStringExpr{Parts: parts, P: pos}, nil
}
