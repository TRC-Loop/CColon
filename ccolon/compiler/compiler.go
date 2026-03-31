package compiler

import (
	"fmt"

	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
)

type Local struct {
	Name  string
	Depth int
	Slot  int
}

type LoopContext struct {
	StartOffset      int
	BreakPatches     []int
	ContinuePatches  []int
	HasContinueTarget bool
	ContinueTarget   int
}

type Compiler struct {
	function   *FuncObject
	locals     []Local
	scopeDepth int
	loops      []LoopContext
}

func New() *Compiler {
	fn := &FuncObject{Name: "<script>"}
	return &Compiler{
		function: fn,
	}
}

func (c *Compiler) Compile(program *parser.Program) (*FuncObject, error) {
	for _, stmt := range program.Stmts {
		if err := c.compileStmt(stmt); err != nil {
			return nil, err
		}
	}
	c.emit(byte(OP_HALT), 0)
	c.function.LocalCount = len(c.locals)
	return c.function, nil
}

func (c *Compiler) emit(b byte, line int) {
	c.function.Code = append(c.function.Code, b)
	c.function.Lines = append(c.function.Lines, line)
}

func (c *Compiler) emitOp(op OpCode, line int) {
	c.emit(byte(op), line)
}

func (c *Compiler) emitBytes(line int, bytes ...byte) {
	for _, b := range bytes {
		c.emit(b, line)
	}
}

func (c *Compiler) addConstant(val interface{}) int {
	// dedup strings and numbers
	for i, existing := range c.function.Constants {
		if existing == val {
			return i
		}
	}
	c.function.Constants = append(c.function.Constants, val)
	return len(c.function.Constants) - 1
}

func (c *Compiler) emitConstant(val interface{}, line int) {
	idx := c.addConstant(val)
	c.emitOp(OP_CONST, line)
	c.emitUint16(idx, line)
}

func (c *Compiler) emitUint16(val int, line int) {
	c.emit(byte(val>>8), line)
	c.emit(byte(val&0xFF), line)
}

func (c *Compiler) emitJump(op OpCode, line int) int {
	c.emitOp(op, line)
	c.emit(0xFF, line)
	c.emit(0xFF, line)
	return len(c.function.Code) - 2
}

func (c *Compiler) patchJump(offset int) {
	jump := len(c.function.Code) - offset - 2
	c.function.Code[offset] = byte(jump >> 8)
	c.function.Code[offset+1] = byte(jump & 0xFF)
}

func (c *Compiler) emitLoop(loopStart int, line int) {
	c.emitOp(OP_LOOP, line)
	offset := len(c.function.Code) - loopStart + 2
	c.emit(byte(offset>>8), line)
	c.emit(byte(offset&0xFF), line)
}

func (c *Compiler) beginScope() {
	c.scopeDepth++
}

func (c *Compiler) endScope(line int) {
	c.scopeDepth--
	for len(c.locals) > 0 && c.locals[len(c.locals)-1].Depth > c.scopeDepth {
		c.emitOp(OP_POP, line)
		c.locals = c.locals[:len(c.locals)-1]
	}
}

func (c *Compiler) addLocal(name string) int {
	slot := len(c.locals)
	c.locals = append(c.locals, Local{Name: name, Depth: c.scopeDepth, Slot: slot})
	return slot
}

func (c *Compiler) resolveLocal(name string) (int, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if c.locals[i].Name == name {
			return c.locals[i].Slot, true
		}
	}
	return 0, false
}

func (c *Compiler) compileStmt(stmt parser.Stmt) error {
	switch s := stmt.(type) {
	case *parser.ImportStmt:
		return c.compileImport(s)
	case *parser.VarDecl:
		return c.compileVarDecl(s)
	case *parser.FuncDecl:
		return c.compileFuncDecl(s)
	case *parser.ExprStmt:
		if err := c.compileExpr(s.Expression); err != nil {
			return err
		}
		c.emitOp(OP_POP, s.P.Line)
		return nil
	case *parser.AssignStmt:
		return c.compileAssign(s)
	case *parser.IfStmt:
		return c.compileIf(s)
	case *parser.WhileStmt:
		return c.compileWhile(s)
	case *parser.ForInStmt:
		return c.compileForIn(s)
	case *parser.ReturnStmt:
		return c.compileReturn(s)
	case *parser.BreakStmt:
		return c.compileBreak(s)
	case *parser.ContinueStmt:
		return c.compileContinue(s)
	default:
		return fmt.Errorf("unknown statement type %T", stmt)
	}
}

