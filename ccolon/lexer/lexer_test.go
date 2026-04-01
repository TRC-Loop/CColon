package lexer

import (
	"testing"
)

func TestTokenizeIntegers(t *testing.T) {
	l := New("42 0 123456")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expect := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_INT_LIT, "42"},
		{TOKEN_INT_LIT, "0"},
		{TOKEN_INT_LIT, "123456"},
		{TOKEN_EOF, ""},
	}
	for i, e := range expect {
		if tokens[i].Type != e.typ || tokens[i].Literal != e.lit {
			t.Errorf("token %d: got %v %q, want %v %q", i, tokens[i].Type, tokens[i].Literal, e.typ, e.lit)
		}
	}
}

func TestTokenizeFloats(t *testing.T) {
	l := New("3.14 0.5 100.0")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range tokens[:3] {
		if tok.Type != TOKEN_FLOAT_LIT {
			t.Errorf("expected FLOAT_LIT, got %v for %q", tok.Type, tok.Literal)
		}
	}
}

func TestTokenizeStrings(t *testing.T) {
	l := New(`"hello" "world\n" "tab\there" "escaped\"quote"`)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expect := []string{"hello", "world\n", "tab\there", "escaped\"quote"}
	for i, e := range expect {
		if tokens[i].Type != TOKEN_STRING_LIT {
			t.Errorf("token %d: expected STRING_LIT, got %v", i, tokens[i].Type)
		}
		if tokens[i].Literal != e {
			t.Errorf("token %d: got %q, want %q", i, tokens[i].Literal, e)
		}
	}
}

func TestTokenizeUnterminatedString(t *testing.T) {
	l := New(`"hello`)
	_, err := l.Tokenize()
	if err == nil {
		t.Fatal("expected error for unterminated string")
	}
}

func TestTokenizeOperators(t *testing.T) {
	l := New("+ - * / % == != < > <= >= =")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expect := []TokenType{
		TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT,
		TOKEN_EQ, TOKEN_NEQ, TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE,
		TOKEN_ASSIGN, TOKEN_EOF,
	}
	for i, e := range expect {
		if tokens[i].Type != e {
			t.Errorf("token %d: got %v, want %v", i, tokens[i].Type, e)
		}
	}
}

func TestTokenizeKeywords(t *testing.T) {
	l := New("var function if else while for return import class extends self super public private try catch throw with as")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expect := []TokenType{
		TOKEN_VAR, TOKEN_FUNCTION, TOKEN_IF, TOKEN_ELSE, TOKEN_WHILE,
		TOKEN_FOR, TOKEN_RETURN, TOKEN_IMPORT, TOKEN_CLASS, TOKEN_EXTENDS,
		TOKEN_SELF, TOKEN_SUPER, TOKEN_PUBLIC, TOKEN_PRIVATE,
		TOKEN_TRY, TOKEN_CATCH, TOKEN_THROW, TOKEN_WITH, TOKEN_AS,
	}
	for i, e := range expect {
		if tokens[i].Type != e {
			t.Errorf("token %d: got %v, want %v for %q", i, tokens[i].Type, e, tokens[i].Literal)
		}
	}
}

func TestTokenizeTypeKeywords(t *testing.T) {
	l := New("int float string bool list array dict")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expect := []TokenType{
		TOKEN_TYPE_INT, TOKEN_TYPE_FLOAT, TOKEN_TYPE_STRING, TOKEN_TYPE_BOOL,
		TOKEN_TYPE_LIST, TOKEN_TYPE_ARRAY, TOKEN_TYPE_DICT,
	}
	for i, e := range expect {
		if tokens[i].Type != e {
			t.Errorf("token %d: got %v, want %v", i, tokens[i].Type, e)
		}
	}
}

func TestTokenizeDelimiters(t *testing.T) {
	l := New("( ) { } [ ] , . : ;")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expect := []TokenType{
		TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_LBRACE, TOKEN_RBRACE,
		TOKEN_LBRACKET, TOKEN_RBRACKET, TOKEN_COMMA, TOKEN_DOT,
		TOKEN_COLON, TOKEN_SEMICOLON,
	}
	for i, e := range expect {
		if tokens[i].Type != e {
			t.Errorf("token %d: got %v, want %v", i, tokens[i].Type, e)
		}
	}
}

func TestTokenizeLineComment(t *testing.T) {
	l := New("42 // this is a comment\n43")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Literal != "42" || tokens[1].Literal != "43" {
		t.Errorf("comments not skipped: got %v %v", tokens[0], tokens[1])
	}
}

func TestTokenizeBlockComment(t *testing.T) {
	l := New("42 /* block\ncomment */ 43")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Literal != "42" || tokens[1].Literal != "43" {
		t.Errorf("block comments not skipped: got %v %v", tokens[0], tokens[1])
	}
}

func TestTokenizeIdentifiers(t *testing.T) {
	l := New("foo bar_baz _private myVar123")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range tokens[:4] {
		if tok.Type != TOKEN_IDENT {
			t.Errorf("expected IDENT for %q, got %v", tok.Literal, tok.Type)
		}
	}
}

func TestTokenizeEmpty(t *testing.T) {
	l := New("")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 || tokens[0].Type != TOKEN_EOF {
		t.Errorf("expected single EOF, got %v", tokens)
	}
}

func TestTokenizeBooleans(t *testing.T) {
	l := New("true false")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != TOKEN_TRUE || tokens[1].Type != TOKEN_FALSE {
		t.Errorf("got %v %v", tokens[0].Type, tokens[1].Type)
	}
}

func TestTokenizeLogicalOps(t *testing.T) {
	l := New("and or not")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != TOKEN_AND || tokens[1].Type != TOKEN_OR || tokens[2].Type != TOKEN_NOT {
		t.Errorf("got %v %v %v", tokens[0].Type, tokens[1].Type, tokens[2].Type)
	}
}

func TestTokenizeLineTracking(t *testing.T) {
	l := New("a\nb\nc")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Line != 1 || tokens[1].Line != 2 || tokens[2].Line != 3 {
		t.Errorf("lines: %d %d %d", tokens[0].Line, tokens[1].Line, tokens[2].Line)
	}
}
