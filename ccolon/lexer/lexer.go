package lexer

import (
	"fmt"
	"unicode"
)

type Lexer struct {
	source []rune
	pos    int
	line   int
	col    int
}

func New(source string) *Lexer {
	return &Lexer{
		source: []rune(source),
		pos:    0,
		line:   1,
		col:    1,
	}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

func (l *Lexer) peekNext() rune {
	if l.pos+1 >= len(l.source) {
		return 0
	}
	return l.source[l.pos+1]
}

func (l *Lexer) advance() rune {
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.source) {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.advance()
		} else if ch == '/' && l.peekNext() == '/' {
			l.skipLineComment()
		} else if ch == '/' && l.peekNext() == '*' {
			l.skipBlockComment()
		} else {
			break
		}
	}
}

func (l *Lexer) skipLineComment() {
	l.advance() // /
	l.advance() // /
	for l.pos < len(l.source) && l.peek() != '\n' {
		l.advance()
	}
}

func (l *Lexer) skipBlockComment() {
	l.advance() // /
	l.advance() // *
	for l.pos < len(l.source) {
		if l.peek() == '*' && l.peekNext() == '/' {
			l.advance() // *
			l.advance() // /
			return
		}
		l.advance()
	}
}

func (l *Lexer) readString() (Token, error) {
	line, col := l.line, l.col
	l.advance() // opening "
	var result []rune
	for l.pos < len(l.source) {
		ch := l.advance()
		if ch == '"' {
			return Token{TOKEN_STRING_LIT, string(result), line, col}, nil
		}
		if ch == '\\' {
			if l.pos >= len(l.source) {
				return Token{}, fmt.Errorf("line %d:%d: unterminated string escape", line, col)
			}
			esc := l.advance()
			switch esc {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case '"':
				result = append(result, '"')
			case '\\':
				result = append(result, '\\')
			default:
				result = append(result, '\\', esc)
			}
		} else {
			result = append(result, ch)
		}
	}
	return Token{}, fmt.Errorf("line %d:%d: unterminated string", line, col)
}

func (l *Lexer) readNumber() Token {
	line, col := l.line, l.col
	start := l.pos
	isFloat := false
	for l.pos < len(l.source) && (unicode.IsDigit(l.peek()) || l.peek() == '.') {
		if l.peek() == '.' {
			if isFloat {
				break
			}
			if l.pos+1 < len(l.source) && unicode.IsDigit(l.peekNext()) {
				isFloat = true
			} else {
				break // it's a dot operator, not decimal
			}
		}
		l.advance()
	}
	lit := string(l.source[start:l.pos])
	if isFloat {
		return Token{TOKEN_FLOAT_LIT, lit, line, col}
	}
	return Token{TOKEN_INT_LIT, lit, line, col}
}

func (l *Lexer) readIdentifier() (Token, error) {
	line, col := l.line, l.col
	start := l.pos
	for l.pos < len(l.source) && (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
		l.advance()
	}
	lit := string(l.source[start:l.pos])

	// f-string: f"..."
	if lit == "f" && l.pos < len(l.source) && l.peek() == '"' {
		return l.readFString(line, col)
	}

	if tok, ok := LookupKeyword(lit); ok {
		return Token{tok, lit, line, col}, nil
	}
	return Token{TOKEN_IDENT, lit, line, col}, nil
}

func (l *Lexer) readFString(line, col int) (Token, error) {
	l.advance() // opening "
	var result []rune
	for l.pos < len(l.source) {
		ch := l.advance()
		if ch == '"' {
			return Token{TOKEN_FSTRING_LIT, string(result), line, col}, nil
		}
		if ch == '\\' {
			if l.pos >= len(l.source) {
				return Token{}, fmt.Errorf("line %d:%d: unterminated f-string escape", line, col)
			}
			esc := l.advance()
			switch esc {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case '"':
				result = append(result, '"')
			case '\\':
				result = append(result, '\\')
			case '{':
				result = append(result, '{')
			case '}':
				result = append(result, '}')
			default:
				result = append(result, '\\', esc)
			}
		} else {
			result = append(result, ch)
		}
	}
	return Token{}, fmt.Errorf("line %d:%d: unterminated f-string", line, col)
}

func (l *Lexer) NextToken() (Token, error) {
	l.skipWhitespace()
	if l.pos >= len(l.source) {
		return Token{TOKEN_EOF, "", l.line, l.col}, nil
	}

	line, col := l.line, l.col
	ch := l.peek()

	if ch == '"' {
		return l.readString()
	}
	if unicode.IsDigit(ch) {
		return l.readNumber(), nil
	}
	if unicode.IsLetter(ch) || ch == '_' {
		return l.readIdentifier()
	}

	l.advance()
	switch ch {
	case '+':
		return Token{TOKEN_PLUS, "+", line, col}, nil
	case '-':
		return Token{TOKEN_MINUS, "-", line, col}, nil
	case '*':
		return Token{TOKEN_STAR, "*", line, col}, nil
	case '/':
		return Token{TOKEN_SLASH, "/", line, col}, nil
	case '%':
		return Token{TOKEN_PERCENT, "%", line, col}, nil
	case '=':
		if l.peek() == '=' {
			l.advance()
			return Token{TOKEN_EQ, "==", line, col}, nil
		}
		return Token{TOKEN_ASSIGN, "=", line, col}, nil
	case '!':
		if l.peek() == '=' {
			l.advance()
			return Token{TOKEN_NEQ, "!=", line, col}, nil
		}
		return Token{}, fmt.Errorf("line %d:%d: unexpected character '!'", line, col)
	case '<':
		if l.peek() == '=' {
			l.advance()
			return Token{TOKEN_LTE, "<=", line, col}, nil
		}
		return Token{TOKEN_LT, "<", line, col}, nil
	case '>':
		if l.peek() == '=' {
			l.advance()
			return Token{TOKEN_GTE, ">=", line, col}, nil
		}
		return Token{TOKEN_GT, ">", line, col}, nil
	case '(':
		return Token{TOKEN_LPAREN, "(", line, col}, nil
	case ')':
		return Token{TOKEN_RPAREN, ")", line, col}, nil
	case '{':
		return Token{TOKEN_LBRACE, "{", line, col}, nil
	case '}':
		return Token{TOKEN_RBRACE, "}", line, col}, nil
	case '[':
		return Token{TOKEN_LBRACKET, "[", line, col}, nil
	case ']':
		return Token{TOKEN_RBRACKET, "]", line, col}, nil
	case ',':
		return Token{TOKEN_COMMA, ",", line, col}, nil
	case '.':
		return Token{TOKEN_DOT, ".", line, col}, nil
	case ':':
		return Token{TOKEN_COLON, ":", line, col}, nil
	case ';':
		return Token{TOKEN_SEMICOLON, ";", line, col}, nil
	}

	return Token{}, fmt.Errorf("line %d:%d: unexpected character '%c'", line, col, ch)
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		tok, err := l.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens, nil
}
