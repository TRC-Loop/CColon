package parser

import (
	"math/big"

	"github.com/TRC-Loop/ccolon/lexer"
)

type Position struct {
	Line int
	Col  int
}

type Node interface {
	Pos() Position
}

type Expr interface {
	Node
	exprNode()
}

type Stmt interface {
	Node
	stmtNode()
}

// --- Expressions ---

type IntLiteral struct {
	Value int64
	P     Position
}

type SintLiteral struct {
	Value *big.Int
	P     Position
}

type FloatLiteral struct {
	Value float64
	P     Position
}

type StringLiteral struct {
	Value string
	P     Position
}

type BoolLiteral struct {
	Value bool
	P     Position
}

type Identifier struct {
	Name string
	P    Position
}

type BinaryExpr struct {
	Left  Expr
	Op    lexer.TokenType
	Right Expr
	P     Position
}

type UnaryExpr struct {
	Op      lexer.TokenType
	Operand Expr
	P       Position
}

type CallExpr struct {
	Callee Expr
	Args   []Expr
	P      Position
}

type MethodCallExpr struct {
	Object Expr
	Method string
	Args   []Expr
	P      Position
}

type IndexExpr struct {
	Object Expr
	Index  Expr
	P      Position
}

type ListLiteral struct {
	Elements []Expr
	P        Position
}

type FixedArrayLiteral struct {
	Elements []Expr
	P        Position
}

type RangeExpr struct {
	Start Expr
	End   Expr
	P     Position
}

type SelfExpr struct {
	P Position
}

type FieldAccessExpr struct {
	Object Expr
	Field  string
	P      Position
}

type SuperCallExpr struct {
	Method string
	Args   []Expr
	P      Position
}

type DictLiteral struct {
	Keys   []Expr
	Values []Expr
	P      Position
}

func (n *IntLiteral) Pos() Position        { return n.P }
func (n *SintLiteral) Pos() Position       { return n.P }
func (n *FloatLiteral) Pos() Position      { return n.P }
func (n *StringLiteral) Pos() Position     { return n.P }
func (n *BoolLiteral) Pos() Position       { return n.P }
func (n *Identifier) Pos() Position        { return n.P }
func (n *BinaryExpr) Pos() Position        { return n.P }
func (n *UnaryExpr) Pos() Position         { return n.P }
func (n *CallExpr) Pos() Position          { return n.P }
func (n *MethodCallExpr) Pos() Position    { return n.P }
func (n *IndexExpr) Pos() Position         { return n.P }
func (n *ListLiteral) Pos() Position       { return n.P }
func (n *FixedArrayLiteral) Pos() Position { return n.P }
func (n *RangeExpr) Pos() Position         { return n.P }
func (n *SelfExpr) Pos() Position           { return n.P }
func (n *FieldAccessExpr) Pos() Position    { return n.P }
func (n *SuperCallExpr) Pos() Position      { return n.P }
func (n *DictLiteral) Pos() Position        { return n.P }

func (n *IntLiteral) exprNode()        {}
func (n *SintLiteral) exprNode()       {}
func (n *FloatLiteral) exprNode()      {}
func (n *StringLiteral) exprNode()     {}
func (n *BoolLiteral) exprNode()       {}
func (n *Identifier) exprNode()        {}
func (n *BinaryExpr) exprNode()        {}
func (n *UnaryExpr) exprNode()         {}
func (n *CallExpr) exprNode()          {}
func (n *MethodCallExpr) exprNode()    {}
func (n *IndexExpr) exprNode()         {}
func (n *ListLiteral) exprNode()       {}
func (n *FixedArrayLiteral) exprNode() {}
func (n *RangeExpr) exprNode()         {}
func (n *SelfExpr) exprNode()           {}
func (n *FieldAccessExpr) exprNode()    {}
func (n *SuperCallExpr) exprNode()      {}
func (n *DictLiteral) exprNode()        {}

// --- Statements ---

type Param struct {
	TypeName string
	Name     string
	Default  Expr
}

type FieldDecl struct {
	Visibility string
	TypeName   string
	Name       string
	Default    Expr
	P          Position
}

type MethodDecl struct {
	Visibility string
	Name       string
	Params     []Param
	ReturnType string
	Body       []Stmt
	P          Position
}

type VarDecl struct {
	TypeName string
	Name     string
	Value    Expr
	P        Position
}

type AssignStmt struct {
	Target Expr
	Value  Expr
	P      Position
}

type ExprStmt struct {
	Expression Expr
	P          Position
}

type ReturnStmt struct {
	Value Expr
	P     Position
}

type IfStmt struct {
	Cond     Expr
	Body     []Stmt
	ElseBody []Stmt
	P        Position
}

type WhileStmt struct {
	Cond Expr
	Body []Stmt
	P    Position
}

type ForInStmt struct {
	VarName  string
	Iterable Expr
	Body     []Stmt
	P        Position
}

type BreakStmt struct {
	P Position
}

type ContinueStmt struct {
	P Position
}

type FuncDecl struct {
	Name       string
	Params     []Param
	ReturnType string
	Body       []Stmt
	P          Position
}

type ImportStmt struct {
	Module string
	IsFile bool
	P      Position
}

type ClassDecl struct {
	Name      string
	SuperName string
	Fields    []FieldDecl
	Methods   []MethodDecl
	P         Position
}

type TryCatchStmt struct {
	TryBody   []Stmt
	CatchType string
	CatchName string
	CatchBody []Stmt
	P         Position
}

type ThrowStmt struct {
	Value Expr
	P     Position
}

type WithStmt struct {
	Expr    Expr
	VarName string
	Body    []Stmt
	P       Position
}

type Program struct {
	Stmts []Stmt
}

func (n *VarDecl) Pos() Position      { return n.P }
func (n *AssignStmt) Pos() Position   { return n.P }
func (n *ExprStmt) Pos() Position     { return n.P }
func (n *ReturnStmt) Pos() Position   { return n.P }
func (n *IfStmt) Pos() Position       { return n.P }
func (n *WhileStmt) Pos() Position    { return n.P }
func (n *ForInStmt) Pos() Position    { return n.P }
func (n *BreakStmt) Pos() Position    { return n.P }
func (n *ContinueStmt) Pos() Position { return n.P }
func (n *FuncDecl) Pos() Position     { return n.P }
func (n *ImportStmt) Pos() Position    { return n.P }
func (n *ClassDecl) Pos() Position     { return n.P }
func (n *TryCatchStmt) Pos() Position  { return n.P }
func (n *ThrowStmt) Pos() Position     { return n.P }
func (n *WithStmt) Pos() Position      { return n.P }

func (n *VarDecl) stmtNode()      {}
func (n *AssignStmt) stmtNode()   {}
func (n *ExprStmt) stmtNode()     {}
func (n *ReturnStmt) stmtNode()   {}
func (n *IfStmt) stmtNode()       {}
func (n *WhileStmt) stmtNode()    {}
func (n *ForInStmt) stmtNode()    {}
func (n *BreakStmt) stmtNode()    {}
func (n *ContinueStmt) stmtNode() {}
func (n *FuncDecl) stmtNode()     {}
func (n *ImportStmt) stmtNode()    {}
func (n *ClassDecl) stmtNode()     {}
func (n *TryCatchStmt) stmtNode()  {}
func (n *ThrowStmt) stmtNode()     {}
func (n *WithStmt) stmtNode()      {}