func (c *Compiler) compileImport(s *parser.ImportStmt) error {
	idx := c.addConstant(s.Module)
	c.emitOp(OP_IMPORT, s.P.Line)
	c.emitUint16(idx, s.P.Line)
	return nil
}

func (c *Compiler) compileVarDecl(s *parser.VarDecl) error {
	if err := c.compileExpr(s.Value); err != nil {
		return err
	}

	if c.scopeDepth > 0 {
		c.addLocal(s.Name)
		return nil
	}

	idx := c.addConstant(s.Name)
	c.emitOp(OP_STORE_GLOBAL, s.P.Line)
	c.emitUint16(idx, s.P.Line)
	return nil
}

func (c *Compiler) compileFuncDecl(s *parser.FuncDecl) error {
	fnCompiler := &Compiler{
		function:   &FuncObject{Name: s.Name, Arity: len(s.Params)},
		scopeDepth: 1,
	}

	for _, param := range s.Params {
		fnCompiler.addLocal(param.Name)
	}

	for _, stmt := range s.Body {
		if err := fnCompiler.compileStmt(stmt); err != nil {
			return err
		}
	}

	// implicit nil return
	fnCompiler.emitOp(OP_NIL, s.P.Line)
	fnCompiler.emitOp(OP_RETURN, s.P.Line)
	fnCompiler.function.LocalCount = len(fnCompiler.locals)

	idx := c.addConstant(fnCompiler.function)
	c.emitOp(OP_CONST, s.P.Line)
	c.emitUint16(idx, s.P.Line)

	nameIdx := c.addConstant(s.Name)
	c.emitOp(OP_STORE_GLOBAL, s.P.Line)
	c.emitUint16(nameIdx, s.P.Line)
	return nil
}

func (c *Compiler) compileAssign(s *parser.AssignStmt) error {
	switch target := s.Target.(type) {
	case *parser.Identifier:
		if err := c.compileExpr(s.Value); err != nil {
			return err
		}
		if slot, ok := c.resolveLocal(target.Name); ok {
			c.emitOp(OP_STORE_LOCAL, s.P.Line)
			c.emit(byte(slot), s.P.Line)
			return nil
		}
		idx := c.addConstant(target.Name)
		c.emitOp(OP_STORE_GLOBAL, s.P.Line)
		c.emitUint16(idx, s.P.Line)
		return nil
	case *parser.IndexExpr:
		return c.compileIndexAssign(target, s.Value, s.P.Line)
	default:
		return fmt.Errorf("line %d:%d: invalid assignment target", s.P.Line, s.P.Col)
	}
}

func (c *Compiler) compileIndexAssign(target *parser.IndexExpr, value parser.Expr, line int) error {
	if err := c.compileExpr(target.Object); err != nil {
		return err
	}
	if err := c.compileExpr(target.Index); err != nil {
		return err
	}
	if err := c.compileExpr(value); err != nil {
		return err
	}
	c.emitOp(OP_INDEX_SET, line)
	return nil
}

func (c *Compiler) compileIf(s *parser.IfStmt) error {
	if err := c.compileExpr(s.Cond); err != nil {
		return err
	}

	falseJump := c.emitJump(OP_JUMP_IF_FALSE, s.P.Line)

	c.beginScope()
	for _, stmt := range s.Body {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}
	c.endScope(s.P.Line)

	if len(s.ElseBody) > 0 {
		endJump := c.emitJump(OP_JUMP, s.P.Line)
		c.patchJump(falseJump)
		c.beginScope()
		for _, stmt := range s.ElseBody {
			if err := c.compileStmt(stmt); err != nil {
				return err
			}
		}
		c.endScope(s.P.Line)
		c.patchJump(endJump)
	} else {
		c.patchJump(falseJump)
	}

	return nil
}

