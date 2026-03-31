package lexer

type TokenType uint8

const (
	TOKEN_EOF TokenType = iota

	// Literals
	TOKEN_INT_LIT
	TOKEN_FLOAT_LIT
	TOKEN_STRING_LIT

	// Identifiers
	TOKEN_IDENT

	// Keywords
	TOKEN_VAR
	TOKEN_FUNCTION
	TOKEN_RETURN
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_FOR
	TOKEN_IN
	TOKEN_WHILE
	TOKEN_IMPORT
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_BREAK
	TOKEN_CONTINUE

	// Type keywords
	TOKEN_TYPE_STRING
	TOKEN_TYPE_INT
	TOKEN_TYPE_FLOAT
	TOKEN_TYPE_BOOL
	TOKEN_TYPE_LIST
	TOKEN_TYPE_ARRAY

	// Built-in functions
	TOKEN_RANGE
	TOKEN_FIXED

	// Operators
	TOKEN_PLUS
	TOKEN_MINUS
	TOKEN_STAR
	TOKEN_SLASH
	TOKEN_PERCENT
	TOKEN_ASSIGN
	TOKEN_EQ
	TOKEN_NEQ
	TOKEN_LT
	TOKEN_GT
	TOKEN_LTE
	TOKEN_GTE

	// Delimiters
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_LBRACE
	TOKEN_RBRACE
	TOKEN_LBRACKET
	TOKEN_RBRACKET
	TOKEN_COMMA
	TOKEN_DOT
	TOKEN_SEMICOLON
)

var keywords = map[string]TokenType{
	"var":      TOKEN_VAR,
	"function": TOKEN_FUNCTION,
	"return":   TOKEN_RETURN,
	"if":       TOKEN_IF,
	"else":     TOKEN_ELSE,
	"for":      TOKEN_FOR,
	"in":       TOKEN_IN,
	"while":    TOKEN_WHILE,
	"import":   TOKEN_IMPORT,
	"and":      TOKEN_AND,
	"or":       TOKEN_OR,
	"not":      TOKEN_NOT,
	"true":     TOKEN_TRUE,
	"false":    TOKEN_FALSE,
	"break":    TOKEN_BREAK,
	"continue": TOKEN_CONTINUE,
	"string":   TOKEN_TYPE_STRING,
	"int":      TOKEN_TYPE_INT,
	"float":    TOKEN_TYPE_FLOAT,
	"bool":     TOKEN_TYPE_BOOL,
	"list":     TOKEN_TYPE_LIST,
	"array":    TOKEN_TYPE_ARRAY,
	"range":    TOKEN_RANGE,
	"fixed":    TOKEN_FIXED,
}

func LookupKeyword(ident string) (TokenType, bool) {
	tok, ok := keywords[ident]
	return tok, ok
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

var tokenNames = map[TokenType]string{
	TOKEN_EOF: "EOF", TOKEN_INT_LIT: "INT", TOKEN_FLOAT_LIT: "FLOAT",
	TOKEN_STRING_LIT: "STRING", TOKEN_IDENT: "IDENT",
	TOKEN_VAR: "var", TOKEN_FUNCTION: "function", TOKEN_RETURN: "return",
	TOKEN_IF: "if", TOKEN_ELSE: "else", TOKEN_FOR: "for", TOKEN_IN: "in",
	TOKEN_WHILE: "while", TOKEN_IMPORT: "import", TOKEN_AND: "and",
	TOKEN_OR: "or", TOKEN_NOT: "not", TOKEN_TRUE: "true", TOKEN_FALSE: "false",
	TOKEN_BREAK: "break", TOKEN_CONTINUE: "continue",
	TOKEN_TYPE_STRING: "string", TOKEN_TYPE_INT: "int", TOKEN_TYPE_FLOAT: "float",
	TOKEN_TYPE_BOOL: "bool", TOKEN_TYPE_LIST: "list", TOKEN_TYPE_ARRAY: "array",
	TOKEN_RANGE: "range", TOKEN_FIXED: "fixed",
	TOKEN_PLUS: "+", TOKEN_MINUS: "-", TOKEN_STAR: "*", TOKEN_SLASH: "/",
	TOKEN_PERCENT: "%", TOKEN_ASSIGN: "=", TOKEN_EQ: "==", TOKEN_NEQ: "!=",
	TOKEN_LT: "<", TOKEN_GT: ">", TOKEN_LTE: "<=", TOKEN_GTE: ">=",
	TOKEN_LPAREN: "(", TOKEN_RPAREN: ")", TOKEN_LBRACE: "{", TOKEN_RBRACE: "}",
	TOKEN_LBRACKET: "[", TOKEN_RBRACKET: "]", TOKEN_COMMA: ",", TOKEN_DOT: ".",
	TOKEN_SEMICOLON: ";",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

func IsTypeKeyword(t TokenType) bool {
	return t >= TOKEN_TYPE_STRING && t <= TOKEN_TYPE_ARRAY
}
