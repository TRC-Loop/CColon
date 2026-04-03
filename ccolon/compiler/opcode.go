package compiler

type OpCode byte

const (
	OP_CONST OpCode = iota + 1
	OP_TRUE
	OP_FALSE
	OP_NIL

	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD
	OP_NEG
	OP_NOT

	OP_EQ
	OP_NEQ
	OP_LT
	OP_GT
	OP_LTE
	OP_GTE

	OP_JUMP
	OP_JUMP_IF_FALSE
	OP_LOOP

	OP_LOAD_LOCAL
	OP_STORE_LOCAL
	OP_LOAD_GLOBAL
	OP_STORE_GLOBAL

	OP_CALL
	OP_RETURN

	OP_LIST_NEW
	OP_ARRAY_NEW
	OP_INDEX_GET
	OP_INDEX_SET

	OP_METHOD_CALL

	OP_IMPORT

	OP_POP
	OP_DUP

	OP_HALT

	OP_GET_FIELD
	OP_SET_FIELD
	OP_INHERIT

	OP_DICT_NEW

	OP_TRY_BEGIN
	OP_TRY_END
	OP_THROW

	OP_IMPORT_FILE

	OP_BREAK_PLACEHOLDER
	OP_CONTINUE_PLACEHOLDER

	OP_MARK_CONST
	OP_CALL_KW
	OP_FROM_IMPORT
)

type FuncObject struct {
	Name       string
	Arity      int           // minimum required args
	MaxArity   int           // total params including optional
	Defaults   []interface{} // default values for optional params (nil for required)
	ParamNames []string      // parameter names for kwargs resolution
	Code       []byte
	Constants  []interface{}
	Lines      []int
	LocalCount int
}

type FieldDef struct {
	Visibility string
	TypeName   string
	Default    interface{}
}

type MethodDef struct {
	Visibility string
	Fn         *FuncObject
}

type ClassDef struct {
	Name      string
	SuperName string
	Fields    map[string]*FieldDef
	Methods   map[string]*MethodDef
	InitArity int
	MaxArity  int
	InitDefs  []interface{}
}

var opNames = map[OpCode]string{
	OP_CONST: "CONST", OP_TRUE: "TRUE", OP_FALSE: "FALSE", OP_NIL: "NIL",
	OP_ADD: "ADD", OP_SUB: "SUB", OP_MUL: "MUL", OP_DIV: "DIV", OP_MOD: "MOD",
	OP_NEG: "NEG", OP_NOT: "NOT",
	OP_EQ: "EQ", OP_NEQ: "NEQ", OP_LT: "LT", OP_GT: "GT", OP_LTE: "LTE", OP_GTE: "GTE",
	OP_JUMP: "JUMP", OP_JUMP_IF_FALSE: "JUMP_IF_FALSE", OP_LOOP: "LOOP",
	OP_LOAD_LOCAL: "LOAD_LOCAL", OP_STORE_LOCAL: "STORE_LOCAL",
	OP_LOAD_GLOBAL: "LOAD_GLOBAL", OP_STORE_GLOBAL: "STORE_GLOBAL",
	OP_CALL: "CALL", OP_RETURN: "RETURN",
	OP_LIST_NEW: "LIST_NEW", OP_ARRAY_NEW: "ARRAY_NEW",
	OP_INDEX_GET: "INDEX_GET", OP_INDEX_SET: "INDEX_SET",
	OP_METHOD_CALL: "METHOD_CALL", OP_IMPORT: "IMPORT",
	OP_POP: "POP", OP_DUP: "DUP", OP_HALT: "HALT",
	OP_GET_FIELD: "GET_FIELD", OP_SET_FIELD: "SET_FIELD", OP_INHERIT: "INHERIT",
	OP_DICT_NEW: "DICT_NEW",
	OP_TRY_BEGIN: "TRY_BEGIN", OP_TRY_END: "TRY_END", OP_THROW: "THROW",
	OP_IMPORT_FILE: "IMPORT_FILE",
	OP_MARK_CONST:  "MARK_CONST",
	OP_CALL_KW:      "CALL_KW",
	OP_FROM_IMPORT:  "FROM_IMPORT",
}

func (op OpCode) String() string {
	if name, ok := opNames[op]; ok {
		return name
	}
	return "UNKNOWN"
}