func (c *Compiler) compileWhile(s *parser.WhileStmt) error {
	loopStart := len(c.function.Code)

	loop := LoopContext{StartOffset: loopStart, HasContinueTarget: true, ContinueTarget: loopStart}
	c.loops = append(c.loops, loop)

	if err := c.compileExpr(s.Cond); err != nil {
		return err
	}
	exitJump := c.emitJump(OP_JUMP_IF_FALSE, s.P.Line)

	c.beginScope()
	for _, stmt := range s.Body {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}
	c.endScope(s.P.Line)

	c.emitLoop(loopStart, s.P.Line)
	c.patchJump(exitJump)

	lctx := c.loops[len(c.loops)-1]
	for _, bp := range lctx.BreakPatches {
		c.patchJump(bp)
	}
	c.loops = c.loops[:len(c.loops)-1]

	return nil
}

func (c *Compiler) compileForIn(s *parser.ForInStmt) error {
	rangeExpr, ok := s.Iterable.(*parser.RangeExpr)
	if !ok {
		// for-in over a list: compile differently
		return c.compileForInList(s)
	}

	// for i in range(start, end) -> desugar to while
	c.beginScope()

	// compile start value and create loop variable
	if err := c.compileExpr(rangeExpr.Start); err != nil {
		return err
	}
	iterSlot := c.addLocal(s.VarName)

	// compile end value into a hidden local
	if err := c.compileExpr(rangeExpr.End); err != nil {
		return err
	}
	limitSlot := c.addLocal("$limit")

	loopStart := len(c.function.Code)

	loop := LoopContext{StartOffset: loopStart}
	c.loops = append(c.loops, loop)

	// condition: i < limit
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(iterSlot), s.P.Line)
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(limitSlot), s.P.Line)
	c.emitOp(OP_LT, s.P.Line)
	exitJump := c.emitJump(OP_JUMP_IF_FALSE, s.P.Line)

	// body
	for _, stmt := range s.Body {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}

	// patch continue jumps to here (the increment step)
	lctx := &c.loops[len(c.loops)-1]
	for _, cp := range lctx.ContinuePatches {
		c.patchJump(cp)
	}

	// increment: i = i + 1
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(iterSlot), s.P.Line)
	c.emitConstant(int64(1), s.P.Line)
	c.emitOp(OP_ADD, s.P.Line)
	c.emitOp(OP_STORE_LOCAL, s.P.Line)
	c.emit(byte(iterSlot), s.P.Line)

	c.emitLoop(loopStart, s.P.Line)
	c.patchJump(exitJump)

	for _, bp := range lctx.BreakPatches {
		c.patchJump(bp)
	}
	c.loops = c.loops[:len(c.loops)-1]

	c.endScope(s.P.Line)
	return nil
}

func (c *Compiler) compileForInList(s *parser.ForInStmt) error {
	c.beginScope()

	// compile the list
	if err := c.compileExpr(s.Iterable); err != nil {
		return err
	}
	listSlot := c.addLocal("$list")

	// index counter = 0
	c.emitConstant(int64(0), s.P.Line)
	idxSlot := c.addLocal("$idx")

	// the iteration variable
	c.emitOp(OP_NIL, s.P.Line)
	iterSlot := c.addLocal(s.VarName)

	loopStart := len(c.function.Code)

	loop := LoopContext{StartOffset: loopStart}
	c.loops = append(c.loops, loop)

	// condition: $idx < $list.length()
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(idxSlot), s.P.Line)
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(listSlot), s.P.Line)
	methodIdx := c.addConstant("length")
	c.emitOp(OP_METHOD_CALL, s.P.Line)
	c.emitUint16(methodIdx, s.P.Line)
	c.emit(0, s.P.Line) // 0 args
	c.emitOp(OP_LT, s.P.Line)
	exitJump := c.emitJump(OP_JUMP_IF_FALSE, s.P.Line)

	// iter = list[idx]
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(listSlot), s.P.Line)
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(idxSlot), s.P.Line)
	c.emitOp(OP_INDEX_GET, s.P.Line)
	c.emitOp(OP_STORE_LOCAL, s.P.Line)
	c.emit(byte(iterSlot), s.P.Line)

	for _, stmt := range s.Body {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}

	// patch continue jumps to here (the increment step)
	lctx := &c.loops[len(c.loops)-1]
	for _, cp := range lctx.ContinuePatches {
		c.patchJump(cp)
	}

	// idx = idx + 1
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(idxSlot), s.P.Line)
	c.emitConstant(int64(1), s.P.Line)
	c.emitOp(OP_ADD, s.P.Line)
	c.emitOp(OP_STORE_LOCAL, s.P.Line)
	c.emit(byte(idxSlot), s.P.Line)

	c.emitLoop(loopStart, s.P.Line)
	c.patchJump(exitJump)

	for _, bp := range lctx.BreakPatches {
		c.patchJump(bp)
	}
	c.loops = c.loops[:len(c.loops)-1]

	c.endScope(s.P.Line)
	return nil
}

func (c *Compiler) compileReturn(s *parser.ReturnStmt) error {
	if s.Value != nil {
		if err := c.compileExpr(s.Value); err != nil {
			return err
		}
	} else {
		c.emitOp(OP_NIL, s.P.Line)
	}
	c.emitOp(OP_RETURN, s.P.Line)
	return nil
}

func (c *Compiler) compileBreak(s *parser.BreakStmt) error {
	if len(c.loops) == 0 {
		return fmt.Errorf("line %d:%d: break outside of loop", s.P.Line, s.P.Col)
	}
	// pop locals in loop scope
	bp := c.emitJump(OP_JUMP, s.P.Line)
	c.loops[len(c.loops)-1].BreakPatches = append(c.loops[len(c.loops)-1].BreakPatches, bp)
	return nil
}

func (c *Compiler) compileContinue(s *parser.ContinueStmt) error {
	if len(c.loops) == 0 {
		return fmt.Errorf("line %d:%d: continue outside of loop", s.P.Line, s.P.Col)
	}
	loop := &c.loops[len(c.loops)-1]
	if loop.HasContinueTarget {
		c.emitLoop(loop.ContinueTarget, s.P.Line)
	} else {
		// forward jump to be patched later (for-in loops where increment is after body)
		bp := c.emitJump(OP_JUMP, s.P.Line)
		loop.ContinuePatches = append(loop.ContinuePatches, bp)
	}
	return nil
}

func (c *Compiler) compileExpr(expr parser.Expr) error {
	switch e := expr.(type) {
	case *parser.IntLiteral:
		c.emitConstant(e.Value, e.P.Line)
	case *parser.FloatLiteral:
		c.emitConstant(e.Value, e.P.Line)
	case *parser.StringLiteral:
		c.emitConstant(e.Value, e.P.Line)
	case *parser.BoolLiteral:
		if e.Value {
			c.emitOp(OP_TRUE, e.P.Line)
		} else {
			c.emitOp(OP_FALSE, e.P.Line)
		}
	case *parser.Identifier:
		return c.compileIdentifier(e)
	case *parser.BinaryExpr:
		return c.compileBinary(e)
	case *parser.UnaryExpr:
		return c.compileUnary(e)
	case *parser.CallExpr:
		return c.compileCall(e)
	case *parser.MethodCallExpr:
		return c.compileMethodCall(e)
	case *parser.IndexExpr:
		return c.compileIndex(e)
	case *parser.ListLiteral:
		return c.compileList(e)
	case *parser.FixedArrayLiteral:
		return c.compileFixedArray(e)
	case *parser.RangeExpr:
		// range as an expression (used in for-in, but could be standalone)
		return fmt.Errorf("line %d:%d: range() can only be used in for loops", e.P.Line, e.P.Col)
	default:
		return fmt.Errorf("unknown expression type %T", expr)
	}
	return nil
}

func (c *Compiler) compileIdentifier(e *parser.Identifier) error {
	if slot, ok := c.resolveLocal(e.Name); ok {
		c.emitOp(OP_LOAD_LOCAL, e.P.Line)
		c.emit(byte(slot), e.P.Line)
		return nil
	}
	idx := c.addConstant(e.Name)
	c.emitOp(OP_LOAD_GLOBAL, e.P.Line)
	c.emitUint16(idx, e.P.Line)
	return nil
}

func (c *Compiler) compileBinary(e *parser.BinaryExpr) error {
	// short-circuit for and/or
	if e.Op == lexer.TOKEN_AND {
		if err := c.compileExpr(e.Left); err != nil {
			return err
		}
		falseJump := c.emitJump(OP_JUMP_IF_FALSE, e.P.Line)
		if err := c.compileExpr(e.Right); err != nil {
			return err
		}
		endJump := c.emitJump(OP_JUMP, e.P.Line)
		c.patchJump(falseJump)
		c.emitOp(OP_FALSE, e.P.Line)
		c.patchJump(endJump)
		return nil
	}
	if e.Op == lexer.TOKEN_OR {
		if err := c.compileExpr(e.Left); err != nil {
			return err
		}
		// if true, skip right side
		trueJump := c.emitJump(OP_JUMP_IF_FALSE, e.P.Line)
		c.emitOp(OP_TRUE, e.P.Line)
		endJump := c.emitJump(OP_JUMP, e.P.Line)
		c.patchJump(trueJump)
		if err := c.compileExpr(e.Right); err != nil {
			return err
		}
		c.patchJump(endJump)
		return nil
	}

	if err := c.compileExpr(e.Left); err != nil {
		return err
	}
	if err := c.compileExpr(e.Right); err != nil {
		return err
	}

	switch e.Op {
	case lexer.TOKEN_PLUS:
		c.emitOp(OP_ADD, e.P.Line)
	case lexer.TOKEN_MINUS:
		c.emitOp(OP_SUB, e.P.Line)
	case lexer.TOKEN_STAR:
		c.emitOp(OP_MUL, e.P.Line)
	case lexer.TOKEN_SLASH:
		c.emitOp(OP_DIV, e.P.Line)
	case lexer.TOKEN_PERCENT:
		c.emitOp(OP_MOD, e.P.Line)
	case lexer.TOKEN_EQ:
		c.emitOp(OP_EQ, e.P.Line)
	case lexer.TOKEN_NEQ:
		c.emitOp(OP_NEQ, e.P.Line)
	case lexer.TOKEN_LT:
		c.emitOp(OP_LT, e.P.Line)
	case lexer.TOKEN_GT:
		c.emitOp(OP_GT, e.P.Line)
	case lexer.TOKEN_LTE:
		c.emitOp(OP_LTE, e.P.Line)
	case lexer.TOKEN_GTE:
		c.emitOp(OP_GTE, e.P.Line)
	default:
		return fmt.Errorf("unknown binary operator %s", e.Op.String())
	}
	return nil
}

func (c *Compiler) compileUnary(e *parser.UnaryExpr) error {
	if err := c.compileExpr(e.Operand); err != nil {
		return err
	}
	switch e.Op {
	case lexer.TOKEN_MINUS:
		c.emitOp(OP_NEG, e.P.Line)
	case lexer.TOKEN_NOT:
		c.emitOp(OP_NOT, e.P.Line)
	default:
		return fmt.Errorf("unknown unary operator %s", e.Op.String())
	}
	return nil
}

func (c *Compiler) compileCall(e *parser.CallExpr) error {
	if err := c.compileExpr(e.Callee); err != nil {
		return err
	}
	for _, arg := range e.Args {
		if err := c.compileExpr(arg); err != nil {
			return err
		}
	}
	c.emitOp(OP_CALL, e.P.Line)
	c.emit(byte(len(e.Args)), e.P.Line)
	return nil
}

func (c *Compiler) compileMethodCall(e *parser.MethodCallExpr) error {
	if err := c.compileExpr(e.Object); err != nil {
		return err
	}
	for _, arg := range e.Args {
		if err := c.compileExpr(arg); err != nil {
			return err
		}
	}
	methodIdx := c.addConstant(e.Method)
	c.emitOp(OP_METHOD_CALL, e.P.Line)
	c.emitUint16(methodIdx, e.P.Line)
	c.emit(byte(len(e.Args)), e.P.Line)
	return nil
}

func (c *Compiler) compileIndex(e *parser.IndexExpr) error {
	if err := c.compileExpr(e.Object); err != nil {
		return err
	}
	if err := c.compileExpr(e.Index); err != nil {
		return err
	}
	c.emitOp(OP_INDEX_GET, e.P.Line)
	return nil
}

func (c *Compiler) compileList(e *parser.ListLiteral) error {
	for _, elem := range e.Elements {
		if err := c.compileExpr(elem); err != nil {
			return err
		}
	}
	c.emitOp(OP_LIST_NEW, e.P.Line)
	c.emitUint16(len(e.Elements), e.P.Line)
	return nil
}

func (c *Compiler) compileFixedArray(e *parser.FixedArrayLiteral) error {
	for _, elem := range e.Elements {
		if err := c.compileExpr(elem); err != nil {
			return err
		}
	}
	c.emitOp(OP_ARRAY_NEW, e.P.Line)
	c.emitUint16(len(e.Elements), e.P.Line)
	return nil
}
